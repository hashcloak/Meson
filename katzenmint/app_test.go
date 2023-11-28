package katzenmint

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"net/url"
	"testing"

	dbm "github.com/cometbft/cometbft-db"
	costypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/hashcloak/Meson/katzenmint/s11n"
	"github.com/hashcloak/Meson/katzenmint/testutil"
	"github.com/katzenpost/core/crypto/eddsa"
	"github.com/katzenpost/core/crypto/rand"
	"github.com/katzenpost/core/pki"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	cryptoenc "github.com/tendermint/tendermint/crypto/encoding"
	"github.com/tendermint/tendermint/crypto/merkle"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/rpc/client/mock"
)

func newDiscardLogger() (logger log.Logger) {
	logger = log.NewTMLogger(log.NewSyncWriter(ioutil.Discard))
	return
}

func TestGetEpoch(t *testing.T) {
	require := require.New(t)

	// setup application
	db := dbm.NewMemDB()
	defer db.Close()
	logger := newDiscardLogger()
	app := NewKatzenmintApplication(kConfig, db, testDBCacheSize, logger)
	m := mock.ABCIApp{
		App: app,
	}
	m.App.BeginBlock(abcitypes.RequestBeginBlock{})
	m.App.Commit()

	// fetch abci info
	appinfo, err := m.ABCIInfo(context.Background())
	require.Nil(err)

	// advance block height
	m.App.BeginBlock(abcitypes.RequestBeginBlock{})
	m.App.Commit()

	// get epoch
	query, err := EncodeJson(&Query{
		Version: protocolVersion,
		Epoch:   0,
		Command: GetEpoch,
		Payload: "",
	})
	if err != nil {
		t.Fatalf("Failed to marshal query: %+v\n", err)
	}
	resp := m.App.Query(abcitypes.RequestQuery{Data: query})
	require.True(resp.IsOK(), resp.Log)
	expectEpoch, _ := binary.Uvarint(resp.Value[:8])
	require.Equal(appinfo.Response.Data, fmt.Sprint(expectEpoch))
}

func TestAddAuthority(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	// setup application
	db := dbm.NewMemDB()
	defer db.Close()
	logger := newDiscardLogger()
	app := NewKatzenmintApplication(kConfig, db, testDBCacheSize, logger)
	m := mock.ABCIApp{
		App: app,
	}

	// create authority transaction
	privKey, err := eddsa.NewKeypair(rand.Reader)
	require.NoError(err, "eddsa.NewKeypair()")
	authority := &Authority{
		Auth:    "katzenmint",
		Power:   1,
		PubKey:  privKey.PublicKey().Bytes(),
		KeyType: privKey.KeyType(),
	}
	rawAuth, err := EncodeJson(authority)
	if err != nil {
		t.Fatalf("Failed to marshal authority: %+v\n", err)
	}
	tx, err := FormTransaction(AddNewAuthority, 1, string(rawAuth), privKey)
	if err != nil {
		t.Fatalf("Error forming transaction: %+v\n", err)
	}

	// post transaction to app
	m.App.BeginBlock(abcitypes.RequestBeginBlock{})
	res, err := m.BroadcastTxCommit(context.Background(), tx)
	require.Nil(err)
	assert.True(res.CheckTx.IsOK())
	require.NotNil(res.DeliverTx)
	assert.True(res.DeliverTx.IsOK())

	// commit once
	m.App.Commit()

	// make checks
	validator := abcitypes.UpdateValidator(authority.PubKey, authority.Power, authority.KeyType)
	pubkey, err := cryptoenc.PubKeyFromProto(validator.PubKey)
	if err != nil {
		t.Fatalf("Failed to decode public key: %v\n", err)
	}
	key := storageKey(authoritiesBucket, pubkey.Address(), 0)
	_, err = app.state.get(key)
	if err != nil {
		t.Fatalf("Failed to get authority from database: %+v\n", err)
	}
	if len(app.state.validatorUpdates) <= 0 {
		t.Fatal("Failed to update authority\n")
	}
}

