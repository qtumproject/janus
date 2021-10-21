package transformer

import (
	"encoding/json"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"github.com/qtumproject/janus/pkg/eth"
)

type Web3Sha3 struct{}

func (p *Web3Sha3) Method() string {
	return "web3_sha3"
}

func (p *Web3Sha3) Request(rawreq *eth.JSONRPCRequest, c echo.Context) (interface{}, error) {
	var err error
	var req eth.Web3Sha3Request
	if err = json.Unmarshal(rawreq.Params, &req); err != nil {
		return nil, err
	}

	message := req.Message
	var decoded []byte
	// zero length should return "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"
	if len(message) != 0 {
		decoded, err = hexutil.Decode(string(message))
		if err != nil {
			return nil, errors.Wrap(err, "Failed to decode")
		}
	}

	return hexutil.Encode(crypto.Keccak256(decoded)), nil
}
