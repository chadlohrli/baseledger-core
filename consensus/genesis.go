package consensus

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/providenetwork/tendermint/crypto/ed25519"
	"github.com/providenetwork/tendermint/types"
	"github.com/provideplatform/provide-go/api"
)

const defaultGenesisValidatorVotingPower = 1

// GenesisFactory initializes and returns the genesis state, which
// defines the initial conditions for a tendermint blockchain, including
// its validator set and application state
func GenesisFactory(cfg *Config) (*types.GenesisDoc, error) {
	genesis, err := GenesisDocFactory(cfg)
	if err != nil {
		return nil, err
	}

	// write the latest config to disk
	genesisJSON, err := json.MarshalIndent(genesis, "", "    ")
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
func GenesisDocFactory(cfg *Config) (*types.GenesisDoc, error) {
	if cfg.GenesisURL != nil {
		genesisJSON, err := fetchGenesis(cfg)
		if err != nil {
			return nil, err
		}

		var genesis *types.GenesisDoc
		err = json.Unmarshal(genesisJSON, &genesis)
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
		ConsensusParams: &types.ConsensusParams{
			Block: types.BlockParams{
				MaxBytes: 22020096, // 21 MiB
				MaxGas:   -1,
			},
			Evidence: types.EvidenceParams{
				MaxAgeNumBlocks: 100000,
				MaxAgeDuration:  time.Hour * 24 * 30, // 30 days
				MaxBytes:        1048576,             // 1 MiB
			},
			Validator: types.ValidatorParams{
				PubKeyTypes: []string{
					types.ABCIPubKeyTypeEd25519,
				},
			},
		},
		GenesisTime:   genesisTime,
		InitialHeight: int64(0),
		Validators:    genesisValidatorsFactory(cfg),
	}, nil
}

func fetchGenesis(cfg *Config) (json.RawMessage, error) {
	client := &api.Client{
		Host:   cfg.GenesisURL.Host,
		Scheme: cfg.GenesisURL.Scheme,
		Path:   "/",
	}

	_, resp, err := client.Get(cfg.GenesisURL.Path, map[string]interface{}{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch genesis JSON at url: %s; %s", cfg.GenesisURL.String(), err.Error())
	}

	raw, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse genesis JSON; %s", err.Error())
	}

	return json.RawMessage(raw), nil
}

func fetchGenesisState(cfg *Config) (json.RawMessage, error) {
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

func genesisValidatorsFactory(cfg *Config) []types.GenesisValidator {
	validators := make([]types.GenesisValidator, 0)

	pubKey := &ed25519.VaultedPublicKey{
		VaultID:           *cfg.VaultID,
		VaultKeyID:        *cfg.VaultKeyID,
		VaultRefreshToken: *cfg.VaultRefreshToken,
	}

	validators = append(validators, types.GenesisValidator{
		Address: pubKey.Address(),
		PubKey:  pubKey,
		Power:   int64(defaultGenesisValidatorVotingPower),
		Name:    "",
	})

	return validators
}
