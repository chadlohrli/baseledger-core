# baseledger-core

_Baseledger core consensus client for running a validator, full or seed node._

⚠️ WARNING: this code has not been audited and is not ready for use in production. The Baseledger mainnet is scheduled to launch in Q1 2022.

This package depends on a modified version of Tendermint v0.34.11 which requires a Vault for storing private key material. We also compile the Baseledger Application Blockchain Interface ([ABCI](https://docs.tendermint.com/master/spec/abci)) and [tendermint core](https://docs.tendermint.com) dependencies together within a single binary for ease-of-use, cross-platform portability and optimal performance.

## Quickstart

1. Install golang
2. Build the project from source
3. Create a Vault and Ed25519 keypair using the `prvd` CLI
4. Run a full or validator node
5. (Optional) Deposit UBT in the appropriate staking contract

## Roadmap

[Baseledger](https://baseledger.net) was created by [Unibright](https://unibright.io) and [Provide](https://provide.network) and is under active development.

### Testnets

In the spirit of developing a proof of concept implementation to experiment with network validation in tendermint (including staking and delegation), native opcodes and a community block explorer, the team built the ["lakewood" testnet](https://github.com/baseledger/lakewood). This testnet was created using Cosmos SDK.

The ["peachtree" testnet](https://explorer.peachtree.baseledger.net) was created from scratch using tendermint for BFT consensus and the [Provide stack](https://docs.provide.services) for subscribing to events emitted by the Baseledger governance and staking contracts, broadcasting _baseline proofs_ to the network and otherwise interacting with the [Baseline Protocol](https://github.com/eea-oasis/baseline). As a result of this design, `baseledger-core` can be built as a single container and added to existing deployments of the Provide stack for increased security. `baseledger-core` can also run standalone (i.e., outside the context of a Provide stack). Baseledger nodes running outside the context of a Provide stack are not restricted from operating as validator, full or seed nodes. Organizations implementing the _baseline_ pattern in commercial multiparty workflows benefit from running a local Baseledger node because it provides additional security to the cryptographic commitments (proofs) stored within the Provide stack without sacrificing any privacy guarantees inherent to _baselining_.

### Mainnet

The Baseledger mainnet is currently scheduled to launch in Q1 2022. More information will be made available in Q4 2021 about the governance council and how you can apply to become a validator on the mainnet to earn block rewards in UBT.

## Prerequisites

- golang (1.16 recommended); only required when building from source

### Hardware Requirements

The following are the minimum hardware requirements recommended for running a Baseledger validator or null node:

- 2+ CPU cores
- 4GB RAM minimum
- SSD; recommended minimum free disk space >= 10GB as of September 2021

## Build

```
git clone git@github.com:baseledger/baseledger-core.git
make build
```

## Creating a Vault

Using the [Provide CLI](https://github.com/provideplatform/provide-cli), you can easily setup a secure vault instance to secure your Baseledger keys. Baseledger currently supports Ed25519 keys for peer-to-peer authorization and validator keys.

If you have not previously created a Provide user, first create one:

```
prvd users create
```

Next, authenticate using those credentials and create a vault and Ed25519 key:

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
| ropsten | UBTR | `0xa9a466b8f415bcc5883934eda70016f8b23ea776` | `0x0B5FC75192F8EE3B4795AB44b3B455aB3d97A6dF` |
| rinkeby | -- | -- | -- |
| kovan | -- | -- | -- |
| goerli | -- | -- | -- |

### Faucet

A faucet is being added to [Provide Payments](https://docs.provide.services/payments) to make it easy to request UBT on supported testnets via the CLI.

### Methods

The core purpose of the staking contract is to enable deposits and withdrawals of UBT on the Ethereum mainnet,
or "test UBT" (such as [UBTR](https://ropsten.etherscan.io/token/0xa9a466b8f415bcc5883934eda70016f8b23ea776), on the Ropsten testnet).

#### `deposit(address beneficiary, bytes32 validator, uint256 amount) external`

Become a depositor to the configured staking contract or increase an existing position.

Prior to making your first deposit into the staking contract from any address, or if a subsequent amount you wish to deposit exceeds the value of the remaining approved tokens,
you must call `approve(address spender, uint25¸amount)` on the token contract to allow it to transfer UBT on your behalf when you call `deposit()`.

The following example contract call to the UBTR token contract (`0xa9a466b8f415bcc5883934eda70016f8b23ea776`) approves the staking contract (`0x0B5FC75192F8EE3B4795AB44b3B455aB3d97A6dF`), enabling you to deposit up to 250,000 UBTR:

```
Function: approve(address spender, uint256 amount)

MethodID: 0x095ea7b3
[0]:  0000000000000000000000000b5fc75192f8ee3b4795ab44b3b455ab3d97a6df
[1]:  000000000000000000000000000000000000000000000000000016bcc41e9000
```

Call the `deposit()` method. The following example contract call to the staking contract on the "peachtree" testnet (`0x0B5FC75192F8EE3B4795AB44b3B455aB3d97A6dF`) results in 25,000 UBTR transferred and placed on deposit for benefit of sender.

```
Function: deposit(address beneficiary, bytes32 validator, uint256 amount) ***

MethodID: 0xeb2243f8
[0]:  000000000000000000000000bee25e36774dc2baeb14342f1e821d5f765e2739
[1]:  eacbbc154c8373d7cb9134ed2a2fa2a4bdaf8bfef27b91299b8dce4042bd0000
[2]:  00000000000000000000000000000000000000000000000000000246139ca800
```

This method emits a `Deposit(address addr, address beneficiary, bytes32 validator, uint256 amount)` event from the EVM/mainnet contract when a validator deposit succeeds, either by
way of governance approval or, in primitive/testnet setups, implicit approval.

Staking contract source can be found [here](https://github.com/Baseledger/baseledger-contracts/blob/master/contracts/Staking.sol#L42).
Example transaction on Ropsten can be found [here](https://ropsten.etherscan.io/tx/0x6f222afc6ee868f09c2738a13ef4508ba1022fd5f2f3d06c4dc63e5901fd4997).

#### `withdraw(uint256 amount) external` 

Initiate the withdrawal of a portion, or all, of a previously deposited stake from the
configured staking contract; the following example contract call to the staking contract on the "peachtree: testnet (`0x0B5FC75192F8EE3B4795AB44b3B455aB3d97A6dF`) results in 10,000 UBTR being withdrawn from our depositor account on the staking contract and returned.

```
Function: withdraw(uint256 value) ***

MethodID: 0x2e1a7d4d
[0]:  000000000000000000000000000000000000000000000000000000e8d4a51000
```

This method emits a `Withdraw(address addr, bytes32 validator, uint256 amount)` event from the EVM/mainnet contract when a validator withdrawal succeeds, either by
way of governance approval or, in primitive/testnet setups, implicit approval.

Staking contract source can be found [here](https://github.com/Baseledger/baseledger-contracts/blob/master/contracts/Staking.sol#L61).
Example transaction on Ropsten can be found [here](https://ropsten.etherscan.io/tx/0x3fb753d45038b38c1cb503fdfa06942c24958e8601a9356983cc0a6968096b99).

### Proxy Staking

An abstract proxy staking mechanism is being developed to add composable delegation functionality to the staking contract (see relevant placeholder in the source code [here](https://github.com/Baseledger/baseledger-contracts/blob/master/contracts/Staking.sol#L17)). Validators will be able to create competitive proxy staking offerings and implementations.

💡 _This is a great idea for a hackathon project at the upcoming [EthAtlanta](https://ethatl.com) hackathon, happening October 1-3._

_Additional documentation forthcoming._
