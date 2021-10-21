package transformer

import (
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"github.com/qtumproject/janus/pkg/eth"
	"github.com/qtumproject/janus/pkg/notifier"
	"github.com/qtumproject/janus/pkg/qtum"
)

// ETHSubscribe implements ETHProxy
type ETHSubscribe struct {
	*qtum.Qtum
	*notifier.Agent
}

func (p *ETHSubscribe) Method() string {
	return "eth_subscribe"
}

func (p *ETHSubscribe) Request(rawreq *eth.JSONRPCRequest, c echo.Context) (interface{}, error) {
	notifier := getNotifier(c)
	if notifier == nil {
		p.GetLogger().Log("msg", "eth_subscribe only supported over websocket")
		/*
			// TODO
			{
				"jsonrpc": "2.0",
				"id": 580,
				"error": {
					"code": -32601,
					"message": "The method eth_subscribe does not exist/is not available"
				}
			}
		*/
		return nil, errors.New("The method eth_subscribe does not exist/is not available")
	}

	var req eth.EthSubscriptionRequest
	if err := unmarshalRequest(rawreq.Params, &req); err != nil {
		return nil, err
	}

	return p.request(&req, notifier)
}

func (p *ETHSubscribe) request(req *eth.EthSubscriptionRequest, notifier *notifier.Notifier) (*eth.EthSubscriptionResponse, error) {
	notifier.ResponseRequired()
	id, err := p.NewSubscription(notifier, req)
	response := eth.EthSubscriptionResponse(id)
	return &response, err
}
