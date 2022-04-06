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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

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
  Url = "localhost:8545"
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
	currencyRequest := common.NewRequest(ticker, hexBlob)
	ethRequest := currencyRequest.ToJson()
	id := uint64(123)
	hasSURB := true
	reply, err := p.OnRequest(id, ethRequest, hasSURB)
	assert.NoError(err)

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
  Url = "localhost:8545"
`, logDir, ticker))
	tmpfn := filepath.Join(logDir, "currency.toml")
	err = ioutil.WriteFile(tmpfn, content, 0666)
	assert.NoError(err)

	cfg, err := config.LoadFile(tmpfn)
	assert.NoError(err)
	p, err := New(cfg)
	assert.NoError(err)

	hexBlob := "deadbeef"
	currencyRequest := common.NewRequest(ticker, hexBlob)
	ethRequest := currencyRequest.ToJson()
	id := uint64(123)
	hasSURB := true
	reply, err := p.OnRequest(id, ethRequest, hasSURB)
	assert.NoError(err)

	t.Logf("reply: %s", reply)
}
