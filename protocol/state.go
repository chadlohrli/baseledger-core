package protocol

import (
	"bytes"
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

	Name       string         `json:"name"`
	Height     int64          `json:"height"`
	Root       []byte         `json:"root"`
	Staking    *StakingParams `json:"staking"`
	Validators []*Validator   `json:"validators"`
}

// GetValidator returns the validator if it exists in the state instance, or nil
func (s *State) GetValidator(address []byte) *Validator {
	for _, validator := range s.Validators {
		if bytes.Equal(address, []byte(*validator.Address)) {
			return validator
		}
	}

	return nil
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

// TotalVotingPower returns the total validator votinmg power, as it exists in the state instance
func (s *State) TotalVotingPower() int64 {
	power := int64(0)

	for _, validator := range s.Validators {
		power += validator.VotingPower()
	}

	return power
}

func stateFactory(cfg *common.Config, name string, genesis *types.GenesisDoc) (*State, error) {
	stateMutex.Lock()
	defer stateMutex.Unlock()

	path := fmt.Sprintf("%s%sabci-state-%s.json", cfg.RootDir, string(os.PathSeparator), name)
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

	var stateParams *StateParams
	var staking *StakingParams
	err := json.Unmarshal(genesis.AppState, &stateParams)
	if err == nil {
		common.Log.Debug("unmarshaled genesis state to state params")
		staking = stateParams.Staking
	}

	return &State{
		path:       path,
		Name:       name,
		Height:     0,
		Root:       []byte{},
		Staking:    staking,
		Validators: make([]*Validator, 0),
	}, nil
}
