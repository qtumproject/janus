package notifier

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"sync"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/pkg/errors"
)

var UnsubSignal = new(struct{})

type UnsubscribeCallback func(string)

type Subscription struct {
	*Notifier
	id          string
	once        sync.Once
	unsubscribe UnsubscribeCallback
}

func NewSubscription(notifier *Notifier, callback UnsubscribeCallback) (*Subscription, error) {
	id, err := getRandomSubscriptionId()
	if err != nil {
		return nil, err
	}
	return &Subscription{
		notifier,
		id,
		sync.Once{},
		func(id string) {
			callback(id)
			// call in goroutine as this can be called from Unsubscribe and end in a deadlock
			go notifier.Unsubscribe(id)
		},
	}, nil
}

func getRandomSubscriptionId() (string, error) {
	var subid [16]byte
	n, _ := rand.Read(subid[:])
	if n != 16 {
		return "", errors.New("Unable to generate subscription id")
	}
	return "0x" + hex.EncodeToString(subid[:]), nil
}

func (s *Subscription) Unsubscribe() {
	s.once.Do(func() {
		s.unsubscribe(s.id)
	})
}

type Notifier struct {
	runMutex              sync.Mutex
	mutex                 sync.RWMutex
	ctx                   context.Context
	close                 func()
	send                  func([]byte) error
	logger                log.Logger
	queue                 chan interface{}
	subscriptionIdPending *chan interface{}
	subscriptionsFlushed  *chan interface{}
	subscriptions         map[string]*Subscription
}

func NewNotifier(ctx context.Context, close func(), send func([]byte) error, logger log.Logger) *Notifier {
	pending := make(chan interface{}, 10)
	flushed := make(chan interface{}, 10)
	notifier := &Notifier{
		runMutex:              sync.Mutex{},
		mutex:                 sync.RWMutex{},
		ctx:                   ctx,
		close:                 close,
		send:                  send,
		logger:                log.WithPrefix(logger, "component", "notifier"),
		queue:                 make(chan interface{}, 50),
		subscriptionIdPending: &pending,
		subscriptionsFlushed:  &flushed,
		subscriptions:         make(map[string]*Subscription),
	}
	go notifier.run()
	return notifier
}

func (n *Notifier) Context() context.Context {
	return n.ctx
}

func (n *Notifier) Subscribe(unsubscribeCallback UnsubscribeCallback) (*Subscription, error) {
	sub, err := NewSubscription(n, unsubscribeCallback)
	if err != nil {
		return nil, err
	}

	n.mutex.Lock()
	n.subscriptions[sub.id] = sub
	n.mutex.Unlock()

	return sub, nil
}

// internal function to expose subscription directly to test
func (n *Notifier) test_getSubscription(id string) *Subscription {
	return n.subscriptions[id]
}

func (n *Notifier) Unsubscribe(id string) bool {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	subscription, ok := n.subscriptions[id]
	if ok {
		subscription.Unsubscribe()
		delete(n.subscriptions, id)
		n.logger.Log("subscriptionId", id, "msg", "Subscription id unsubscribed")
	} else {
		n.logger.Log("subscriptionId", id, "msg", "Unknown subscription id to unsubscribe from")
	}

	return ok
}

func (n *Notifier) ResponseSent() {
	n.mutex.RLock()
	subscriptionIdPending := n.subscriptionIdPending
	n.mutex.RUnlock()
	if subscriptionIdPending == nil {
		return
	}
	n.mutex.Lock()
	defer n.mutex.Unlock()
	if n.subscriptionIdPending != nil {
		close(*n.subscriptionIdPending)
		n.subscriptionIdPending = nil
	}
}

func (n *Notifier) ResponseRequired() {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	pending := make(chan interface{}, 10)
	n.subscriptionIdPending = &pending
}

func (n *Notifier) Send(event interface{}) {
	n.queue <- event
}

func (n *Notifier) closeSubscriptionsFlushed() {
	if n.subscriptionsFlushed != nil {
		close(*n.subscriptionsFlushed)
		n.subscriptionsFlushed = nil
	}
}

func (n *Notifier) run() {
	n.runMutex.Lock()
	defer n.runMutex.Unlock()

	log.With(level.Debug(n.logger)).Log("msg", "Entering notifier loop")

	defer func() {
		n.mutex.Lock()
		defer n.mutex.Unlock()
		defer func() {
			log.With(level.Debug(n.logger)).Log("msg", "Notifier loop exited")
		}()

		n.close()
		close(n.queue)
		n.closeSubscriptionsFlushed()
		for _, sub := range n.subscriptions {
			sub.Unsubscribe()
		}
	}()

	for {
		select {
		case <-n.ctx.Done():
			return
		case event := <-n.queue:
			b, _ := json.Marshal(event)
			log.With(level.Debug(n.logger)).Log("notifier event", string(b))
			n.mutex.RLock()
			subscriptionIdPending := n.subscriptionIdPending
			n.mutex.RUnlock()
			if subscriptionIdPending != nil {
				select {
				case <-*subscriptionIdPending:
					log.With(level.Debug(n.logger)).Log("msg", "subscription pending complete")
				case <-n.ctx.Done():
					return
				}
			}
			if event == UnsubSignal {
				n.mutex.Lock()
				n.closeSubscriptionsFlushed()
				n.mutex.Unlock()
			} else {
				bytes, err := json.Marshal(event)
				if err != nil {
					panic(err)
				}
				err = n.send(bytes)
				if err != nil {
					// write failure, close connection and unsubscribe
					log.With(level.Debug(n.logger)).Log("msg", "Error writing response to websocket, closing it", "err", err)
					n.close()
					return
				}
			}
		}
	}
}
