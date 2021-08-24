package protocol

import (
	"fmt"

	"github.com/providenetwork/baseledger/common"
	"github.com/provideplatform/provide-go/api/nchain"
)

const nchainNetworkIDRopsten = "66d44f30-9092-4182-a3c4-bc02736d6ae5"
const networkRopsten = "ropsten"

type StateParams struct {
	Staking *StakingParams `json:"staking"`
}

type StakingParams struct {
	Contract *nchain.CompiledArtifact `json:"contract"`
	Network  *Network                 `json:"network"`
}

type Network struct {
	Ropsten *NetworkParams `json:"ropsten"`
}

func (n *Network) GetParams(network string) (*NetworkParams, error) {
	switch network {
	case networkRopsten:
		return n.Ropsten, nil
	}

	return nil, fmt.Errorf("failed to get params for unrecognized network: %s", network)
}

type NetworkParams struct {
	Address *string       `json:"address"`
	Argv    []interface{} `json:"argv"`
}

func (n *NetworkParams) NChainID(network string) (*string, error) {
	switch network {
	case networkRopsten:
		return common.StringOrNil(nchainNetworkIDRopsten), nil
	}

	return nil, fmt.Errorf("failed to get nchain id for unrecognized network: %s", network)
}
