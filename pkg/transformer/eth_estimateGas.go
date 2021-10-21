package transformer

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"github.com/qtumproject/janus/pkg/eth"
	"github.com/qtumproject/janus/pkg/qtum"
)

// 22000
var NonContractVMGasLimit = "0x55f0"
var ErrExecutionReverted = errors.New("execution reverted")

// ProxyETHEstimateGas implements ETHProxy
type ProxyETHEstimateGas struct {
	*ProxyETHCall
}

func (p *ProxyETHEstimateGas) Method() string {
	return "eth_estimateGas"
}

func (p *ProxyETHEstimateGas) Request(rawreq *eth.JSONRPCRequest, c echo.Context) (interface{}, error) {
	var ethreq eth.CallRequest
	if err := unmarshalRequest(rawreq.Params, &ethreq); err != nil {
		return nil, err
	}

	if ethreq.Data == "" {
		response := eth.EstimateGasResponse(NonContractVMGasLimit)
		return &response, nil
	}

	// when supplying this parameter to callcontract to estimate gas in the qtum api
	// if there isn't enough gas specified here, the result will be an exception
	// Excepted = "OutOfGasIntrinsic"
	// Gas = "the supplied value"
	// this is different from geth's behavior
	// which will return a used gas value that is higher than the incoming gas parameter
	// so we set this to nil so that callcontract will return the actual gas estimate
	ethreq.Gas = nil

	// eth req -> qtum req
	qtumreq, err := p.ToRequest(&ethreq)
	if err != nil {
		return nil, err
	}

	// qtum [code: -5] Incorrect address occurs here
	qtumresp, err := p.CallContract(qtumreq)
	if err != nil {
		return nil, err
	}

	return p.toResp(qtumresp)
}

func (p *ProxyETHEstimateGas) toResp(qtumresp *qtum.CallContractResponse) (*eth.EstimateGasResponse, error) {
	if qtumresp.ExecutionResult.Excepted != "None" {
		// TODO: Return code -32000
		return nil, ErrExecutionReverted
	}
	gas := eth.EstimateGasResponse(hexutil.EncodeUint64(uint64(qtumresp.ExecutionResult.GasUsed)))
	p.GetDebugLogger().Log(p.Method(), gas)
	return &gas, nil
}
