# Meson
[![Build Status](https://travis-ci.com/hashcloak/Meson.svg?branch=master)](https://travis-ci.com/hashcloak/Meson)

This is the main repository related to the Meson project. 
Meson is a mixnet for cryptocurrency transactions. Meson is based on the [Katzenpost software project](https://katzenpost.mixnetworks.org/).

## Docs
These are the basics for joining the Meson mixnet as a provider, authority or mix. For more in-depth documentation on the Meson project, visit our [project site](https://hashcloak.com/Meson).

## Usage

__⚠️ Meson is still in alpha and for this reason we are using a centralized non voting authority. This is also the reason why we need manual exchange of the public keys of the nodes that get added to the network.__

#### How to Run a Provider or Mix Node

A provider node is essentially the same as a mix node just that it has more capabilities. Specifically it can provide services or capabilities in the form of [plugins](https://github.com/katzenpost/docs/blob/master/handbook/mix_server.rst#external-kaetzchen-plugin-configuration). It also acts as the edge nodes of a mixnet in which traffic either enters or leaves.

All of our infrastructure uses docker setups. You will first need to generate a provider config and its PKI keys. The easiest way to do that is by using our [genconfig](https://github.com/hashcloak/genconfig/#genconfig) script:

```bash
go get github.com/hashcloak/genconfig
genconfig \
  -a 138.197.57.19 \ # Current ip address of authority
  -authID RJWGWCjof2GLLhekd6KsvN+LvHq9sxgcpra/J59/X8A= \ # Current public key of authority
  -name provider-name \ # Your provider name
  -ipv4 1.1.1.1 \ # Your public ipv4 address
  -provider # Flag to indicate you only want a provider config
```

This will make a directory called `output/provider-name` with a file called `identity.public.pem`. Send us your public key to our email [info@hashcloak.com](info@hashcloak.com). We will then help you in getting your node added to the mixnet (look at the [warning](#usage)). Once you give us your public key you can get your node running with:

```bash
docker service create \
  --name meson -d \
  -p 30001:30001 \ # Mixnet port
  -p 40001:40001 \ # User registration port
  --mount type=bind,source=`pwd`/output/provider-name,destination=/conf \
  hashcloak/meson:master
```

__Note:__ You will have to wait for about 10 minutes before your node is being used in the mixnet. It has to wait for a [new epoch](https://hashcloak.com/Meson/docs/#waiting-for-epoch).

To run a mix node please take a look at the [docs](https://hashcloak.com/Meson/docs/#running-meson).

#### How to send transactions

__⚠️ WARNING ⚠️__: The mixnet is not ready for strong anonymity since it is still being worked on. The privacy features are not ready for production use. There is currently support for both `Goerli` and `Rinkeby` testnets.

You can easily add support for other Ethereum based [chains](https://hashcloak.com/Meson/docs/#other-blockchains) by correctly configuring a provider node and the plugin. The following example shows you how to use our wallet demo to send an Ethereum transaction on the `Goerli` testnet. It uses our [Meson-client](https://github.com/hashcloak/Meson-client) library to connect to mixnet and send the transaction through the mixnet. This demo is mostly to show that the mixnet is  working properly.

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
    Address = "138.197.57.19:30000"
    PublicKey = "RJWGWCjof2GLLhekd6KsvN+LvHq9sxgcpra/J59/X8A="
```

## Donations
If you want to support the development of the Meson mixnet project, you can send us some ETH or any token you like at meson.eth or you can donate directly on [our gitcoin grants page](https://gitcoin.co/grants/290/meson?tab=description)
