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
	case "KOT":
		return &ETHChain{ticker: "KOT", chainID: 6}, nil
	case "SEP":
		return &ETHChain{ticker: "SEP", chainID: 11155111}, nil
	case "HOL":
		return &ETHChain{ticker: "HOL", chainID: 17000}, nil
	case "TBC":
		return &CosmosChain{ticker: "TBNB", chainID: 0}, nil
	case "BC":
		return &CosmosChain{ticker: "BNB", chainID: 1}, nil
	case "TBSC":
		return &ETHChain{ticker: "TBNB", chainID: 97}, nil
	case "BSC":
		return &ETHChain{ticker: "BNB", chainID: 56}, nil
	case "TMAT":
		return &ETHChain{ticker: "TMATIC", chainID: 80001}, nil
	case "MAT":
		return &ETHChain{ticker: "MATIC", chainID: 137}, nil
	case "TARB":
		return &ETHChain{ticker: "TARB", chainID: 421611}, nil
	case "ARB":
		return &ETHChain{ticker: "ARB", chainID: 42161}, nil
	case "TOPT":
		return &ETHChain{ticker: "TOPT", chainID: 69}, nil
	case "OPT":
		return &ETHChain{ticker: "OPT", chainID: 10}, nil
	case "BTC":
		return &BTCChain{ticker: "BTC", testnet: false}, nil
	case "TBTC":
		return &BTCChain{ticker: "BTC", testnet: true}, nil
	default:
		return nil, fmt.Errorf("unsupported chain")
	}
}
