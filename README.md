# Meson
This is the main repository related to the Meson project. 
Meson is a mixnet for cryptocurrency transactions. Meson is based on the [Katzenpost software project](https://katzenpost.mixnetworks.org/).

## Docs
These are the basics for joining the Meson mixnet as a provider, authority or mix. For more in-depth documentation on the Meson project, visit out [project site](hashcloak.com/Meson).

### How to run a provider
TODO

### How to run a mix
TODO

### How to join the Authority PKI
TODO

### How to send transactions

__⚠️ WARNING ⚠️__: The mixnet is not ready for strong anonymity since it is still being worked on. The privacy features are not ready for production use.  There is currently is support for both goerli and rinkeby.

Requirements:
- `go` version 1.13
- a private key that has goerli or rinkeby balance __OR__ a signed raw transaction blob

```bash
# Clone the github repo with the demo wallet
$ git clone https://github.com/hashcloak/Meson-wallet-demo
$ cd Meson-wallet-demo


# Send a rinkeby transaction using a private key
$ go run ./cmd/wallet/main.go \
  -t rin \ # rin is the ethereum chain identifier for the rinkeby testnet
  -s rin \ # the Meson service name
  -pk 0x9e23c88a0ef6745f55d92eb63f063d2e267f07222bfa5cb9efb0cfc698198997 \ # the private key 
  -c client.toml \ # the config file
  -chain 4 \ # Chain id for rinkeby
  -rpc https://rinkeby.hashcloak.com # An rpc endpoint to obtain the latest nonce count and gas price. Only necesary when using a private key.


# Send a goerli transaction using a transaction blob
RAW_TXN=0xf8640284540be40083030d409400b1c66f34d680cb8bf82c64dcc1f39be5d6e77501802ca0c434f4d4b894b7cce2d880c250f7a67e4ef64cf0a921e3e4859219dff7b086fda0375a6195e221be77afda1d7c9e7d91bf39845065e9c56f7b5154e077a1ef8a77
$ go run ./cmd/wallet/main.go \
  -t gor \ # gor is the ethereum chain identifier for the goerli testnet
  -s gor \ # the Meson service name
  -rt $RAW_TXN \ # The raw transaction blob
  -c client.toml \ # the config file
  -chain 5 \ # Chain id for goerli
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
  DisableDecoyTraffic = true # Not a safe parameter for privacy
  CaseSensitiveUserIdentifiers = false
  PollingInterval = 1
[NonvotingAuthority]
    Address = "134.209.46.0:30000"
    PublicKey = "3mAR/JpJzSqHKm0gcupG00gbT+kB52wrckA6i+sjXy8="
```

## Donations
If you want to support the development of the Meson mixnet project, you can send us some ETH or any token you like at meson.eth.
