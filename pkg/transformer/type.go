package transformer

import (
	"errors"

	"github.com/labstack/echo"
	"github.com/qtumproject/janus/pkg/eth"
)

var UnmarshalRequestErr = errors.New("Input is invalid")

type Option func(*Transformer) error

type ETHProxy interface {
	Request(*eth.JSONRPCRequest, echo.Context) (interface{}, error)
	Method() string
}
