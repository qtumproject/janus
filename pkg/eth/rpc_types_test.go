package eth

import (
	"encoding/json"
	"testing"
)

func TestEthLogSubscriptionRequestSerialization(t *testing.T) {
	jsonValue := `["logs",{"address":"0x8320fe7702b96808f7bbc0d4a888ed1468216cfd","topics":["0xd78a0cb8bb633d06981248b816e7bd33c2a35a6089241d099fa519e361cab902"]}]`
	var request EthSubscriptionRequest
	err := json.Unmarshal([]byte(jsonValue), &request)
	if err != nil {
		t.Fatal(err)
	}
	asJson, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	if string(asJson) != jsonValue {
		t.Fatalf(`"%s" != "%s"\n`, string(asJson), jsonValue)
	}
}

func TestEthNewPendingTransactionsRequestSerialization(t *testing.T) {
	jsonValue := `["newPendingTransactions"]`
	var request EthSubscriptionRequest
	err := json.Unmarshal([]byte(jsonValue), &request)
	if err != nil {
		t.Fatal(err)
	}
	asJson, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	if string(asJson) != jsonValue {
		t.Fatalf(`"%s" != "%s"\n`, string(asJson), jsonValue)
	}
}
