package katzenmint

import (
	"bytes"
	"context"
	"encoding/binary"
	"io/ioutil"
	"net/url"
	"testing"

	"github.com/cosmos/iavl"
	"github.com/hashcloak/katzenmint-pki/testutil"
	"github.com/katzenpost/core/crypto/eddsa"
	"github.com/katzenpost/core/crypto/rand"
	"github.com/katzenpost/core/pki"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/merkle"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/rpc/client/mock"
	dbm "github.com/tendermint/tm-db"
)

func newDiscardLogger() (logger log.Logger) {
	logger = log.NewTMLogger(log.NewSyncWriter(ioutil.Discard))
	return
}

func TestAddAuthority(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	// setup application
	db := dbm.NewMemDB()
	defer db.Close()
	logger := newDiscardLogger()
	app := NewKatzenmintApplication(kConfig, db, logger)
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
		KeyType: "",
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
	protoPubKey, err := validator.PubKey.Marshal()
	if err != nil {
		t.Fatalf("Failed to encode public with protobuf: %v\n", err)
	}
	key := storageKey(authoritiesBucket, protoPubKey, 0)
	_, err = app.state.Get(key)
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
	app := NewKatzenmintApplication(kConfig, db, logger)
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

	// commit through the epoch
	for i := int64(0); i <= epochInterval; i++ {
		m.App.Commit()
	}

	// test the doc is formed and exists in state
	loaded, _, err := app.state.documentForEpoch(epoch, app.state.blockHeight)
	require.Nil(err, "Failed to get pki document from state: %+v\n", err)
	require.NotNil(loaded, "Failed to get pki document from state: wrong key")
	// test against the expected doc?
	/* require.Equal(sDoc, loaded, "App state contains an erroneous pki document") */

	// prepare verification metadata
	appinfo, err = m.ABCIInfo(context.Background())
	require.Nil(err)
	apphash := appinfo.Response.LastBlockAppHash
	e := make([]byte, 8)
	binary.PutUvarint(e, epoch)
	key := storageKey(documentsBucket, e, epoch)
	path := "/" + url.PathEscape(string(key))

	m.App.Commit()

	// make a query for the doc
	query, err := EncodeJson(Query{
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
	verifier.RegisterOpDecoder(iavl.ProofOpIAVLValue, iavl.ValueOpDecoder)
	err = verifier.VerifyValue(rsp.Response.ProofOps, apphash, path, rsp.Response.Value)
	require.Nil(err, "Invalid proof for app responses")
}
