package blockhash

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/qtumproject/ethereum-block-processor/cache"
	"github.com/qtumproject/ethereum-block-processor/db"
	"github.com/qtumproject/ethereum-block-processor/dispatcher"
	"github.com/qtumproject/ethereum-block-processor/eth"
	"github.com/qtumproject/ethereum-block-processor/jsonrpc"
	blockHashLog "github.com/qtumproject/ethereum-block-processor/log"
)

var ErrDatabaseNotConfigured = errors.New("database not connected")

type BlockHash struct {
	ctx   context.Context
	mutex sync.RWMutex

	qtumDB    *db.QtumDB
	getLogger func() log.Logger

	chainId      int
	chainIdMutex sync.RWMutex
}

type DatabaseConfig struct {
	Host             string
	Port             int
	User             string
	Password         string
	DatabaseName     string
	SSL              bool
	ConnectionString string
}

func (config *DatabaseConfig) String() string {
	if config.ConnectionString != "" {
		return config.ConnectionString
	}

	ssl := "disable"
	if config.SSL {
		ssl = "enable"
	}
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", config.Host, config.Port, config.User, config.Password, config.DatabaseName, ssl)
}

func NewBlockHash(ctx context.Context, getLogger func() log.Logger) (*BlockHash, error) {
	return &BlockHash{
		ctx:       ctx,
		getLogger: getLogger,
	}, nil
}

func (bh *BlockHash) GetQtumBlockHash(ethereumBlockHash string) (*string, error) {
	return bh.GetQtumBlockHashContext(nil, ethereumBlockHash)
}

func (bh *BlockHash) GetQtumBlockHashContext(ctx context.Context, ethereumBlockHash string) (*string, error) {
	var qtumBlockHash string
	bh.mutex.RLock()
	qtumDB := bh.qtumDB
	bh.mutex.RUnlock()
	if qtumDB == nil {
		return &qtumBlockHash, ErrDatabaseNotConfigured
	}

	bh.chainIdMutex.RLock()
	chainId := bh.chainId
	bh.chainIdMutex.RUnlock()

	if chainId == 0 {
		panic("Got invalid chainId")
	}

	if !strings.HasPrefix(ethereumBlockHash, "0x") {
		ethereumBlockHash = fmt.Sprintf("0x%s", ethereumBlockHash)
	}

	if ctx == nil {
		return qtumDB.GetQtumHash(chainId, ethereumBlockHash)
	} else {
		return qtumDB.GetQtumHashContext(ctx, chainId, ethereumBlockHash)
	}
}

func (bh *BlockHash) Start(databaseConfig *DatabaseConfig, chainIdChan <-chan int) error {
	numWorkers := runtime.NumCPU() * 2
	bh.chainIdMutex.Lock()

	// logger.Info("Number of workers: ", numWorkers)
	// channel to receive errors from goroutines
	errChan := make(chan error, numWorkers+1)
	// channel to pass blocks to workers
	blockChan := make(chan int64, numWorkers)
	completedBlockChan := make(chan int64, numWorkers)
	// channel to pass results from workers to DB
	resultChan := make(chan jsonrpc.HashPair, numWorkers)

	connectionString := databaseConfig.String()

	qdb, err := db.NewQtumDB(bh.ctx, connectionString, resultChan, errChan)
	if err != nil {
		bh.chainIdMutex.Unlock()
		// Quick fail if database connection fails
		return err
	}

	bh.mutex.Lock()
	bh.qtumDB = qdb
	bh.mutex.Unlock()

	go func() {
		var chainId int
		for chainId == 0 {
			t := time.After(5 * time.Second)
			select {
			case chainId = <-chainIdChan:
				bh.chainIdMutex.Unlock()
				bh.getLogger().Log("msg", "Got chain id: will write to database", "chaindId", chainId)
				break
			case <-t:
				bh.getLogger().Log("msg", "Waiting for chain id from eth_chainId before writing to database")
				continue
			case <-bh.ctx.Done():
				bh.chainIdMutex.Unlock()
				return
			}
		}
		bh.mutex.Lock()
		bh.chainId = chainId
		bh.mutex.Unlock()
		dbCloseChan := make(chan error)
		qdb.Start(bh.ctx, chainId, dbCloseChan)
		// channel to signal  work completion to main from dispatcher
		done := make(chan struct{})
		// channel to receive os signals
		// sigs := make(chan os.Signal, 1
		// signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		// dispatch blocks to block channel
		// ctx, cancelFunc := context.WithCancel(context.Background())

		// janus, err := url.Parse("https://janus.qiswap.com")
		janus, _ := url.Parse("http://localhost:23889")
		providers := []*url.URL{janus}

		dispatchLogger, _ := blockHashLog.GetLogger()

		blockCacheLogger := dispatchLogger.WithField("module", "dispatcher")

		blockCache := cache.NewBlockCache(
			bh.ctx,
			func(ctx context.Context) ([]int64, error) {
				latestBlock, err := eth.GetLatestBlock(ctx, blockCacheLogger, (providers)[0].String())
				if err != nil {
					return nil, err
				}

				return qdb.GetMissingBlocks(ctx, chainId, latestBlock)
			},
		)

		go func() {
			for {
				select {
				case <-completedBlockChan:
				case <-bh.ctx.Done():
					return
				}
			}
		}()

		// dispatcher.NewDispatcher(blockChan, resultChan, completedBlockChan, providers, 10, 1, done, errChan, blockCache)

		d := dispatcher.NewDispatcher(blockChan, resultChan, completedBlockChan, providers, 10, 1, done, errChan, blockCache)
		d.Start(bh.ctx, numWorkers, providers, true)
		// start workers
		// wg.Add(numWorkers)
		// workerState := workers.StartWorkers(bh.ctx, numWorkers, blockChan, resultChan, providers, &wg, errChan)
		// start = time.Now()

		go func() {
			var status int
			select {
			case <-done:
				qdb.Shutdown()
				bh.mutex.Lock()
				bh.qtumDB = nil
				bh.mutex.Unlock()
				status = 0
			// case <-sigs:
			// panic("Received ^C ... exiting")
			// logger.Warn("Canceling block dispatcher and stopping workers")
			// cancelFunc()
			// status = 1
			case err := <-errChan:
				panic(err)
				// logger.Warn("Canceling block dispatcher and stopping workers")
				// cancelFunc()
				status = 1
			}
			fmt.Println(status)
			// panic(status)
		}()
	}()

	return nil
}
