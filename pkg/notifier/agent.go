package notifier

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"github.com/qtumproject/janus/pkg/eth"
	"github.com/qtumproject/janus/pkg/qtum"
	"github.com/qtumproject/janus/pkg/utils"
)

var agentConfigNewHeadsKey = "newHeadsInterval"
var agentConfigNewHeadsInterval = 10 * time.Second

// Allows dependency injection of eth rpc calls as the transformer package imports this package
type Transformer interface {
	Transform(req *eth.JSONRPCRequest, c echo.Context) (interface{}, error)
}

func NewAgent(ctx context.Context, qtum *qtum.Qtum, transformer Transformer) *Agent {
	return newAgentWithConfiguration(ctx, qtum, transformer, make(map[string]interface{}))
}

func newAgentWithConfiguration(ctx context.Context, qtum *qtum.Qtum, transformer Transformer, configuration map[string]interface{}) *Agent {
	if ctx == nil {
		panic("ctx cannot be nil")
	}
	if qtum == nil {
		panic("qtum cannot be nil")
	}
	agent := &Agent{
		qtum:          qtum,
		transformer:   transformer,
		ctx:           ctx,
		mutex:         sync.RWMutex{},
		running:       false,
		config:        configuration,
		stop:          make(chan interface{}, 1000),
		newHeads:      newSubscriptionRegistry(),
		logs:          newSubscriptionRegistry(),
		newPendingTxs: newSubscriptionRegistry(),
		syncing:       newSubscriptionRegistry(),
	}

	go agent.run()
	return agent
}

type subscriptionRegistry struct {
	mutex             sync.RWMutex
	subscriptionCount int
	subscriptions     map[string]*subscriptionInformation
}

func newSubscriptionRegistry() *subscriptionRegistry {
	return &subscriptionRegistry{
		mutex:             sync.RWMutex{},
		subscriptionCount: 0,
		subscriptions:     make(map[string]*subscriptionInformation),
	}
}

func (s *subscriptionRegistry) forEach(do func(*subscriptionInformation)) {
	s.mutex.RLock()
	subscriptions := make([]*subscriptionInformation, 0, len(s.subscriptions))
	for _, subscription := range s.subscriptions {
		subscriptions = append(subscriptions, subscription)
	}
	s.mutex.RUnlock()
	for index := range subscriptions {
		subscription := subscriptions[index]
		do(subscription)
	}
}

func (s *subscriptionRegistry) Count() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.subscriptionCount
}

func (s *subscriptionRegistry) SendAll(message interface{}) {
	send := func(s *subscriptionInformation) {
		// send writes to a queue that can block when full if a client has a lot of responses queued up
		// that could potentially affect other clients so we run this in a goroutine
		subscription := &eth.EthSubscription{
			SubscriptionID: s.Subscription.id,
			Result:         message,
		}
		go s.Send(subscription)
	}
	s.forEach(send)
}

type Agent struct {
	qtum          *qtum.Qtum
	transformer   Transformer
	ctx           context.Context
	mutex         sync.RWMutex
	running       bool
	stop          chan interface{}
	config        map[string]interface{}
	newHeads      *subscriptionRegistry
	logs          *subscriptionRegistry
	newPendingTxs *subscriptionRegistry
	syncing       *subscriptionRegistry
}

func (a *Agent) SetTransformer(transformer Transformer) {
	a.mutex.Lock()
	a.transformer = transformer
	a.mutex.Unlock()
}

func (a *Agent) Stop() {
	a.mutex.Lock()
	a.lockAllRegistries(false)
	defer a.unlockAllRegistries(false)
	defer a.mutex.Unlock()

	closeSubscriptionRegistry(a.newHeads)
	closeSubscriptionRegistry(a.logs)
	closeSubscriptionRegistry(a.newPendingTxs)
	closeSubscriptionRegistry(a.syncing)
}

