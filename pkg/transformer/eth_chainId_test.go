package transformer

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/qtumproject/janus/pkg/eth"
	"github.com/qtumproject/janus/pkg/internal"
	"github.com/qtumproject/janus/pkg/qtum"
)

func TestChainIdMainnet(t *testing.T) {
	testChainIdsImpl(t, "main", "0x22b8")
}

func TestChainIdTestnet(t *testing.T) {
	testChainIdsImpl(t, "test", "0x22b9")
}

func TestChainIdRegtest(t *testing.T) {
	testChainIdsImpl(t, "regtest", "0x22ba")
}

func TestChainIdUnknown(t *testing.T) {
	testChainIdsImpl(t, "???", "0x22ba")
}

func testChainIdsImpl(t *testing.T, chain string, expected string) {
	//preparing request
	requestParams := []json.RawMessage{}
	request, err := internal.PrepareEthRPCRequest(1, requestParams)
	if err != nil {
		t.Fatal(err)
	}

	mockedClientDoer := internal.NewDoerMappedMock()
	qtumClient, err := internal.CreateMockedClient(mockedClientDoer)
	if err != nil {
		t.Fatal(err)
	}

	//preparing client response
	getBlockCountResponse := qtum.GetBlockChainInfoResponse{Chain: chain}
	err = mockedClientDoer.AddResponseWithRequestID(2, qtum.MethodGetBlockChainInfo, getBlockCountResponse)
	if err != nil {
		t.Fatal(err)
	}

	//preparing proxy & executing request
	proxyEth := ProxyETHChainId{qtumClient}
	got, err := proxyEth.Request(request, nil)
	if err != nil {
		t.Fatal(err)
	}

	want := eth.ChainIdResponse(expected)
	if !reflect.DeepEqual(got, want) {
		t.Errorf(
			"error\ninput: %s\nwant: %s\ngot: %s",
			request,
			want,
			got,
		)
	}
}
