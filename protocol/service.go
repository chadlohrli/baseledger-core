package protocol

import (
	"encoding/json"

	"github.com/providenetwork/baseledger/common"
	"github.com/providenetwork/tendermint/types"
	"github.com/provideplatform/provide-go/api/baseline"
	"github.com/provideplatform/provide-go/api/ident"
	"github.com/provideplatform/provide-go/api/nchain"
	"github.com/provideplatform/provide-go/api/privacy"
	"github.com/provideplatform/provide-go/api/vault"
)

// Service instance exposes a compliant implementation of the Baseline protocol
type Service struct {
	baseline *baseline.Service
	ident    *ident.Service
	nchain   *nchain.Service
	privacy  *privacy.Service
	vault    *vault.Service
}

func authorizeAccessToken(refreshToken string) (*ident.Token, error) {
	token, err := ident.CreateToken(refreshToken, map[string]interface{}{
		"grant_type": "refresh_token",
	})
	if err != nil {
		return nil, err
	}

	return token, nil
}

func serviceFactory(cfg *common.Config, genesis *types.GenesisDoc) (*Service, error) {
	if cfg.ProvideRefreshToken == nil {
		common.Log.Debug("baseline protocol service implementation not configured; no bearer refresh token provided")
		return nil, nil
	}

	token, err := authorizeAccessToken(*cfg.ProvideRefreshToken)
	if err != nil {
		common.Log.Panicf("failed to initialize baseline protocol service implementation; bearer access token not authorized; %s", err.Error())
	}

	srvc := &Service{
		baseline: baseline.InitBaselineService(*token.AccessToken),
		ident:    ident.InitIdentService(token.AccessToken),
		nchain:   nchain.InitNChainService(*token.AccessToken),
		privacy:  privacy.InitPrivacyService(*token.AccessToken),
		vault:    vault.InitVaultService(token.AccessToken),
	}

	var stateParams *StateParams
	err = json.Unmarshal(genesis.AppState, &stateParams)
	if err != nil {
		common.Log.Warningf("failed to unmarshal genesis state; %s", err.Error())
	}

	err = srvc.initStaking(*token.AccessToken, cfg, stateParams)
	if err != nil {
		common.Log.Panicf("failed to initialize baseline protocol service implementation; state not initialized; %s", err.Error())
	}

	return srvc, nil
}

func (s *Service) initStaking(token string, cfg *common.Config, params *StateParams) error {

	stakingContractConfigured := cfg.StakingContractAddress != nil && cfg.StakingNetwork != nil
	if !stakingContractConfigured && params != nil && params.Staking != nil && params.Staking.Contract != nil && cfg.StakingNetwork != nil {
		stakingNetworkParams, err := params.Staking.Network.GetParams(*cfg.StakingNetwork)
		if err != nil {
			return err
		}
		stakingContractConfigured = stakingNetworkParams.Address != nil
	}

	if stakingContractConfigured {
		_, err := s.requireStakingContract(
			token,
			*cfg.StakingNetwork,
			params.Staking,
		)
		if err != nil {
			common.Log.Errorf("failed to create staking contract reference; contract address: %s; %s", *cfg.StakingContractAddress, err.Error())
			return err
		}

		err = s.startStakingSubscriptions(token, *cfg.StakingContractAddress)
		if err != nil {
			common.Log.Warningf("failed to subscribe to configured staking contract address: %s; %s", *cfg.StakingContractAddress, err.Error())
			return err
		}
	} else {
		common.Log.Warning("no staking contract address configured; consensus limited to static validator set")
	}

	return nil
}

func (s *Service) requireStakingContract(token, networkName string, params *StakingParams) (*nchain.Contract, error) {
	var contract *nchain.Contract
	var err error

	network, err := params.Network.GetParams(networkName)
	if err != nil {
		return nil, err
	}

	networkID, err := network.NChainID(networkName)
	if err != nil {
		return nil, err
	}

	contract, err = nchain.GetContractDetails(token, *network.Address, map[string]interface{}{})
	if err != nil {
		contract, err = nchain.CreatePublicContract(token, map[string]interface{}{
			"network_id": networkID,
			"name":       params.Contract.Name,
			"address":    *network.Address,
			"params": map[string]interface{}{
				"compiled_artifact": params.Contract,
				"argv":              network.Argv,
			},
		})
	}
	if err != nil {
		return nil, err
	}

	common.Log.Debugf("resolved staking contract: %s", *network.Address)
	return contract, nil
}

func (s *Service) startStakingSubscriptions(token, contractAddress string) error {
	tkn, err := nchain.VendContractSubscriptionToken(token, contractAddress, map[string]interface{}{})
	if err != nil {
		common.Log.Errorf("failed to subscribe to staking contract events; contract address: %s; %s", contractAddress, err.Error())
		return err
	}

	common.Log.Debugf("vended contract subscription token: %s", tkn.Token)

	return nil
}
