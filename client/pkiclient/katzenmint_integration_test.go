package pkiclient

import (
	"context"
	"io/ioutil"
	stdlog "log"
	"os"
	"path/filepath"
	"testing"
	"time"

	kpki "github.com/hashcloak/katzenmint-pki"
	"github.com/hashcloak/katzenmint-pki/config"
	"github.com/hashcloak/katzenmint-pki/testutil"
	katlog "github.com/katzenpost/core/log"

	"github.com/stretchr/testify/require"

	log "github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/light"
	"github.com/tendermint/tendermint/light/provider/http"
	"github.com/tendermint/tendermint/rpc/client/local"
	rpctest "github.com/tendermint/tendermint/rpc/test"
	dbm "github.com/tendermint/tm-db"
)

var (
	abciClient *local.Local
)

func newDiscardLogger() (logger log.Logger) {
	logger = log.NewTMLogger(log.NewSyncWriter(ioutil.Discard))
	return
}

func testCreateClient(t *testing.T, dbname string) *PKIClient {
	var (
		require    = require.New(t)
		config     = rpctest.GetConfig()
		chainID    = config.ChainID()
		rpcAddress = config.RPC.ListenAddress
	)

	// Give Tendermint time to generate some blocks
	time.Sleep(5 * time.Second)

	// Get an initial trusted block
	primary, err := http.New(chainID, rpcAddress)
	require.NoError(err)

	block, err := primary.LightBlock(context.Background(), 0)
	require.NoError(err)

	trustOptions := light.TrustOptions{
		Period: 10 * time.Minute,
		Height: block.Height,
		Hash:   block.Hash(),
	}

	// Setup a pki client
	logPath := filepath.Join(testDir, "pkiclient_log")
	logBackend, err := katlog.New(logPath, "INFO", true)
	require.NoError(err)

	pkiClient, err := NewPKIClient(&PKIClientConfig{
		LogBackend:         logBackend,
		ChainID:            chainID,
		TrustOptions:       trustOptions,
		PrimaryAddress:     rpcAddress,
		WitnessesAddresses: []string{rpcAddress},
		DatabaseName:       dbname,
		DatabaseDir:        testDir,
		RPCAddress:         rpcAddress,
	})
	require.NoError(err)

	return pkiClient
}

/*
// TestGetDocument tests the functionality of Meson universe
func TestGetDocument(t *testing.T) {
	var (
		require   = require.New(t)
		pkiClient = testCreateClient(t, "integration_test1")
	)

	// Get the current epoch
	epoch, _, err := pkiClient.GetEpoch(context.Background())
	require.NoError(err)

	// Create a document
	_, docSer := testutil.CreateTestDocument(require, epoch)
	docTest, err := s11n.VerifyAndParseDocument(docSer)
	require.NoError(err)

	// Create an add document transaction
	privKey, err := eddsa.NewKeypair(rand.Reader)
	require.NoError(err)
	tx, err := kpki.FormTransaction(kpki.AddConsensusDocument, epoch, string(docSer), privKey)
	require.NoError(err)

	// Upload the document
	resp, err := pkiClient.PostTx(context.Background(), tx)
	require.NoError(err)
	require.NotNil(resp)

	// Get the document and verify
	err = rpcclient.WaitForHeight(abciClient, resp.Height+1, nil)
	require.NoError(err)

	doc, _, err := pkiClient.GetDoc(context.Background(), epoch)
	require.NoError(err)
	require.Equal(docTest, doc, "Got an incorrect document")

	// Try getting a non-existing document
	_, _, err = pkiClient.GetDoc(context.Background(), epoch+1)
	require.NotNil(err, "Got a document that should not exist")
}
*/

func TestPostDescriptor(t *testing.T) {
	var (
		require   = require.New(t)
		pkiClient = testCreateClient(t, "integration_test2")
	)

	// Get the upcoming epoch
	epoch, _, err := pkiClient.GetEpoch(context.Background())
	require.NoError(err)
	epoch += 1

	// Post test descriptor
	desc, _, privKey := testutil.CreateTestDescriptor(require, 0, 0, epoch)
	err = pkiClient.Post(context.Background(), epoch, &privKey, desc)
	require.NoError(err)
}

// TestMain tests the katzenmint-pki
func TestMain(m *testing.M) {
	var err error

	// set up test directory
	testDir, err = ioutil.TempDir("", "pkiclient_test")
	if err != nil {
		stdlog.Fatal(err)
	}

	// start katzenmint node in the background to test against
	kcfg := config.DefaultConfig()
	db := dbm.NewMemDB()
	logger := newDiscardLogger()
	app := kpki.NewKatzenmintApplication(kcfg, db, logger)
	node := rpctest.StartTendermint(app, rpctest.SuppressStdout)
	abciClient = local.New(node)

	code := m.Run()

	// and shut down properly at the end
	rpctest.StopTendermint(node)
	db.Close()
	os.RemoveAll(testDir)
	os.Exit(code)
}
