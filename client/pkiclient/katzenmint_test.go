package pkiclient

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"path/filepath"
	"testing"

	kpki "github.com/hashcloak/katzenmint-pki"
	"github.com/hashcloak/katzenmint-pki/s11n"
	"github.com/hashcloak/katzenmint-pki/testutil"

	katlog "github.com/katzenpost/core/log"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	// ics23 "github.com/confio/ics23/go"
	"github.com/cosmos/iavl"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/merkle"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	lightrpc "github.com/tendermint/tendermint/light/rpc"
	lcmock "github.com/tendermint/tendermint/light/rpc/mocks"
	tmcrypto "github.com/tendermint/tendermint/proto/tendermint/crypto"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	rpcmock "github.com/tendermint/tendermint/rpc/client/mocks"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"
)

var (
	testDir string
)

func getEpoch(abciClient rpcclient.Client, require *require.Assertions) uint64 {
	appInfo, err := abciClient.ABCIInfo(context.Background())
	require.NoError(err)
	epochByt := kpki.DecodeHex(appInfo.Response.Data)
	epoch, err := binary.ReadUvarint(bytes.NewReader(epochByt))
	require.NoError(err)
	return epoch
}

type testOp struct {
	Key   []byte
	Proof *iavl.RangeProof
}

func (op testOp) GetKey() []byte {
	return op.Key
}

func (op testOp) ProofOp() tmcrypto.ProofOp {
	proof := iavl.NewValueOp(op.Key, op.Proof)
	return proof.ProofOp()
}

func (op testOp) Run(args [][]byte) ([][]byte, error) {
	root := op.Proof.ComputeRootHash()
	switch len(args) {
	case 0:
		if err := op.Proof.Verify(root); err != nil {
			return nil, fmt.Errorf("root did not verified: %+v", err)
		}
	case 1:
		if err := op.Proof.VerifyAbsence(args[0]); err != nil {
			return nil, fmt.Errorf("proof did not verified: %+v", err)
		}
	default:
		return nil, fmt.Errorf("args must be length 0 or 1, got: %d", len(args))
	}

	return [][]byte{root}, nil
}

// TestMockPKIClientGetDocument tests PKI Client get document and verifies proofs.
func TestMockPKIClientGetDocument(t *testing.T) {
	var (
		require            = require.New(t)
		epoch       uint64 = 1
		blockHeight int64  = 1
	)

	// create a test document
	_, docSer := testutil.CreateTestDocument(require, epoch)
	testDoc, err := s11n.VerifyAndParseDocument(docSer)
	require.NoError(err)

	var (
		key   = []byte{1}
		value = make([]byte, len(docSer))
	)

	n := copy(value, docSer)
	require.Equal(n, len(docSer))

	// create iavl tree
	tree, err := iavl.NewMutableTree(dbm.NewMemDB(), 100)
	require.NoError(err)

	tree.Set(key, value)

	rawDoc, proof, err := tree.GetWithProof(key)
	require.NoError(err)
	require.Equal(rawDoc, docSer)

	testOp := &testOp{
		Key:   key,
		Proof: proof,
	}

	query := kpki.Query{
		Version: kpki.ProtocolVersion,
		Epoch:   epoch,
		Command: kpki.GetConsensus,
		Payload: "",
	}
	rawQuery, err := kpki.EncodeJson(query)
	require.NoError(err)

	// moke the abci query
	next := &rpcmock.Client{}
	next.On(
		"ABCIQueryWithOptions",
		context.Background(),
		mock.AnythingOfType("string"),
		tmbytes.HexBytes(rawQuery),
		mock.AnythingOfType("client.ABCIQueryOptions"),
	).Return(&ctypes.ResultABCIQuery{
		Response: abci.ResponseQuery{
			Code:   0,
			Key:    testOp.GetKey(),
			Value:  value,
			Height: blockHeight,
			ProofOps: &tmcrypto.ProofOps{
				Ops: []tmcrypto.ProofOp{testOp.ProofOp()},
			},
		},
	}, nil)

	// mock the abci info
	epochByt := make([]byte, 8)
	binary.PutUvarint(epochByt[:], epoch)
	next.On(
		"ABCIInfo",
		context.Background(),
	).Return(&ctypes.ResultABCIInfo{
		Response: abci.ResponseInfo{
			Data:            kpki.EncodeHex(epochByt),
			LastBlockHeight: blockHeight,
		},
	}, nil)

	// initialize pki client with light client
	lc := &lcmock.LightClient{}
	rootHash, err := testOp.Run(nil)
	require.NoError(err)
	lc.On("VerifyLightBlockAtHeight", context.Background(), int64(2), mock.AnythingOfType("time.Time")).Return(
		&types.LightBlock{
			SignedHeader: &types.SignedHeader{
				Header: &types.Header{AppHash: rootHash[0]},
			},
		},
		nil,
	)

	c := lightrpc.NewClient(next, lc,
		lightrpc.KeyPathFn(func(_ string, key []byte) (merkle.KeyPath, error) {
			kp := merkle.KeyPath{}
			kp = kp.AppendKey(key, merkle.KeyEncodingURL)
			return kp, nil
		}))

	logPath := filepath.Join(testDir, "pkiclient_log")
	logBackend, err := katlog.New(logPath, "INFO", true)
	require.NoError(err)

	pkiClient, err := NewPKIClientFromLightClient(c, logBackend)
	require.NoError(err)
	require.NotNil(pkiClient)

	// test get abci info
	e := getEpoch(next, require)
	require.Equal(e, epoch)

	// test get document with pki client
	doc, _, err := pkiClient.GetDoc(context.Background(), epoch)
	require.NoError(err)
	require.Equal(doc, testDoc)
}

