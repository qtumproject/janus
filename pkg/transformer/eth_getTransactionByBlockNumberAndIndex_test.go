package transformer

import (
	"encoding/json"
	"testing"

	"github.com/qtumproject/janus/pkg/internal"
	"github.com/qtumproject/janus/pkg/qtum"
)

func initializeProxyETHGetTransactionByBlockNumberAndIndex(qtumClient *qtum.Qtum) ETHProxy {
	return &ProxyETHGetTransactionByBlockNumberAndIndex{qtumClient}
}

func TestGetTransactionByBlockNumberAndIndex(t *testing.T) {
	testETHProxyRequest(
		t,
		initializeProxyETHGetTransactionByBlockNumberAndIndex,
		[]json.RawMessage{[]byte(`"` + internal.GetTransactionByHashBlockNumberHex + `"`), []byte(`"0x0"`)},
		internal.GetTransactionByHashResponseData,
	)
}
