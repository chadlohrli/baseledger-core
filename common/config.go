package common

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	uuid "github.com/kthomas/go.uuid"
	"github.com/providenetwork/tendermint/config"
	"github.com/provideplatform/provide-go/common"
	prvdutil "github.com/provideplatform/provide-go/common"
)

const baseledgerModeFull = "full"
const baseledgerModeSeed = "seed"
const baseledgerModeValidator = "validator"

const defaultABCIConnectionType = "socket"
const defaultBlockTime = time.Second * 5
const defaultConfigFilePath = "config.json"
const defaultChainID = "peachtree"
const defaultFastSync = true
const defaultFastSyncVersion = "v2"
const defaultFilterPeers = false
const defaultLogFormat = "plain"
const defaultMode = "validator"
const defaultDBBackend = "goleveldb"
const defaultGenesisFilePath = "genesis.json"
const defaultGenesisURL = ""
const defaultGenesisStateURL = "https://s3.amazonaws.com/static.provide.services/capabilities/baseledger-genesis-state.json"
const defaultMempoolCacheSize = 256
const defaultMempoolSize = 1024
const defaultNetworkName = "Baseledger"
const defaultP2PListenAddress = "tcp://0.0.0.0:33333"
const defaultP2PMaxConnections = uint16(32)
const defaultP2PMaxPacketMessagePayloadSize = 22020096
const defaultP2PPersistentPeerMaxDialPeriod = time.Second * 10
const defaultRPCMaxSubscriptionsPerClient = 32
const defaultRPCMaxSubscriptionClients = 1024
const defaultPeerAlias = "prvd"
const defaultRPCCORSOrigins = "*"
const defaultRPCListenAddress = "tcp://0.0.0.0:1337"
const defaultRPCMaxOpenConnections = 1024
const defaultTxIndexer = "kv"

// Config is the baseledger configuration
type Config struct {
	config.Config

	ChainID         string   `json:"chain_id"`
	GenesisURL      *url.URL `json:"genesis_url"`
	GenesisStateURL *url.URL `json:"genesis_state_url"`

	VaultID    *uuid.UUID `json:"vault_id"`
	VaultKeyID *uuid.UUID `json:"vault_key_id"`

	VaultRefreshToken   *string `json:"-"`
	ProvideRefreshToken *string `json:"-"`
}

func (c *Config) IsFullNode() bool {
	return strings.ToLower(c.Mode) == baseledgerModeFull
}

func (c *Config) IsValidatorNode() bool {
	return strings.ToLower(c.Mode) == baseledgerModeValidator
}

func (c *Config) IsSeedNode() bool {
	return strings.ToLower(c.Mode) == baseledgerModeSeed
}

