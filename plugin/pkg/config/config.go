// config.go - Crypto currency transaction submition configuration.
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

package config

import (
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/BurntSushi/toml"
	"github.com/hashcloak/Meson/plugin/pkg/chain"
)

type RPCMetadata struct {
	Url, User, Pass string
}

// Config is the configuration for this currency transaction proxy service.
type Config struct {
	Rpc      map[string]RPCMetadata
	LogDir   string
	LogLevel string
}

// Validate returns nil if the config is valid
// and otherwise an error is returned.
func (cfg *Config) Validate() error {
	if len(cfg.Rpc) == 0 {
		return errors.New("config: No ticker being set")
	}
	for ticker, rpc := range cfg.Rpc {
		_, err := chain.GetChain(ticker)
		if err != nil {
			return err
		}
		if rpc.Url == "" {
			return errors.New("config: Missing rpc url of ticker")
		}
	}
	return nil
}

// Load parses and validates the provided buffer b as a config file body and
// returns the Config.
func Load(b []byte) (*Config, error) {
	cfg := new(Config)
	md, err := toml.Decode(string(b), cfg)
	if err != nil {
		return nil, err
	}
	if undecoded := md.Undecoded(); len(undecoded) != 0 {
		return nil, fmt.Errorf("config: Undecoded keys in config file: %v", undecoded)
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// LoadFile loads, parses and validates the provided file and returns the
// Config.
func LoadFile(f string) (*Config, error) {
	b, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, err
	}
	return Load(b)
}
