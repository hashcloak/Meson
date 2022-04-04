## Overview

We propose a mixnet PKI system using Tendermint as its consensus engine. Tendermint is a general purpose, state machine replication system. Validators in the Tendermint system vote on blocks of transactions and upon receiving enough votes, the block gets commits to a hash chain of blocks, known as a blockchain.

In this system, authorities act as Tendermint validators and handle the chain's consensus. In addition to their responsibilities as validators, they still carry out their responsibilities as outlined in (insert link to katzenpost pki spec).

Valid consensus documents, mix descriptors and authority set changes are different transactions types that are batched into blocks and voted upon by authorities. 

Mix nodes and clients use the Tendermint light client system to retrieve information from the blockchain without having the responsilities of a full, Tendermint node. This reduces the communication mix nodes and clients need to do with validators. Mix nodes and clients' responsibilities remain the same as outlined in the katzenpost PKI spec.

Providers responsibilities' are reduced in this PKI system. They no longer need to cache consensus documents for clients to fetch. Instead, they can (perhaps MAY is a better term here) serve as full nodes for the overall availability of the Tendermint blockchain.

## Description

### Security Goals
The security goals of this Directory Authority system remain the same with the addition of the following goals and features:

- Byzantine-Fault tolerance: It allows for consensus faults between the directory authorities. Further, it is possible to find badly behaving operators in the system.
- The Directory Authority servers form a peer to peer gossip amongst themselves.

### Transaction Format
We format the Tendermint transaction as
```
type Transaction struct {
	// version
	Version string

	// Epoch
	Epoch uint64

	// command
	Command Command

	// hex encoded ed25519 public key (should not be 0x prefxied)
	PublicKey string

	// hex encoded ed25519 signature (should not be 0x prefixed)
	Signature string

	// json encoded payload (eg. mix descriptor/authority)
	Payload string
}
```

We will be considering the following commands:
- `PublishMixDescriptor`: A command to publish mix descriptors by mix nodes. 
    - Payload is the hex representation of the json-encoded form of the `MixDescriptor` Go struct.
	```
	type MixDescriptor struct {
		// Name is the human readable (descriptive) node identifier.
		Name string
	
		// IdentityKey is the node's identity (signing) key.
		IdentityKey *eddsa.PublicKey
	
		// LinkKey is the node's wire protocol public key.
		LinkKey *ecdh.PublicKey
	
		// MixKeys is a map of epochs to Sphinx keys.
		MixKeys map[uint64]*ecdh.PublicKey
	
		// Addresses is the map of transport to address combinations that can
		// be used to reach the node.
		Addresses map[Transport][]string
	
		// Kaetzchen is the map of provider autoresponder agents by capability
		// to parameters.
		Kaetzchen map[string]map[string]interface{} `json:",omitempty"`
	
		// RegistrationHTTPAddresses is a slice of HTTP URLs used for Provider
		// user registration. Providers of course may choose to set this to nil.
		RegistrationHTTPAddresses []string
	
		// Layer is the topology layer.
		Layer uint8
	
		// LoadWeight is the node's load balancing weight (unused).
		LoadWeight uint8
	}
	```

- `AddNewAuthority`: A command to add a new authority node to the PKI System.
    - Payload is the json-encoded form of the `Authority` Go struct.
    ```
	type Authority struct {
		// Auth is the prefix of the authority.
		Auth string

		// PubKey is the validator's public key.
		PubKey []byte

		// KeyType is the validator's key type.
		KeyType string

		// Power is the voting power of the authority.
		Power int64
	}
	```

The consensus document will be generated using all received mix descriptors at the end of each epoch.
```
type Document struct {
	Version           string
	Epoch             uint64
	GenesisEpoch      uint64
	SendRatePerMinute uint64

	Mu              float64
	MuMaxDelay      uint64
	LambdaP         float64
	LambdaPMaxDelay uint64
	LambdaL         float64
	LambdaLMaxDelay uint64
	LambdaD         float64
	LambdaDMaxDelay uint64
	LambdaM         float64
	LambdaMMaxDelay uint64

	Topology  [][][]byte
	Providers [][]byte
}
```

### Configuration

#### Initialization
In order to define the behavior of this chain at startup, one needs to define the following parameters. 

##### Parameters to set in the `genesis.json` file
- `genesis_time`: Time the blockchain starts. For our pursposes, this can be the time the mixnet starts.
- `chain_id`: ID of the blockchain. Effectively, this can change for major changes made to the blockchain. Can be used to delineate different versions of the chain.
- `consensus_params`:
    - `block`:
        - `time_iota_ms`: Minimum time increments between consecutive blocks. For our purposes, this will be time between epoch. The current katzenpost spec has this set to 3 hrs. However, the current implementation is set to 30 mins. TODO: Finalize epoch duration length
- `validators`: List of initial validators (authorities). We can set this at genesis or initialize it when we deploy the Tendermint-based PKI Authority.
    - `pub_key`: Ed25519 public keys where the first byte specifies the kind of key and the rest of the bytes specify the public key. TODO: Ensure that these keys can be converted to a more katzenpost friendly format.
    - `power`: validator's voting power. Initially, we can set this to 1. To remove an authority, set the voting power to 0. TODO: Determine ways to leverage this in the Katzenpost PKI authority.
- `app_hash`: expected application hash. Meant as a way to authenticate the application
- `app_state`: Application state. May not be directly relevant for our purposes as we don't have a token.

For more information about `genesis.json`, see https://github.com/tendermint/tendermint/blob/master/types/genesis.go

##### Parameters to set in the katzenmint system
- `epochInterval` in `state.go`: Number of blocks that consists of an epoch.
- `Layers` in `katzemint.toml`: Number of layers of the mix network.
- `MinNodesPerLayer` in `katzemint.toml`: Minimum number of mix nodes in every layer.

### Differences from current Katzenpost PKI system

The main differences between the current PKI system and this proposed system are:
- Authorities are selected in a round robin fashion to propose blocks as part of the tendermint consensus protocol.
- There is no randomness generation (NOTE: This can be added either through using Core Star's Tendermint fork or having a transaction that outputs the result of the regular shared randomness beacon)
- This tendermint-based authority system favors consistency over availability in a distributed systems sense.
- This protocol tolerates up to a 1/3 of authorities being byzantine.

## Privacy Considerations

- The list of authorities, mix descriptors and consensus documents are publicly posted on a public blockchain. Anyone can look at these transactions.
- Information retrieval using the light client system and transaction broadcasting are not privacy-preserving, by default, in Tendermint.


## Implementation Considerations
- Due to the blockchain structure, we might need to replace BoltDB, that is currently used in the voting PKI, with another DB that is optimized for both reads and writes (RocksDB and BadgerDB are good contenders for this). 
- Mix nodes and clients now need a tendermint light client in order to retrieve the latest view of the chain. 

## Future Considerations
- Incentivization via an external cryptocurrency (e.g. Zcash)
- Slashing penalties for misbehavior
- More permissionless enrollment of authorities
    - A Sybil-resistance mechanism for enrolling authorities
- PIR-like techniques for light clients
- Using [Core Star's Tendermint fork with an embedded BLS random beacon](https://github.com/corestario/tendermint)

## References
TODO: Add references
