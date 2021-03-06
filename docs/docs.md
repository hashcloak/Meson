# Docs
This is the documentation related to the Meson mixnet project. Here, you can find out how to deploy a provider, authority or mix node and learn how to use our client libraries.

## Running Meson

##### __⚠️ WARNING ⚠️ These instructions for joining or running a mixnet are only for the current alpha version of a Katzenpost mixnet. The alpha version is not ready for production usage and relies on manual configuration of the PKI.__

Requirements:

- `go` version >= __1.13__
- `docker swarm`. You will need to have [initialized](https://docs.docker.com/engine/reference/commandline/swarm_init) your docker swarm.
- `python` version >= __3.5__
- `make`

### How to Run a Provider Node

All of our infrastructure uses docker to run the mixnet nodes. You will first need to generate a provider config and its PKI keys. The easiest way to do that is by using our [genconfig](https://github.com/hashcloak/genconfig/#genconfig) script:

```bash
go get github.com/hashcloak/genconfig
genconfig \
  -a 157.245.41.154 \ # Current ip address of authority
  -authID qVhmF/rOHVbHwhHBP6oOOP7fE9oPg4IuEoxac+RaCHk= \ # Current public key of authority
  -name provider-name \ # Your provider name
  -ipv4 1.1.1.1 \ # Your public ipv4 address
  -provider # Flag to indicate you only want a provider config
```

This will make a directory called `output/provider-name` with a file called `identity.public.pem`. Send us your public key to [info@hashcloak.com](info@hashcloak.com). We will then get your node added to the mixnet (this is not a decentralized step, please look at the [warning](#running-meson) at the top of the page). Once you give us your public key you can get your node running with:

```bash
docker service create \
  --name meson -d \
  -p 30001:30001 \ # Mixnet port
  -p 40001:40001 \ # User registration port
  --mount type=bind,source=`pwd`/output/provider-name,destination=/conf \
  hashcloak/meson:master
```

It is important that the IPv4 address you use is reachable by the authority node. To look at the logs see [Logs](#log-files). Also be aware that your provider will have to wait for a [new epoch](#waiting-for-katzenpost-epoch).

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

The Meson plugin defined above is handling a serviced called `gor` which Meson sends to the Ethereum chain `Goerli`. This is what the [wallet demo](#sending-transactions) application uses as the `-t` and `-s` flags.

The Meson plugin uses an additional configuration file to be able to connect to the RPC endpoint of a blockchain node. This file is called `/conf/currency.toml` in the `Provider.CBORPluginKaetzchen.Config` section of `katzenpost.toml` and it has the following parameters:

```toml
# currency.toml is the configuration of Meson
Ticker = "gor" # This is the name of service provided by Meson
RPCUser = "rpcuser" # HTTP login for the blockchain node.
RPCPass = "rpcpassword" # HTTP password
RPCURL = "https://goerli.hashcloak.com" # The RPC url of the node that will receive the transaction
LogDir = "/conf" # Location of the logs
LogLevel = "DEBUG" # Log level of the Meson plugin.
```

The `Ticker` parameter has to match the `Capability` and `Endpoint` parameters of `Provider.CBORPluginKaetchen` in `katzenpost.toml`.

__Note__ that to maximize the privacy of the mixnet users it is best if the RPC endpoint in the `currency.toml` file is a blockchain node that you control.

#### Check that Meson is running

Due to how plugin processes are spawned from the main katzenpost server program, the Docker container does not have information on the exit status of the plugins inside of the container. This leads to a situation in which the container is running but the Meson plugin is not.

One can check if this is the case by running the following command:

```
$ docker exec nonvoting_testnet_provider1_1 top # nonvoting_testnet_provider1_1 is the name of the container

Mem: 13169904K used, 1050596K free, 581716K shrd, 969380K buff, 6457840K cached
CPU:   2% usr   1% sys   0% nic  96% idle   0% io   0% irq   0% sirq
Load average: 0.31 0.66 0.80 2/1636 111
  PID  PPID USER     STAT   VSZ %VSZ CPU %CPU COMMAND
    1     0 root     S     314m   2%   2   0% /go/bin/server -f /conf/katzenpost
   29     1 root     S     111m   1%   6   0% /go/bin/memspool -data_store /conf
   22     1 root     S     111m   1%   7   0% /go/bin/panda_server -log_dir /con
   16     1 root     S     109m   1%   2   0% /go/bin/echo_server -log_level DEB
   58     0 root     S     1572   0%   2   0% top
   36     1 root     Z        0   0%   5   0% [Meson]
```

This shows Meson at the bottom with no cpu nor memory allocated to it. This is an indicator that Meson exited with and error and you can find the error in the currency log file:

```log
#currency.36.log
INFO 001 currency server started
ERRO 002 Failed to load config file '/conf/currency.toml: config: RPCUrl is not set
ERRO 003 Exiting
```

In this case, the solution is self explanatory. We just need to add the `RPCurl` value to `currency.toml`.

### How to Run a Mix Node

To run a mix node we have to run `genconfig` to generate the config file. The only difference is changing the `-provider` flag with `-node`.

```bash
genconfig \
  -a 157.245.41.154 \ # Current ip address of authority
  -authID qVhmF/rOHVbHwhHBP6oOOP7fE9oPg4IuEoxac+RaCHk= \ # Current public key of authority
  -name mix-node-name \ # Your provider name
  -ipv4 1.1.1.1 \ # Your public ipv4 address
  -node # Flag to indicate you only want a mix node config
```

This will make a directory called `output/mix-node-name` with a file called `identity.public.pem`. Send us your public key to [info@hashcloak.com](info@hashcloak.com). We will then help you to get added as a mix (please look at the [warning](#running-meson) at the top of this page).

```bash
docker service create \
  --name meson-mix -d \
  -p 30001:30001 \
  --mount type=bind,source=`pwd`/output/mix-node-name,destination=/conf \
  hashcloak/meson:master
```

__Notice__ that the ports that docker exposes are the same as the provider node instructions from above. If the container is running on the same host then you will need to change the port number. To change the port value you need to edit the following file `output/mix-node-name/katzenpost.toml` and change the ports numbers under the `[Server]` section:

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

After changing the port numbers you can run the docker service command with `-p 30002:30002`. Also be aware that your provider or node will have to wait for a [new epoch](#waiting-for-katzenpost-epoch).


### How to Run an Nonvoting Authority

Only one nonvoting authority is needed per nonvoting mixnet. Once you have a valid `authority.toml` file you can use the following docker command to run a mixnet. Look at the output of `genconfig` or at [katznepost docs](https://github.com/katzenpost/docs/blob/master/handbook/nonvoting_pki.rst#configuring-the-non-voting-directory-authority) for more information on how to create the configuration of the authority.

```
docker service create --name authority -d \
  -p 30000:30000 \
  --mount type=bind,source=$HOME/configs/nonvoting,destination=/conf \
  hashcloak/authority:master
``` 

Hashcloak is maintaining a [docker container](https://hub.docker.com/repository/docker/hashcloak/katzenpost-auth) of [katzenpost/authority](https://github.com/katzenpost/authority).

#### Updating Authority Config

When a node wants to join the non voting mixnet it needs to get added to the `authority.toml`:

```toml
# authority.toml
...
[[Mixes]]
  Identifier = ""
  IdentityKey = "RVAjV/p1azndjGUjuyOUq2p5X46tva2DmXJhGo84DUk=" # Mix public key

[[Providers]]
  Identifier = "provider-name" # The name of the newly added provider
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
  -chain 4 \ # Chain id for rinkeby. only necessary when using a private key
  -rpc https://rinkeby.hashcloak.com # An rpc endpoint to obtain the latest nonce count and gas price. Only necessary when using a private key.
```

Another way of sending a transaction with our wallet is by using a presigned raw transaction. Like this:

```bash
RAW_TXN=0xf8640284540be40083030d409400b1c66f34d680cb8bf82c64dcc1f39be5d6e77501802ca0c434f4d4b894b7cce2d880c250f7a67e4ef64cf0a921e3e4859219dff7b086fda0375a6195e221be77afda1d7c9e7d91bf39845065e9c56f7b5154e077a1ef8a77
go run ./cmd/wallet/main.go \
  -t gor \ # gor is the ethereum chain ticker for the goerli testnet
  -s gor \ # Meson service name
  -rt $RAW_TXN \ # Signed raw transaction blob
  -chain 5 \ # ChainID that cross checks with Meson
  -c client.toml \ # Config file
```

The contents of `client.toml` are:

```toml
#client.toml
[Logging]
  Disable = false # Enables logging
  Level = "DEBUG" # Log level. Possible values are: ERROR, WARNING, NOTICE, INFO, DEBUG
  File = "" # No file name means the logs are displayed in stdout

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

After the transaction is successfully received by Meson at the egress provider node you will see this message in the logs: 

```
reply: {"Message":"success","StatusCode":0,"Version":0}
```

This doesn't mean that the transaction was accepted by the blockchain node that Meson is using. It just means that Meson successfully forwarded the transaction to the RPC endpoint. 


## Log files

Because of the way our docker services are being created, all of the log files are saved to the mounted docker volume, thus all of the log files will be located in the mount the directory of the docker host. If the docker volume is mounted `$HOME/configs/nonvoting` the logs of the authority will be saved in that directory. The same goes for all of the nodes.

If you are running a full mixnet or a single node this command might be useful for you:

```
find ./configs -name "*.log" | xargs tail -f
```

## Environment variables

This is a list of environment variables that is mostly used in the `ops/` directory.

- `BUILD`: Forces the build of the containers instead of pulling them from docker hub. Default: off. Enable with `BUILD=1`
- `LOG`: Enables logging. Default off. Enable with `LOG=1`
- `DOCKER_BUILDKIT`: Enables additional logs during the docker build steps. Enable with `DOCKER_BUILDKIT=1`.
- `REPOS_AUTH_BRANCH`: The branch to use for the authority container. Can also be a commit hash. Default: `master`.
- `REPOS_AUTH_CONTAINER`: The docker repository for the authority container. Default: `hashcloak/authority`.
- `REPOS_AUTH_GITHASH`: The git commit to use for building authority.
- `REPOS_AUTH_HASHTAG`: The docker tag to use for container. Default is the value of `REPOS_AUTH_GITHASH`.
- `REPOS_AUTH_NAMEDTAG`: The docker tag of the authority container. Default is the name of the branch.
- `REPOS_AUTH_REPOSITORY`: The repository from which to build the authority from. Default: `github.com/katzenpost/authority`
- `REPOS_SERVER_BRANCH`: The branch to use for the server container. Can also be a commit hash. Default: `master`.
- `REPOS_SERVER_CONTAINER`: The repository of the authority container. Default: `hashcloak/server`.
- `REPOS_SERVER_GITHASH`: The git commit to use for building server.
- `REPOS_SERVER_HASHTAG`: The docker tag to use for container. Default is the value of `REPOS_SERVER_GITHASH`.
- `REPOS_SERVER_NAMEDTAG`: The docker tag of the server container. Default is the name of the branch.
- `REPOS_SERVER_REPOSITORY`: The repository from which to build server from. Default: `github.com/katzenpost/server`
- `REPOS_MESON_BRANCH`: The branch to use for the meson container. Can also be a commit hash. Default: The current branch of the repository.
- `REPOS_MESON_CONTAINER`: The docker repository for the meson container. Default: `hashcloak/meson`.
- `REPOS_MESON_GITHASH`: The git commit to use for building meson. Defaults to the latest commit of the current working branch.
- `REPOS_MESON_HASHTAG`: The docker tag to use for meson container. Default is the value of `REPOS_MESON_GITHASH`.
- `REPOS_MESON_NAMEDTAG`: The docker tag of the container meson. Default is the name of the branch.
- `TEST_ATTEMPTS`: The amount of retries for the integration tests until a transaction is found. Default: `3`.
- `TEST_CLIENTCOMMIT`: The commit to use for the integration tests. Default: `master`.
- `TEST_NODES`: The amount of mix nodes to spawn. Default: `2`.
- `TEST_PROVIDERS`: The amount of provider nodes to spawn. Default: `2`.
- `TEST_PKS_BINANCE`: The private key to use for the binance tests.
- `TEST_PKS_ETHEREUM`: The private key to use for the ethereum tests.
- `WARPED`: This flag is turned on by default in any non `master` branch or when `WARPED=false` is used as an environment variable. This flag will also add the `warped_` suffix to all the container tags. For example: `hashcloak/server:warped_51881a5`. `WARPED` also builds the container with a warped build flag. This means that the epoch times for the mixnet are down from 20 minutes to 2 minutes.

## Waiting for Katzenpost Epoch

Due to how katzenpost is designed, when you join the mixnet you will have to wait for a new epoch to publish your node descriptor. An epoch right now is 10 minutes. While you wait for a new epoch you will see this message appear in the log files of your node.

```
01:40:35.977 WARN pki: Authority rejected upload for epoch: 138107 (Conflict/Late)
01:40:35.977 WARN pki: Failed to post to PKI: pki: post for epoch will never succeed
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

We intend to add support for other chains but, for now, only Ethereum based transactions are supported. We are currently only running `Goerli` and `Rinkeby` testnets but you can run a provider with access to an RPC node of any Ethereum compatible chain such as `ETC` or `Mordor`. If you want help setting up a provider for another chain please get in contact with us at info@hashcloak.com!

The steps needed to add a new Ethereum based chain are:

- Obtain access to an RPC endpoint of the new chain.
- Change the `Provider.CBORPluginKaetzchen` to use the new service ticker.
- Configure `currency.toml` with the new ticker.

After updating those configuration files, running a provider node should follow the same steps as detailed [above](#running-meson).
