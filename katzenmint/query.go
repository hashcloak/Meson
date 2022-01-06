package katzenmint

// Query represents the Query request
type Query struct {
	// version
	Version string

	// Epoch
	Epoch uint64

	// command
	Command Command

	// payload
	Payload string
}
