package command

const (
	PostTransaction uint8 = 0x00
)

// Request Types
type PostTransactionRequest struct {
	TxHex string
}

// Response Types
type PostTransactionResponse struct {
	TxHash string
}
