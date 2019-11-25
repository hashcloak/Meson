// common.go - Crypto currency common client and server code.
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

package common

import (
	"bytes"
	"errors"

	"github.com/ugorji/go/codec"
)

const (
	CurrencyVersion    = 0
	CurrencyCapability = "currency"
	CurrencyTicker     = "ticker"
	CurrencyChainId    = "chainId"
)

var (
	jsonHandle                codec.JsonHandle
	ErrInvalidCurrencyRequest = errors.New("kaetzchen/meson: invalid request")
	errInvalidJson            = errors.New("meson: bad json")
	errWrongVersion           = errors.New("meson: request version mismatch")
	errWrongTicker            = errors.New("meson: request ticker mismatch")
	errWrongChainID           = errors.New("meson: request chaindId mismatch")
)

type CurrencyRequest struct {
	Version int
	Tx      string
	Ticker  string
	ChainID int
}

func NewRequest(ticker string, hexBlob string, chaindId int) *CurrencyRequest {
	return &CurrencyRequest{
		Version: CurrencyVersion,
		Ticker:  ticker,
		Tx:      hexBlob,
		ChainID: chaindId,
	}
}

func RequestFromJson(expectedTicker string, expectedChainId int, rawRequest []byte) (*CurrencyRequest, error) {
	// Parse out the request payload.
	req := CurrencyRequest{}
	dec := codec.NewDecoderBytes(bytes.TrimRight(rawRequest, "\x00"), &jsonHandle)
	if err := dec.Decode(&req); err != nil {
		return nil, errInvalidJson
	}

	// Sanity check the request.
	if req.Version != CurrencyVersion {
		return nil, errWrongVersion
	}
	if req.Ticker != expectedTicker {
		return nil, errWrongTicker
	}
	if req.ChainID != expectedChainId {
		return nil, errWrongChainID
	}
	return &req, nil
}

func (c *CurrencyRequest) ToJson() []byte {
	var request []byte
	enc := codec.NewEncoderBytes(&request, &jsonHandle)
	enc.Encode(c)
	return request
}

type CurrencyResponse struct {
	Version    int
	StatusCode int
	Message    string
}

func NewResponse(status int, message string) *CurrencyResponse {
	return &CurrencyResponse{
		Version:    CurrencyVersion,
		StatusCode: status,
		Message:    message,
	}
}

func (c *CurrencyResponse) ToJson() []byte {
	var response []byte
	enc := codec.NewEncoderBytes(&response, &jsonHandle)
	enc.Encode(c)
	return response
}
