package transformer

import (
	"fmt"
	"math/big"

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

func (p *ProxyETHCall) Request(rawreq *eth.JSONRPCRequest) (interface{}, error) {
	var req eth.CallRequest
	if err := unmarshalRequest(rawreq.Params, &req); err != nil {
		return nil, err
	}

	return p.request(&req)
}

func (p *ProxyETHCall) request(ethreq *eth.CallRequest) (interface{}, error) {
	// eth req -> qtum req
	qtumreq, err := p.ToRequest(ethreq)
	if err != nil {
		return nil, err
	}

	qtumresp, err := p.CallContract(qtumreq)
	if err != nil {
		return nil, err
	}

	// qtum res -> eth res
	return p.ToResponse(qtumresp), nil
}

func (p *ProxyETHCall) ToRequest(ethreq *eth.CallRequest) (*qtum.CallContractRequest, error) {
	from := ethreq.From
	var err error
	if utils.IsEthHexAddress(from) {
		from, err = p.FromHexAddress(from)
		if err != nil {
			return nil, err
		}
	}

	return &qtum.CallContractRequest{
		To:       ethreq.To,
		From:     from,
		Data:     ethreq.Data,
		// TODO: qtum [code: -3] Invalid value for gasLimit (Minimum is: 10000)
		// Incorrect gas format
		GasLimit: big.NewInt(10000),
		//GasLimit: ethreq.Gas.Int,
	}, nil
}

func (p *ProxyETHCall) ToResponse(qresp *qtum.CallContractResponse) interface{} {
	excepted := qresp.ExecutionResult.Excepted
	if excepted != "None" {
		return &eth.JSONRPCError{
			Code:    -32000,
			Message: fmt.Sprintf("VM exception: %s", excepted),
			// To see how eth_call supports revert reason, see:
			// https://gist.github.com/hayeah/795bc18a683053218fb3ff5032d31144
			//
			// Data: ...
		}
	}

	data := utils.AddHexPrefix(qresp.ExecutionResult.Output)
	qtumresp := eth.CallResponse(data)
	return &qtumresp
}
