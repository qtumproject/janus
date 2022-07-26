package qtum

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/qtumproject/janus/pkg/utils"
)

type Qtum struct {
	*Client
	*Method
	chainMutex       sync.RWMutex
	queryingChain    bool
	queryingComplete chan bool
	chain            string

	errorState *errorState
}

const (
	ChainMain    = "main"
	ChainTest    = "test"
	ChainRegTest = "regtest"
	ChainAuto    = "auto"
	ChainUnknown = ""
)

var AllChains = []string{ChainMain, ChainRegTest, ChainTest, ChainAuto, ChainUnknown}

func New(c *Client, chain string) (*Qtum, error) {
	if !utils.InStrSlice(AllChains, chain) {
		return nil, errors.Errorf("Invalid qtum chain: '%s'", chain)
	}

	qtum := &Qtum{
		Client:     c,
		Method:     &Method{Client: c},
		chain:      chain,
		errorState: newErrorState(),
	}

	c.SetErrorHandler(func(ctx context.Context, err error) error {
		if errorHandler, ok := errorHandlers[err]; ok {
			return errorHandler(ctx, qtum.errorState, qtum.Method)
		}
		return nil
	})

	go qtum.detectChain()

	return qtum, nil
}

func (c *Qtum) detectChain() {
	c.chainMutex.Lock()
	if c.queryingChain || // already querying
		(c.chain != ChainAuto && c.chain != "") { // specified in command line arguments
		c.chainMutex.Unlock()
		return
	}
	c.queryingChain = true
	c.queryingComplete = make(chan bool, 1000)
	c.chainMutex.Unlock()

	// detect chain we are pointing at
	for i := 0; ; i++ {
		blockchainInfo, err := c.GetBlockChainInfo(c.ctx)
		if err == nil {
			chain := strings.ToLower(blockchainInfo.Chain)
			if utils.InStrSlice(AllChains, chain) {
				c.chainMutex.Lock()
				c.chain = chain
				c.queryingChain = false
				if c.queryingComplete != nil {
					queryingComplete := c.queryingComplete
					c.queryingComplete = nil
					close(queryingComplete)
				}
				c.chainMutex.Unlock()
				c.GetDebugLogger().Log("msg", "Detected chain type", "chain", chain)
				return
			} else {
				c.GetErrorLogger().Log("msg", "Unknown chain type in getblockchaininfo", "chain", chain)
			}
		}

		interval := 250 * time.Millisecond
		backoff := time.Duration(math.Min(float64(i), 10)) * interval
		c.GetDebugLogger().Log("msg", "Failed to detect chain type, backing off", "backoff", backoff)
		// TODO check if this works as expected
		// time.Sleep(backoff)
		var done <-chan struct{}
		if c.ctx != nil {
			done = c.ctx.Done()
		} else {
			done = context.Background().Done()
		}
		select {
		case <-time.After(backoff):
		case <-done:
			return
		}
	}
}

func (c *Qtum) Chain() string {
	c.chainMutex.RLock()
	queryingChain := c.queryingChain
	queryingComplete := c.queryingComplete
	c.chainMutex.RUnlock()

	if queryingChain && queryingComplete != nil {
		select {
		case <-c.ctx.Done():
		case <-queryingComplete:
		}
	}

	c.chainMutex.RLock()
	defer c.chainMutex.RUnlock()

	return c.chain
}

func (c *Qtum) ChainId() int {
	var chainId int
	switch strings.ToLower(c.Chain()) {
	case "main":
		chainId = 81
	case "test":
		chainId = 8889
	case "regtest":
		chainId = 8890
	default:
		chainId = 8890
		c.GetDebugLogger().Log("msg", fmt.Sprintf("Unknown chain %d", chainId))
	}

	return chainId
}

func (c *Qtum) GetMatureBlockHeight() int {
	blockHeightOverride := c.GetFlagInt(FLAG_MATURE_BLOCK_HEIGHT_OVERRIDE)
	if blockHeightOverride != nil {
		return *blockHeightOverride
	}

	return 2000
}

func (c *Qtum) CanGenerate() bool {
	return c.Chain() == ChainRegTest
}

func (c *Qtum) GenerateIfPossible() {
	if !c.CanGenerate() {
		return
	}

	if _, generateErr := c.Generate(c.ctx, 1, nil); generateErr != nil {
		c.GetErrorLogger().Log("Error generating new block", generateErr)
	}
}

// Presents hexed address prefix of a specific chain without
// `0x` prefix, this is a ready to use hexed string
type HexAddressPrefix string

const (
	PrefixMainChainAddress    HexAddressPrefix = "3a"
	PrefixTestChainAddress    HexAddressPrefix = "78"
	PrefixRegTestChainAddress HexAddressPrefix = PrefixTestChainAddress
)

// Returns decoded hexed string prefix, as ready to use slice of bytes
func (prefix HexAddressPrefix) AsBytes() ([]byte, error) {
	bytes, err := hex.DecodeString(string(prefix))
	if err != nil {
		return nil, errors.Wrap(err, "couldn't decode hexed string")
	}
	return bytes, nil
}

// Returns first 4 bytes of a double sha256 hash of the provided `prefixedAddrBytes`,
// which must be already prefixed with a specific chain prefix
func CalcAddressChecksum(prefixedAddr []byte) []byte {
	hash := sha256.Sum256(prefixedAddr)
	hash = sha256.Sum256(hash[:])
	return hash[:4]
}
