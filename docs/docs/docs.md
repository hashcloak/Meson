# Docs
This is the documentation related to the Meson mixnet project. Here, you can find out how to deploy a provider, authority or mix node and learn how to use our client libraries.

## Running Meson

##### __⚠️ WARNING ⚠️ These instructions for joining or running a mixnet are only for the current alpha version of a Katzenpost mixnet. The alpha version is not ready for production usage.__

Requirements:

- `go` version 1.13
- `docker swarm`

You will need to have [initialized](https://docs.docker.com/engine/reference/commandline/swarm_init) your docker swarm.

### How to Run a Provider Node

All of our infrastructure uses docker to run the mixnet nodes. You will first need to generate a provider config and its PKI keys. The easiest way to do that is by using our [genconfig](https://github.com/hashcloak/genconfig/#genconfig) script:

```bash
go get github.com/hashcloak/genconfig
genconfig \
  -a 138.197.57.19 \ # Current ip address of authority
  -authID RJWGWCjof2GLLhekd6KsvN+LvHq9sxgcpra/J59/X8A= \ # Current public key of authority
  -name provider-name \ # Your provider name
  -ipv4 1.1.1.1 \ # Your public ipv4 address
  -provider # Flag to indicate you only want a provider config
```

This will make a directory called `output/provider-name` with a file called `identity.public.pem`. Send us your public key to [info@hashcloak.com](info@hashcloak.com). We will then get your node added to the mixnet. Once you give is your public key you can get your node running with:

```bash
docker service create \
  --name meson -d \
  -p 30001:30001 \ # Mixnet port
  -p 40001:40001 \ # User registration port
  --mount type=bind,source=`pwd`/output/provider-name,destination=/conf \
  hashcloak/meson:master
```

It is important that the IPv4 address you use is reachable by the authority node. To look at the logs see [Logs](#log-files)

#### Currency Service for the Provider 

A service is the capability of a Katzenpost plugin. Each Katzenpost plugin has a capability that is advertised to the mixnet during each epoch. In `katzentpost.toml` of the provider node, there is a section called `CBORPluginsKaetzchen`. This is where the different services can be configured.

```toml
# katzenpost.toml 
# This is not a complete configuration file for a katzenpost server. 
# Please look at the output of the genconfig tool to see what a complete 
# katzenpost.toml looks like
[Provider]
  [[Provider.CBORPluginKaetzchen]]
    Capability = "gor" # The service advertised by the provider
    Endpoint = "+gor" # The API endpoint path where clients connect to be forward to the plugin.
    Command = "/go/bin/Meson" # The plugin executable path
    MaxConcurrency = 1 # Amount of plugin programs to spawn
    Disable = false # Disables the plugin if true
    [Provider.CBORPluginKaetzchen.Config]
      f = "/conf/currency.toml" # Configuration file for Meson
      log_dir = "/conf" # Log directory.
      log_level = "DEBUG" # Log level
```

The Meson plugin defined above is handling the Ethereum chain of `Goerli` in this configuration. This is what the [wallet demo](#sending-transactions) application uses as the `-t` and `-s` flags.

The Meson plugin uses an additional configuration file to be able to connect to the rpc endpoint of a blockchain node. This file is called `/conf/currency.toml` in the `Provider.CBORPluginKaetzchen.Config` section of `katzenpost.toml` and it has the following parameters:

```toml
# currency.toml
Ticker = "gor"
ChainID = 5
RPCUser = "rpcuser"
RPCPass = "rpcpassword"
RPCURL = "https://goerli.hashcloak.com"
LogDir = "/conf"
LogLevel = "DEBUG"
```

The `ticker` parameter has to match the `Capability` and `Endpoint` parameters of `Provider.CBORPluginKaetchen` in `katzenpost.toml`.

__Note__ that to maximize the users it is best if the rpc endpoint in the `currency.toml` file is a blockchain node that you control.

### How to Run a Mix Node

To run a mix node we have to run the same command to generate the config file. The only difference is changing the `-provider` flag with `-node`.

```bash
genconfig \
  -a 138.197.57.19 \ # Current ip address of authority
  -authID RJWGWCjof2GLLhekd6KsvN+LvHq9sxgcpra/J59/X8A= \ # Current public key of authority
  -name mix-node-name \ # Your provider name
  -ipv4 1.1.1.1 \ # Your public ipv4 address
  -node # Flag to indicate you only want a mix node config
```

This will make a directory called `output/mix-node-name` with a file called `identity.public.pem`. Send us your public key to [info@hashcloak.com](info@hashcloak.com). We will then help you to get added as a mix. Once you give is your public key you can get your node running with:

```bash
docker service create \
  --name meson-mix -d \
  -p 30001:30001 \
  --mount type=bind,source=`pwd`/output/mix-node-name,destination=/conf \
  hashcloak/meson:master
```

__Notice__ that the ports that docker exposes are the same as the provider node container. If the container is running on the same host then you will need to change the port number. To change the port value you need to edit the following file `output/mix-node-name/katzenpost.toml` and change the ports numbers under the `[Server]` section:

```toml
# output/mix-node-name/katzenpost.toml
[Server]
  Identifier = "mix-node-name"
  Addresses = ["0.0.0.0:30002"] # <- Here from 30001 to 300002
  OnlyAdvertiseAltAddresses = true
  DataDir = "/conf"
  IsProvider = false
  [Server.AltAddresses]
    tcp4 = ["1.1.1.1:30002"] # <- Here from 30001 to 300002
```

After changing the port numbers you can run the docker service command with `-p 30002:30002`.

### How to Run an Nonvoting Authority

Only one non voting authority is needed per nonvoting mixnet. Once you have a valid `authority.toml` file you can use the following docker command to run a mixnet. Take note of the docker tag at the end.

```
docker service create --name authority -d \
  -p 30000:30000 \
  --mount type=bind,source=$HOME/configs/nonvoting,destination=/conf \
  hashcloak/katzenpost-auth:1c00188
```

#### Updating Authority Config

When a node wants to join the non voting mixnet it needs to get added to the `authority.toml`:

```toml
# authority.toml
...
[[Mixes]]
  Identifier = ""
  IdentityKey = "RVAjV/p1azndjGUjuyOUq2p5X46tva2DmXJhGo84DUk=" # Mix public key

[[Providers]]
  Identifier = "provider-name" # The name of your provider. It needs ot be the same.   
  IdentityKey = "92gxXY/Y8BaCWoDMCFERWGxQBMensH9v/kVLLwBFFg8=" # Provider public key
...
```

__Note__ that the `Identifier` value for mixes needs to be an empty string because of [this](https://github.com/katzenpost/authority/blob/master/nonvoting/server/config/config.go#L299-#L304).

Once the new keys are added to `authority.toml`, you need to restart your authority by running `docker service rm authority` and restarting the docker service of the authority.

## Sending Transactions

Currently, the way we send transactions is by using our wallet demo [application](https://github.com/hashcloak/Meson-wallet-demo).

```bash
git clone https://github.com/hascloak/Meson-wallet-demo
cd Meson-wallet-demo
go run ./cmd/wallet/main.go \
  -t rin \ # rin is the ethereum chain identifier for the rinkeby testnet
  -s rin \ # Meson service name
  -pk 0x9e23c88a0ef6745f55d92eb63f063d2e267f07222bfa5cb9efb0cfc698198997 \ # Private key 
  -c client.toml \ # Config file
  -chain 4 \ # Chain id for rinkeby. Needed only when using a private key
  -rpc https://rinkeby.hashcloak.com # An rpc endpoint to obtain the latest nonce count and gas price. Only necessary when using a private key.
```

Another way of sending a transaction with our wallet is by using a presigned raw transaction. Like this:

```bash
RAW_TXN=0xf8640284540be40083030d409400b1c66f34d680cb8bf82c64dcc1f39be5d6e77501802ca0c434f4d4b894b7cce2d880c250f7a67e4ef64cf0a921e3e4859219dff7b086fda0375a6195e221be77afda1d7c9e7d91bf39845065e9c56f7b5154e077a1ef8a77
go run ./cmd/wallet/main.go \
  -t gor \ # gor is the ethereum chain ticker for the goerli testnet
  -s gor \ # Meson service name
  -rt $RAW_TXN \ # Signed raw transaction blob
  -c client.toml \ # Config file
```

The contents of `client.toml` are:

```toml
#client.toml
[Logging]
  Disable = false # Enables logging
  Level = "DEBUG" # Log level. Possible values are: ERROR, WARNING, NOTICE, INFO, DEBUG
  File = "" # No file name output logs to stdout

[UpstreamProxy]
  Type = "none" # Proxy to connect to before connecting to mixnet

[Debug]
  DisableDecoyTraffic = true # Disables the decoy traffic of the mixnet
  CaseSensitiveUserIdentifiers = true # Checks for correct capitalization of provider identifiers
  PollingInterval = 1 # Interval in seconds that will be used to poll the receive queue.

[NonvotingAuthority]
    Address = "138.197.57.19:30000" # The address of the authority
    PublicKey = "RJWGWCjof2GLLhekd6KsvN+LvHq9sxgcpra/J59/X8A=" # Public key of the authority
```

## Log files

Because of the way our docker services are being created, all of the log files are saved to the mounted docker volume, thus all of the log files will be located in the mount the directory of the docker host. If the docker volume is mounted `$HOME/configs/nonvoting` the logs of the authority will be saved in that directory. The same goes for all of the nodes.

If you are running a full mixnet this little command will be useful for looking at all of the logs of the mixnet. If you are running a single node then this command also works for you.

```
find ./configs -name "*.log" | xargs tail -f
```

## Waiting for Katzenpost Epoch

Due to how katzenpost is designed, when you join the mixnet you will have to wait for a new epoch to publish your node descriptor. An epoch right now is 10 minutes. While you wait for a new epoch you will see this message appear in the log files o your provider or mix node:

```
01:40:35.977 WARN pki: Authority rejected upload for epoch: 138107 (Conflict/Late)
01:40:35.977 WARN pki: Failed to post to PKI: pki: post for epoch will never succeeed
```

The above log occurs every time your node tries to post a new epoch description to the authority. In the authority's logs you will see this:

```
18:36:31.688 ERRO authority: Peer 10.0.0.2:57660: Rejected probably a conflict: state: Node oRh8boMS6VzJW57m5lMNqfK8EZ+LYfkfV0eJXKAJcJc=: Late descriptor upload for for epoch 138207
```

Once you your node has successfully published its descriptor to the authority you will get a message that starts with this:

```
01:51:37.821 DEBU pki/nonvoting/client: Document: &{Epoch:138108...
```

## Other Blockchains

We intend to add support for other chains but, for now, only Ethereum based transactions are supported. We are currently only running `Goerli` and `Rinkeby` testnets but you can run a provider with access to an rpc node of any Ethereum compatible chain such as `ETC` or `Mordor`. If you want help setting up a provider for another chain please get in contact with us at info@hashcloak.com!

The steps needed to add a new Ethereum based chain are:

- Obtain access to an rpc endpoint of the new chain
- Change the `Provider.CBORPluginKaetzchen` to use the new service ticker
- Configure `currency.toml` with the new service.

After updating those configuration files running a provider node should follow the same steps as detailed [above](#running-meson).