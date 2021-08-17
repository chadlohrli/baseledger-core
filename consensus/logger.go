package consensus

import (
	"os"

	"github.com/providenetwork/baseledger/common"
	"github.com/providenetwork/tendermint/libs/log"
)

// LogFactory returns the tendermint logger
func LogFactory(cfg *common.Config) (*log.Logger, error) {
	logger := log.NewTMJSONLogger(os.Stdout)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to initialize tendermint logger; %s", err.Error())
	// }

	return &logger, nil
}
