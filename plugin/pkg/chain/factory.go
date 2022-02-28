package chain

import (
	"fmt"
	"strings"
)

// GetChain takes a ticker symbol for a supported chain and returns an interface
// for that chain
func GetChain(ticker string) (IChain, error) {
	switch strings.ToUpper(ticker) {
	case "ETH":
		return &ETHChain{ticker: "ETH", chainID: 1}, nil
	case "ETC":
		return &ETHChain{ticker: "ETC", chainID: 61}, nil
	case "GOR":
		return &ETHChain{ticker: "GOR", chainID: 5}, nil
	case "RIN":
		return &ETHChain{ticker: "RIN", chainID: 4}, nil
	case "KOT":
		return &ETHChain{ticker: "KOT", chainID: 6}, nil
	case "TBC":
		return &CosmosChain{ticker: "TBNB", chainID: 0}, nil
	case "BC":
		return &CosmosChain{ticker: "BNB", chainID: 1}, nil
	case "TBSC":
		return &ETHChain{ticker: "TBNB", chainID: 97}, nil
	case "BSC":
		return &ETHChain{ticker: "BNB", chainID: 56}, nil
	default:
		return nil, fmt.Errorf("Unsupported chain")
	}
}
