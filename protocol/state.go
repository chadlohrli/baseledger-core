package protocol

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/providenetwork/baseledger/common"
	"github.com/providenetwork/tendermint/types"
)

var stateMutex *sync.Mutex

func init() {
	stateMutex = &sync.Mutex{}
}

// State represents the last-known state of the underlying consensus
type State struct {
	path string `json:"-"`

	Height     int64              `json:"height"`
	Root       []byte             `json:"root"`
	Validators []*types.Validator `json:"validators"`
}

func (s *State) Save() error {
	stateMutex.Lock()
	defer stateMutex.Unlock()

	// write the latest state to disk
	stateJSON, err := json.MarshalIndent(s, "", "    ")
	if err != nil {
		return err
	}

	err = os.WriteFile(s.path, stateJSON, 0644)
	if err != nil {
		return err
	}

	return nil
}

func stateFactory(cfg *common.Config, genesis *types.GenesisDoc) (*State, error) {
	stateMutex.Lock()
	defer stateMutex.Unlock()

	path := fmt.Sprintf("%s%sabci-state.json", cfg.RootDir, string(os.PathSeparator))
	if _, err := os.Stat(path); err == nil {
		stateJSON, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		var state *State
		err = json.Unmarshal(stateJSON, &state)
		if err != nil {
			return nil, err
		}

		state.path = path
		return state, nil
	}

	return &State{
		path:       path,
		Height:     0,
		Root:       []byte{},
		Validators: make([]*types.Validator, 0),
	}, nil
}
