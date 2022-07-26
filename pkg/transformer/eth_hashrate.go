package transformer

import (
	"context"
	"math"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/labstack/echo"
	"github.com/qtumproject/janus/pkg/eth"
	"github.com/qtumproject/janus/pkg/qtum"
)

//ProxyETHGetHashrate implements ETHProxy
type ProxyETHHashrate struct {
	*qtum.Qtum
}

func (p *ProxyETHHashrate) Method() string {
	return "eth_hashrate"
}

func (p *ProxyETHHashrate) Request(_ *eth.JSONRPCRequest, c echo.Context) (interface{}, eth.JSONRPCError) {
	return p.request(c.Request().Context())
}

func (p *ProxyETHHashrate) request(ctx context.Context) (*eth.HashrateResponse, eth.JSONRPCError) {
	qtumresp, err := p.Qtum.GetHashrate(ctx)
	if err != nil {
		return nil, eth.NewCallbackError(err.Error())
	}

	// qtum res -> eth res
	return p.ToResponse(qtumresp), nil
}

func (p *ProxyETHHashrate) ToResponse(qtumresp *qtum.GetHashrateResponse) *eth.HashrateResponse {
	hexVal := hexutil.EncodeUint64(math.Float64bits(qtumresp.Difficulty))
	ethresp := eth.HashrateResponse(hexVal)
	return &ethresp
}
