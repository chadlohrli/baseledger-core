package main

import (
	"context"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/providenetwork/baseledger/common"
	"github.com/providenetwork/baseledger/consensus"
	"github.com/providenetwork/baseledger/protocol"
	"github.com/providenetwork/tendermint/libs/service"
)

const runloopSleepInterval = 250 * time.Millisecond
const runloopTickInterval = 5000 * time.Millisecond

var (
	cancelF     context.CancelFunc
	closing     uint32
	shutdownCtx context.Context
	sigs        chan os.Signal

	baseline   *protocol.Baseline
	tendermint service.Service
)

func main() {
	common.Log.Debugf("starting baseledger node")
	installSignalHandlers()

	startConsensus()

	timer := time.NewTicker(runloopTickInterval)
	defer timer.Stop()

	for !shuttingDown() {
		select {
		case <-timer.C:
			// no-op for now...
		case sig := <-sigs:
			common.Log.Debugf("received signal: %s", sig)
			stopConsensus()
			shutdown()
		case <-shutdownCtx.Done():
			close(sigs)
		default:
			time.Sleep(runloopSleepInterval)
		}
	}

	common.Log.Debug("exiting baseledger node")
	cancelF()
}

func startConsensus() {
	cfg, err := consensus.ConfigFactory()
	if err != nil {
		common.Log.Panicf("failed to initialize baseledger core consensus; failed to load configuration; %s", err.Error())
	}

	logger, err := consensus.LogFactory(cfg)
	if err != nil {
		common.Log.Panicf("failed to initialize baseledger core consensus; failed to initialize logger; %s", err.Error())
	}

	genesis, err := consensus.GenesisFactory(cfg)
	if err != nil {
		common.Log.Panicf("failed to initialize baseledger core consensus; failed to initialize genesis; %s", err.Error())
	}

	baseline = protocol.BaselineProtocolFactory(&cfg.Config, genesis)
	tendermint, err = consensus.InitTendermint(cfg, logger, genesis, baseline)
	if err != nil {
		common.Log.Panicf("failed to initialize baseledger core consensus; %s", err.Error())
	}

	err = tendermint.Start()
	if err != nil {
		common.Log.Panicf("failed to start baseledger core consensus; %s", err.Error())
	}

	common.Log.Debugf("initialized baseledger core consensus; %v", tendermint.String())
}

func stopConsensus() {
	defer func() {
		if r := recover(); r != nil {
			common.Log.Warningf("recovered while stopping baseledger core consensus; %s", r)
		}
	}()

	tendermint.Stop()
}

func installSignalHandlers() {
	common.Log.Debug("installing signal handlers for baseledger node")
	sigs = make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	shutdownCtx, cancelF = context.WithCancel(context.Background())
}

func shutdown() {
	if atomic.AddUint32(&closing, 1) == 1 {
		common.Log.Debug("shutting down baseledger node")
		cancelF()
	}
}

func shuttingDown() bool {
	return (atomic.LoadUint32(&closing) > 0)
}
