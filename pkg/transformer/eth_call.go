package transformer

import (
	"context"
	"math/big"

	"github.com/labstack/echo"
	"github.com/qtumproject/janus/pkg/eth"
	"github.com/qtumproject/janus/pkg/qtum"
	"github.com/qtumproject/janus/pkg/utils"
)

// ProxyETHCall implements ETHProxy
type ProxyETHCall struct {
	*qtum.Qtum
}

func (p *ProxyETHCall) Method() string {
	return "eth_call"
}

func (p *ProxyETHCall) Request(rawreq *eth.JSONRPCRequest, c echo.Context) (interface{}, eth.JSONRPCError) {
	var req eth.CallRequest
	if err := unmarshalRequest(rawreq.Params, &req); err != nil {
		// TODO: Is this correct error code?
		return nil, eth.NewInvalidParamsError(err.Error())
	}

	return p.request(c.Request().Context(), &req)
}

func (p *ProxyETHCall) request(ctx context.Context, ethreq *eth.CallRequest) (interface{}, eth.JSONRPCError) {
	// eth req -> qtum req
	qtumreq, jsonErr := p.ToRequest(ethreq)
	if jsonErr != nil {
		return nil, jsonErr
	}
	if qtumreq.GasLimit != nil && qtumreq.GasLimit.Cmp(big.NewInt(40000000)) > 0 {
		qtumresp := eth.CallResponse("0x")
		p.Qtum.GetLogger().Log("msg", "Caller gas above allowance, capping", "requested", qtumreq.GasLimit.Int64(), "cap", "40,000,000")
		return &qtumresp, nil
	}

	qtumresp, err := p.CallContract(ctx, qtumreq)
	if err != nil {
		if err == qtum.ErrInvalidAddress {
			qtumresp := eth.CallResponse("0x")
			return &qtumresp, nil
		}

		return nil, eth.NewCallbackError(err.Error())
	}

	// qtum res -> eth res
	return p.ToResponse(qtumresp), nil
}

func (p *ProxyETHCall) ToRequest(ethreq *eth.CallRequest) (*qtum.CallContractRequest, eth.JSONRPCError) {
	from := ethreq.From
	var err error
	if utils.IsEthHexAddress(from) {
		from, err = p.FromHexAddress(from)
		if err != nil {
			return nil, eth.NewCallbackError(err.Error())
		}
	}

	var gasLimit *big.Int
	if ethreq.Gas != nil {
		gasLimit = ethreq.Gas.Int
	}

	if gasLimit != nil && gasLimit.Int64() < MinimumGasLimit {
		p.GetLogger().Log("msg", "Gas limit is too low", "gasLimit", gasLimit.String())
	}

	return &qtum.CallContractRequest{
		To:       ethreq.To,
		From:     from,
		Data:     ethreq.Data,
		GasLimit: gasLimit,
	}, nil
}

func (p *ProxyETHCall) ToResponse(qresp *qtum.CallContractResponse) interface{} {
	if qresp.ExecutionResult.Output == "" {
		return eth.NewJSONRPCError(
			-32000,
			"Revert: executionResult output is empty",
			nil,
		)
	}

	data := utils.AddHexPrefix(qresp.ExecutionResult.Output)
	qtumresp := eth.CallResponse(data)
	return &qtumresp

}
