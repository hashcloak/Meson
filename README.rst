
Meson mixnet microservice for Ethereum-based transactions
=====================================================

Meson is a fork of the `katzenpost currency plugin
<https://github.com/katzenpost/currency>`_.
It proxies transactions to a daemon via the ETH JSON HTTP RPC. As such, as long as a cryptocurrency
conforms to the ETH JSON HTTP RPC, it can make use of this plugin. One needs to only submit a raw transaction
and this mixnet service will submit them to the respective blockchain.


Dependecies
-----------

::

  docker-compose
  make

Usage
-----

It's a plugin. You are not supposed to run it yourself on the commandline.
See the handbook to learn how to configure external plugins:

* https://github.com/katzenpost/docs/blob/master/handbook/index.rst#external-kaetzchen-plugin-configuration

( if that's not enough then read our spec: https://github.com/katzenpost/docs/blob/master/specs/kaetzchen.rst )

Running a local docker-compose tesnet:
::

   make up

Bring down the local docker-compose tesnet:
::

   make down


Configuration
-------------

In order to use this plugin your Katzenpost server will need
a configuration section that looks like this:

::

  [[Provider.CBORPluginKaetzchen]]
    Disable = false
    Capability = "eth"
    Endpoint = "+eth"
    Command = "/go/bin/Meson"
    MaxConcurrency = 1
    [Provider.CBORPluginKaetzchen.Config]
      log_dir = "/conf/service_logs"
      log_level = "DEBUG"
      f = "/conf/currency.toml"


And you will also need a `currency.toml` file as the config file for Meson:

::

   Ticker = "ETH"
   RPCUser = "rpcuser"
   RPCPass = "rpcpassword"
   RPCURL = "http://127.0.0.1:9545"
   ChainId = 4
   LogDir = "/conf/service_logs"
   LogLevel = "DEBUG"


C bindings
----------

Firstly, build the client_bindings as documented here:

* https://github.com/katzenpost/client_bindings

And then build the currency common bindings:

::

   cd common/bindings
   go build -o meson_bindings.so -buildmode=c-shared bindings.go

Finally, we can build our example C wallet:

::


   gcc ./examples/wallet.c ./common/bindings/meson_bindings.so ../../client_bindings/client_bindings.so -I /home/user/gopath/src/github.com/hashcloak/Meson/common/bindings/ -I /home/user/gopath/src/github.com/katzenpost/client_bindings/ -o wallet


license
=======

AGPL: see LICENSE file for details.
