package consensus

import (
	"fmt"

	"github.com/providenetwork/tendermint/libs/log"
)

func LogFactory(cfg *Config) (*log.Logger, error) {
	logger, err := log.NewDefaultLogger(cfg.LogFormat, cfg.LogLevel, true)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize tendermint logger; %s", err.Error())
	}

	return &logger, nil
}
