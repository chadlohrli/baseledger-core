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
)

const runloopSleepInterval = 250 * time.Millisecond
const runloopTickInterval = 5000 * time.Millisecond

var (
	cancelF     context.CancelFunc
	closing     uint32
	shutdownCtx context.Context
	sigs        chan os.Signal

	baseledger *consensus.Tendermint
)

func init() {
	var err error
	baseledger, err = consensus.TendermintFactory()
	if err != nil {
		panic(err)
	}
}

func main() {
	common.Log.Debugf("starting baseledger node")
	installSignalHandlers()

	err := baseledger.Start()
	if err != nil {
		panic(err)
	}

	timer := time.NewTicker(runloopTickInterval)
	defer timer.Stop()

	for !shuttingDown() {
		select {
		case <-timer.C:
			// no-op for now...
		case sig := <-sigs:
			common.Log.Debugf("received signal: %s", sig)
			baseledger.Stop()
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
