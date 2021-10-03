package protocol

import (
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"
	//"os"

	"github.com/providenetwork/baseledger/common"
	abcitypes "github.com/providenetwork/tendermint/abci/types"
	nchain "github.com/provideplatform/provide-go/api/nchain"
)

const queryBlockLatest = "latest"
const queryRegexEntropyFetch = `^\/baseline\/entropy\/fetch\/(.*)$`

const queryRegexPeerAddressFilter = `^\/p2p\/filter\/addr\/(.*)$`
const peerAddressFilterResponseCode = 1
const peerAddressFilterResponseTimeout = time.Millisecond * 100

type QueryHandlers struct {
	expressions map[string]*regexp.Regexp
	handlers    map[string]func(abcitypes.RequestQuery) abcitypes.ResponseQuery
	nchain   	*nchain.Service
}

func queryHandlersFactory(nchain *nchain.Service) *QueryHandlers {
	return &QueryHandlers{
		expressions: map[string]*regexp.Regexp{
			queryRegexEntropyFetch:      regexp.MustCompile(queryRegexEntropyFetch),
			queryRegexPeerAddressFilter: regexp.MustCompile(queryRegexPeerAddressFilter),
		},
		handlers: map[string]func(abcitypes.RequestQuery) abcitypes.ResponseQuery{
			queryRegexEntropyFetch:      fetchEntropy,
			queryRegexPeerAddressFilter: filterPeerQuery,
		},
		nchain: nchain,
	}
}

func (q *QueryHandlers) handle(query abcitypes.RequestQuery) (*abcitypes.ResponseQuery, error) {
	for exp, regexp := range q.expressions {
		if regexp.Match([]byte(query.Path)) {
			resp := q.handlers[exp](query)
			return &resp, nil
		}
	}

	return nil, fmt.Errorf("%d-byte query did not match a registered handler", len(query.Data))
}

// handler implementations

func filterPeerQuery(req abcitypes.RequestQuery) abcitypes.ResponseQuery {
	path := strings.Split(string(req.Path), "/")
	addr := path[len(path)-1]
	conn, err := net.DialTimeout("tcp", addr, peerAddressFilterResponseTimeout)
	if err != nil {
		common.Log.Tracef("filtering unreachable peer: %s", addr)
		return abcitypes.ResponseQuery{
			Code: peerAddressFilterResponseCode,
		}
	}
	conn.Close()

	common.Log.Tracef("peer reachable: %s", addr)
	return abcitypes.ResponseQuery{
		Code: 0,
	}
}

func fetchEntropy(req abcitypes.RequestQuery) abcitypes.ResponseQuery {
	// TODO: the work... query ethereum, chainlink network, etc....
	common.Log.Debugf("in query.fetchEntropy")
	// 1. Call randomResult to get randomness
	// 2. Tx getRandomNumber to generate new randomness for next validator 
	// Edge case -- contract may not generate randomness in time for first validator
	// Should be called during initChain?

	// nchain.ExecuteContract(token, contractID string, params map[string]interface{})
	// How to specify the smart contract method from params map[string]interface{}?
	nchain.ExecuteContract(os.Getenv("ProvideRefreshToken"), os.Getenv("ENTROPY_CONTRACT_ADDRESS"),getRandomNumber)
	abcitypes.ResponseQuery.Value  := nchain.ExecuteContract(os.Getenv("ProvideRefreshToken"), os.Getenv("ENTROPY_CONTRACT_ADDRESS"), randomResult)

	return abcitypes.ResponseQuery{
		Code: 0,
	}
}
