package protocol

import (
	"encoding/json"
	"fmt"

	"github.com/kthomas/go-natsutil"
	"github.com/nats-io/nats.go"

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

	stakingContractSubscription *nats.Subscription
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

// Shutdown handles the graceful shutdown of all service resources
func (s *Service) Shutdown() error {
	err := s.unsubscribeStakingSubscription()
	if err != nil {
		return err
	}

	return nil
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
		contract, err := s.requireStakingContract(
			token,
			*cfg.StakingNetwork,
			params.Staking,
		)
		if err != nil {
			common.Log.Errorf("failed to create staking contract reference; contract address: %s; %s", *cfg.StakingContractAddress, err.Error())
			return err
		}

		// assert(*cfg.StakingContractAddress == *contract.Address)

		s.stakingContractSubscription, err = s.startStakingSubscriptions(token, contract)
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
			"address":    *network.Address,
			"name":       params.Contract.Name,
			"network_id": networkID,
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

// unsubscribeStakingSubscription gracefully shuts down and removes any active
// subscription to events emitted by the configured staking contract
func (s *Service) unsubscribeStakingSubscription() error {
	if s.stakingContractSubscription != nil {
		err := s.stakingContractSubscription.Unsubscribe()
		if err != nil {
			common.Log.Warningf("failed to unsubscribe from staking contract")
			return err
		}
	}

	return nil
}

// vendContractSubscriptionBearerToken users the nchain api to vend a VC authorizing access
// to a dedicated subject where events will be delivere
func (s *Service) vendContractSubscriptionBearerToken(token string, contract *nchain.Contract) (*string, error) {
	if contract.Address == nil {
		return nil, fmt.Errorf("failed to vend contract subscription bearer token; nil contract address")
	}

	tkn, err := nchain.VendContractSubscriptionToken(token, *contract.Address, map[string]interface{}{})
	if err != nil {
		common.Log.Errorf("failed to vend contract subscription bearer token; contract address: %s; %s", *contract.Address, err.Error())
		return nil, err
	}

	common.Log.Debugf("vended contract subscription token: %s", tkn.Token)
	return tkn.Token, nil
}

func (s *Service) startStakingSubscriptions(token string, contract *nchain.Contract) (*nats.Subscription, error) {
	if contract.Address == nil {
		return nil, fmt.Errorf("failed to subscribe to staking contract events; nil contract address")
	}

	_, err := s.vendContractSubscriptionBearerToken(token, contract)
	if err != nil {
		common.Log.Debugf("FIXME-- nchain token vending machine api is currently issuing a 500; please add integration test")
	}

	conn, _ := natsutil.GetSharedNatsConnection(&token)
	subject := fmt.Sprintf("network.%s.contracts.%s", contract.NetworkID, *contract.Address)
	subscription, err := conn.Subscribe(subject, func(msg *nats.Msg) {
		common.Log.Debugf("consuming %d-byte NATS contract event on subject: %s", len(msg.Data), msg.Subject)

		// TODO: unmarshal to StakingContractEvent to handle the following staking contract events:

		// Deposit/stake
		//
		// Become a depositor in the configured staking contract or increase an existing position.
		//
		// sig: Deposit (address addr, address beneficiary, bytes32 validator, uint256 amount)
		// raw: 0x000000000000000000000000bee25e36774dc2baeb14342f1e821d5f765e2739000000000000000000000000bee25e36774dc2baeb14342f1e821d5f765e2739eacbbc154c8373d7cb9134ed2a2fa2a4bdaf8bfef27b91299b8dce4042bd00000000000000000000000000000000000000000000000000000000000005f5e100
		//
		// This event is emitted from EVM/mainnet when a validator deposit succeeds, either by way of
		// governance approval or, in primitive/testnet setups, simply calling the external deposit()
		// method on the staking contract.
		//
		// A governance contract architecture is being developed which will, among other things,
		// make the staking contract upgradable by way of the governance council.
		//
		// Staking contract source: https://github.com/Baseledger/baseledger-contracts/blob/master/contracts/Staking.sol#L42
		// Example transaction on Ropsten: https://ropsten.etherscan.io/tx/0xbe4f32e51074830622d2fe553c59fb08611faa7bfdb37667e1a67f5374a6df14

		// Withdraw
		//
		// Initiate the withdrawal of a portion, or all, of a previously deposited stake from the
		// configured staking contract. If this transaction affects the withdrawal of 100% of the
		// amount on deposit, the validator will cease to participate in block rewards effective
		// after some number of block confirmations. The number of confirmations required prior to
		// the Baseledger network recognizing any associated updates to the validator set is
		// determined based on which EVM-based network is hosting the staking and token contracts:
		//
		// Network			Block Confirmations
		// -------			-------------------
		// mainnet			[30]
		// ropsten			[3]
		//
		// sig: Withdraw (address addr, bytes32 validator, uint256 amount)
		// raw: 0x000000000000000000000000bee25e36774dc2baeb14342f1e821d5f765e2739eacbbc154c8373d7cb9134ed2a2fa2a4bdaf8bfef27b91299b8dce4042bd00000000000000000000000000000000000000000000000000000000000000000929
		//
		// This event is emitted from EVM/mainnet when a validator withdrawal succeeds, either by way of
		// governance approval or, in primitive/testnet setups, simply calling the external withdraw()
		// method on the staking contract.
		//
		// A governance contract architecture is being developed which will, among other things,
		// make the staking contract upgradable by way of the governance council.
		//
		// Staking contract source: https://github.com/Baseledger/baseledger-contracts/blob/master/contracts/Staking.sol#L61
		// Example transaction on Ropsten: https://ropsten.etherscan.io/tx/0xd85f15cd13749b7572485f4cbccc197743e9078ac5f60e4a2aa9a55122427412

	})

	if err != nil {
		return nil, err
	}

	common.Log.Debugf("established NATS subscription on subject: %s", subscription.Subject)
	return subscription, nil
}
