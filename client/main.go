// main.go - Katzenpost wallet client for Ethereum
// Copyright (C) 2018  David Stainton
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

	"github.com/katzenpost/client"
	"github.com/katzenpost/client/config"
	"github.com/hashcloak/Meson/common"
	"github.com/ugorji/go/codec"
)

var (
	jsonHandle codec.JsonHandle
)

func main() {
	cfgFile := flag.String("f", "katzenpost.toml", "Path to the server config file.")
	hexBlob := flag.String("h", "", "Transaction hex blob to send.")
	ticker := flag.String("t", "", "Ticker.")
	chainID := flag.Int("c", 0, "ChainID")
	service := flag.String("s", "", "Service Name")
	flag.Parse()

	if *hexBlob == "" {
		panic("must specify tx hex blob")
	}

	cfg, err := config.LoadFile(*cfgFile)
	if err != nil {
		panic(err)
	}

	cfg, linkKey := client.AutoRegisterRandomClient(cfg)
	c, err := client.New(cfg)
	if err != nil {
		panic(err)
	}

	session, err := c.NewSession(linkKey)
	if err != nil {
		panic(err)
	}

	// serialize our transaction inside an ethereum kaetzpost request message
	req := common.NewRequest(*ticker, *hexBlob, *chainID)
	currencyRequest := req.ToJson()

	currencyService, err := session.GetService(*service)
	if err != nil {
		panic(err)
	}

	reply, err := session.BlockingSendUnreliableMessage(currencyService.Name, currencyService.Provider, currencyRequest)
	if err != nil {
		panic(err)
	}
	fmt.Printf("reply: %s\n", reply)
	fmt.Println("Done. Shutting down.")
	c.Shutdown()
}
