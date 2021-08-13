package protocol

import (
	"github.com/providenetwork/baseledger/common"
	"github.com/providenetwork/tendermint/types"
)

// State represents the last-known state of the underlying consensus
type State struct {
	Height int64
	Root   []byte
}

func stateFactory(cfg *common.Config, genesis *types.GenesisDoc) (*State, error) {
	var root []byte // TODO-- load last-known state and calculate this

	return &State{
		Height: genesis.InitialHeight,
		Root:   root,
	}, nil
}
