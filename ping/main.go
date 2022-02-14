// main.go - Katzenpost ping tool
// Copyright (C) 2018, 2019  David Stainton
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

package main

import (
	"flag"
	"fmt"

	client "github.com/hashcloak/Meson/client"
	"github.com/hashcloak/Meson/client/config"
	"github.com/katzenpost/core/crypto/ecdh"
)

func register(configFile string) (*config.Config, *ecdh.PrivateKey) {
	cfg, err := config.LoadFile(configFile)
	if err != nil {
		panic(err)
	}
	_ = cfg.UpdateTrust()
	_ = cfg.SaveConfig(configFile)
	linkKey := client.AutoRegisterRandomClient(cfg)
	return cfg, linkKey
}

func main() {
	var configFile string
	var service string
	flag.StringVar(&configFile, "c", "client.toml", "configuration file")
	flag.StringVar(&service, "s", "echo", "service name")
	flag.Parse()

	if service == "" {
		panic("must specify service name with -s")
	}

	cfg, linkKey := register(configFile)

	// create a client and connect to the mixnet Provider
	c, err := client.NewFromConfig(cfg, service)
	if err != nil {
		panic(err)
	}

	s, err := c.NewSession(linkKey)
	if err != nil {
		panic(err)
	}

	serviceDesc, err := s.GetService(service)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Sending Sphinx packet payload to: %s@%s\n", serviceDesc.Name, serviceDesc.Provider)
	resp, err := s.BlockingSendUnreliableMessage(serviceDesc.Name, serviceDesc.Provider, []byte(`Data encryption is used widely today!`))
	if err != nil {
		panic(err)
	}
	payload, err := client.ValidateReply(resp)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Return: %s\n", payload)

	c.Shutdown()
}
