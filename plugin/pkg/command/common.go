package command

const (
	PostTransaction uint8 = 0x00
	DirectPost      uint8 = 0x01 // Directly send payload to rpc
)

// Request Types
type PostTransactionRequest struct {
	TxHex string
}

// Response Types
type PostTransactionResponse struct {
	TxHash string
}
