// proxy_tests.go - Katzenpost currency serice plugin proxy tests
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

package proxy

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashcloak/Meson/plugin/pkg/command"
	"github.com/hashcloak/Meson/plugin/pkg/common"
	"github.com/hashcloak/Meson/plugin/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestProxy(t *testing.T) {
	assert := assert.New(t)

	logDir, err := ioutil.TempDir("", "example")
	assert.NoError(err)
	defer os.RemoveAll(logDir) // clean up
	ticker := "eth"
	content := []byte(fmt.Sprintf(`
LogDir = "%s"
LogLevel = "DEBUG"

[Rpc.%s]
  Url = "http://localhost:8545"
  User = "somerpcusername"
  Pass = "somepassword"
`, logDir, ticker))
	tmpfn := filepath.Join(logDir, "currency.toml")
	err = ioutil.WriteFile(tmpfn, content, 0666)
	assert.NoError(err)

	cfg, err := config.LoadFile(tmpfn)
	assert.NoError(err)
	p, err := New(cfg)
	assert.NoError(err)

	hexBlob := "deadbeef"
	payload, _ := json.Marshal(command.PostTransactionRequest{TxHex: hexBlob})
	currencyRequest := common.NewRequest(command.PostTransaction, ticker, payload)
	ethRequest := currencyRequest.ToJson()
	id := uint64(123)
	hasSURB := true
	reply, err := p.OnRequest(id, ethRequest, hasSURB)
	if err != nil {
		if err.Error() != "failed to send currency request: Post \"http://localhost:8545\": dial tcp [::1]:8545: connect: connection refused" {
			t.Fatal(err)
		}
		// There is no Ethereum RPC working in the background.
		// Skip the rest of the test
	}

	t.Logf("reply: %s", reply)
}

func TestProxyWithoutAuth(t *testing.T) {
	assert := assert.New(t)

	logDir, err := ioutil.TempDir("", "example")
	assert.NoError(err)
	defer os.RemoveAll(logDir) // clean up
	ticker := "eth"
	content := []byte(fmt.Sprintf(`
LogDir = "%s"
LogLevel = "DEBUG"

[Rpc.%s]
  Url = "http://localhost:8545"
`, logDir, ticker))
	tmpfn := filepath.Join(logDir, "currency.toml")
	err = ioutil.WriteFile(tmpfn, content, 0666)
	assert.NoError(err)

	cfg, err := config.LoadFile(tmpfn)
	assert.NoError(err)
	p, err := New(cfg)
	assert.NoError(err)

	hexBlob := "deadbeef"
	payload, _ := json.Marshal(command.PostTransactionRequest{TxHex: hexBlob})
	currencyRequest := common.NewRequest(command.PostTransaction, ticker, payload)
	ethRequest := currencyRequest.ToJson()
	id := uint64(123)
	hasSURB := true
	reply, err := p.OnRequest(id, ethRequest, hasSURB)
	if err != nil {
		if err.Error() != "failed to send currency request: Post \"http://localhost:8545\": dial tcp [::1]:8545: connect: connection refused" {
			t.Fatal(err)
		}
		// There is no Ethereum RPC working in the background.
		// Skip the rest of the test
	}

	t.Logf("reply: %s", reply)
}
