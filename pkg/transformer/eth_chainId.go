package transformer

import (
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/labstack/echo"
	"github.com/qtumproject/janus/pkg/eth"
	"github.com/qtumproject/janus/pkg/qtum"
)

type ProxyETHChainId struct {
	*qtum.Qtum
}

func (p *ProxyETHChainId) Method() string {
	return "eth_chainId"
}

func (p *ProxyETHChainId) Request(req *eth.JSONRPCRequest, c echo.Context) (interface{}, error) {
	var qtumresp *qtum.GetBlockChainInfoResponse
	if err := p.Qtum.Request(qtum.MethodGetBlockChainInfo, nil, &qtumresp); err != nil {
		return nil, err
	}

	var chainId *big.Int
	switch strings.ToLower(qtumresp.Chain) {
	case "main":
		chainId = big.NewInt(8888)
	case "test":
		chainId = big.NewInt(8889)
	case "regtest":
		chainId = big.NewInt(8890)
	default:
		chainId = big.NewInt(8890)
		p.GetDebugLogger().Log("method", p.Method(), "msg", "Unknown chain "+qtumresp.Chain)
	}

	return eth.ChainIdResponse(hexutil.EncodeBig(chainId)), nil
}
