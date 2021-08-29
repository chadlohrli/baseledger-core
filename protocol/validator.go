package protocol

import (
	"encoding/base64"

	"github.com/providenetwork/baseledger/common"
	abcitypes "github.com/providenetwork/tendermint/abci/types"
	"github.com/providenetwork/tendermint/crypto"
	"github.com/providenetwork/tendermint/crypto/tmhash"
	tmcrypto "github.com/providenetwork/tendermint/proto/tendermint/crypto"
	"github.com/providenetwork/tendermint/types"
)

// Validator represents a network validator with a stake and voting power
type Validator struct {
	Name      *string `json:"name,omitempty"`
	Address   *string `json:"address,omitempty"`
	PublicKey []byte  `json:"public_key"`
	Stake     *int64  `json:"stake"`
}

func defaultValidatorsFactory(genesis *types.GenesisDoc) []abcitypes.ValidatorUpdate {
	validators := make([]abcitypes.ValidatorUpdate, 0)

	for _, validator := range genesis.Validators {
		validators = append(validators, abcitypes.ValidatorUpdate{
			PubKey: tmcrypto.PublicKey{
				Sum: &tmcrypto.PublicKey_Ed25519{
					Ed25519: validator.PubKey.Bytes(),
				},
			},
			Power: validator.Power,
		})
	}

	if len(validators) == 0 {
		// FIXME-- look this up- in the genesis state
		pubkey, _ := base64.StdEncoding.DecodeString("6su8FUyDc9fLCRNODSovoqS9r4v+8ge5Epm43OQAQr0=")
		validators = append(validators, abcitypes.ValidatorUpdate{
			PubKey: tmcrypto.PublicKey{
				Sum: &tmcrypto.PublicKey_Ed25519{
					Ed25519: pubkey,
				},
			},
			Power: 1,
		})
	}

	return validators
}

func validatorFactory(publicKey []byte, stake int64) *Validator {
	return &Validator{
		Address:   common.StringOrNil(crypto.Address(tmhash.SumTruncated(publicKey)).String()),
		PublicKey: publicKey,
		Stake:     &stake,
	}
}

func (v *Validator) AdjustStake(delta int64) int64 {
	*v.Stake += delta

	if *v.Stake < 0 {
		common.Log.Warningf("staking delta for validator %s resulted in negative stake (%d); stake will be set to zero", v.Address, v.Stake)
		stake := int64(0)
		v.Stake = &stake
	}

	return *v.Stake
}

func (v *Validator) AsValidatorUpdate() abcitypes.ValidatorUpdate {
	return abcitypes.ValidatorUpdate{
		PubKey: tmcrypto.PublicKey{
			Sum: &tmcrypto.PublicKey_Ed25519{
				Ed25519: v.PublicKey,
			},
		},
		Power: *v.Stake,
	}
}

func (v *Validator) VotingPower() int64 {
	if v.Stake == nil {
		return int64(0)
	}

	return *v.Stake
}
