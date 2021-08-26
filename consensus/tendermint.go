package consensus

import (
	"fmt"

	"github.com/providenetwork/baseledger/common"
	"github.com/providenetwork/baseledger/protocol"
	"github.com/providenetwork/tendermint/libs/log"
	"github.com/providenetwork/tendermint/libs/service"
	"github.com/providenetwork/tendermint/node"
	"github.com/providenetwork/tendermint/proxy"
	"github.com/providenetwork/tendermint/types"
)

type Tendermint struct {
	baseline *protocol.Baseline
	genesis  *types.GenesisDoc
	logger   *log.Logger
	service  service.Service
}

// TendermintFactory initializes and returns the baseledger tendermint consensus service
func TendermintFactory() (*Tendermint, error) {
	cfg, err := common.ConfigFactory()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize baseledger core consensus; failed to load configuration; %s", err.Error())
	}

	logger, err := LogFactory(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize baseledger core consensus; failed to initialize logger; %s", err.Error())
	}

	genesis, err := GenesisFactory(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize baseledger core consensus; failed to initialize genesis; %s", err.Error())
	}

	baseline, err := protocol.BaselineProtocolFactory(cfg, genesis)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize baseledger core consensus; failed to initialize baseline protocol service implementation; %s", err.Error())
	}

	service, err := initTendermint(cfg, logger, genesis, baseline)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize baseledger core consensus; %s", err.Error())
	}

	return &Tendermint{
		baseline: baseline,
		genesis:  genesis,
		logger:   logger,
		service:  service,
	}, nil
}

// Start attempts to start the consensus engine
func (t *Tendermint) Start() error {
	err := t.service.Start()
	if err != nil {
		return fmt.Errorf("failed to start baseledger core consensus; %s", err.Error())
	}

	common.Log.Debugf("initialized baseledger core consensus; %v", t.service.String())
	return nil
}

// Stop gracefully stops the consensus engine
func (t *Tendermint) Stop() {
	defer func() {
		if r := recover(); r != nil {
			common.Log.Warningf("recovered while attempting to stop baseledger core consensus; %s", r)
		}
	}()

	t.baseline.Shutdown()
	t.service.Stop()
}

func initTendermint(
	cfg *common.Config,
	logger *log.Logger,
	genesis *types.GenesisDoc,
	baseline *protocol.Baseline,
) (service.Service, error) {
	service, err := node.New(
		&cfg.Config,
		*logger,
		proxy.NewLocalClientCreator(baseline),
		genesis,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize baseledger node: %s", err.Error())
	}

	return service, nil
}
