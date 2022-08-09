# katzenmint-pki
[![Katzenmint](https://github.com/hashcloak/Meson/actions/workflows/katzenmint.yml/badge.svg)](https://github.com/hashcloak/Meson/actions/workflows/katzenmint.yml)

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
2. Initialize the identity credential identity `TMHOME=`pwd`/chain tendermint init`
3. Build katzenmint `make build`
4. Execute `./katzenmint -config ./chain/config/config.toml`
