#!/bin/bash
identifier="${IDENTIFIER:=provider1}"
port="${PORT:=59484}"
registrationPort="${REGISTRATION_PORT:=36968}"
dockerNetworkIP=$(ip addr show eth0 | grep inet | grep -v inet6 | cut -d' ' -f6 | cut -d\/ -f1)
address="${ADDRESS:=$dockerNetworkIP}"
configFile="${CONFIG_FILE:=/conf/katzenpost.toml}"
dataDir="${DATA_DIR:=/conf/data}"
mkdir -p $dataDir
chmod -R 700 $dataDir
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
  DataDir = "${dataDir}"
  IsProvider = ${isProvider}

[PKI]
  [PKI.Nonvoting]
    Address = "${authorityAddress}:${authorityPort}"
    PublicKey = "${authorityPubKey}"

[Logging]
  Disable = ${disableLogging}
  File = "${dataDir}/katzenpost.log"
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
      log_dir = "${dataDir}"
      log_level = "${logLevel}"

   [[Provider.CBORPluginKaetzchen]]
     Disable = false
     Capability = "panda"
     Endpoint = "+panda"
     Command = "/go/bin/panda_server"
     MaxConcurrency = 1
     [Provider.CBORPluginKaetzchen.Config]
      log_dir = "${dataDir}"
      log_level = "${logLevel}"
      fileStore = "${dataDir}/panda.storage"

  [[Provider.CBORPluginKaetzchen]]
    Disable = false
    Capability = "spool"
    Endpoint = "+spool"
    Command = "/go/bin/memspool"
    MaxConcurrency = 1
    [Provider.CBORPluginKaetzchen.Config]
      data_store = "${dataDir}/memspool.storage"
      log_dir = "${dataDir}"

  [[Provider.CBORPluginKaetzchen]]
    Disable = false
    Capability = "${service}"
    Endpoint = "+${service}"
    Command = "/go/bin/Meson"
    MaxConcurrency = 1
    [Provider.CBORPluginKaetzchen.Config]
      log_dir = "${dataDir}"
      log_level = "${logLevel}"
      f = "$currencyFile"
EOF

cat - > $currencyFile <<EOF
Ticker = "${service}"
RPCUser = "rpcuser"
RPCPass = "rpcpassword"
RPCURL = "${rpcURL}"
ChainId = ${chainID}
LogDir = "${dataDir}"
LogLevel = "${logLevel}"
EOF


if [ ${IDENTIFIER} == "provider1" ]; then
pk='-----BEGIN ED25519 PRIVATE KEY-----
ndkH4sYkxVXPFU7OF6wryVar5cWxsZEBcFXWOHnEM3/aSvB80N9tqRkJJNRRlgpf
B127OUSBL0l/Cbt7JnSwKA==
-----END ED25519 PRIVATE KEY-----'
pub='-----BEGIN ED25519 PUBLIC KEY-----
2krwfNDfbakZCSTUUZYKXwdduzlEgS9Jfwm7eyZ0sCg=
-----END ED25519 PUBLIC KEY-----'
link='-----BEGIN X25519 PRIVATE KEY-----
iboFtIVykmdzQoL3rDha5Vs0RtxfmQJT8CyWRbT09Xg=
-----END X25519 PRIVATE KEY-----'
fi

if [ ${IDENTIFIER} == "provider2" ]; then
pk='-----BEGIN ED25519 PRIVATE KEY-----
xun4XQfdsh+w8pD+e1Rml9RCaOTfCoZHG3s6OOek9v2KaKDMjbq1NFfJgte6MsQ8
j1Cs1g4SALgMSWwV0/1pxA==
-----END ED25519 PRIVATE KEY-----'
pub='-----BEGIN ED25519 PUBLIC KEY-----
imigzI26tTRXyYLXujLEPI9QrNYOEgC4DElsFdP9acQ=
-----END ED25519 PUBLIC KEY-----'
link='-----BEGIN X25519 PRIVATE KEY-----
04XSX85Ov4jDTb0vqEMv3McYME7weayniGLKtmF6UwQ=
-----END X25519 PRIVATE KEY-----'
fi

cat - > ${dataDir}/identity.private.pem << EOF
${pk}
EOF

cat - > ${dataDir}/identity.public.pem << EOF
${pub}
EOF

cat - > ${dataDir}/link.private.pem << EOF
${link}
EOF


exec /go/bin/server -f $configFile