func (a *Agent) setConfigValue(key string, value interface{}) {
	a.config[key] = value
}

func (a *Agent) getConfigValue(key string, defaultValue interface{}) interface{} {
	if value, ok := a.config[key]; ok {
		return value
	}
	return defaultValue
}

func closeSubscriptionRegistry(registry *subscriptionRegistry) {
	for _, sub := range registry.subscriptions {
		sub.cancelFunc()
	}
}

func (a *Agent) lockAllRegistries(readOnly bool) {
	if readOnly {
		a.newHeads.mutex.RLock()
		a.logs.mutex.RLock()
		a.newPendingTxs.mutex.RLock()
		a.syncing.mutex.RLock()
	} else {
		a.newHeads.mutex.Lock()
		a.logs.mutex.Lock()
		a.newPendingTxs.mutex.Lock()
		a.syncing.mutex.Lock()
	}
}

func (a *Agent) unlockAllRegistries(readOnly bool) {
	if readOnly {
		a.newHeads.mutex.RUnlock()
		a.logs.mutex.RUnlock()
		a.newPendingTxs.mutex.RUnlock()
		a.syncing.mutex.RUnlock()
	} else {
		a.newHeads.mutex.Unlock()
		a.logs.mutex.Unlock()
		a.newPendingTxs.mutex.Unlock()
		a.syncing.mutex.Unlock()
	}
}

func (a *Agent) subscriptionCount(acquireLocks bool) int {
	if acquireLocks {
		a.lockAllRegistries(true)
		defer a.unlockAllRegistries(true)
	}

	return a.newHeads.subscriptionCount +
		a.logs.subscriptionCount +
		a.newPendingTxs.subscriptionCount +
		a.syncing.subscriptionCount
}

func (a *Agent) unsubscribe(id string) {
	removeSubscription(id, a.newHeads)
	removeSubscription(id, a.logs)
	removeSubscription(id, a.newPendingTxs)
	removeSubscription(id, a.syncing)
}

func addSubscription(subscription *subscriptionInformation, registry *subscriptionRegistry) {
	registry.mutex.Lock()
	defer registry.mutex.Unlock()

	_, collision := registry.subscriptions[subscription.id]
	registry.subscriptions[subscription.id] = subscription
	if !collision {
		registry.subscriptionCount = registry.subscriptionCount + 1
	}

	go subscription.run()
}

func removeSubscription(id string, registry *subscriptionRegistry) {
	registry.mutex.RLock()
	sub, exists := registry.subscriptions[id]
	registry.mutex.RUnlock()
	if exists {
		registry.mutex.Lock()
		sub, exists = registry.subscriptions[id]
		if exists {
			delete(registry.subscriptions, id)
			registry.subscriptionCount = registry.subscriptionCount - 1
		}
		registry.mutex.Unlock()
	}

	if sub != nil {
		sub.cancelFunc()
	}
}

func (a *Agent) NewSubscription(notifier *Notifier, params *eth.EthSubscriptionRequest) (string, error) {
	subscription, err := notifier.Subscribe(a.unsubscribe)
	if err != nil {
		return "", err
	}

	wrappedContext, cancel := context.WithCancel(notifier.Context())

	wrappedSubscription := &subscriptionInformation{
		subscription,
		params,
		sync.RWMutex{},
		wrappedContext,
		cancel,
		false,
		a.qtum,
	}

	switch strings.ToLower(params.Method) {
	case "logs":
		addSubscription(wrappedSubscription, a.logs)
	case "newheads":
		addSubscription(wrappedSubscription, a.newHeads)
	case "newpendingtransactions":
		addSubscription(wrappedSubscription, a.newPendingTxs)
	case "syncing":
		addSubscription(wrappedSubscription, a.syncing)
	default:
		return "", errors.New(fmt.Sprintf("Unknown subscription type %s", params.Method))
	}

	a.mutex.RLock()
	if !a.running {
		// start processing subscriptions if nothing is running
		// only one routine will run at once so if multiple startup they will exit so only one runs
		go a.run()
	}
	a.mutex.RUnlock()

	return subscription.id, nil
}

