package transformer

import (
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"github.com/qtumproject/janus/pkg/eth"
	"github.com/qtumproject/janus/pkg/notifier"
	"github.com/qtumproject/janus/pkg/qtum"
)

// ETHUnsubscribe implements ETHProxy
type ETHUnsubscribe struct {
	*qtum.Qtum
	*notifier.Agent
}

func (p *ETHUnsubscribe) Method() string {
	return "eth_unsubscribe"
}

func (p *ETHUnsubscribe) Request(rawreq *eth.JSONRPCRequest, c echo.Context) (interface{}, error) {
	notifier := getNotifier(c)
	if notifier == nil {
		p.GetLogger().Log("msg", "eth_unsubscribe only supported over websocket")
		/*
			// TODO
			{
				"jsonrpc": "2.0",
				"id": 580,
				"error": {
					"code": -32601,
					"message": "The method eth_unsubscribe does not exist/is not available"
				}
			}
		*/
		return nil, errors.New("The method eth_subscribe does not exist/is not available")
	}

	var req eth.EthUnsubscribeRequest
	if err := unmarshalRequest(rawreq.Params, &req); err != nil {
		return nil, err
	}

	return p.request(&req, notifier)
}

func (p *ETHUnsubscribe) request(req *eth.EthUnsubscribeRequest, notifier *notifier.Notifier) (eth.EthUnsubscribeResponse, error) {
	if len(*req) != 1 {
		// TODO: Proper error response
		return false, errors.New("requires one parameter")
	}
	param := (*req)[0]
	success := notifier.Unsubscribe(param)
	return eth.EthUnsubscribeResponse(success), nil
}
