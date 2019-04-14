package proxy

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarshalRequest(t *testing.T) {
	request, err := marshalRequest(1, []string{"0xd46e8dd67c5d32be8d46e8dd67c5d32be8058bb8eb970870f072445675058bb8eb970870f072445675"})

	type Request struct {
		Id      int       `json:"id"`
		Jsonrpc string    `json:"jsonrpc"`
		Method  string	  `json:"method"`
		Params  []string  `json:"params"`
	}

	r := Request {
		Id:      1,
		Jsonrpc: "2.0",
		Method:  "eth_sendRawTransaction",
		Params:  []string{"0xd46e8dd67c5d32be8d46e8dd67c5d32be8058bb8eb970870f072445675058bb8eb970870f072445675"},
	}

	expectedRequest, expectedError := json.Marshal(r)
	assert.Nil(t, err)
	assert.Equal(t, expectedRequest, request)
	assert.Equal(t, expectedError, err)
}
