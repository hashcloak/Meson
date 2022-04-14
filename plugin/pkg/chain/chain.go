package chain

import "github.com/hashcloak/Meson/plugin/pkg/common"

// HttpData is the common struct containing the body and url
type HttpData struct {
	Body []byte
	URL  string
}

// IChain is an abstraction for a cryptocurrency
// It creates raw transactions
type IChain interface {
	WrapPostRequest(rpcURL string, req *common.PostRequest) ([]HttpData, error)
	WrapQueryRequest(rpcURL string, req *common.QueryRequest) ([]HttpData, error)
}
