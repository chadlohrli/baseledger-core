package protocol

import (
	"fmt"

	"github.com/providenetwork/baseledger/common"
	abcitypes "github.com/providenetwork/tendermint/abci/types"
	"github.com/providenetwork/tendermint/proto/tendermint/crypto"
	"github.com/providenetwork/tendermint/types"
)

const defaultABCISemanticVersion = "v1.0.0"

var _ abcitypes.Application = (*Baseline)(nil)

// Baseline is the tendermint Application Blockchain Interface (ABCI);
// it must conform to the ABCI specification. Use extreme care when
// making use of the underlying Service, as all interactions with the
// ABCI must be deterministic across the entire network
type Baseline struct {
	Config  *common.Config
	Genesis *types.GenesisDoc
	Service *Service
	State   *State
	Version string
}

func BaselineProtocolFactory(cfg *common.Config, genesis *types.GenesisDoc) (*Baseline, error) {
	service, err := serviceFactory(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize service; %s", err.Error())
	}

	state, err := stateFactory(cfg, genesis)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize ABCI state; %s", err.Error())
	}

	return &Baseline{
		Config:  cfg,
		Genesis: genesis,
		Service: service,
		State:   state,
		Version: defaultABCISemanticVersion,
	}, nil
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

// Info should return a response comprised of the following:
//
// AppVersion (uint64): The application protocol version
// Data (string): Some arbitrary information
// LastBlockAppHash ([]byte): Latest result of Commit
// LastBlockHeight (int64): Latest block for which the app has called Commit
// Version (string): The application software semantic version
func (b *Baseline) Info(req abcitypes.RequestInfo) abcitypes.ResponseInfo {
	return abcitypes.ResponseInfo{
		AppVersion:       b.Genesis.ConsensusParams.Version.AppVersion,
		Data:             "hello world",
		LastBlockAppHash: b.State.Root,
		LastBlockHeight:  b.State.Height,
		Version:          b.Version,
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
