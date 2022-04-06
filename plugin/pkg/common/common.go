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
)

var (
	jsonHandle                codec.JsonHandle
	ErrInvalidCurrencyRequest = errors.New("kaetzchen/meson: invalid request")
	errInvalidJson            = errors.New("meson: bad json")
	errWrongVersion           = errors.New("meson: request version mismatch")
)

type CurrencyRequest struct {
	Version int
	Tx      string
	Ticker  string
}

func NewRequest(ticker string, hexBlob string) *CurrencyRequest {
	return &CurrencyRequest{
		Version: CurrencyVersion,
		Ticker:  ticker,
		Tx:      hexBlob,
	}
}

func RequestFromJson(rawRequest []byte) (*CurrencyRequest, error) {
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
	return &req, nil
}

func (c *CurrencyRequest) ToJson() []byte {
	var request []byte
	enc := codec.NewEncoderBytes(&request, &jsonHandle)
	_ = enc.Encode(c)
	return request
}

type CurrencyResponse struct {
	Version int
	Message string
	Error   string
}

func NewResponse(message string, errMsg string) *CurrencyResponse {
	return &CurrencyResponse{
		Version: CurrencyVersion,
		Message: message,
		Error:   errMsg,
	}
}

func (c *CurrencyResponse) ToJson() []byte {
	var response []byte
	enc := codec.NewEncoderBytes(&response, &jsonHandle)
	_ = enc.Encode(c)
	return response
}

func RespondSuccess(message string) []byte {
	return NewResponse(message, "").ToJson()
}

func RespondFailure(err error) []byte {
	return NewResponse("", err.Error()).ToJson()
}

func ResponseFromJson(rawResponse []byte) (string, error) {
	resp := CurrencyResponse{}
	dec := codec.NewDecoderBytes(rawResponse, &jsonHandle)
	if err := dec.Decode(&resp); err != nil {
		return "", errInvalidJson
	}
	if resp.Error != "" {
		return "", errors.New(resp.Error)
	}
	return resp.Message, nil
}
