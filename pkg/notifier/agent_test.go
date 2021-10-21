package notifier

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/qtumproject/janus/pkg/eth"
	"github.com/qtumproject/janus/pkg/internal"
	"github.com/qtumproject/janus/pkg/qtum"
)

func TestAgentAddSubscriptionLogs(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	expectedSubscriptionID := "0x08e2af779d38a09e4c11442d9de22413"
	want := `{"subscription":"` + expectedSubscriptionID + `","result":{"address":"0x0000000000000000000000000000000000000000","blockHash":"0xbba11e1bacc69ba535d478cf1f2e542da3735a517b0b8eebaf7e6bb25eeb48c5","blockNumber":"0xf8f","data":"0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000","logIndex":"0x0","topics":["0xtopic1"],"transactionHash":"0x11e97fa5877c5df349934bafc02da6218038a427e8ed081f048626fa6eb523f5","transactionIndex":"0x2"}}`

	doer := internal.NewDoerMappedMock()
	topic1 := "topic1"

	doer.AddResponse(qtum.MethodWaitForLogs, qtum.WaitForLogsResponse{
		Entries: []qtum.WaitForLogsEntry{
			internal.QtumWaitForLogsEntry(qtum.Log{
				Address: internal.QtumTransactionReceipt(nil).ContractAddress,
				Topics:  []string{topic1},
				Data:    "00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
			}),
		},
		Count:     1,
		NextBlock: internal.QtumTransactionReceipt(nil).BlockNumber + 1,
	})

	mockedClient, err := internal.CreateMockedClient(doer)
	if err != nil {
		t.Fatal(err)
	}
	agent := NewAgent(ctx, mockedClient, nil)

	notifierContext, cancelNotifierContext := context.WithCancel(ctx)

	sentValuesChannel := make(chan interface{}, 10)
	sentValuesMutex := sync.Mutex{}
	sentValues := [][]byte{}

	send := func(v []byte) error {
		sentValuesMutex.Lock()
		defer sentValuesMutex.Unlock()
		sentValues = append(sentValues, v)
		defer func() { sentValuesChannel <- nil }()
		return nil
	}

	notifier := NewNotifier(notifierContext, cancelNotifierContext, send, log.NewLogfmtLogger(os.Stdout))

	id, err := agent.NewSubscription(notifier, &eth.EthSubscriptionRequest{
		Method: "logs",
		Params: &eth.EthLogSubscriptionParameter{
			Address: internal.QtumTransactionReceipt(nil).ContractAddress,
			Topics: []interface{}{
				topic1,
			},
		},
	})

	if err != nil {
		t.Fatal(err)
	}

	if id == "" {
		t.Fatal("Empty subscription id")
	}

	if len(id) != 34 {
		t.Fatalf("Subscription id incorrect length: %s", id)
	}

	select {
	case <-time.After(150 * time.Millisecond):
		// good
	case <-sentValuesChannel:
		t.Fatal("Received subscription result before sending subscription id to client")
	}

	internalSubscription := notifier.test_getSubscription(id)
	if internalSubscription == nil {
		for subscriptionId := range notifier.subscriptions {
			t.Errorf("Have %s\n", subscriptionId)
		}
		t.Fatalf("Couldn't get internal subscription object %s\n", id)
	}

	notifier.ResponseSent()

	select {
	case <-sentValuesChannel:
		gotBytes := sentValues[0]
		got := string(gotBytes)
		var jsonRpcNotification eth.JSONRPCNotification
		err = json.Unmarshal(gotBytes, &jsonRpcNotification)
		method := "eth_subscription"
		if jsonRpcNotification.Method != method {
			t.Fatalf(
				"unexpected method\nwant: %s\ngot: %s",
				method,
				string(jsonRpcNotification.Method),
			)
		}
		var receivedEthSubscription eth.EthSubscription
		err = json.Unmarshal(jsonRpcNotification.Params, &receivedEthSubscription)
		if err != nil {
			t.Fatalf("Failed to unmarshal: %s: %s", got, err)
		}
		receivedEthSubscription.SubscriptionID = expectedSubscriptionID
		gotBytes, err = json.Marshal(receivedEthSubscription)
		if err != nil {
			t.Fatalf("Failed to marshal: %s", err)
		}
		got = string(gotBytes)
		if got != want {
			t.Fatalf(
				"logs subscription error\nwant: %s\ngot: %s",
				want,
				got,
			)
		}
	case <-time.After(350 * time.Millisecond):
		t.Fatalf("Timed out waiting for subscription")
	}

	internalSubscription = notifier.test_getSubscription(id)
	if internalSubscription == nil {
		for subscriptionId := range notifier.subscriptions {
			t.Errorf("Have %s\n", subscriptionId)
		}
		t.Fatalf("Couldn't get internal subscription object %s\n", id)
	}

	if !notifier.Unsubscribe(id) {
		t.Fatalf("Failed to unsubscribe to subscription %s", id)
	}

	internalSubscriptionAfterUnsubscribe := notifier.test_getSubscription(id)

	if internalSubscriptionAfterUnsubscribe != nil {
		t.Fatal("Internal subscription object should be nil")
	}

	if notifier.Unsubscribe(id) {
		t.Fatalf("Double unsubscribe to subscription %s worked, it shouldn't", id)
	}
}

