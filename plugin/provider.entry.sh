#!/bin/bash
identifier="${IDENTIFIER:=provider1}"
port="${PORT:=59484}"
registrationPort="${REGISTRATION_PORT:=36968}"
dockerNetworkIP=$(ip addr show eth0 | grep inet | grep -v inet6 | cut -d' ' -f6 | cut -d\/ -f1)
address="${ADDRESS:=$dockerNetworkIP}"
configFile="${CONFIG_FILE:=/conf/katzenpost.toml}"
currencyFile="${CURRENCY_FILE:=/conf/currency.toml}"
isProvider="${IS_PROVIDER:=false}"
authorityAddress="${AUTH_ADDRESS:=172.28.1.2}"
authorityPort="${AUTH_PORT:=29483}"
authorityPubKey="${AUTH_PUBK:=o4w1Nyj/nKNwho5SWfAIfh7SMU8FRx52nMHGgYsMHqQ=}"
# Level specifies the log level out of `ERROR`, `WARNING`, `NOTICE`,
# `INFO` and `DEBUG`.
logLevel="${LOG_LEVEL:=ERROR}"
disableLogging=${DISABLE_LOGS:=true}
disableRateLimit=${DISABLE_RATE_LIM:=true}
enableUserRegistrationHTTP=${ALLOW_HTTP_USER_REGISTRATION:=true}
management=${ALLOW_MANAGEMENT:=false}
service="${SERVICE_TICKER:=gor}"
chainID="${CHAIN_ID:=5}"
rpcURL="${RPC_URL:=http://172.28.1.10:9545}"

cat - > $configFile <<EOF
# Katzenpost server configuration file.
[Server]
  Identifier = "${identifier}"
  Addresses = [ "${address}:${port}"]
  DataDir = "/conf/provider_data"
  IsProvider = ${isProvider}

[PKI]
  [PKI.Nonvoting]
    Address = "${authorityAddress}:${authorityPort}"
    PublicKey = "${authorityPubKey}"

[Logging]
  Disable = ${disableLogging}
  File = "/conf/provider_data/katzenpost.log"
  Level = "${logLevel}"

[Debug]
  DisableRateLimit = ${disableRateLimit}

[Provider]
EnableUserRegistrationHTTP = ${enableUserRegistrationHTTP}
UserRegistrationHTTPAddresses = ["${address}:${registrationPort}"]
AdvertiseUserRegistrationHTTPAddresses = ["http://${address}:${registrationPort}"]

[Management]
  Enable = ${management}

[[Provider.Kaetzchen]]
    Disable = false
    Capability = "loop"
    Endpoint = "+loop"

   [[Provider.CBORPluginKaetzchen]]
     Disable = false
     Capability = "echo"
     Endpoint = "+echo"
     Command = "/go/bin/echo_server"
     MaxConcurrency = 1
     [Provider.CBORPluginKaetzchen.Config]
      log_dir = "/conf/service_logs"
      log_level = "${logLevel}"

   [[Provider.CBORPluginKaetzchen]]
     Disable = false
     Capability = "panda"
     Endpoint = "+panda"
     Command = "/go/bin/panda_server"
     MaxConcurrency = 1
     [Provider.CBORPluginKaetzchen.Config]
      log_dir = "/conf/service_logs"
      log_level = "${logLevel}"
      fileStore = "/conf/service_data/panda.storage"

  [[Provider.CBORPluginKaetzchen]]
    Disable = false
    Capability = "spool"
    Endpoint = "+spool"
    Command = "/go/bin/memspool"
    MaxConcurrency = 1
    [Provider.CBORPluginKaetzchen.Config]
      data_store = "/conf/service_data/memspool.storage"
      log_dir = "/conf/service_logs"

  [[Provider.CBORPluginKaetzchen]]
    Disable = false
    Capability = "${service}"
    Endpoint = "+${service}"
    Command = "/go/bin/Meson"
    MaxConcurrency = 1
    [Provider.CBORPluginKaetzchen.Config]
      log_dir = "/conf/service_logs"
      log_level = "${logLevel}"
      f = "$currencyFile"
EOF

cat - > $currencyFile <<EOF
Ticker = "${service}"
RPCUser = "rpcuser"
RPCPass = "rpcpassword"
RPCURL = "${rpcURL}"
ChainId = ${chainID}
LogDir = "/conf/service_logs"
LogLevel = "${logLevel}"
EOF

exec /go/bin/server -f $configFile
