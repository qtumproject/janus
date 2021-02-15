package transformer

import (
	"errors"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/qtumproject/janus/pkg/eth"
	"github.com/qtumproject/janus/pkg/qtum"
	"github.com/qtumproject/janus/pkg/utils"
	"github.com/shopspring/decimal"
)

// ProxyETHGetBalance implements ETHProxy
type ProxyETHGetBalance struct {
	*qtum.Qtum
}

func (p *ProxyETHGetBalance) Method() string {
	return "eth_getBalance"
}

func (p *ProxyETHGetBalance) Request(rawreq *eth.JSONRPCRequest) (interface{}, error) {
	var req eth.GetBalanceRequest
	if err := unmarshalRequest(rawreq.Params, &req); err != nil {
		return nil, err
	}

	addr := utils.RemoveHexPrefix(req.Address)
	{
		// is address a contract or an account?
		qtumreq := qtum.GetAccountInfoRequest(addr)
		qtumresp, err := p.GetAccountInfo(&qtumreq)

		// the address is a contract
		if err == nil {
			// the unit of the balance Satoshi
			return hexutil.EncodeUint64(uint64(qtumresp.Balance)), nil
		}
	}

	{
		// try account
		base58Addr, err := p.FromHexAddress(addr)
		if err != nil {
			return nil, err
		}

		qtumreq := qtum.NewListUnspentRequest(base58Addr)
		qtumresp, err := p.ListUnspent(qtumreq)
		if err != nil {
			return nil, err
		}


		balance := decimal.NewFromFloat(0)
		for _, utxo := range *qtumresp {
			balance = balance.Add(utxo.Amount)
		}

		// 1 QTUM = 10 ^ 8 Satoshi
		balance = balance.Mul(decimal.NewFromFloat(1e18))
		floatBalance, exact := balance.Float64()
		
		if exact != true {
			return exact, errors.New("precision error:  float64 value does not represent the original decimal precisely")
		}
		
		return hexutil.EncodeUint64(uint64(floatBalance)), nil
	}
}
