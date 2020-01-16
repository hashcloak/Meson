# Docs
This is documentation related to the Meson mixnet project. Here, you can find out how to deploy a provider, authority or mix node and learn how to use our client libraries.


## Running Meson
### How to Run a Provider Node

All of our infrastructure uses docker setups. You will first need to generate a provider config and its PKI keys. The easiest way to do that is by using our [genconfig](https://github.com/hashcloak/genconfig/#genconfig) script:

```bash
git clone https://github.com/hashcloak/genconfig.git
cd genconfig
go run main.go \
  -a 138.197.57.19 \ # current ip address of authority
  -authID RJWGWCjof2GLLhekd6KsvN+LvHq9sxgcpra/J59/X8A= \ # current public key of authority
  -name provider-name \ # your provider name
  -ipv4 1.1.1.1 \ # your public ipv4 address
  -provider \ # flag to indicate you only want a provider config
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

### How to Run a Mix Node

To run a mix node we have to run the same command to generate the config file. The only difference is changing the `-provider` flag with `-node`.

```bash
go run main.go \
  -a 138.197.57.19 \ # current ip address of authority
  -authID RJWGWCjof2GLLhekd6KsvN+LvHq9sxgcpra/J59/X8A= \ # current public key of authority
  -name mix-node-name \ # your provider name
  -ipv4 1.1.1.1 \ # your public ipv4 address
  -node \ # flag to indicate you only want a mix node config
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
#output/mixn-node-name/katzenpost.toml
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


### Updating Authority Config

### Sending Transactions

```bash
o run ./cmd/wallet/main.go \
  -t rin \ # rin is the ethereum chain identifier for the rinkeby testnet
  -s rin \ # the Meson service name
  -pk 0x9e23c88a0ef6745f55d92eb63f063d2e267f07222bfa5cb9efb0cfc698198997 \ # the private key 
  -c client.toml \ # the config file
  -chain 4 \ # Chain id for rinkeby
  -rpc https://rinkeby.hashcloak.com # An rpc endpoint to obtain the latest nonce count and gas price. Only necessary when using a private key.
  ```



### Waiting for Katzenpost Epoch

Due to how katzenpost is designed, when you join the mixnet you will have to wait for about 10 minutes before your node can send its   

```
01:40:35.977 WARN pki: Authority rejected upload for epoch: 138106 (Conflict/Late)
01:40:35.977 WARN pki: Failed to post to PKI: pki: post for epoch will never succeeed
```


01:51:37.821 DEBU pki/nonvoting/client: Document: &{Epoch:138108...
```