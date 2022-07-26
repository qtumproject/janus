package transformer

import (
	"context"
	"encoding/json"

	"github.com/labstack/echo"
	"github.com/qtumproject/janus/pkg/conversion"
	"github.com/qtumproject/janus/pkg/eth"
	"github.com/qtumproject/janus/pkg/qtum"
	"github.com/qtumproject/janus/pkg/utils"
)

// ProxyETHGetLogs implements ETHProxy
type ProxyETHGetLogs struct {
	*qtum.Qtum
}

func (p *ProxyETHGetLogs) Method() string {
	return "eth_getLogs"
}

func (p *ProxyETHGetLogs) Request(rawreq *eth.JSONRPCRequest, c echo.Context) (interface{}, eth.JSONRPCError) {
	var req eth.GetLogsRequest
	if err := unmarshalRequest(rawreq.Params, &req); err != nil {
		// TODO: Correct error code?
		return nil, eth.NewInvalidParamsError(err.Error())
	}

	// TODO: Graph Node is sending the topic
	// if len(req.Topics) != 0 {
	// 	return nil, errors.New("topics is not supported yet")
	// }

	// Calls ToRequest in order transform ETH-Request to a Qtum-Request
	qtumreq, err := p.ToRequest(c.Request().Context(), &req)
	if err != nil {
		return nil, err
	}

	return p.request(c.Request().Context(), qtumreq)
}

func (p *ProxyETHGetLogs) request(ctx context.Context, req *qtum.SearchLogsRequest) (*eth.GetLogsResponse, eth.JSONRPCError) {
	receipts, err := conversion.SearchLogsAndFilterExtraTopics(ctx, p.Qtum, req)
	if err != nil {
		return nil, err
	}

	logs := make([]eth.Log, 0)
	for _, receipt := range receipts {
		r := qtum.TransactionReceipt(receipt)
		logs = append(logs, conversion.ExtractETHLogsFromTransactionReceipt(r, r.Log)...)
	}

	resp := eth.GetLogsResponse(logs)
	return &resp, nil
}

func (p *ProxyETHGetLogs) ToRequest(ctx context.Context, ethreq *eth.GetLogsRequest) (*qtum.SearchLogsRequest, eth.JSONRPCError) {
	//transform EthRequest fromBlock to QtumReq fromBlock:
	from, err := getBlockNumberByRawParam(ctx, p.Qtum, ethreq.FromBlock, true)
	if err != nil {
		return nil, err
	}

	//transform EthRequest toBlock to QtumReq toBlock:
	to, err := getBlockNumberByRawParam(ctx, p.Qtum, ethreq.ToBlock, true)
	if err != nil {
		return nil, err
	}

	//transform EthReq address to QtumReq address:
	var addresses []string
	if ethreq.Address != nil {
		if isBytesOfString(ethreq.Address) {
			var addr string
			if jsonErr := json.Unmarshal(ethreq.Address, &addr); jsonErr != nil {
				return nil, eth.NewInvalidParamsError(jsonErr.Error())
			}
			addresses = append(addresses, addr)
		} else {
			if jsonErr := json.Unmarshal(ethreq.Address, &addresses); jsonErr != nil {
				return nil, eth.NewInvalidParamsError(jsonErr.Error())
			}
		}
		for i := range addresses {
			addresses[i] = utils.RemoveHexPrefix(addresses[i])
		}
	}

	//transform EthReq topics to QtumReq topics:
	topics, topicsErr := eth.TranslateTopics(ethreq.Topics)
	if topicsErr != nil {
		return nil, eth.NewCallbackError(topicsErr.Error())
	}

	return &qtum.SearchLogsRequest{
		Addresses: addresses,
		FromBlock: from,
		ToBlock:   to,
		Topics:    qtum.NewSearchLogsTopics(topics),
	}, nil
}
