// proxy.go - Crypto currency transaction proxy.
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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"

	"github.com/hashcloak/Meson/plugin/pkg/common"
	"github.com/hashcloak/Meson/plugin/pkg/config"
	"github.com/ugorji/go/codec"
	"gopkg.in/op/go-logging.v1"
)

var logFormat = logging.MustStringFormatter(
	"%{level:.4s} %{id:03x} %{message}",
)

const (
	// ResponseSuccess : Indicates whether a Response was successful
	ResponseSuccess = 0
	// ResponseError : Indicates whether a Response was unsuccessful
	ResponseError = 1
)

func stringToLogLevel(level string) (logging.Level, error) {
	switch level {
	case "DEBUG":
		return logging.DEBUG, nil
	case "INFO":
		return logging.INFO, nil
	case "NOTICE":
		return logging.NOTICE, nil
	case "WARNING":
		return logging.WARNING, nil
	case "ERROR":
		return logging.ERROR, nil
	case "CRITICAL":
		return logging.CRITICAL, nil
	}
	return -1, fmt.Errorf("invalid logging level %s", level)
}

func setupLoggerBackend(level logging.Level, writer io.Writer) logging.LeveledBackend {
	format := logFormat
	backend := logging.NewLogBackend(writer, "", 0)
	formatter := logging.NewBackendFormatter(backend, format)
	leveler := logging.AddModuleLevel(formatter)
	leveler.SetLevel(level, "echo-go")
	return leveler
}

// An ethereum request abstraction.
// Only need it for one method, though.
type ethRequest struct {
	// ChainId to indicate which Ethereum-based network
	ID int `json:"id"`
	// Indicates which version of JSON RPC to use
	// Since all networks support JSON RPC 2.0,
	// this attribute is a constant
	JSONRPC string `json:"jsonrpc"`
	// Which method you want to call
	METHOD string `json:"method"`
	// Params for the method you want to call
	Params []string `json:"params"`
}

// Takes a chainId and signed transaction data as parameters
// Returns a new ethereum request
func newEthRequest(id int, params []string) ethRequest {
	return ethRequest{
		ID:      id,
		JSONRPC: "2.0",
		METHOD:  "eth_sendRawTransaction",
		Params:  params,
	}
}

// Currency :  Handles logging and RPC details. Implements the ServicePlugin interface
type Currency struct {
	log        *logging.Logger
	jsonHandle codec.JsonHandle

	params map[string]string

	ticker   string
	chaindID int
	rpcUser  string
	rpcPass  string
	rpcURL   string
}

// GetParameters : Returns params from Currency struct
func (k *Currency) GetParameters() map[string]string {
	return k.params
}

// OnRequest : Request Handler
func (k *Currency) OnRequest(id uint64, payload []byte, hasSURB bool) ([]byte, error) {
	k.log.Debugf("Handling request %d", id)

	// Send request to HTTP RPC.
	req, err := common.RequestFromJson(k.ticker, k.chaindID, payload)
	if err != nil {
		k.log.Debug("Failed to send currency transaction request: (%v)", err)
		return common.NewResponse(ResponseError, err.Error()).ToJson(), nil
	}

	err = k.sendTransaction(req.Tx)
	if err != nil {
		k.log.Debug("Failed to send currency transaction request: (%v)", err)
		return common.NewResponse(ResponseError, err.Error()).ToJson(), nil
	}
	message := "success"
	return common.NewResponse(ResponseSuccess, message).ToJson(), nil
}

// Halt : Stops the plugin
func (k *Currency) Halt() {

}

func (k *Currency) sendTransaction(txHex string) error {
	k.log.Debug("sendTransaction")

	// marshall new transaction blob
	// allowHighFees := true
	// cmd := btcjson.NewSendRawTransactionCmd(txHex, &allowHighFees)
	// txId := 0 // this txId is not important
	ethRequest := newEthRequest(k.chaindID, []string{txHex})
	marshalledRequest, err := json.Marshal(ethRequest)
	bodyReader := bytes.NewReader(marshalledRequest)

	// create an http request
	httpReq, err := http.NewRequest("POST", k.rpcURL, bodyReader)
	if err != nil {
		return err
	}
	httpReq.Close = true
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.SetBasicAuth(k.rpcUser, k.rpcPass)

	// send http request
	client := http.Client{}
	httpResponse, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	k.log.Debugf("currency RPC response status: %s", httpResponse.Status)

	return nil
}

// New : Returns a pointer to a newly instantiated Currency struct
func New(cfg *config.Config) (*Currency, error) {
	currency := &Currency{
		ticker:   cfg.Ticker,
		chaindID: cfg.ChainID,
		rpcUser:  cfg.RPCUser,
		rpcPass:  cfg.RPCPass,
		rpcURL:   cfg.RPCURL,
		params:   make(map[string]string),
	}
	currency.jsonHandle.Canonical = true
	currency.jsonHandle.ErrorIfNoField = true
	currency.params = map[string]string{
		"name":    "currency_trickle",
		"version": "0.0.0",
	}

	// Ensure that the log directory exists.
	s, err := os.Stat(cfg.LogDir)
	if err != nil {
		return nil, err
	}
	if !s.IsDir() {
		return nil, errors.New("must be a directory")
	}

	// Log to a file.
	level, err := stringToLogLevel(cfg.LogLevel)
	logFile := path.Join(cfg.LogDir, fmt.Sprintf("meson-go.%d.log", os.Getpid()))
	f, err := os.Create(logFile)
	if err != nil {
		return nil, err
	}
	logBackend := setupLoggerBackend(level, f)
	currency.log = logging.MustGetLogger("meson-go")
	currency.log.SetBackend(logBackend)

	return currency, nil
}
