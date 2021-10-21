package transformer

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/qtumproject/janus/pkg/internal"
)

func TestGetCompilersReturnsEmptyArray(t *testing.T) {
	//preparing the request
	requestParams := []json.RawMessage{} //eth_getCompilers has no params
	request, err := internal.PrepareEthRPCRequest(1, requestParams)
	if err != nil {
		t.Fatal(err)
	}

	proxyEth := ETHGetCompilers{}
	got, err := proxyEth.Request(request, nil)
	if err != nil {
		t.Fatal(err)
	}

	if fmt.Sprintf("%v", got) != "[]" {
		t.Errorf(
			"error\nwant: '[]'\ngot: '%v'",
			got,
		)
	}
}
