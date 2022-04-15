package chain

import (
	"encoding/json"
	"testing"

	"github.com/hashcloak/Meson/plugin/pkg/command"
)

const (
	broadcastTxAsync = "/broadcast_tx_async?tx=0x"
)

func TestChainFactoryError(t *testing.T) {
	_, err := GetChain("SOMETHING")
	if err == nil {
		t.Fatalf("Should return an error")
	}
}

func TestChainFactoryErrorEmpty(t *testing.T) {
	_, err := GetChain("")
	if err == nil {
		t.Fatalf("Should return an error")
	}
}

func TestEthereumChainURLEmptyValue(t *testing.T) {
	chainInterface, _ := GetChain("ETH")
	req, _ := json.Marshal(command.PostTransactionRequest{TxHex: ""})
	_, err := chainInterface.WrapRequest("", command.PostTransaction, req)
	if err == nil {
		t.Fatalf("Should return an error")
	}
	expectedErrorValue := "non existent RPC URL for Ethereum chain"
	if err.Error() != expectedErrorValue {
		t.Fatalf("Not the right error value.\nExpected: %s\nGot: %s", expectedErrorValue, err.Error())
	}
}

func TestEthereumChainURLValue(t *testing.T) {
	chainInterface, _ := GetChain("ETH")
	expectedURL := "EXPECTED_URL"
	req, _ := json.Marshal(command.PostTransactionRequest{TxHex: ""})
	postRequest, _ := chainInterface.WrapRequest(expectedURL, command.PostTransaction, req)
	if postRequest[0].URL != expectedURL {
		t.Fatalf("Expected %s, got %s", expectedURL, postRequest[0].URL)
	}
}

func TestEthereumChainTxnInBody(t *testing.T) {
	chainInterface, _ := GetChain("ETH")
	txn := `"TXN"`
	req, _ := json.Marshal(command.PostTransactionRequest{TxHex: txn})
	postRequest, _ := chainInterface.WrapRequest("URL", command.PostTransaction, req)
	var gotValue ethRequest
	err := json.Unmarshal(postRequest[0].Body, &gotValue)
	if err != nil {
		t.Fatalf("err unmarshal: %v\n", err)
	}
	gotParams, ok := gotValue.Params.([]interface{})
	if ok != true {
		t.Fatalf("err unmarshal Param")
	}
	if len(gotParams) != 1 {
		t.Fatalf("Length expected to be %d, got %d", 1, len(gotParams))
	}
	if gotParams[0].(string) != txn {
		t.Fatalf("Expected %s, got %s", txn, gotParams[0].(string))
	}
}

func TestCosmosChainURLEmpty(t *testing.T) {
	chainInterface, _ := GetChain("TBC")
	req, _ := json.Marshal(command.PostTransactionRequest{TxHex: ""})
	_, err := chainInterface.WrapRequest("", command.PostTransaction, req)
	if err == nil {
		t.Fatalf("Should return an error when passed empty URL")
	}
}
func TestCosmosChainBody(t *testing.T) {
	chainInterface, _ := GetChain("TBC")
	req, _ := json.Marshal(command.PostTransactionRequest{TxHex: ""})
	postRequest, _ := chainInterface.WrapRequest("URL", command.PostTransaction, req)
	if len(postRequest[0].Body) > 0 {
		t.Fatalf("Body should be empty for cosmos request")
	}
}
func TestCosmosChainURLAppend(t *testing.T) {
	chainInterface, _ := GetChain("TBC")
	inputTxn := "EXPECTED_TXN"
	inputURL := "URL"
	expectedResult := inputURL + broadcastTxAsync + inputTxn
	req, _ := json.Marshal(command.PostTransactionRequest{TxHex: inputTxn})
	postRequest, _ := chainInterface.WrapRequest(inputURL, command.PostTransaction, req)
	if postRequest[0].URL != expectedResult {
		t.Fatalf("URL should have value %s, got %s", broadcastTxAsync, postRequest[0].URL)
	}
}
func TestCosmosChainURL(t *testing.T) {
	chainInterface, _ := GetChain("TBC")
	expectedURL := "EXPECTED_URL"
	req, _ := json.Marshal(command.PostTransactionRequest{TxHex: ""})
	postRequest, _ := chainInterface.WrapRequest(expectedURL, command.PostTransaction, req)
	if postRequest[0].URL != expectedURL+broadcastTxAsync {
		t.Fatalf("URL should have value %s, got %s", broadcastTxAsync, postRequest[0].URL)
	}
}
