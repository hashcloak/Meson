# Meson-plugin

[![Go](https://github.com/hashcloak/Meson-plugin/actions/workflows/go.yml/badge.svg)](https://github.com/hashcloak/Meson-plugin/actions/workflows/go.yml)

A Library for adding cryptocurrencies to Meson mixnet.

## Currently Supported Chains

- Major ETH-based forks and their testnets
  - Ethereum (ETH)
  - Ethereum Classic (ETC)
  - Goerli Testnet (GOR)
  - Rinkeby Testnet (RIN)
  - Kotti Testnet (KOT)

## Add a New Chain

To add support for a new chain, the following needs to be done:
1. In the chain package, create a new file named `($NEW_SUPPORTED_CHAIN)_chain.go`
2. In this new file, create a new struct in which the attributes are properties that are needed to create an appropriate JSON object. 
3. This struct must conform to the `IChain` interface defined in chain.go
4. Once this is done, you need to add your chain to `factory.go` in the `GetChain` function

## Usage

__⚠️ Meson is still in alpha and for this reason we are using a centralized non voting authority. This is also the reason why we need manual exchange of the public keys of the nodes that get added to the network.__

### Dependencies:

- `git`
- `docker swarm`. You will need to have [initialized](https://docs.docker.com/engine/reference/commandline/swarm_init) your docker swarm.
- `go` version >= 1.13
- `python3`
- `make`

To start a testnet run the following command:

```
make tesnet
```

You can then use the [Wallet demo](https://github.com/hashcloak/Meson-wallet-demo) with this tesnet that just got spawned or take a look at [Sending transactions](#how-to-send-transactions). You can see the containers that are running with: `docker service ls`


### Integrations tests

The `ops/` directory contains several python scripts that can help with creating a testnet. It relies extensively on environment variables to set the different conditions of the mixnet that gets spawned. For a full list of environment variables see [this](https://meson.hashcloak.com/docs/#environment-variables).

There are four environment variables that matter for the integration tests:

- `TEST_PKS_ETHEREUM`: The ethereum private key to use 
- `LOG`: Turns on logging
- `BUILD`: Forces the build of the containers instead of pulling form docker hub.
- `TEST_CLIENTCOMMIT`: The commit hash or branch to run the integration tests. Default master. Useful for when there is divergence between the master branches of both Meson-plugin and Meson-client repositories.


An example command looks like this:

```
LOG=1 \
TEST_CLIENTCOMMIT=42b868252e09f49837802d3123c8c8cce2dbe630 \ # can also be TEST_CLIENTCOMMIT=development-branch
TEST_PKS_ETHEREUM=b7fdfefea39820f2ae25f0b47c1143e197d87ac3a1cb25c304603abcbe0834e9 \
BUILD=1 \
make integration_test
```

### How to Run a Provider or Mix Node

A provider node is essentially the same as a mix node just that it has more capabilities. Specifically it can provide services or capabilities in the form of [plugins](https://github.com/katzenpost/docs/blob/master/handbook/mix_server.rst#external-kaetzchen-plugin-configuration). It also acts as the edge nodes of a mixnet in which traffic either enters or leaves.

All of our infrastructure uses docker setups (but docker is not neccesary if you want to use Meson without docker). You will first need to generate a provider config and its PKI keys. The easiest way to do that is by using our [genconfig](https://github.com/hashcloak/genconfig/#genconfig) script:

```bash
go get github.com/hashcloak/genconfig
genconfig \
  -a 157.245.41.154 \ # Current ip address of authority
  -authID qVhmF/rOHVbHwhHBP6oOOP7fE9oPg4IuEoxac+RaCHk= \ # Current public key of authority
  -name provider-name \ # Your provider name
  -ipv4 1.1.1.1 \ # Your public ipv4 address, needs to be reachable from the authority
  -provider # Flag to indicate you only want a provider config
```

This will make a directory called `output/provider-name` with a file called `identity.public.pem`. Send us your public key to our email [info@hashcloak.com](info@hashcloak.com). We will then help you in getting your node added to the mixnet (look at the [warning](#Usage)). Once you give us your public key you can get your node running with:

```bash
docker service create \
  --name meson -d \
  -p 30001:30001 \ # Mixnet port
  -p 40001:40001 \ # User registration port
  --mount type=bind,source=`pwd`/output/provider-name,destination=/conf \
  hashcloak/meson:master
```

__Note:__ You will have to wait for about 20 minutes before your node is being used in the mixnet. It has to wait for a [new epoch](https://hashcloak.com/Meson/docs/#waiting-for-katzenpost-epoch).

To run a mix node please take a look at the [docs](https://hashcloak.com/Meson/docs/#how-to-run-a-mix-node).

### How to send transactions

__⚠️ WARNING ⚠️__: The mixnet is not ready for strong anonymity since it is still being worked on. The privacy features are not ready for production use. There is currently support for both `Goerli` and `Rinkeby` testnets.

You can easily add support for other Ethereum based [chains](https://hashcloak.com/Meson/docs/#other-blockchains) by correctly configuring a provider node and the plugin. The following example shows you how to use our wallet demo to send an Ethereum transaction on the `Goerli` testnet. It uses our [Meson-client](https://github.com/hashcloak/Meson-client) library to connect to mixnet and send the transaction through the mixnet. This demo is mostly to show that the mixnet is working properly.

```bash
git clone https://github.com/hashcloak/Meson-wallet-demo
cd Meson-wallet-demo

RAW_TXN=0xf8640284540be40083030d409400b1c66f34d680cb8bf82c64dcc1f39be5d6e77501802ca0c434f4d4b894b7cce2d880c250f7a67e4ef64cf0a921e3e4859219dff7b086fda0375a6195e221be77afda1d7c9e7d91bf39845065e9c56f7b5154e077a1ef8a77
go run ./cmd/wallet/main.go \
  -t gor \ # gor is the ethereum chain ticker for the goerli testnet
  -s gor \ # Meson service name
  -rt $RAW_TXN \ # Signed raw transaction blob
  -chain 5 \ # ChainID that crosschecks with Meson
  -c client.toml \ # Config file
```
The contents of `client.toml` are:

```toml
[Logging]
  Disable = false
  Level = "DEBUG"
  File = ""
[UpstreamProxy]
  Type = "none"
[Debug]
  DisableDecoyTraffic = true # Not a safe value for privacy
  CaseSensitiveUserIdentifiers = false
  PollingInterval = 1
[NonvotingAuthority]
    Address = "157.245.41.154:30000"
    PublicKey = "qVhmF/rOHVbHwhHBP6oOOP7fE9oPg4IuEoxac+RaCHk="
```

## Docs
These are the basics for joining the Meson mixnet as a provider, authority or mix. For more in-depth documentation on the Meson project, visit our [project site](https://hashcloak.com/Meson).

## Donations
If you want to support the development of the Meson mixnet project, you can send us some ETH or any token you like at meson.eth or you can donate directly on [our gitcoin grants page](https://gitcoin.co/grants/290/meson?tab=description)
