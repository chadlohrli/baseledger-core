package protocol

import (
	"math/big"

	abcitypes "github.com/providenetwork/tendermint/abci/types"
	"github.com/providenetwork/tendermint/config"
	"github.com/providenetwork/tendermint/proto/tendermint/crypto"
	"github.com/providenetwork/tendermint/types"
)

const defaultProtocolVersion = "v1.0.0"

var _ abcitypes.Application = (*Baseline)(nil)

type Baseline struct {
	Config  *config.Config
	Genesis *types.GenesisDoc

	Version string
}

func BaselineProtocolFactory(cfg *config.Config, genesis *types.GenesisDoc) *Baseline {
	return &Baseline{
		Config:  cfg,
		Genesis: genesis,
		Version: defaultProtocolVersion,
	}
}

func (b *Baseline) ApplySnapshotChunk(req abcitypes.RequestApplySnapshotChunk) abcitypes.ResponseApplySnapshotChunk {
	return abcitypes.ResponseApplySnapshotChunk{}
}

func (b *Baseline) BeginBlock(req abcitypes.RequestBeginBlock) abcitypes.ResponseBeginBlock {
	return abcitypes.ResponseBeginBlock{}
}

func (b *Baseline) CheckTx(req abcitypes.RequestCheckTx) abcitypes.ResponseCheckTx {
	tx, _ := TransactionFromRaw(req.Tx)
	code := tx.isValid()
	return abcitypes.ResponseCheckTx{Code: code, GasWanted: 1}
}

func (b *Baseline) Commit() abcitypes.ResponseCommit {
	return abcitypes.ResponseCommit{}
}

func (b *Baseline) DeliverTx(req abcitypes.RequestDeliverTx) abcitypes.ResponseDeliverTx {
	return abcitypes.ResponseDeliverTx{Code: 0}
}

func (b *Baseline) EndBlock(req abcitypes.RequestEndBlock) abcitypes.ResponseEndBlock {
	return abcitypes.ResponseEndBlock{}
}

func (b *Baseline) Info(req abcitypes.RequestInfo) abcitypes.ResponseInfo {
	return abcitypes.ResponseInfo{
		AppVersion: new(big.Int).SetBytes([]byte(b.Version)).Uint64(),
		Version:    req.Version,
	}
}

func (b *Baseline) InitChain(req abcitypes.RequestInitChain) abcitypes.ResponseInitChain {
	validators := make([]abcitypes.ValidatorUpdate, 0)
	for _, validator := range b.Genesis.Validators {
		validators = append(validators, abcitypes.ValidatorUpdate{
			PubKey: crypto.PublicKey{
				Sum: &crypto.PublicKey_Ed25519{
					Ed25519: validator.PubKey.Bytes(),
				},
			},
			Power: validator.Power,
		})
	}

	return abcitypes.ResponseInitChain{
		AppHash:         req.AppStateBytes,
		ConsensusParams: req.ConsensusParams,
		Validators:      validators,
	}
}

func (b *Baseline) ListSnapshots(req abcitypes.RequestListSnapshots) abcitypes.ResponseListSnapshots {
	return abcitypes.ResponseListSnapshots{}
}

func (b *Baseline) LoadSnapshotChunk(req abcitypes.RequestLoadSnapshotChunk) abcitypes.ResponseLoadSnapshotChunk {
	return abcitypes.ResponseLoadSnapshotChunk{}
}

func (b *Baseline) OfferSnapshot(req abcitypes.RequestOfferSnapshot) abcitypes.ResponseOfferSnapshot {
	return abcitypes.ResponseOfferSnapshot{}
}

func (b *Baseline) Query(req abcitypes.RequestQuery) abcitypes.ResponseQuery {
	return abcitypes.ResponseQuery{Code: 0}
}