func TestPostDescriptorAndCommit(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	// setup application
	db := dbm.NewMemDB()
	defer db.Close()
	logger := newDiscardLogger()
	app := NewKatzenmintApplication(kConfig, db, testDBCacheSize, logger)
	m := mock.ABCIApp{
		App: app,
	}

	// fetch current epoch
	appinfo, err := m.ABCIInfo(context.Background())
	require.Nil(err)
	epochBytes := DecodeHex(appinfo.Response.Data)
	epoch, err := binary.ReadUvarint(bytes.NewReader(epochBytes))
	require.Nil(err)

	// create descriptors of providers and mixs
	descriptors := make([][]byte, 0)
	for i := 0; i < app.state.minNodesPerLayer; i++ {
		_, rawDesc, _ := testutil.CreateTestDescriptor(require, i, pki.LayerProvider, epoch)
		descriptors = append(descriptors, rawDesc)
	}
	for layer := 0; layer < app.state.layers; layer++ {
		for i := 0; i < app.state.minNodesPerLayer; i++ {
			_, rawDesc, _ := testutil.CreateTestDescriptor(require, i, 0, epoch)
			descriptors = append(descriptors, rawDesc)
		}
	}

	// create transaction for each descriptor
	transactions := make([][]byte, 0)
	privKey, err := eddsa.NewKeypair(rand.Reader)
	require.NoError(err, "GenerateKey()")
	for _, rawDesc := range descriptors {
		packedTx, err := FormTransaction(PublishMixDescriptor, epoch, EncodeHex(rawDesc), privKey)
		require.NoError(err)
		transactions = append(transactions, packedTx)
	}

	// post descriptor transactions to app
	m.App.BeginBlock(abcitypes.RequestBeginBlock{})
	for _, tx := range transactions {
		res, err := m.BroadcastTxCommit(context.Background(), tx)
		require.Nil(err)
		assert.True(res.CheckTx.IsOK(), res.CheckTx.Log)
		require.NotNil(res.DeliverTx)
		assert.True(res.DeliverTx.IsOK(), res.DeliverTx.Log)
	}
	m.App.Commit()

	// commit through the epoch, and add one more
	for i := 0; i < int(EpochInterval); i++ {
		m.App.BeginBlock(abcitypes.RequestBeginBlock{})
		m.App.Commit()
	}

	// test that the epoch proceeds
	query, err := EncodeJson(&Query{
		Version: protocolVersion,
		Epoch:   0,
		Command: GetEpoch,
		Payload: "",
	})
	if err != nil {
		t.Fatalf("Failed to marshal query: %+v\n", err)
	}
	resp := m.App.Query(abcitypes.RequestQuery{Data: query})
	require.True(resp.IsOK(), resp.Log)
	gotEpoch, _ := binary.Uvarint(resp.Value[:8])
	require.Equal(epoch+1, gotEpoch)

	// test the doc is formed and exists in state
	loaded, _, err := app.state.GetDocument(epoch, app.state.blockHeight-1)
	require.Nil(err, "Failed to get pki document from state: %+v\n", err)
	require.NotNil(loaded, "Failed to get pki document from state: wrong key")
	_, err = s11n.VerifyAndParseDocument(loaded)
	require.Nil(err, "Failed to parse pki document: %+v\n", err)

	// prepare verification metadata (in an old block)
	appinfo, err = m.ABCIInfo(context.Background())
	require.Nil(err)
	apphash := appinfo.Response.LastBlockAppHash
	key := storageKey(documentsBucket, []byte{}, epoch)
	keyPath := "/" + url.PathEscape(string(key))
	m.App.BeginBlock(abcitypes.RequestBeginBlock{})
	m.App.Commit()

	// make a query for the doc
	query, err = EncodeJson(Query{
		Version: protocolVersion,
		Epoch:   epoch,
		Command: GetConsensus,
		Payload: "",
	})
	require.Nil(err)

	rsp, err := m.ABCIQuery(context.Background(), "", query)
	require.Nil(err)
	require.True(rsp.Response.IsOK(), rsp.Response.Log)
	require.Equal(loaded, rsp.Response.Value, "App responses with an erroneous pki document")

	// verify query proof
	verifier := merkle.NewProofRuntime()
	verifier.RegisterOpDecoder(costypes.ProofOpIAVLCommitment, costypes.CommitmentOpDecoder)
	err = verifier.VerifyValue(rsp.Response.ProofOps, apphash, keyPath, rsp.Response.Value)
	require.Nil(err, "Invalid proof for app responses")
}
