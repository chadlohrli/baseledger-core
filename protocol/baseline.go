package protocol

import (
	"encoding/json"
	"fmt"
	"sync"
	//"os"

	"github.com/providenetwork/baseledger/common"
	abcitypes "github.com/providenetwork/tendermint/abci/types"
	"github.com/providenetwork/tendermint/types"
)

const abciStateCheckTx = "check_tx"
const abciStateDeliverTx = "deliver_tx"
const abciStateCommit = "commit"

const eventTypeBlock = "block"
const eventNewHeader = "header"

const defaultABCISemanticVersion = "v1.0.0"
const defaultEntropyBlockInterval = 100

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

	mutex         *sync.Mutex
	queryHandlers *QueryHandlers
}

func BaselineProtocolFactory(cfg *common.Config, genesis *types.GenesisDoc) (*Baseline, error) {
	service, err := serviceFactory(cfg, genesis)
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
		Version: defaultABCISemanticVersion,

		CheckTxState:   checkTxState,
		DeliverTxState: deliverTxState,
		CommitState:    commitState,

		mutex:         &sync.Mutex{},
		queryHandlers: queryHandlersFactory(service.nchain),
	}, nil
}

func (b *Baseline) ApplySnapshotChunk(req abcitypes.RequestApplySnapshotChunk) abcitypes.ResponseApplySnapshotChunk {
	common.Log.Debugf("ApplySnapshotChunk; %s", req)
	return abcitypes.ResponseApplySnapshotChunk{}
}

func (b *Baseline) BeginBlock(req abcitypes.RequestBeginBlock) abcitypes.ResponseBeginBlock {
	common.Log.Debugf("BeginBlock; %v", req)
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
		RetainHeight: 0,
	}
}

func (b *Baseline) DeliverTx(req abcitypes.RequestDeliverTx) abcitypes.ResponseDeliverTx {
	common.Log.Debugf("DeliverTx; %s", req)
	return abcitypes.ResponseDeliverTx{Code: 0}
}

func (b *Baseline) EndBlock(req abcitypes.RequestEndBlock) abcitypes.ResponseEndBlock {
	common.Log.Debugf("EndBlock; %v", req)

	// check config to see if node is a validator vs full node
	err := b.resolveRandomBeaconEntropy(req)
	if err != nil {
		common.Log.Warningf("failed to resolve random beacon entropy; %s", err.Error())
	}

	validatorUpdates := b.resolveValidatorUpdates(req)

	return abcitypes.ResponseEndBlock{
		ValidatorUpdates: validatorUpdates,
	}
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
	validators := defaultValidatorsFactory(b.Genesis)
	for _, validator := range validators {
		b.CommitState.Validators = append(
			b.CommitState.Validators,
			validatorFactory(validator.PubKey.GetEd25519(), validator.Power),
		)
	}

	// TODO: call to entropy should probably happen here 

	return abcitypes.ResponseInitChain{
		AppHash:         b.CommitState.Root,
		ConsensusParams: req.ConsensusParams,
		Validators:      validators,
	}
}

func (b *Baseline) ListSnapshots(req abcitypes.RequestListSnapshots) abcitypes.ResponseListSnapshots {
	common.Log.Debugf("ListSnapshots; %s", req)
	return abcitypes.ResponseListSnapshots{}
}

func (b *Baseline) LoadSnapshotChunk(req abcitypes.RequestLoadSnapshotChunk) abcitypes.ResponseLoadSnapshotChunk {
	common.Log.Debugf("LoadSnapshotChunk; %s", req)
	return abcitypes.ResponseLoadSnapshotChunk{}
}

func (b *Baseline) OfferSnapshot(req abcitypes.RequestOfferSnapshot) abcitypes.ResponseOfferSnapshot {
	common.Log.Debugf("OfferSnapshot; %s", req)
	return abcitypes.ResponseOfferSnapshot{}
}

func (b *Baseline) SetOption(req abcitypes.RequestSetOption) abcitypes.ResponseSetOption {
	common.Log.Debugf("SetOption; %s", req)
	return abcitypes.ResponseSetOption{}
}

func (b *Baseline) Query(req abcitypes.RequestQuery) abcitypes.ResponseQuery {
	resp, err := b.queryHandlers.handle(req)
	if resp != nil && err == nil {
		return *resp
	}

	return abcitypes.ResponseQuery{
		Code: 0,
	}
}

// Shutdown handles the consolidated shutdown of all ABCI-owned resources
func (b *Baseline) Shutdown() error {
	err := b.Service.unsubscribeStakingSubscription()
	if err != nil {
		return err
	}

	return nil
}

// resolveBeaconEntropy resolves entropy for a random beacon and dispatches
// a transaction to store this entropy as part of the next block
func (b *Baseline) resolveRandomBeaconEntropy(req abcitypes.RequestEndBlock) error {

	if req.Height%defaultEntropyBlockInterval == 0 {
		// store latest L1-derived entropy...
		// create a query randomness query
		reqQuery := abcitypes.RequestQuery {
			Path: "/baseline/entropy/fetch",
			Height: 0,
			Prove: false,
		}
		resQuery := b.Query(reqQuery)
		common.Log.Debugf("%s", resQuery)
		//os.Exit(1)

		randomness := resQuery.Value
		common.Log.Debugf("%s", randomness)

		// 1.  TODO: parse  randomness
		// ex. randomness = 0: uint256: 11274928295812345  (as bytes)
		
		// 2. store in block header
		common.Log.Debugf("TODO-- fetch and store entropy at height... %d", req.Height)
	}

	return nil
}

// resolveValidatorUpdates for the given block
func (b *Baseline) resolveValidatorUpdates(req abcitypes.RequestEndBlock) []abcitypes.ValidatorUpdate {
	validatorUpdates := make([]abcitypes.ValidatorUpdate, 0)

	read := true
	for read {
		select {
		case delta := <-b.Service.validatorDeltasChannel:
			validator := b.CommitState.GetValidator(delta.Address)
			if validator == nil {
				validator = validatorFactory(delta.PublicKey, 0)
				b.CommitState.Validators = append(b.CommitState.Validators, validator)
				common.Log.Debugf("adding new validator %s in block %d", validator.Address, req.Height)
			}

			common.Log.Debugf("applying validator staking delta to validator %s in block %d", validator.Address, req.Height)
			validator.AdjustStake(delta.StakingDelta)
			validatorUpdates = append(validatorUpdates, validator.AsValidatorUpdate())
		default:
			// no more buffered updates
			read = false
		}
	}

	if b.CommitState.TotalVotingPower() == 0 {
		common.Log.Debugf("all validator staking power withdrawn as of block %d; reverting to default validator set", req.Height)
		validatorUpdates = append(validatorUpdates, defaultValidatorsFactory(b.Genesis)...)
	}

	return validatorUpdates
}
