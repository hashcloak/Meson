#!/bin/bash
port="${PORT:=29485}"
dockerNetworkIP=$(ip addr show eth0 | grep inet | grep -v inet6 | cut -d' ' -f6 | cut -d\/ -f1)
address="${ADDRESS:=$dockerNetworkIP}"
isProvider="${IS_PROVIDER:=false}"
authorityAddress="${AUTH_ADDRESS:=172.28.1.2}"
authorityPort="${AUTH_PORT:=29483}"
authorityPubKey="${AUTH_PUBK:=o4w1Nyj/nKNwho5SWfAIfh7SMU8FRx52nMHGgYsMHqQ=}"
configFile="${CONFIG_FILE:=/conf/katzenpost.toml}"
# Level specifies the log level out of `ERROR`, `WARNING`, `NOTICE`,
# `INFO` and `DEBUG`.
logLevel="${LOG_LEVEL:=ERROR}"
disablelogging="${DISABLE_LOGS:=true}"

cat - > $configFile <<EOF
# Katzenpost server configuration file.
[Server]
  Identifier = "${IDENTIFIER}"
  Addresses = [ "${address}:${port}"]
  DataDir = "/conf/mix_data"
  IsProvider = ${isProvider}

[PKI]
  [PKI.Nonvoting]
    Address = "${authorityAddress}:${authorityPort}"
    PublicKey = "${authorityPubKey}"

[Logging]
  Disable = ${disablelogging}
  File = "/conf/mix_data/katzenpost.log"
  Level = "${logLevel}"
EOF

exec /go/bin/server -f $configFile
