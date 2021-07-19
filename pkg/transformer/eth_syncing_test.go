package transformer

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/qtumproject/janus/pkg/eth"
	"github.com/qtumproject/janus/pkg/qtum"
)

func TestETHSyncRequest(t *testing.T) {
	//preparing the request
	requestParams := []json.RawMessage{} //eth_syncing has no params
	request, err := prepareEthRPCRequest(1, requestParams)
	if err != nil {
		t.Fatal(err)
	}

	mockedClientDoer := newDoerMappedMock()
	qtumClient, err := createMockedClient(mockedClientDoer)
	if err != nil {
		t.Fatal(err)
	}

	getETHSyncResponse := qtum.GetMiningResponse{Staking: true}
	err = mockedClientDoer.AddResponse(qtum.MethodGetStakingInfo, getETHSyncResponse)
	if err != nil {
		t.Fatal(err)
	}

	proxyEth := ProxyETHSyncing{qtumClient}
	got, err := proxyEth.Request(request)
	if err != nil {
		t.Fatal(err)
	}

	want := eth.SyncingResponse(true)
	if !reflect.DeepEqual(got, &want) {
		t.Errorf(
			"error\ninput: %s\nwant: %t\ngot: %t",
			request,
			want,
			got,
		)
	}

}