func ConfigFactory() (*Config, error) {
	chainID := defaultChainID
	if os.Getenv("BASELEDGER_CHAIN_ID") != "" {
		chainID = os.Getenv("BASELEDGER_CHAIN_ID")
	}

	var genesisURL *url.URL
	if os.Getenv("BASELEDGER_GENESIS_URL") != "" {
		_url, err := url.Parse(os.Getenv("BASELEDGER_GENESIS_URL"))
		if err != nil {
			panic(err)
		}
		genesisURL = _url
	}

	genesisStateURL, _ := url.Parse(defaultGenesisStateURL)
	if os.Getenv("BASELEDGER_GENESIS_STATE_URL") != "" {
		stateURL, err := url.Parse(os.Getenv("BASELEDGER_GENESIS_STATE_URL"))
		if err != nil {
			panic(err)
		}
		genesisStateURL = stateURL
	}

	mode := defaultMode
	if os.Getenv("BASELEDGER_MODE") != "" {
		mode = os.Getenv("BASELEDGER_MODE")
	}

	networkName := defaultNetworkName
	if os.Getenv("BASELEDGER_NETWORK_NAME") != "" {
		networkName = os.Getenv("BASELEDGER_NETWORK_NAME")
	}

	logLevel := defaultLogLevel
	if os.Getenv("BASELEDGER_LOG_LEVEL") != "" {
		logLevel = os.Getenv("BASELEDGER_LOG_LEVEL")
	}

	logFormat := defaultLogFormat
	if os.Getenv("BASELEDGER_LOG_FORMAT") != "" {
		logFormat = os.Getenv("BASELEDGER_LOG_FORMAT")
	}

	dbBackend := defaultDBBackend
	if os.Getenv("BASELEDGER_DB_BACKEND") != "" {
		dbBackend = os.Getenv("BASELEDGER_DB_BACKEND")
	}

	fastSync := defaultFastSync
	if os.Getenv("BASELEDGER_FAST_SYNC") != "" {
		fastSync = os.Getenv("BASELEDGER_FAST_SYNC") == "true"
	}

	fastSyncVersion := defaultFastSyncVersion
	if os.Getenv("BASELEDGER_FAST_SYNC_VERSION") != "" {
		fastSyncVersion = os.Getenv("BASELEDGER_FAST_SYNC_VERSION")
	}

	txIndexer := defaultTxIndexer
	if os.Getenv("BASELEDGER_TX_INDEXER") != "" {
		txIndexer = os.Getenv("BASELEDGER_TX_INDEXER")
	}

	// abciConnectionType := defaultABCIConnectionType
	// if os.Getenv("BASELEDGER_ABCI_CONNECTION_TYPE") != "" {
	// 	abciConnectionType = os.Getenv("BASELEDGER_ABCI_CONNECTION_TYPE")
	// }

	filterPeers := defaultFilterPeers
	if os.Getenv("BASELEDGER_FILTER_PEERS") != "" {
		filterPeers = os.Getenv("BASELEDGER_FILTER_PEERS") == "true"
	}

	blockTime := defaultBlockTime
	if os.Getenv("BASELEDGER_BLOCK_TIME") != "" {
		time, err := time.ParseDuration(os.Getenv("BASELEDGER_BLOCK_TIME"))
		if err != nil {
			panic(err)
		}
		blockTime = time
	}

	mempoolSize := defaultMempoolSize
	if os.Getenv("BASELEDGER_MEMPOOL_SIZE") != "" {
		size, err := strconv.ParseInt(os.Getenv("BASELEDGER_MEMPOOL_SIZE"), 10, 64)
		if err != nil {
			panic(err)
		}
		mempoolSize = int(size)
	}

	mempoolCacheSize := defaultMempoolCacheSize
	if os.Getenv("BASELEDGER_MEMPOOL_CACHE_SIZE") != "" {
		size, err := strconv.ParseInt(os.Getenv("BASELEDGER_MEMPOOL_CACHE_SIZE"), 10, 64)
		if err != nil {
			panic(err)
		}
		mempoolCacheSize = int(size)
	}

	rpcListenAddress := defaultRPCListenAddress
	if os.Getenv("BASELEDGER_RPC_LISTEN_ADDRESS") != "" {
		rpcListenAddress = os.Getenv("BASELEDGER_RPC_LISTEN_ADDRESS")
	}

	rpcCORSOrigins := strings.Split(defaultRPCCORSOrigins, ",")
	if os.Getenv("BASELEDGER_RPC_CORS_ORIGINS") != "" {
		rpcCORSOrigins = strings.Split(os.Getenv("BASELEDGER_RPC_CORS_ORIGINS"), ",")
	}

	rpcMaxOpenConnections := defaultRPCMaxOpenConnections
	if os.Getenv("BASELEDGER_RPC_MAX_OPEN_CONNECTIONS") != "" {
		maxConnections, err := strconv.ParseInt(os.Getenv("BASELEDGER_RPC_MAX_OPEN_CONNECTIONS"), 10, 64)
		if err != nil {
			panic(err)
		}
		rpcMaxOpenConnections = int(maxConnections)
	}

	rpcMaxSubscriptionClients := defaultRPCMaxSubscriptionClients
	if os.Getenv("BASELEDGER_RPC_MAX_SUBSCRIPTION_CLIENTS") != "" {
		maxClients, err := strconv.ParseInt(os.Getenv("BASELEDGER_RPC_MAX_SUBSCRIPTION_CLIENTS"), 10, 64)
		if err != nil {
			panic(err)
		}
		rpcMaxSubscriptionClients = int(maxClients)
	}

	rpcMaxSubscriptionsPerClient := defaultRPCMaxSubscriptionsPerClient
	if os.Getenv("BASELEDGER_RPC_MAX_CLIENT_SUBSCRIPTIONS") != "" {
		maxSubscriptions, err := strconv.ParseInt(os.Getenv("BASELEDGER_RPC_MAX_CLIENT_SUBSCRIPTIONS"), 10, 64)
		if err != nil {
			panic(err)
		}
		rpcMaxSubscriptionsPerClient = int(maxSubscriptions)
	}

	peerAlias := defaultPeerAlias
	if os.Getenv("BASELEDGER_PEER_ALIAS") != "" {
		peerAlias = os.Getenv("BASELEDGER_PEER_ALIAS")
	}

	p2pListenAddress := defaultP2PListenAddress
	if os.Getenv("BASELEDGER_P2P_LISTEN_ADDRESS") != "" {
		p2pListenAddress = os.Getenv("BASELEDGER_P2P_LISTEN_ADDRESS")
	}

	p2pMaxConnections := defaultP2PMaxConnections
	if os.Getenv("BASELEDGER_P2P_MAX_CONNECTIONS") != "" {
		maxConnections, err := strconv.ParseInt(os.Getenv("BASELEDGER_P2P_MAX_CONNECTIONS"), 10, 16)
		if err != nil {
			panic(err)
		}
		p2pMaxConnections = uint16(maxConnections)
	}

	p2pPersistentPeerMaxDialPeriod := defaultP2PPersistentPeerMaxDialPeriod
	if os.Getenv("BASELEDGER_P2P_PERSISTENT_PEER_MAX_DIAL_PERIOD") != "" {
		duration, err := time.ParseDuration(os.Getenv("BASELEDGER_P2P_PERSISTENT_PEER_MAX_DIAL_PERIOD"))
		if err != nil {
			panic(err)
		}
		p2pPersistentPeerMaxDialPeriod = duration
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	rootPath := fmt.Sprintf("%s%s.baseledger%s%s", homeDir, string(os.PathSeparator), string(os.PathSeparator), chainID)
	err = os.MkdirAll(rootPath, 0700)
	if err != nil {
		panic(err)
	}

	p2pBroadcastAddress := os.Getenv("BASELEDGER_PEER_BROADCAST_ADDRESS")
	if p2pBroadcastAddress == "" {
		addr, err := prvdutil.ResolvePublicIP()
		if err != nil {
			panic(err)
		}

		p2pListenAddrParts := strings.Split(p2pListenAddress, ":")
		p2pListenPort := p2pListenAddrParts[len(p2pListenAddrParts)-1]
		p2pBroadcastAddress = fmt.Sprintf("%s:%s", *addr, p2pListenPort)
	}

	p2pMaxPacketMessagePayloadSize := defaultP2PMaxPacketMessagePayloadSize
	if os.Getenv("BASELEDGER_P2P_MAX_PACKET_MESSAGE_PAYLOAD_SIZE") != "" {
		size, err := strconv.ParseInt(os.Getenv("BASELEDGER_P2P_MAX_PACKET_MESSAGE_PAYLOAD_SIZE"), 10, 64)
		if err != nil {
			panic(err)
		}
		p2pMaxPacketMessagePayloadSize = int(size)
	}

	p2pSeedPeers := os.Getenv("BASELEDGER_SEEDS")
	// p2pBootstrapPeers := os.Getenv("BASELEDGER_BOOTSTRAP_PEERS")
	p2pPersistentPeers := os.Getenv("BASELEDGER_PERSISTENT_PEERS")

	var provideRefreshToken *string
	if os.Getenv("PROVIDE_REFRESH_TOKEN") != "" {
		provideRefreshToken = common.StringOrNil(os.Getenv("PROVIDE_REFRESH_TOKEN"))
	}

	var vaultRefreshToken string
	if os.Getenv("VAULT_REFRESH_TOKEN") != "" {
		vaultRefreshToken = os.Getenv("VAULT_REFRESH_TOKEN")
	}

	var vaultID *uuid.UUID
	var vaultIDStr string
	if os.Getenv("VAULT_ID") != "" {
		vaultUUID, err := uuid.FromString(os.Getenv("VAULT_ID"))
		if err != nil {
			return nil, fmt.Errorf("failed to parse VAULT_ID as valid uuid; %s", err.Error())
		}

		vaultID = &vaultUUID
		vaultIDStr = vaultID.String()
	}

	var vaultKeyID *uuid.UUID
	var vaultKeyIDStr string
	if os.Getenv("VAULT_KEY_ID") != "" {
		keyUUID, err := uuid.FromString(os.Getenv("VAULT_KEY_ID"))
		if err != nil {
			return nil, fmt.Errorf("failed to parse VAULT_KEY_ID as valid uuid; %s", err.Error())
		}

		vaultKeyID = &keyUUID
		vaultKeyIDStr = vaultKeyID.String()
	}

	cfg := &Config{
		Config: config.Config{
			BaseConfig: config.BaseConfig{
				// The root directory for all data.
				// This should be set in viper so it can unmarshal into this struct
				RootDir: rootPath,

				// TCP or UNIX socket address of the ABCI application,
				// or the name of an ABCI application compiled in with the Tendermint binary
				ProxyApp: networkName,

				// Mode of Node: full | validator | seed
				// * validator
				//   - all reactors
				//   - with priv_validator_key.json, priv_validator_state.json
				// * full
				//   - all reactors
				//   - No priv_validator_key.json, priv_validator_state.json
				// * seed
				//   - only P2P, PEX Reactor
				//   - No priv_validator_key.json, priv_validator_state.json
				Mode: mode,

				// Path to the JSON file containing the private key to use as a validator in the consensus protocol
				PrivValidatorKey: fmt.Sprintf("%s%svalidator.json", rootPath, string(os.PathSeparator)),

				// Path to the JSON file containing the last sign state of a validator
				PrivValidatorState: fmt.Sprintf("%s%svalidator-state.json", rootPath, string(os.PathSeparator)),

				// A custom human readable name for this node
				Moniker: peerAlias,

				// If this node is many blocks behind the tip of the chain, FastSync
				// allows them to catchup quickly by downloading blocks in parallel
				// and verifying their commits
				FastSyncMode: fastSync,

				// Database backend: goleveldb | cleveldb | boltdb | rocksdb
				// * goleveldb (github.com/syndtr/goleveldb - most popular implementation)
				//   - pure go
				//   - stable
				// * cleveldb (uses levigo wrapper)
				//   - fast
				//   - requires gcc
				//   - use cleveldb build tag (go build -tags cleveldb)
				// * boltdb (uses etcd's fork of bolt - github.com/etcd-io/bbolt)
				//   - EXPERIMENTAL
				//   - may be faster is some use-cases (random reads - indexer)
				//   - use boltdb build tag (go build -tags boltdb)
				// * rocksdb (uses github.com/tecbot/gorocksdb)
				//   - EXPERIMENTAL
				//   - requires gcc
				//   - use rocksdb build tag (go build -tags rocksdb)
				// * badgerdb (uses github.com/dgraph-io/badger)
				//   - EXPERIMENTAL
				//   - use badgerdb build tag (go build -tags badgerdb)
				DBBackend: dbBackend,

				// Path to the JSON file containing the initial validator set and other meta data
				Genesis: fmt.Sprintf("%s%sgenesis.json", rootPath, string(os.PathSeparator)),

				// Database directory
				DBPath: fmt.Sprintf("%s%sdb", rootPath, string(os.PathSeparator)),

				// Output level for logging
				LogLevel: logLevel,

				// Output format: 'plain' (colored text) or 'json'
				LogFormat: logFormat,

				// NodeKey json path
				NodeKey: fmt.Sprintf("%s%snode.json", rootPath, string(os.PathSeparator)),

				// Mechanism to connect to the ABCI application: socket | grpc
				ABCI: "asdf",

				// If true, query the ABCI app on connecting to a new peer
				// so the app can decide if we should keep the connection or not
				FilterPeers: filterPeers,

				VaultID:           vaultIDStr,
				VaultKeyID:        vaultKeyIDStr,
				VaultRefreshToken: vaultRefreshToken,
			},

			FastSync: &config.FastSyncConfig{
				Version: fastSyncVersion,
			},

			Consensus: &config.ConsensusConfig{
				RootDir: rootPath,
				WalPath: fmt.Sprintf("%s%swrite-ahead.log", rootPath, string(os.PathSeparator)),

				// How long we wait for a proposal block before prevoting nil
				TimeoutPropose: blockTime,

				// How much timeout_propose increases with each round
				// TimeoutProposeDelta time.Duration `mapstructure:"timeout_propose_delta"`

				// How long we wait after receiving +2/3 prevotes for “anything” (ie. not a single block or nil)
				TimeoutPrevote: blockTime,

				// How much the timeout_prevote increases with each round
				// TimeoutPrevoteDelta time.Duration `mapstructure:"timeout_prevote_delta"`

				// How long we wait after receiving +2/3 precommits for “anything” (ie. not a single block or nil)
				TimeoutPrecommit: blockTime,

				// How much the timeout_precommit increases with each round
				// TimeoutPrecommitDelta time.Duration `mapstructure:"timeout_precommit_delta"`

				// How long we wait after committing a block, before starting on the new
				// height (this gives us a chance to receive some more precommits, even
				// though we already have +2/3).
				// TimeoutCommit: time.Millisecond * 25,

				// Make progress as soon as we have all the precommits (as if TimeoutCommit = 0)
				SkipTimeoutCommit: true,

				// EmptyBlocks mode and possible interval between empty blocks
				CreateEmptyBlocks:         true,
				CreateEmptyBlocksInterval: blockTime,

				// Reactor sleep duration parameters
				// PeerGossipSleepDuration     time.Duration `mapstructure:"peer_gossip_sleep_duration"`
				// PeerQueryMaj23SleepDuration time.Duration `mapstructure:"peer_query_maj23_sleep_duration"`

				// DoubleSignCheckHeight int64 `mapstructure:"double_sign_check_height"`
			},

			Instrumentation: &config.InstrumentationConfig{},

			Mempool: &config.MempoolConfig{
				RootDir:   rootPath,
				Recheck:   true,
				Broadcast: true,

				// Maximum number of transactions in the mempool
				Size: mempoolSize,

				// Limit the total size of all txs in the mempool.
				// This only accounts for raw transactions (e.g. given 1MB transactions and
				// max-txs-bytes=5MB, mempool will only accept 5 transactions).
				// MaxTxsBytes int64 `mapstructure:"max-txs-bytes"`

				// Size of the cache (used to filter transactions we saw earlier) in transactions
				CacheSize: mempoolCacheSize,

				// Do not remove invalid transactions from the cache (default: false)
				// Set to true if it's not possible for any invalid transaction to become
				// valid again in the future.
				// KeepInvalidTxsInCache bool `mapstructure:"keep-invalid-txs-in-cache"`

				// Maximum size of a single transaction
				// NOTE: the max size of a tx transmitted over the network is {max-tx-bytes}.
				// MaxTxBytes int `mapstructure:"max-tx-bytes"`

				// Maximum size of a batch of transactions to send to a peer
				// Including space needed by encoding (one varint per transaction).
				// XXX: Unused due to https://github.com/providenetwork/tendermint/issues/5796
				// MaxBatchBytes int `mapstructure:"max-batch-bytes"`

				// TTLDuration, if non-zero, defines the maximum amount of time a transaction
				// can exist for in the mempool.
				//
				// Note, if TTLNumBlocks is also defined, a transaction will be removed if it
				// has existed in the mempool at least TTLNumBlocks number of blocks or if it's
				// insertion time into the mempool is beyond TTLDuration.
				// TTLDuration time.Duration `mapstructure:"ttl-duration"`

				// TTLNumBlocks, if non-zero, defines the maximum number of blocks a transaction
				// can exist for in the mempool.
				//
				// Note, if TTLDuration is also defined, a transaction will be removed if it
				// has existed in the mempool at least TTLNumBlocks number of blocks or if
				// it's insertion time into the mempool is beyond TTLDuration.
				// TTLNumBlocks int64 `mapstructure:"ttl-num-blocks"`
			},

			RPC: &config.RPCConfig{
				RootDir: rootPath,

				// TCP or UNIX socket address for the RPC server to listen on
				ListenAddress: rpcListenAddress,

				// A list of origins a cross-domain request can be executed from.
				// If the special '*' value is present in the list, all origins will be allowed.
				// An origin may contain a wildcard (*) to replace 0 or more characters (i.e.: http://*.domain.com).
				// Only one wildcard can be used per origin.
				CORSAllowedOrigins: rpcCORSOrigins,

				// A list of methods the client is allowed to use with cross-domain requests.
				CORSAllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD"},

				// A list of non simple headers the client is allowed to use with cross-domain requests.
				CORSAllowedHeaders: []string{"Accept", "Content-Type", "Origin", "X-Requested-With", "X-Server-Time", "X-Total-Results-Count"},

				// Activate unsafe RPC commands like /dial-persistent-peers and /unsafe-flush-mempool
				Unsafe: false,

				// Maximum number of simultaneous connections (including WebSocket).
				// Does not include gRPC connections. See grpc-max-open-connections
				// If you want to accept a larger number than the default, make sure
				// you increase your OS limits.
				// 0 - unlimited.
				// Should be < {ulimit -Sn} - {MaxNumInboundPeers} - {MaxNumOutboundPeers} - {N of wal, db and other open files}
				// 1024 - 40 - 10 - 50 = 924 = ~900
				MaxOpenConnections: rpcMaxOpenConnections,

				// Maximum number of unique clientIDs that can /subscribe
				// If you're using /broadcast_tx_commit, set to the estimated maximum number
				// of broadcast_tx_commit calls per block.
				MaxSubscriptionClients: rpcMaxSubscriptionClients,

				// Maximum number of unique queries a given client can /subscribe to
				// If you're using GRPC (or Local RPC client) and /broadcast_tx_commit, set
				// to the estimated maximum number of broadcast_tx_commit calls per block.
				MaxSubscriptionsPerClient: rpcMaxSubscriptionsPerClient,

				// How long to wait for a tx to be committed during /broadcast_tx_commit
				// WARNING: Using a value larger than 10s will result in increasing the
				// global HTTP write timeout, which applies to all connections and endpoints.
				// See https://github.com/providenetwork/tendermint/issues/3435
				// TimeoutBroadcastTxCommit time.Duration `mapstructure:"timeout-broadcast-tx-commit"`

				// Maximum size of request body, in bytes
				// MaxBodyBytes int64 `mapstructure:"max-body-bytes"`

				// Maximum size of request header, in bytes
				// MaxHeaderBytes int `mapstructure:"max-header-bytes"`

				// The path to a file containing certificate that is used to create the HTTPS server.
				// Might be either absolute path or path related to Tendermint's config directory.
				//
				// If the certificate is signed by a certificate authority,
				// the certFile should be the concatenation of the server's certificate, any intermediates,
				// and the CA's certificate.
				//
				// NOTE: both tls-cert-file and tls-key-file must be present for Tendermint to create HTTPS server.
				// Otherwise, HTTP server is run.
				// TLSCertFile string `mapstructure:"tls-cert-file"`

				// The path to a file containing matching private key that is used to create the HTTPS server.
				// Might be either absolute path or path related to tendermint's config directory.
				//
				// NOTE: both tls-cert-file and tls-key-file must be present for Tendermint to create HTTPS server.
				// Otherwise, HTTP server is run.
				// TLSKeyFile string `mapstructure:"tls-key-file"`

				// pprof listen address (https://golang.org/pkg/net/http/pprof)
				// PprofListenAddress string `mapstructure:"pprof-laddr"`
			},

			P2P: &config.P2PConfig{
				RootDir: rootPath,

				// Address to listen for incoming connections
				ListenAddress: p2pListenAddress,

				// Address to advertise to peers for them to dial
				ExternalAddress: p2pBroadcastAddress,

				// Comma separated list of seed nodes to connect to
				// We only use these if we can’t connect to peers in the addrbook
				// NOTE: not used by the new PEX reactor. Please use BootstrapPeers instead.
				// TODO: Remove once p2p refactor is complete
				// ref: https://github.com/providenetwork/tendermint/issues/5670
				Seeds: p2pSeedPeers,

				// Comma separated list of peers to be added to the peer store
				// on startup. Either BootstrapPeers or PersistentPeers are
				// needed for peer discovery
				// BootstrapPeers: p2pBootstrapPeers,

				// Comma separated list of nodes to keep persistent connections to
				PersistentPeers: p2pPersistentPeers,

				// UPNP port forwarding
				UPNP: true,

				// Path to address book
				AddrBook: fmt.Sprintf("%s%saddress-book.json", rootPath, string(os.PathSeparator)),

				// Set true for strict address routability rules
				// Set false for private or local networks
				// AddrBookStrict bool `mapstructure:"addr-book-strict"`

				// Maximum number of inbound peers
				//
				// TODO: Remove once p2p refactor is complete in favor of MaxConnections.
				// ref: https://github.com/providenetwork/tendermint/issues/5670
				MaxNumInboundPeers: int(p2pMaxConnections),

				// Maximum number of outbound peers to connect to, excluding persistent peers.
				//
				// TODO: Remove once p2p refactor is complete in favor of MaxConnections.
				// ref: https://github.com/providenetwork/tendermint/issues/5670
				MaxNumOutboundPeers: int(p2pMaxConnections),

				// MaxConnections defines the maximum number of connected peers (inbound and
				// outbound).
				// MaxConnections: p2pMaxConnections,

				// MaxIncomingConnectionAttempts rate limits the number of incoming connection
				// attempts per IP address.
				// MaxIncomingConnectionAttempts uint `mapstructure:"max-incoming-connection-attempts"`

				// List of node IDs, to which a connection will be (re)established ignoring any existing limits
				// UnconditionalPeerIDs string `mapstructure:"unconditional-peer-ids"`

				// Maximum pause when redialing a persistent peer (if zero, exponential backoff is used)
				PersistentPeersMaxDialPeriod: p2pPersistentPeerMaxDialPeriod,

				// Time to wait before flushing messages out on the connection
				// FlushThrottleTimeout time.Duration `mapstructure:"flush-throttle-timeout"`

				// Maximum size of a message packet payload, in bytes
				MaxPacketMsgPayloadSize: p2pMaxPacketMessagePayloadSize,

				// Rate at which packets can be sent, in bytes/second
				// SendRate int64 `mapstructure:"send-rate"`

				// Rate at which packets can be received, in bytes/second
				// RecvRate int64 `mapstructure:"recv-rate"`

				// Set true to enable the peer-exchange reactor
				PexReactor: true,

				// Comma separated list of peer IDs to keep private (will not be gossiped to
				// other peers)
				// PrivatePeerIDs string `mapstructure:"private-peer-ids"`

				// Toggle to disable guard against peers connecting from the same ip.
				AllowDuplicateIP: false,

				// Peer connection configuration.
				// HandshakeTimeout time.Duration `mapstructure:"handshake-timeout"`
				// DialTimeout      time.Duration `mapstructure:"dial-timeout"`

				// DisableLegacy is used mostly for testing to enable or disable the legacy
				// P2P stack.
				// DisableLegacy bool `mapstructure:"disable-legacy"`

				// Makes it possible to configure which queue backend the p2p
				// layer uses. Options are: "fifo", "priority" and "wdrr",
				// with the default being "fifo".
				// QueueType string `mapstructure:"queue-type"`
			},

			StateSync: &config.StateSyncConfig{
				Enable: false,
				// TempDir             string        `mapstructure:"temp_dir"`
				// RPCServers          []string      `mapstructure:"rpc_servers"`
				// TrustPeriod         time.Duration `mapstructure:"trust_period"`
				// TrustHeight         int64         `mapstructure:"trust_height"`
				// TrustHash           string        `mapstructure:"trust_hash"`
				// DiscoveryTime       time.Duration `mapstructure:"discovery_time"`
				// ChunkRequestTimeout time.Duration `mapstructure:"chunk_request_timeout"`
				// ChunkFetchers       int32         `mapstructure:"chunk_fetchers"`
			},

			TxIndex: &config.TxIndexConfig{
				Indexer: txIndexer,
			},
		},

		ChainID:         chainID,
		GenesisURL:      genesisURL,
		GenesisStateURL: genesisStateURL,

		ProvideRefreshToken: provideRefreshToken,

		VaultID:           vaultID,
		VaultKeyID:        vaultKeyID,
		VaultRefreshToken: &vaultRefreshToken,
	}

	cfgJSON, err := json.MarshalIndent(cfg, "", "    ")
	if err != nil {
		return nil, err
	}

	cfgFilePath := fmt.Sprintf("%s%s%s", rootPath, string(os.PathSeparator), defaultConfigFilePath)
	err = os.WriteFile(cfgFilePath, cfgJSON, 0644)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
