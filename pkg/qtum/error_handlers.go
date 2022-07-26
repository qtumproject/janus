package qtum

import (
	"context"
	"errors"
	"sync"
)

var errorHandlers map[error]errorHandler
var ErrErrorHandlerFailed = errors.New("failed to recover from error")

type errorHandler func(ctx context.Context, state *errorState, method *Method) error

func newErrorState() *errorState {
	return &errorState{
		state: make(map[string]interface{}),
	}
}

type errorState struct {
	mutex sync.RWMutex
	state map[string]interface{}
}

func (e *errorState) Get(variable string) interface{} {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	return e.state[variable]
}

func (e *errorState) Put(variable string, value interface{}) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.state[variable] = value
}

func errWalletNotFoundHandler(ctx context.Context, state *errorState, method *Method) error {
	if state.Get("createwallet") != nil {
		return ErrErrorHandlerFailed
	}

	req := CreateWalletRequest([]string{"wallet"})
	_, err := method.CreateWallet(ctx, &req)

	if err == nil {
		state.Put("createwallet", true)
	}

	if err == ErrWalletError {
		req := LoadWalletRequest([]string{"wallet"})
		_, err = method.LoadWallet(ctx, &req)

		if err == nil {
			state.Put("createwallet", true)
		}
	}

	return err
}

func init() {
	errorHandlers = make(map[error]errorHandler)
	errorHandlers[ErrWalletNotFound] = errWalletNotFoundHandler
}
