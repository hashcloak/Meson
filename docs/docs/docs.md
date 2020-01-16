# Docs
This is documentation related to the Meson mixnet project. Here, you can find out how to deploy a provider, authority or mix node and learn how to use our client libraries.


## Running Meson

Requirements:

- `go` version 1.13
- `docker swarm`

You will need to have initialized your docker swarm.

###### __⚠️ WARNING ⚠️__ These instructions for joining or running a mixnet are only for the current alpha version of a Meson mixnet. The alpha version is not ready for production usage.


#### How to Run a Provider Node

All of our infrastructure uses docker setups. You will first need to generate a provider config and its PKI keys. The easiest way to do that is by using our [genconfig](https://github.com/hashcloak/genconfig/#genconfig) script:

```bash
go get github.com/hashcloak/genconfig
genconfig \
  -a 138.197.57.19 \ # current ip address of authority
  -authID RJWGWCjof2GLLhekd6KsvN+LvHq9sxgcpra/J59/X8A= \ # current public key of authority
  -name provider-name \ # your provider name
  -ipv4 1.1.1.1 \ # your public ipv4 address
  -provider # flag to indicate you only want a provider config
```

This will make a directory called `output/provider-name`. Send us your public key at our [Riot.im](https://riot.im/app/#/room/#hashcloak:matrix.org) room. Your public key will be called `output/provider-name/identity.public.key`. We will then help you to get added as a provider. Once you give is your public key you can get your node running with:

```bash
docker service create \
  --name meson -d \
  -p 30001:30001 \ # Mixnet port
  -p 40001:40001 \ # User registration port
  --mount type=bind,source=`pwd`/output/provider-name,destination=/conf \
  hashcloak/meson:master
```

It is important that the ipv4 address you use is reachable by the authority node.

##### Currency Service for the Provider 

In the `katzentpost.toml` fie of the provider node there is a section that has the `CBORPlugins` configuration.

#### How to Run a Mix Node

To run a mix node we have to run the same command to generate the config file. The only difference is changing the `-provider` flag with `-node`.

```bash
genconfig \
  -a 138.197.57.19 \ # current ip address of authority
  -authID RJWGWCjof2GLLhekd6KsvN+LvHq9sxgcpra/J59/X8A= \ # current public key of authority
  -name mix-node-name \ # your provider name
  -ipv4 1.1.1.1 \ # your public ipv4 address
  -node # flag to indicate you only want a mix node config
```

This will make a directory called `output/mix-node-name`. Again, send us your public key at our [Riot.im](https://riot.im/app/#/room/#hashcloak:matrix.org) room. Your public key is located in `output/provider-name/identity.public.key`. We will then help you to get added as a provider. Once you give is your public key you can get your node running with:


```bash
docker service create \
  --name meson-mix -d \
  -p 30001:30001 \
  --mount type=bind,source=`pwd`/output/mix-node-name,destination=/conf \
  hashcloak/meson:master
```

__Notice__ above that the ports above are the same as the provider node. If the container is running on the same host then you will need to change the port number. To change the port value you need to edit the following file `output/mixn-node-name/katzenpost.toml` and change the ports numbers under the `[Server]` section:

```toml
# output/mix-node-name/katzenpost.toml
[Server]
  Identifier = "mix-node-name"
  Addresses = ["0.0.0.0:30002"] # <- Here
  OnlyAdvertiseAltAddresses = true
  DataDir = "/conf"
  IsProvider = false
  [Server.AltAddresses]
    tcp4 = ["1.1.1.1:30002"] # <- Here
```

Then you can run the docker service command with the new port numbers.


#### How to Run an Nonvoting Authority

Only one non voting authority is needed per mixnet. Once you have a valid authority.toml file you can user the following docker command to run a mixnet. Take note of the docker tag at the end.

```
docker service create --name authority -d \
  -p 30000:30000 \
  --mount type=bind,source=$HOME/configs/nonvoting,destination=/conf \
  hashcloak/katzenpost-auth:1c00188
```

##### Updating Authority Config

When a node wants to join the non voting mixnet it needs to get added to the `authority.toml`.

```toml
# authority.toml
...
[[Mixes]]
  Identifier = ""
  IdentityKey = "RVAjV/p1azndjGUjuyOUq2p5X46tva2DmXJhGo84DUk="

[[Providers]]
  Identifier = "provider-name"
  IdentityKey = "92gxXY/Y8BaCWoDMCFERWGxQBMensH9v/kVLLwBFFg8="
...
```

__Note__ that the `Identifier` key for the Mixes is empty. From production usage we recommend you leave this value as an empty string.

Once the new keys are added to `authority.toml` you need to restart your authority by running `docker service rm authority` and restarting the docker service of the authority.

## Sending Transactions

Currently, the way we send transactions is by using our wallet demo [application](https://github.com/hashcloak/Meson-wallet-demo).

```bash
git clone https://github.com/hascloak/Meson-wallet-demo
cd Meson-wallet-demo
go run ./cmd/wallet/main.go \
  -t rin \ # rin is the ethereum chain identifier for the rinkeby testnet
  -s rin \ # the Meson service name
  -pk 0x9e23c88a0ef6745f55d92eb63f063d2e267f07222bfa5cb9efb0cfc698198997 \ # the private key 
  -c client.toml \ # the config file
  -chain 4 \ # Chain id for rinkeby
  -rpc https://rinkeby.hashcloak.com # An rpc endpoint to obtain the latest nonce count and gas price. Only necessary when using a private key.
```

Another way of sending a transaction with our wallet is by using a presigned raw transaction. Like this:

```bash
RAW_TXN=0xf8640284540be40083030d409400b1c66f34d680cb8bf82c64dcc1f39be5d6e77501802ca0c434f4d4b894b7cce2d880c250f7a67e4ef64cf0a921e3e4859219dff7b086fda0375a6195e221be77afda1d7c9e7d91bf39845065e9c56f7b5154e077a1ef8a77
go run ./cmd/wallet/main.go \
  -t gor \ # gor is the ethereum chain ticker for the goerli testnet
  -s gor \ # the Meson service name
  -rt $RAW_TXN \ # The signed raw transaction blob
  -c client.toml \ # the config file
  -chain 5 \ # Chain id for goerli
```

## Log files

The way our docker services are being ran mounts all of the log files output to the host docker host. All of the log files will be located in the mount the directory of the docker service. So if you you mount the docker container to `./mixnet/nonvoting` the logs of the authority will be saved in that location.

If you are running a full mixnet this little script will be useful to looking at all of the logs.

```
find ./configs -name "*.log" | sudo xargs tail -f
```

## Waiting for Katzenpost Epoch

Due to how katzenpost is designed, when you join the mixnet you will have to wait a new epoch to publish your node descriptor. An epoch right is about 10 minutes. While you wait for a new epoch you will see this message appear in the log files:

```
01:40:35.977 WARN pki: Authority rejected upload for epoch: 138106 (Conflict/Late)
01:40:35.977 WARN pki: Failed to post to PKI: pki: post for epoch will never succeeed
```

And in the authority you will see this log appear:

```
18:36:31.688 ERRO authority: Peer 10.0.0.2:57660: Rejected probably a conflict: state: Node oRh8boMS6VzJW57m5lMNqfK8EZ+LYfkfV0eJXKAJcJc=: Late descriptor upload for for epoch 138207
```

Once you your node has published its descriptor to the authority you will get a message that starts with this:

```
01:51:37.821 DEBU pki/nonvoting/client: Document: &{Epoch:138108...
```

## Other Blockchains

We intend to add support for other chains but, for now, only Ethereum based transactions are supported. We are currently only running `Goerli` and `Rinkeby` testnets but you can run a provider with access to an rpc node of any Ethereum compatible chain such as `ETC` or `Mordor`. If you want help setting up a provider for another chain please get in contact with us at info@hashcloak.com!

The steps needed to add a new Ethereum based chain are:

- Obtain access to an rpc endpoint
- Configure currency.toml with the new service.