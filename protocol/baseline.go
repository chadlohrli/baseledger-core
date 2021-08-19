package protocol

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/providenetwork/baseledger/common"
	abcitypes "github.com/providenetwork/tendermint/abci/types"
	"github.com/providenetwork/tendermint/proto/tendermint/crypto"
	"github.com/providenetwork/tendermint/types"
)

const abciStateCheckTx = "check_tx"
const abciStateDeliverTx = "deliver_tx"
const abciStateCommit = "commit"

const eventTypeBlock = "block"
const eventNewHeader = "header"

const defaultABCISemanticVersion = "v1.0.0"

// Baseline is the tendermint Application Blockchain Interface (ABCI);
// it must conform to the ABCI specification. Use extreme care when
// making use of the underlying Service, as all interactions with the
// ABCI must be deterministic across the entire network
type Baseline struct {
	Config  *common.Config
	Genesis *types.GenesisDoc
	Service *Service
	Version string

	CheckTxState   *State
	DeliverTxState *State
	CommitState    *State
}

func BaselineProtocolFactory(cfg *common.Config, genesis *types.GenesisDoc) (*Baseline, error) {
	service, err := serviceFactory(cfg)
	if err != nil {
		return nil, err
	}

	checkTxState, err := stateFactory(cfg, abciStateCheckTx, genesis)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize ABCI check tx state; %s", err.Error())
	}

	deliverTxState, err := stateFactory(cfg, abciStateDeliverTx, genesis)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize ABCI deliver tx state; %s", err.Error())
	}

	commitState, err := stateFactory(cfg, abciStateCommit, genesis)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize ABCI commit state; %s", err.Error())
	}

	return &Baseline{
		Config:  cfg,
		Genesis: genesis,
		Service: service,

		CheckTxState:   checkTxState,
		DeliverTxState: deliverTxState,
		CommitState:    commitState,

		Version: defaultABCISemanticVersion,
	}, nil
}

func (b *Baseline) ApplySnapshotChunk(req abcitypes.RequestApplySnapshotChunk) abcitypes.ResponseApplySnapshotChunk {
	return abcitypes.ResponseApplySnapshotChunk{}
}

func (b *Baseline) BeginBlock(req abcitypes.RequestBeginBlock) abcitypes.ResponseBeginBlock {
	resp := abcitypes.ResponseBeginBlock{}

	rawHeader, err := json.Marshal(req.Header)
	if err == nil {
		resp.Events = append(resp.Events, abcitypes.Event{
			Type: eventTypeBlock,
			Attributes: []abcitypes.EventAttribute{
				{
					Key:   []byte(eventNewHeader),
					Value: rawHeader,
				},
			},
		})
	}

	return resp
}

func (b *Baseline) CheckTx(req abcitypes.RequestCheckTx) abcitypes.ResponseCheckTx {
	tx, _ := TransactionFromRaw(req.Tx)
	code := tx.isValid()
	return abcitypes.ResponseCheckTx{
		Code:      code,
		GasWanted: tx.calculateGas(),
	}
}

func (b Baseline) Commit() abcitypes.ResponseCommit {
	b.CommitState.Height++
	b.CommitState.Save() // TODO-- buffer this
	return abcitypes.ResponseCommit{
		RetainHeight: b.CommitState.Height,
	}
}

func (b *Baseline) DeliverTx(req abcitypes.RequestDeliverTx) abcitypes.ResponseDeliverTx {
	return abcitypes.ResponseDeliverTx{Code: 0}
}

func (b *Baseline) EndBlock(req abcitypes.RequestEndBlock) abcitypes.ResponseEndBlock {
	common.Log.Debugf("END BLOCK; %v", req)
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
		LastBlockAppHash: b.CommitState.Root,
		LastBlockHeight:  b.CommitState.Height,
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

	pubkey, _ := base64.StdEncoding.DecodeString("6su8FUyDc9fLCRNODSovoqS9r4v+8ge5Epm43OQAQr0=")
	validators = append(validators, abcitypes.ValidatorUpdate{
		PubKey: crypto.PublicKey{
			Sum: &crypto.PublicKey_Ed25519{
				Ed25519: pubkey,
			},
		},
		Power: 1,
	})

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

func (b *Baseline) SetOption(req abcitypes.RequestSetOption) abcitypes.ResponseSetOption {
	return abcitypes.ResponseSetOption{}
}

func (b *Baseline) Query(req abcitypes.RequestQuery) abcitypes.ResponseQuery {
	common.Log.Debugf("QUERY; %v", req)
	return abcitypes.ResponseQuery{Code: 0}
}
