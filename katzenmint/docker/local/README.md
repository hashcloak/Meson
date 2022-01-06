# Docker
This docker-compose configuration is meant to be used in combination with the [Meson server](https://github.com/hashcloak/Meson-server/tree/develop) and [Katzenmint PKI](https://github.com/hashcloak/katzenmint-pki/tree/develop) repositories. It is meant for testing client and server mix network components as part of the core Meson developer work flow. It should be obvious that this docker-compose situation is not meant for production use.

1. clone meson server and build docker image of meson server
```BASH
$ git clone https://github.com/hashcloak/Meson-server.git
$ cd Meson-server
$ docker build --no-cache -t meson/server .
```

2. build docker image for katzenmint-pki
```BASH
$ cd katzenmint-pki
$ docker build --no-cache -t katzenmint/pki .
```

3. start three katzenmint pki nodes / three mix nodes / two providers.
```BASH
$ docker-compose up
```

4. checkout information of katzenmint pki nodes with curl command
```BASH
# node1
$ curl http://127.0.0.1:21483/net_info

# node2
$ curl http://127.0.0.1:21484/net_info

# node3
$ curl http://127.0.0.1:21485/net_info
```

# Examples

## Ping
[Ping](https://github.com/sc0Vu/mixnet-examples/tree/master/ping) is a mixnet tool for testing and debugging. After start local mixnets, you can test with ping tool.

1. clone
```BASH
$ git clone https://github.com/sc0Vu/mixnet-examples.git
$ cd mixnet-examples/ping
```

2. update `katzenpost.toml
```BASH
$ vim katzenpost.toml
```

```TOML
# update Katzenmint settings, or you can simply use genconfig/updateconfig to update, see: https://github.com/hashcloak/genconfig/tree/add-cilint/update

[Katzenmint]
    ChainID = "test-chain-QvpdAC"
    PrimaryAddress = "tcp://127.0.0.1:26657"
    WitnessesAddresses = [ "127.0.0.1:26657" ]
    DatabaseName = "ping"
    DatabaseDir = "/tmp/meson/ping"
    RPCAddress = "tcp://127.0.0.1:26657"

    [Katzenmint.TrustOptions]
      Period = 600000000000
      Height = 13
      Hash = [54, 57, 47, 184, 187, 218, 235, 146, 116, 42, 17, 26, 39, 62, 241, 55, 25, 108, 10, 156, 59, 217, 7, 96, 76, 102, 248, 214, 37, 129, 9, 106]
```

3. ping
```BASH
$ go run main.go -n 1 -s echo -c ./katzenmint.toml
```

# Clean up chain/mix/provider data and restart

You can simply cleanup chaindata in one command.
```BASH
$ sh cleanup.sh
```

Then, restart three katzenmint pki nodes.
```BASH
$ docker-compose up
```

# TBD
Add meson.
Add catshadow.
Add more examples.