func TestAgentAddSubscriptionNewHeads(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	expectedSubscriptionID := "0x08e2af779d38a09e4c11442d9de22413"
	want := `{"subscription":"` + expectedSubscriptionID + `","result":{"difficulty":"0x4","extraData":"0x0000000000000000000000000000000000000000000000000000000000000000","gasLimit":"0x5208","gasUsed":"0x0","logsBloom":"0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000","miner":"0x0000000000000000000000000000000000000000","nonce":"0x0000000000000000","number":"0xf8f","parentHash":"0x6d7d56af09383301e1bb32a97d4a5c0661d62302c06a778487d919b7115543be","receiptRoot":"0x0b5f03dc9d456c63c587cc554b70c1232449be43d1df62bc25a493b04de90334","sha3Uncles":"0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347","stateRoot":"","timestamp":"0x5b95ebd0","transactionsRoot":"0x0b5f03dc9d456c63c587cc554b70c1232449be43d1df62bc25a493b04de90334"}}`

	doer := internal.NewDoerMappedMock()

	for i := int64(1); i < 10; i++ {
		doer.AddResponse(qtum.MethodGetBlockChainInfo, qtum.GetBlockChainInfoResponse{
			Blocks:        i,
			Bestblockhash: "0x1",
		})
	}

	mockedClient, err := internal.CreateMockedClient(doer)
	if err != nil {
		t.Fatal(err)
	}
	agentTestConfig := make(map[string]interface{})
	// adjust newHeads interval to tick quicker for unit tests
	agentTestConfig[agentConfigNewHeadsKey] = 250 * time.Millisecond

	agent := newAgentWithConfiguration(ctx, mockedClient, nil, agentTestConfig)
	getBlockByHashOriginalResponse := internal.CreateTransactionByHashResponse()
	agent.SetTransformer(internal.NewMockTransformer([]internal.ETHProxy{
		internal.NewMockETHProxy(
			"eth_getBlockByHash",
			&getBlockByHashOriginalResponse,
		),
	}))

	go func() {
		for i := 0; true; i++ {
			select {
			case <-ctx.Done():
				return
			case <-time.After(300 * time.Millisecond):
				response := internal.CreateTransactionByHashResponse()
				response.Nonce = fmt.Sprintf("%s%d", response.Nonce, i)
				agent.SetTransformer(internal.NewMockTransformer([]internal.ETHProxy{
					internal.NewMockETHProxy(
						"eth_getBlockByHash",
						&response,
					),
				}))
			}
		}
	}()

	notifierContext, cancelNotifierContext := context.WithCancel(ctx)

	sentValuesChannel := make(chan interface{}, 10)
	sentValuesMutex := sync.Mutex{}
	sentValues := [][]byte{}

	send := func(v []byte) error {
		sentValuesMutex.Lock()
		defer sentValuesMutex.Unlock()
		sentValues = append(sentValues, v)
		defer func() { sentValuesChannel <- nil }()
		return nil
	}

	notifier := NewNotifier(notifierContext, cancelNotifierContext, send, log.NewLogfmtLogger(os.Stdout))

	id, err := agent.NewSubscription(notifier, &eth.EthSubscriptionRequest{
		Method: "newHeads",
		Params: &eth.EthLogSubscriptionParameter{},
	})

	if err != nil {
		t.Fatal(err)
	}

	if id == "" {
		t.Fatal("Empty subscription id")
	}

	if len(id) != 34 {
		t.Fatalf("Subscription id incorrect length: %s", id)
	}

	select {
	case <-time.After(150 * time.Millisecond):
		// good
	case <-sentValuesChannel:
		t.Fatal("Received subscription result before sending subscription id to client")
	}

	internalSubscription := notifier.test_getSubscription(id)
	if internalSubscription == nil {
		for subscriptionId := range notifier.subscriptions {
			t.Errorf("Have %s\n", subscriptionId)
		}
		t.Fatalf("Couldn't get internal subscription object %s\n", id)
	}

	notifier.ResponseSent()

	select {
	case <-sentValuesChannel:
		gotBytes := sentValues[0]
		got := string(gotBytes)
		var receivedEthSubscription eth.EthSubscription
		err = json.Unmarshal(gotBytes, &receivedEthSubscription)
		if err != nil {
			t.Fatalf("Failed to unmarshal: %s: %s", got, err)
		}
		receivedEthSubscription.SubscriptionID = expectedSubscriptionID
		gotBytes, err = json.Marshal(receivedEthSubscription)
		if err != nil {
			t.Fatalf("Failed to marshal: %s", err)
		}
		got = string(gotBytes)
		if got != want {
			t.Fatalf(
				"logs subscription error\nwant: %s\ngot: %s",
				want,
				got,
			)
		}
	case <-time.After(350 * time.Millisecond):
		t.Fatalf("Timed out waiting for subscription")
	}

	internalSubscription = notifier.test_getSubscription(id)
	if internalSubscription == nil {
		for subscriptionId := range notifier.subscriptions {
			t.Errorf("Have %s\n", subscriptionId)
		}
		t.Fatalf("Couldn't get internal subscription object %s\n", id)
	}

	if !notifier.Unsubscribe(id) {
		t.Fatalf("Failed to unsubscribe to subscription %s", id)
	}

	internalSubscriptionAfterUnsubscribe := notifier.test_getSubscription(id)

	if internalSubscriptionAfterUnsubscribe != nil {
		t.Fatal("Internal subscription object should be nil")
	}

	if notifier.Unsubscribe(id) {
		t.Fatalf("Double unsubscribe to subscription %s worked, it shouldn't", id)
	}

	// check that the agent newHeads run loop has exited
	exited := false
	for i := 0; i < 100; i++ {
		exited = !agent.isRunning()
		if exited {
			break
		}
		select {
		case <-ctx.Done():
			t.Fatal("ctx exited")
		case <-time.After(50 * time.Millisecond):
		}
	}

	if !exited {
		t.Fatalf("agent newHeads loop has not exited yet")
	}
}
