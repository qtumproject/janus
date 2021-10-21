package transformer

import (
	"encoding/json"
	"testing"

	"github.com/qtumproject/janus/pkg/internal"
	"github.com/qtumproject/janus/pkg/qtum"
)

func initializeProxyETHGetTransactionByBlockHashAndIndex(qtumClient *qtum.Qtum) ETHProxy {
	return &ProxyETHGetTransactionByBlockHashAndIndex{qtumClient}
}

func TestGetTransactionByBlockHashAndIndex(t *testing.T) {
	testETHProxyRequest(
		t,
		initializeProxyETHGetTransactionByBlockHashAndIndex,
		[]json.RawMessage{[]byte(`"` + internal.GetTransactionByHashBlockHash + `"`), []byte(`"0x0"`)},
		internal.GetTransactionByHashResponseData,
	)
}
