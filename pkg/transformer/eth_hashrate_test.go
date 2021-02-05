package transformer

import (
	"encoding/json"
	"math/big"
	"reflect"
	"testing"

	"github.com/qtumproject/janus/pkg/eth"
	"github.com/qtumproject/janus/pkg/qtum"
)

func TestHashrateRequest(t *testing.T) {
	//preparing the request
	requestParams := []json.RawMessage{} //eth_hashrate has no params
	request, err := prepareEthRPCRequest(1, requestParams)
	if err != nil {
		panic(err)
	}

	mockedClientDoer := doerMappedMock{make(map[string][]byte)}
	qtumClient, err := createMockedClient(mockedClientDoer)
	if err != nil {
		panic(err)
	}

	getHashrateResponse := qtum.GetHashrateResponse{Difficulty: big.NewInt(457134700)}
	err = mockedClientDoer.AddResponse(2, qtum.MethodGetStakingInfo, getHashrateResponse)
	if err != nil {
		panic(err)
	}

	proxyEth := ProxyETHHashrate{qtumClient}
	got, err := proxyEth.Request(request)
	if err != nil {
		panic(err)
	}

	want := eth.HashrateResponse("0x1b3f526c")
	if !reflect.DeepEqual(got, &want) {
		t.Errorf(
			"error\ninput: %s\nwant: %s\ngot: %s",
			request,
			want,
			got,
		)
	}
}