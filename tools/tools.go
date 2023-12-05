//go:build tools
// +build tools

package tools

import (
	// tools we used
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "honnef.co/go/tools/cmd/staticcheck"
)
