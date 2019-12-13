# Roadmap

We have a vision for Meson and that is to be THE mixnet for cryptocurrency transactions. As such, in order to achieve our vision, we need a plan to get there. Here, we outline various things that we need to work on in order to make Meson the best mixnet it can be.

## Adding support for other chains
Currently, Meson supports Ethereum-based transactions and has the ability to support Bitcoin-based transactions. We want to build Katzenpost plugins that can support other chains. On our immediate list are the following chains:

* ETH2.0
* Cosmos
* Polkadot
* Handshake

## Adding support for Layer 2 projects
Scaling is an important issue in the blockchain space, but often times, comes at the expense of privacy. We want to be able to support various L2 scaling schemes so that they, too, can benefit from the various anonymity properties of mix networks. On our immediate list are the following projects:

* Lightning
* Bolt
* Connext
* ZK Sync
* Fuel

## Easy Deployment of Meson Mixnet components
Meson is made up of 3 kinds of nodes: Authorities, Providers and Mixes. These are components that need to be deployed separately. As such, we want to make it easy for anyone to deploy these components and participate in the network. We are working towards easily configurable deploy scripts and integrating into popular "node-in-a-box" providers like DAppNode and Avado.

## Integration into wallets
In order for Meson to be useful, people need to use it. Towards this end, we want to integrate Meson client software into popular wallets and perhaps even build our own. This will enable anyone the ability to send cryptocurrency transactions over the Meson mixnet. 

## Governance
Even though this started as a HashCloak project, this is overall a blockchain community project. Thus, it needs to be governed by various and diverse stakeholders. Governance is a hard problem for all open source projects. We hope we can attempt to build a strong community of people who want to see a production working mix network for cryptocurrency transactions. The eventual goal is have Meson be its own nonprofit structure with accountability and transparency built-in.

## Contribute to Mix Network research
Meson is an experimental project built on experimental software implementing an experimental anonymous communication protocol. Of course, we are going to help make this stuff a little less experimental by contributing to mix network research. On our immediate list, we want to work on the following research problems:

* A byzantine fault tolerant voting mechanism for the PKI Authority.
* Continuous tuning and parameterization of mix network parameters. 

## A Path to Self-sufficiency
We can't rely on donations and grants forever. As such, we want to find a way to self-sustain the development and maintainance of Meson. We will not do an ICO as there is no need for a token. The current ideas we have around this are:

* Payment channels
* Building a privacy-focused wallet

