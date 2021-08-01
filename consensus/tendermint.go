package consensus

import (
	"fmt"

	"github.com/providenetwork/baseledger/protocol"
	"github.com/providenetwork/tendermint/libs/log"
	"github.com/providenetwork/tendermint/libs/service"
	"github.com/providenetwork/tendermint/node"
	"github.com/providenetwork/tendermint/proxy"
	"github.com/providenetwork/tendermint/types"
)

func InitTendermint(
	cfg *Config,
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