func (a *Agent) isRunning() bool {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return a.running
}

func (a *Agent) run() {
	a.mutex.Lock()
	if a.running {
		a.mutex.Unlock()
		return
	}
	a.running = true
	a.mutex.Unlock()

	defer func() {
		a.mutex.Lock()
		defer a.mutex.Unlock()

		a.qtum.GetDebugLogger().Log("msg", "Agent exited subscription processing thread")

		a.running = false
	}()

	lastBlock := int64(0)

	draining := true
	for draining {
		select {
		case <-a.stop:
			// drain
		case <-a.ctx.Done():
			return
		default:
			draining = false
		}
	}

	newHeadsIntervalValue := a.getConfigValue(agentConfigNewHeadsKey, agentConfigNewHeadsInterval)
	newHeadsInterval, ok := newHeadsIntervalValue.(time.Duration)
	if !ok {
		panic(fmt.Sprintf("Unexpected %s type", agentConfigNewHeadsKey))
	}

	// TODO: newPendingTransactions
	a.qtum.GetDebugLogger().Log("msg", "Agent started subscription processing thread")

	for {
		// infinite loop while we have subscriptions
		newHeadsSubscriptions := a.newHeads.Count()
		if newHeadsSubscriptions == 0 {
			return
		}

		a.mutex.RLock()
		transformer := a.transformer
		a.mutex.RUnlock()
		if transformer == nil {
			a.qtum.GetErrorLogger().Log("msg", "Agent does not have access to eth transformer, cannot process 'newHeads' subscriptions")
		} else {
			blockchainInfo, err := a.qtum.GetBlockChainInfo()
			if err != nil {
				a.qtum.GetErrorLogger().Log("msg", "Failure getting blockchaininfo", "err", err)
			} else {
				latestBlock := blockchainInfo.Blocks
				if lastBlock == 0 {
					// prevent sending the current head to the first client connected
					lastBlock = latestBlock
					a.qtum.GetDebugLogger().Log("msg", "Got getblockchaininfo response for same block", "block", lastBlock)
				} else if latestBlock > lastBlock {
					a.qtum.GetDebugLogger().Log("msg", "New head detected", "block", latestBlock)
					// get the latest block as an eth_getBlockByHash request
					params, err := json.Marshal([]interface{}{
						utils.AddHexPrefix(blockchainInfo.Bestblockhash),
						false,
					})
					if err != nil {
						panic(fmt.Sprintf("Failed to serialize eth_getBlockByHash request parameters: %s", err))
					}
					result, err := transformer.Transform(&eth.JSONRPCRequest{
						JSONRPC: "2.0",
						Method:  "eth_getBlockByHash",
						Params:  params,
					}, nil)
					if err != nil {
						a.qtum.GetErrorLogger().Log("msg", "Failed to eth_getBlockByHash", "hash", blockchainInfo.Bestblockhash, "err", err)
					} else {
						getBlockByHashResponse, ok := result.(*eth.GetBlockByHashResponse)
						if !ok {
							a.qtum.GetErrorLogger().Log("msg", "Failed to eth_getBlockByHash, unexpected response type", "hash", blockchainInfo.Bestblockhash)
						} else {
							lastBlock = latestBlock
							// notify newHead
							newHeadRespose := eth.NewEthSubscriptionNewHeadResponse(getBlockByHashResponse)
							a.newHeads.SendAll(newHeadRespose)
						}
					}
				} else {
					a.qtum.GetDebugLogger().Log("msg", "Detected same head", "block", latestBlock)
				}
			}
		}

		select {
		case <-time.After(newHeadsInterval):
			// continue
		case <-a.ctx.Done():
			return
		case <-a.stop:
			return
		}
	}
}
