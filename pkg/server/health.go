package server

import (
	"time"

	"github.com/pkg/errors"
	"github.com/qtumproject/janus/pkg/qtum"
)

var ErrNoQtumConnections = errors.New("qtumd has no connections")
var ErrCannotGetConnectedChain = errors.New("Cannot detect chain qtumd is connected to")
var ErrBlockSyncingSeemsStalled = errors.New("Block syncing seems stalled")
var ErrLostLotsOfBlocks = errors.New("Lost a lot of blocks, expected block height to be higher")
var ErrLostFewBlocks = errors.New("Lost a few blocks, expected block height to be higher")

func (s *Server) testConnectionToQtumd() error {
	networkInfo, err := s.qtumRPCClient.GetNetworkInfo(s.qtumRPCClient.GetContext())
	if err == nil {
		// chain can theoretically block forever if qtumd isn't up
		// but then GetNetworkInfo would be erroring
		chainChan := make(chan string)
		getChainTimeout := time.NewTimer(10 * time.Second)
		go func(ch chan string) {
			chain := s.qtumRPCClient.Chain()
			chainChan <- chain
		}(chainChan)

		select {
		case chain := <-chainChan:
			if chain == qtum.ChainRegTest {
				// ignore how many connections there are
				return nil
			}
			if networkInfo.Connections == 0 {
				return ErrNoQtumConnections
			}
			break
		case <-getChainTimeout.C:
			return ErrCannotGetConnectedChain
		}
	}
	return err
}

func (s *Server) testLogEvents() error {
	_, err := s.qtumRPCClient.GetTransactionReceipt(s.qtumRPCClient.GetContext(), "0000000000000000000000000000000000000000000000000000000000000000")
	if err == qtum.ErrInternalError {
		return errors.Wrap(err, "-logevents might not be enabled")
	}
	return nil
}

func (s *Server) testBlocksSyncing() error {
	s.blocksMutex.RLock()
	nextBlockCheck := s.nextBlockCheck
	lastBlockStatus := s.lastBlockStatus
	s.blocksMutex.RUnlock()
	now := time.Now()
	if nextBlockCheck == nil {
		nextBlockCheckTime := time.Now().Add(-30 * time.Minute)
		nextBlockCheck = &nextBlockCheckTime
	}
	if nextBlockCheck.After(now) {
		return lastBlockStatus
	}
	s.blocksMutex.Lock()
	if s.nextBlockCheck != nil && nextBlockCheck != s.nextBlockCheck {
		// multiple threads were waiting on write lock
		s.blocksMutex.Unlock()
		return s.testBlocksSyncing()
	}
	defer s.blocksMutex.Unlock()

	blockChainInfo, err := s.qtumRPCClient.GetBlockChainInfo(s.qtumRPCClient.GetContext())
	if err != nil {
		return err
	}

	nextBlockCheckTime := time.Now().Add(5 * time.Minute)
	s.nextBlockCheck = &nextBlockCheckTime

	if blockChainInfo.Blocks == s.lastBlock {
		// stalled
		nextBlockCheckTime = time.Now().Add(15 * time.Second)
		s.nextBlockCheck = &nextBlockCheckTime
		s.lastBlockStatus = ErrBlockSyncingSeemsStalled
	} else if blockChainInfo.Blocks < s.lastBlock {
		// lost some blocks...?
		if s.lastBlock-blockChainInfo.Blocks > 10 {
			// lost a lot of blocks
			// probably a real problem
			s.lastBlock = 0
			nextBlockCheckTime = time.Now().Add(60 * time.Second)
			s.nextBlockCheck = &nextBlockCheckTime
			s.lastBlockStatus = ErrLostLotsOfBlocks
		} else {
			// lost a few blocks
			// could be qtumd nodes out of sync behind a load balancer
			nextBlockCheckTime = time.Now().Add(10 * time.Second)
			s.nextBlockCheck = &nextBlockCheckTime
			s.lastBlockStatus = ErrLostFewBlocks
		}
	} else {
		// got a higher block height than last time
		s.lastBlock = blockChainInfo.Blocks
		nextBlockCheckTime = time.Now().Add(90 * time.Second)
		s.nextBlockCheck = &nextBlockCheckTime
		s.lastBlockStatus = nil
	}

	return s.lastBlockStatus
}
