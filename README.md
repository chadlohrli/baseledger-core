# baseledger-core

_Baseledger core consensus client for running a validator, full or seed node._

⚠️ WARNING: this code is not ready for production. The Baseledger mainnet is currently scheduled to launch in Q1 2022. Use the "peachtree" testnet.

This package depends on a modified version of Tendermint v0.34.11 which requires a Vault for storing private key material. We also bundle the Baseledger ABCI within a single binary for ease-of-use and and cross-platform portability.

## Prerequisites & System Requirements

- golang (1.16 recommended)

## Build

```
git clone git@github.com:baseledger/baseledger-core.git
make build
```

## Creating a Vault

Using the Provide CLI, you can easily setup a secure vault to house your Baseledger keys. Baseledger currently supports Ed25519 keys for peer-to-peer authorization and validator keys.

If you do not have a Provide user, first create one:

```
prvd users create
```

Next, authenticate using your Provide credentials and create a vault and Ed25519 key:

```
prvd authenticate
prvd vaults init --name 'Baseledger Vault'
prvd vaults keys init

# follow the prompts; control-c to bypass selecting an application and organization

✔ Baseledger
✔ asymmetric
✔ sign/verify
Name: Baseledger node key
Description: My first baseledger node key
✔ Ed25519
```

You will see the UUID of the created vault key.

You will also need to authorize a refresh token for the user or organization that is the owner of the vault by running the following:

```
prvd api_tokens init [--organization <org uuid>] --offline-access
```

Note the refresh token, vault id and vault key id. Each of these values will be used to run a validator or full node.

## Running a Full Node

You can use the following command to run a `full` node on the Baseledger "peachtree" testnet:

```
VAULT_REFRESH_TOKEN=<your refresh token> \
VAULT_ID=<vault id> \
VAULT_KEY_ID=<vault key id> \
BASELEDGER_MODE=full \
LOG_LEVEL=debug \
BASELEDGER_LOG_LEVEL='main:info,*:error' \
BASELEDGER_GENESIS_URL=http://genesis.peachtree.baseledger.provide.network:1337/genesis \
BASELEDGER_PERSISTENT_PEERS=e0f0ce7a37be16ede67f70831d5608c5ea6e8540@genesis.peachtree.baseledger.provide.network:33333 \
BASELEDGER_PEER_ALIAS=<your alias> \
./.bin/node
```

## Running a Validator Node

Running a validator node requires the user to be a depositor on the configured staking contract.
See the staking contract `deposit()` method described below.

You can use the following command to run a `validator` node on the Baseledger "peachtree" testnet:

```
VAULT_REFRESH_TOKEN=<your refresh token> \
VAULT_ID=<vault id> \
VAULT_KEY_ID=<vault key id> \
BASELEDGER_MODE=validator \
LOG_LEVEL=debug \
BASELEDGER_LOG_LEVEL='main:info,*:error' \
BASELEDGER_GENESIS_URL=http://genesis.peachtree.baseledger.provide.network:1337/genesis \
BASELEDGER_PERSISTENT_PEERS=e0f0ce7a37be16ede67f70831d5608c5ea6e8540@genesis.peachtree.baseledger.provide.network:33333 \
BASELEDGER_PEER_ALIAS=<your alias> \
./.bin/node
```

## Governance

A governance contract architecture is being developed which will, among other things,
make the staking and other future contracts upgradable by way of the governance council.

### Ethereum Bridge

We have taken a minimalistic approach to the Baseledger node implementation using tendermint.
A critical part of the architecture is maintaining a highly fault-tolerant bridge between a
configured Ethereum network (e.g., mainnet, ropsten, kovan, etc.) and the Baseledger network
(e.g., mainnet or peachtree etc).

#### Latency

Just as crypto exchanges await a number of block confirmations before making deposited assets
available for use, there are a number of block confirmations which must occur on the EVM-based
network which hosts the Baseledger governance and staking contracts prior to any bridged changes
taking effect on the Baseledger network.

For example, if a staking contract `withdraw()` transaction affects the withdrawal of 100%
of the amount on deposit, the validator will cease to participate in block rewards effective
after the number of block confirmations. The number of L1 confirmations required prior to the
Baseledger network recognizing any associated updates (e.g., changes to the validator set) is
determined based on which EVM-based network is hosting the staking and token contracts:

| Network | Block Confirmations |
|--|--|
| mainnet | 30 |
| ropsten | 3 |
| rinkeby | _not supported at this time_ |
| kovan | _not supported at this time_ |
| goerli | _not supported at this time_ |

## Staking Contract

A [staking contract](https://github.com/Baseledger/baseledger-contracts/blob/master/contracts/Staking.sol), initialized with a reference to the UBT token contract address, is deployed on the following Ethereum networks:

| Network | Symbol | Token Contract Address | Staking Contract Address |
|--|--|--|--|
| mainnet | UBT | `0x8400D94A5cb0fa0D041a3788e395285d61c9ee5e` | -- |
| ropsten | UBTR | `0xa9ec5862d3D25caF1eCae6e9d48aDacD8CE5899c` | `0xFD02dAA2BAd85D38C8a4ad6B61afbA0ff92b30F4` |
| rinkeby | -- | -- | -- |
| kovan | -- | -- | -- |
| goerli | -- | -- | -- |

### Proxy Staking

An abstract proxy staking mechanism is being developed to add composable delegation functionality to the staking contract (see relevant placeholder in the source code [here](https://github.com/Baseledger/baseledger-contracts/blob/master/contracts/Staking.sol#L17)). Validators will be able to create competitive proxy staking offerings and implementations.

💡 _This is a great idea for a hackathon project at the upcoming [EthAtlanta](https://ethatl.com) hackathon, happening October 1-3._

### Methods

The core functionality of the staking contract is to enable deposits and withdrawals of UBT on the Ethereum mainnet,
or "test UBT" (such as [UBTR](https://ropsten.etherscan.io/token/0xa9ec5862d3D25caF1eCae6e9d48aDacD8CE5899c), on the Ropsten testnet).

#### `Deposit (address addr, address beneficiary, bytes32 validator, uint256 amount)`

    Become a depositor to the configured staking contract or increase an existing position.

    This method emits events from the EVM/mainnet when a validator deposit succeeds, either by
    way of governance approval or, in primitive/testnet setups, implicit approval.

    Staking contract source can be found [here](https://github.com/Baseledger/baseledger-contracts/blob/master/contracts/Staking.sol#L42).
    Example transaction on Ropsten can be found [here](https://ropsten.etherscan.io/tx/0xbe4f32e51074830622d2fe553c59fb08611faa7bfdb37667e1a67f5374a6df14).

---

#### `Withdraw (address addr, bytes32 validator, uint256 amount)`

    Initiate the withdrawal of a portion, or all, of a previously deposited stake from the
    configured staking contract.

    This method emits events from the EVM/mainnet when a validator withdrawal succeeds, either by
    way of governance approval or, in primitive/testnet setups, implicit approval.
    method on the staking contract.

    Staking contract source can be found [here](https://github.com/Baseledger/baseledger-contracts/blob/master/contracts/Staking.sol#L61).
    Example transaction on Ropsten can be found [here](https://ropsten.etherscan.io/tx/0xd85f15cd13749b7572485f4cbccc197743e9078ac5f60e4a2aa9a55122427412).

---

_Additional documentation forthcoming._
