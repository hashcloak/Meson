
Meson mixnet microservice for Ethereum-based transactions
=====================================================

Meson is a fork of the `katzenpost currency plugin
<https://github.com/katzenpost/currency>`_.
It proxies transactions to a daemon via the ETH JSON HTTP RPC. As such, as long as a cryptocurrency
conforms to the ETH JSON HTTP RPC, it can make use of this plugin. One needs to only submit a raw transaction
and this mixnet service will submit them to the respective blockchain.

usage
-----

It's a plugin. You are not supposed to run it yourself on the commandline.
See the handbook to learn how to configure external plugins:

* https://github.com/katzenpost/docs/blob/master/handbook/index.rst#external-kaetzchen-plugin-configuration

( if that's not enough then read our spec: https://github.com/katzenpost/docs/blob/master/specs/kaetzchen.rst )

::

    ./meson-go -h
      Usage of ./meson-go:
        -f string
            Path to the currency config file. (default "currency.toml")
        -log_dir string
            logging directory
        -log_level string
            logging level could be set to: DEBUG, INFO, NOTICE, WARNING, ERROR, CRITICAL (default "DEBUG")


configuration
-------------

In order to use this plugin your Katzenpost server will need
a configuration section that looks like this:

::

    [[Provider.CBORPluginKaetzchen]]
      Capability = "eth"
      Endpoint = "+eth"
      Disable = false
      Command = "/home/user/test_mixnet/bin/meson-go"
      MaxConcurrency = 10
      [Provider.PluginKaetzchen.Config]
        log_dir = "/home/user/test_mixnet/eth_tx_logs"
        f = "/home/user/test_mixnet/meson/curreny.toml"


Here's a sample configuration file for currency-go to learn it's
Ticker and RPC connection information, currency.toml:

::

   Ticker = "ETH"
   RPCUser = "rpcuser"
   RPCPass = "rpcpassword"
   RPCURL = "http://127.0.0.1:18232"
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
