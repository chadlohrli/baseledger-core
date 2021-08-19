package consensus

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/providenetwork/baseledger/common"
	tmjson "github.com/providenetwork/tendermint/libs/json"
	tmproto "github.com/providenetwork/tendermint/proto/tendermint/types"
	"github.com/providenetwork/tendermint/types"
	"github.com/provideplatform/provide-go/api"
)

const defaultGenesisAppVersion = 0x1
const defaultGenesisValidatorVotingPower = 1

// GenesisFactory initializes and returns the genesis state, which
// defines the initial conditions for a tendermint blockchain, including
// its validator set and application state
func GenesisFactory(cfg *common.Config) (*types.GenesisDoc, error) {
	genesis, err := GenesisDocFactory(cfg)
	if err != nil {
		return nil, err
	}

	// write the latest config to disk
	genesisJSON, err := tmjson.MarshalIndent(genesis, "", "    ")
	if err != nil {
		return nil, err
	}
	err = os.WriteFile(cfg.BaseConfig.Genesis, genesisJSON, 0644)
	if err != nil {
		return nil, err
	}

	return genesis, nil
}

// GenesisDocFactory returns a GenesisDoc defining the initial parameters for a
// tendermint blockchain, in particular its validator set; if a genesis URL is
// provided by the given config, that resource is unmarshaled and returned.
func GenesisDocFactory(cfg *common.Config) (*types.GenesisDoc, error) {
	if cfg.GenesisURL != nil {
		genesisJSON, err := fetchGenesis(cfg)
		if err != nil {
			return nil, err
		}

		var genesis *types.GenesisDoc
		err = tmjson.Unmarshal(genesisJSON, &genesis)
		if err != nil {
			return nil, err
		}

		return genesis, nil
	}

	if _, err := os.Stat(cfg.Genesis); err == nil {
		// TODO-- match this against contents of genesis url if one is provided...?
		genesisJSON, err := os.ReadFile(cfg.Genesis)
		if err != nil {
			return nil, err
		}

		var genesis *types.GenesisDoc
		err = tmjson.Unmarshal(genesisJSON, &genesis)
		if err != nil {
			return nil, err
		}

		return genesis, nil
	}

	genesisTime, err := time.Parse(time.RFC3339Nano, os.Getenv("BASELEDGER_GENESIS_TIMESTAMP"))
	if err != nil {
		genesisTime = time.Now()
	}

	var genesisState json.RawMessage
	if cfg.GenesisStateURL != nil {
		genesisState, err = fetchGenesisState(cfg)
		if err != nil {
			return nil, err
		}
	}

	return &types.GenesisDoc{
		// AppHash         tmbytes.HexBytes         `json:"app_hash"`
		AppState: genesisState,
		ChainID:  cfg.ChainID,
		ConsensusParams: &tmproto.ConsensusParams{
			Block: tmproto.BlockParams{
				MaxBytes:   22020096, // 21 MiB
				MaxGas:     -1,
				TimeIotaMs: 25,
			},
			Evidence: tmproto.EvidenceParams{
				MaxAgeNumBlocks: 100000,
				MaxAgeDuration:  time.Hour * 24 * 30, // 30 days
				MaxBytes:        1048576,             // 1 MiB
			},
			Validator: tmproto.ValidatorParams{
				PubKeyTypes: []string{
					types.ABCIPubKeyTypeEd25519,
				},
			},
			Version: tmproto.VersionParams{
				AppVersion: defaultGenesisAppVersion,
			},
		},
		GenesisTime:   genesisTime,
		InitialHeight: int64(1),
		// Validators:    genesisValidatorsFactory(cfg),
	}, nil
}

func fetchGenesis(cfg *common.Config) (json.RawMessage, error) {
	client := &api.Client{
		Host:   cfg.GenesisURL.Host,
		Scheme: cfg.GenesisURL.Scheme,
		Path:   "/",
	}

	_, resp, err := client.Get(cfg.GenesisURL.Path, map[string]interface{}{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch genesis JSON at url: %s; %s", cfg.GenesisURL.String(), err.Error())
	}

	var genesis map[string]interface{}
	if response, ok := resp.(map[string]interface{}); ok {
		genesis = response

		// handle genesis by way of rpc	response
		if result, resultOk := genesis["result"].(map[string]interface{}); resultOk {
			if resultGenesis, resultGenesisOk := result["genesis"].(map[string]interface{}); resultGenesisOk {
				genesis = resultGenesis
			}
		}
	}

	raw, err := json.Marshal(genesis)
	if err != nil {
		return nil, fmt.Errorf("failed to parse genesis JSON; %s", err.Error())
	}

	return json.RawMessage(raw), nil
}

func fetchGenesisState(cfg *common.Config) (json.RawMessage, error) {
	client := &api.Client{
		Host:   cfg.GenesisStateURL.Host,
		Scheme: cfg.GenesisStateURL.Scheme,
		Path:   "/",
	}

	_, resp, err := client.Get(cfg.GenesisStateURL.Path, map[string]interface{}{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch genesis state at url: %s; %s", cfg.GenesisStateURL.String(), err.Error())
	}

	raw, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse genesis state as JSON; %s", err.Error())
	}

	return json.RawMessage(raw), nil
}
