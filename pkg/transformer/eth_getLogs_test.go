package transformer

import (
	"math/big"
	"encoding/json"
	"testing"

	"github.com/qtumproject/janus/pkg/eth"
	"github.com/qtumproject/janus/pkg/qtum"

)

func TestGetLogs(t *testing.T) {
	//perparing request
	request := eth.GetLogsRequest{
		FromBlock: "latest"
		ToBlock:   "latest"
		Address:   "0x1e6f89d7399081b4f8f8aa1ae2805a5efff2f960"
		Topics:	   ["0x0f6798a560793a54c3bcfe86a93cde1e73087d944c0ea20544137d4121396885",
		            "0x000000000000000000000000cb3cb8375fe457a11f041f9ff55373e1a5a78d19"]
	}
	requestRaw, err := json.Marshal(&request)
	if err != nil {
		panic(err)
	}
	requestParamsArray := []json.RawMessage{requestRaw}
	requestRPC, err :=  prepareEthRPCRequest(1, requestParamsArray)

	clientDoerMock := doerMappedMock{make(map[string][]byte)}
	qtumClient, err := createMockedClient(clientDoerMock)

	//prepare response
	getLogsResponse := qtum.GetLogsResponse{
		[]qtum.TransactionReceipt{
			qtum.TransactionReceipt: struct {
				BlockHash:
				BlockNumber:
				TransactionHash:
				TransactionIndex:
				From:
				To:
				CumulativeGasUsed:
				GasUsed:
				ContractAddress:
				Excepted:
				Log:
				OutputIndex:
			}
		}
	}

	//Add responses

	//Prepare proxy & execute


	//got

	//want

}