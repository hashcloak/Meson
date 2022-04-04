# katzenmint-pki
[![Go](https://github.com/hashcloak/Meson/actions/workflows/go.yml/badge.svg)](https://github.com/hashcloak/Meson/actions/workflows/go.yml)

A BFT PKI for the Katzenpost Authority PKI System using Tendermint

## Overview 

![High-level overview of the architecture](https://github.com/hashcloak/Meson/katzenmint/blob/master/high-level%20katzenmint.png)

## Develop

There is `setup.sh` to help you out setting the develop environment, you can run `make setup` or `sh setup.sh` to start.

Or you can follow these steps:

1. [Install tendermint v0.34.6](https://docs.tendermint.com/master/introduction/install.html)
```BASH
$ git clone https://github.com/tendermint/tendermint.git
$ cd tendermint
$ git checkout v0.34.6
$ make install
```
2. `TMHOME=`pwd`/chain tendermint init`
3. `make build`
4. `./katzenmint -config ./chain/config/config.toml`
5. `curl -s 'localhost:26657/broadcast_tx_commit?tx="tendermint=rocks"'`
6. `curl -s 'localhost:26657/abci_query?data="tendermint"'`
