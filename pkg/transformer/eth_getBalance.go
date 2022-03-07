package transformer

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/labstack/echo"
	"github.com/qtumproject/janus/pkg/eth"
	"github.com/qtumproject/janus/pkg/qtum"
	"github.com/qtumproject/janus/pkg/utils"
)

// ProxyETHGetBalance implements ETHProxy
type ProxyETHGetBalance struct {
	*qtum.Qtum
}

func (p *ProxyETHGetBalance) Method() string {
	return "eth_getBalance"
}

func (p *ProxyETHGetBalance) Request(rawreq *eth.JSONRPCRequest, c echo.Context) (interface{}, eth.JSONRPCError) {
	var req eth.GetBalanceRequest
	if err := unmarshalRequest(rawreq.Params, &req); err != nil {
		// TODO: Correct error code?
		return nil, eth.NewInvalidParamsError(err.Error())
	}

	addr := utils.RemoveHexPrefix(req.Address)
	{
		// is address a contract or an account?
		qtumreq := qtum.GetAccountInfoRequest(addr)
		qtumresp, err := p.GetAccountInfo(&qtumreq)

		// the address is a contract
		if err == nil {
			// the unit of the balance Satoshi
			p.GetDebugLogger().Log("method", p.Method(), "address", req.Address, "msg", "is a contract")

			// 1 QTUM = 10 ^ 8 Satoshi
			balance := new(big.Int).SetUint64(uint64(qtumresp.Balance))

			if !p.GetFlagBool(qtum.FLAG_USE_SATOSHI_INSTEAD_OF_WEI) {
				//Balance for ETH response is represented in Weis (1 QTUM Satoshi = 10 ^ 10 Wei)
				balance = balance.Mul(balance, big.NewInt(10000000000))
			}

			return hexutil.EncodeBig(balance), nil
		}
	}

	{
		// try account
		base58Addr, err := p.FromHexAddress(addr)
		if err != nil {
			p.GetDebugLogger().Log("method", p.Method(), "address", req.Address, "msg", "error parsing address", "error", err)
			return nil, eth.NewCallbackError(err.Error())
		}

		qtumreq := qtum.GetAddressBalanceRequest{Address: base58Addr}
		qtumresp, err := p.GetAddressBalance(&qtumreq)
		if err != nil {
			if err == qtum.ErrInvalidAddress {
				// invalid address should return 0x0
				return "0x0", nil
			}
			p.GetDebugLogger().Log("method", p.Method(), "address", req.Address, "msg", "error getting address balance", "error", err)
			return nil, eth.NewCallbackError(err.Error())
		}

		// 1 QTUM = 10 ^ 8 Satoshi
		balance := new(big.Int).SetUint64(qtumresp.Balance)

		if !p.GetFlagBool(qtum.FLAG_USE_SATOSHI_INSTEAD_OF_WEI) {
			//Balance for ETH response is represented in Weis (1 QTUM Satoshi = 10 ^ 10 Wei)
			balance = balance.Mul(balance, big.NewInt(10000000000))
		}

		return hexutil.EncodeBig(balance), nil
	}
}
