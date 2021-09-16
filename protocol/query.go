package protocol

import (
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/providenetwork/baseledger/common"
	abcitypes "github.com/providenetwork/tendermint/abci/types"
)

const queryRegexPeerAddressFilter = `^\/p2p\/filter\/addr\/(.*)$`
const peerAddressFilterResponseCode = 1
const peerAddressFilterResponseTimeout = time.Millisecond * 2500

type QueryHandlers struct {
	expressions map[string]*regexp.Regexp
	handlers    map[string]func(abcitypes.RequestQuery) abcitypes.ResponseQuery
}

func queryHandlersFactory() *QueryHandlers {
	return &QueryHandlers{
		expressions: map[string]*regexp.Regexp{
			queryRegexPeerAddressFilter: regexp.MustCompile(queryRegexPeerAddressFilter),
		},
		handlers: map[string]func(abcitypes.RequestQuery) abcitypes.ResponseQuery{
			queryRegexPeerAddressFilter: filterPeerQuery,
		},
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