// TestMockPKIClientPostTx tests PKI Client post transaction and verifies proofs.
func TestMockPKIClientPostTx(t *testing.T) {
	var (
		require        = require.New(t)
		epoch   uint64 = 1
	)

	// create a test descriptor
	desc, signed, privKey := testutil.CreateTestDescriptor(require, 0, 0, epoch)
	tx, err := kpki.FormTransaction(kpki.PublishMixDescriptor, epoch, kpki.EncodeHex(signed), &privKey)
	require.NoError(err)

	// moke the abci broadcast commit
	next := &rpcmock.Client{}
	tmtx := make(types.Tx, len(tx))
	n := copy(tmtx, tx)
	require.Equal(len(tx), n)
	next.On(
		"BroadcastTxCommit",
		context.Background(),
		tmtx,
	).Run(
		func(args mock.Arguments) {
			parsed, err := s11n.ParseDescriptorWithoutVerify(signed)
			require.NoError(err)
			require.Equal(desc.IdentityKey, parsed.IdentityKey)
			require.Equal(desc.LinkKey, parsed.LinkKey)
		},
	).Return(
		&ctypes.ResultBroadcastTxCommit{
			CheckTx:   abci.ResponseCheckTx{Code: 0, GasWanted: 1},
			DeliverTx: abci.ResponseDeliverTx{Code: 0},
		},
		nil,
	)

	// initialize light client
	lc := &lcmock.LightClient{}
	require.NoError(err)
	c := lightrpc.NewClient(next, lc,
		lightrpc.KeyPathFn(func(_ string, key []byte) (merkle.KeyPath, error) {
			kp := merkle.KeyPath{}
			return kp, nil
		}))
	logPath := filepath.Join(testDir, "pkiclient_log")
	logBackend, err := katlog.New(logPath, "INFO", true)
	require.NoError(err)

	// initialize pki client
	pkiClient, err := NewPKIClientFromLightClient(c, logBackend)
	require.NoError(err)
	require.NotNil(pkiClient)

	_, err = pkiClient.PostTx(context.Background(), tx)
	require.NoError(err)
}

// TestDeserialize tests PKI Client deserialize document.
func TestDeserialize(t *testing.T) {
	var (
		require        = require.New(t)
		epoch   uint64 = 1
	)

	// create a test document
	_, docSer := testutil.CreateTestDocument(require, epoch)
	testDoc, err := s11n.VerifyAndParseDocument(docSer)
	require.NoError(err)

	// make the abci query
	next := &rpcmock.Client{}

	// initialize pki client with light client
	lc := &lcmock.LightClient{}

	c := lightrpc.NewClient(next, lc,
		lightrpc.KeyPathFn(func(_ string, key []byte) (merkle.KeyPath, error) {
			kp := merkle.KeyPath{}
			return kp, nil
		}))

	logPath := filepath.Join(testDir, "pkiclient_log")
	logBackend, err := katlog.New(logPath, "INFO", true)
	require.NoError(err)

	pkiClient, err := NewPKIClientFromLightClient(c, logBackend)
	require.NoError(err)
	require.NotNil(pkiClient)

	doc, err := pkiClient.Deserialize(docSer)
	require.NoError(err)
	require.Equal(doc, testDoc)
}
