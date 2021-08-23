# baseledger-core

Baseledger core consensus for running validator, full and seed nodes.

This package depends on a modified version of Tendermint v0.34.11 which requires a Vault for storing private key material. We also bundle the Baseledger ABCI within a single binary for ease-of-use and and cross-platform portability.

## Prerequisites

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

You can use the following command to run a full node on the Baseledger "peachtree" testnet:

```
VAULT_REFRESH_TOKEN=<your refresh token>
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

Additional documentation forthcoming.