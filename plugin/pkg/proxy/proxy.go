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
	"io/ioutil"
	"net/http"
	"os"
	"path"

	"github.com/hashcloak/Meson/plugin/pkg/chain"
	"github.com/hashcloak/Meson/plugin/pkg/command"
	"github.com/hashcloak/Meson/plugin/pkg/common"
	"github.com/hashcloak/Meson/plugin/pkg/config"
	"github.com/ugorji/go/codec"
	"gopkg.in/op/go-logging.v1"
)

var (
	logFormat = logging.MustStringFormatter(
		"%{level:.4s} %{id:03x} %{message}",
	)
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

// Currency : Handles logging and RPC details. Implements the ServicePlugin interface
type Currency struct {
	log        *logging.Logger
	jsonHandle codec.JsonHandle
	params     map[string]string
	rpc        map[string]config.RPCMetadata
}

// GetParameters : Returns params from Currency struct
func (k *Currency) GetParameters() map[string]string {
	return k.params
}

// OnRequest : Request Handler
func (k *Currency) OnRequest(id uint64, payload []byte, hasSURB bool) ([]byte, error) {
	k.log.Debugf("Handling request %d", id)

	// Decode request
	req, err := common.RequestFromJson(payload)
	if err != nil {
		k.log.Debug("Failed to decode currency transaction request: (%v)", err)
		return common.RespondFailure(err), nil
	}

	// Get supported chain
	rpc, ok := k.rpc[req.Ticker]
	if !ok {
		return nil, common.ErrWrongTicker
	}
	c, err := chain.GetChain(req.Ticker)
	if err != nil {
		return nil, err
	}

	var sendData *chain.HttpData
	if req.Command == command.DirectPost {
		// Directly send payload to rpc without passing to iChain
		sendData, err = wrapRequest(rpc.Url, req.Command, req.Payload)
		if err != nil {
			return nil, err
		}
	} else {
		// Wrap command into transactions
		sendData, err = c.WrapRequest(rpc.Url, req.Command, req.Payload)
		if err != nil {
			return nil, err
		}
	}

	// Send the transactions
	response, err := k.sendTransaction(rpc, sendData)
	if err != nil {
		return nil, fmt.Errorf("failed to send currency request: %v", err)
	}

	var result []byte
	if req.Command == command.DirectPost {
		// Directly return rpc response without passing to iChain
		result, err = json.Marshal(response)
	} else {
		// Unwrap responses
		result, err = c.UnwrapResponse(req.Command, response)
	}

	if err != nil {
		k.log.Debug("Error response for currency request: (%v)", err)
		return common.RespondFailure(err), nil
	}

	return common.RespondSuccess(string(result)), nil
}

// Called when cmd==0x01 (DirectPost), directly sends payload to rpcURL.
// Payload data is handled by User (Wallet)
func wrapRequest(rpcURL string, cmd uint8, payload []byte) (*chain.HttpData, error) {
	if cmd != command.DirectPost {
		return nil, fmt.Errorf("expect cmd to be 0x01, got %d", cmd)
	}
	return &chain.HttpData{Method: "POST", URL: rpcURL, Body: payload}, nil

}

// Halt : Stops the plugin
func (k *Currency) Halt() {

}

func (k *Currency) sendTransaction(rpc config.RPCMetadata, sendData *chain.HttpData) ([]chain.RPCResponse, error) {
	k.log.Debug("sendTransaction")

	var resp []chain.RPCResponse
	var singleResp chain.RPCResponse
	bodyReader := bytes.NewReader(sendData.Body)
	httpReq, err := http.NewRequest(sendData.Method, sendData.URL, bodyReader)
	if err != nil {
		return nil, err
	}
	httpReq.Close = true
	httpReq.Header.Set("Content-Type", "application/json")
	if rpc.User != "" && rpc.Pass != "" {
		httpReq.SetBasicAuth(rpc.User, rpc.Pass)
	}

	// send http request
	client := http.Client{}
	httpResponse, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	if httpResponse.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("currency RPC error status: %s", httpResponse.Status)
	}
	bodyBytes, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bodyBytes, &resp)
	if err != nil {
		err = json.Unmarshal(bodyBytes, &singleResp)
		if err != nil {
			return nil, err
		}
		resp = append(resp, singleResp)
	}
	return resp, nil
}

// New : Returns a pointer to a newly instantiated Currency struct
func New(cfg *config.Config) (*Currency, error) {
	currency := &Currency{
		rpc:    cfg.RPC,
		params: make(map[string]string),
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
	if err != nil {
		return nil, err
	}
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
