// server_test.go - Katzenpost server tests.
// Copyright (C) 2018  David Stainton.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

// Package server provides the Katzenpost server.
package server

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"

	"github.com/hashcloak/Meson-server/config"
	kpki "github.com/hashcloak/katzenmint-pki"
	"github.com/katzenpost/core/crypto/eddsa"
	"github.com/katzenpost/core/crypto/rand"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmlog "github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/light"
	httpp "github.com/tendermint/tendermint/light/provider/http"
	"github.com/tendermint/tendermint/rpc/client/local"
	rpctest "github.com/tendermint/tendermint/rpc/test"
	dbm "github.com/tendermint/tm-db"
)

var (
	testDir    string
	abciClient *local.Local
)

func newDiscardLogger() (logger tmlog.Logger) {
	logger = tmlog.NewTMLogger(tmlog.NewSyncWriter(ioutil.Discard))
	return
}

func TestServerStartShutdown(t *testing.T) {
	t.Log("New server")
	var (
		assert     = assert.New(t)
		require    = require.New(t)
		rpcCfg     = rpctest.GetConfig()
		chainID    = rpcCfg.ChainID()
		rpcAddress = rpcCfg.RPC.ListenAddress
	)

	// Give Tendermint time to generate some blocks
	time.Sleep(3 * time.Second)

	// Get an initial trusted block
	primary, err := httpp.New(chainID, rpcAddress)
	assert.NoError(err)

	block, err := primary.LightBlock(context.Background(), 0)
	assert.NoError(err)
	require.NotNil(block, "Should not get nil block")

	trustOptions := light.TrustOptions{
		Period: 10 * time.Minute,
		Height: block.Height,
		Hash:   block.Hash(),
	}

	dir, err := ioutil.TempDir("", "server_data_dir")
	assert.NoError(err)

	mixIdKey, err := eddsa.NewKeypair(rand.Reader)
	assert.NoError(err)

	cfg := config.Config{
		Server: &config.Server{
			Identifier: "testserver",
			Addresses:  []string{"127.0.0.1:1234"},
			DataDir:    dir,
			IsProvider: false,
		},
		Logging: &config.Logging{
			Disable: false,
			File:    "",
			Level:   "DEBUG",
		},
		Provider: nil,
		PKI: &config.PKI{
			Voting: &config.Voting{
				ChainID:            chainID,
				TrustOptions:       trustOptions,
				PrimaryAddress:     rpcAddress,
				WitnessesAddresses: []string{rpcAddress},
				DatabaseName:       "test_meson_server_pkiclient_db",
				DatabaseDir:        testDir,
				RPCAddress:         rpcAddress,
			},
		},
		Management: &config.Management{
			Enable: false,
			Path:   "",
		},
		Debug: &config.Debug{
			IdentityKey:                  mixIdKey,
			NumSphinxWorkers:             1,
			NumProviderWorkers:           0,
			NumKaetzchenWorkers:          1,
			SchedulerExternalMemoryQueue: false,
			SchedulerQueueSize:           0,
			SchedulerMaxBurst:            16,
			UnwrapDelay:                  10,
			ProviderDelay:                0,
			KaetzchenDelay:               750,
			SchedulerSlack:               10,
			SendSlack:                    50,
			DecoySlack:                   15 * 1000,
			ConnectTimeout:               60 * 1000,
			HandshakeTimeout:             30 * 1000,
			ReauthInterval:               30 * 1000,
			SendDecoyTraffic:             false,
			DisableRateLimit:             true,
			GenerateOnly:                 false,
		},
	}

	err = cfg.FixupAndValidate()
	assert.NoError(err)

	s, err := New(&cfg)
	assert.NoError(err)
	s.Shutdown()
}

// TestMain tests the katzenmint-pki
func TestMain(m *testing.M) {
	var err error

	// set up test directory
	testDir, err = ioutil.TempDir("", "test_meson_pkiclient_dir")
	if err != nil {
		log.Fatal(err)
	}

	// start katzenmint node in the background to test against
	db := dbm.NewMemDB()
	logger := newDiscardLogger()
	app := kpki.NewKatzenmintApplication(db, logger)
	node := rpctest.StartTendermint(app, rpctest.SuppressStdout)
	abciClient = local.New(node)

	code := m.Run()

	// and shut down properly at the end
	rpctest.StopTendermint(node)
	db.Close()
	os.RemoveAll(testDir)
	os.Exit(code)
}
