#!/bin/bash
set -e
port="${PORT:=59484}"
registrationPort="${REGISTRATION_PORT:=36968}"
dockerNetworkIP=$(ip addr show eth0 | grep inet | grep -v inet6 | cut -d' ' -f6 | cut -d\/ -f1)
address="${ADDRESS:=$dockerNetworkIP}"
configFile="${CONFIG_FILE:=/conf/katzenpost.toml}"
dataDir="${DATA_DIR:=/conf/data}"
mkdir -p $dataDir
chmod -R 700 $dataDir
currencyFile="${CURRENCY_FILE:=/conf/currency.toml}"
isProvider=${IS_PROVIDER:=false}
authorityAddress="${AUTH_ADDRESS:=172.28.1.2:29483}"
authorityPubKey="${AUTH_PUBK:=o4w1Nyj/nKNwho5SWfAIfh7SMU8FRx52nMHGgYsMHqQ=}"
# logLevelValues can be: `ERROR`, `WARNING`, `NOTICE`, `INFO` and `DEBUG`.
logLevel="${LOG_LEVEL:=ERROR}"
disableLogging=${DISABLE_LOGS:=true}
logFile=$dataDir/katzenpost.log
ln -s /dev/stdout $logFile
disableRateLimit=${DISABLE_RATE_LIM:=true}
enableUserRegistrationHTTP=${ALLOW_HTTP_USER_REGISTRATION:=true}
management=${ALLOW_MANAGEMENT:=false}
service="${SERVICE_TICKER:=gor}"
chainID="${CHAIN_ID:=5}"
rpcURL="${RPC_URL:=https://g.sebas.tech}"
plugins="${PLUGINS:=echo panda memspool}"



if $isProvider; then
echo "Setting up provider"
cat - > $configFile <<EOF
# Katzenpost server configuration file.
[Server]
  Identifier = "${IDENTIFIER}"
  Addresses = [ "${address}:${port}"]
  DataDir = "${dataDir}"
  IsProvider = ${isProvider}

[PKI]
  [PKI.Nonvoting]
    Address = "${authorityAddress}"
    PublicKey = "${authorityPubKey}"

[Logging]
  Disable = ${disableLogging}
  File = "${logFile}"
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
    Capability = "${service}"
    Endpoint = "+${service}"
    Command = "/go/bin/Meson"
    MaxConcurrency = 1
    [Provider.CBORPluginKaetzchen.Config]
      log_dir = "${dataDir}"
      log_level = "${logLevel}"
      f = "$currencyFile"
EOF

if [[ $plugins == *"echo"* ]]; then
cat - >> $configFile <<EOF
   [[Provider.CBORPluginKaetzchen]]
     Disable = false
     Capability = "echo"
     Endpoint = "+echo"
     Command = "/go/bin/echo_server"
     MaxConcurrency = 1
     [Provider.CBORPluginKaetzchen.Config]
      log_dir = "${dataDir}"
      log_level = "${logLevel}"
EOF
fi

if [[ $plugins == *"panda"* ]]; then
cat - >> $configFile <<EOF
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
EOF
fi

if [[ $plugins == *"memspool"* ]]; then
cat - >> $configFile <<EOF
  [[Provider.CBORPluginKaetzchen]]
    Disable = false
    Capability = "spool"
    Endpoint = "+spool"
    Command = "/go/bin/memspool"
    MaxConcurrency = 1
    [Provider.CBORPluginKaetzchen.Config]
      data_store = "${dataDir}/memspool.storage"
      log_dir = "${dataDir}"
EOF
fi

cat - > $currencyFile <<EOF
Ticker = "${service}"
RPCUser = "rpcuser"
RPCPass = "rpcpassword"
RPCURL = "${rpcURL}"
ChainId = ${chainID}
LogDir = "${dataDir}"
LogLevel = "${logLevel}"
EOF

else
echo "Setting up mix"
cat - > $configFile <<EOF
# Katzenpost server configuration file.
[Server]
  Identifier = "${IDENTIFIER}"
  Addresses = [ "${address}:${port}"]
  DataDir = "${dataDir}"
  IsProvider = ${isProvider}

[PKI]
  [PKI.Nonvoting]
    Address = "${authorityAddress}"
    PublicKey = "${authorityPubKey}"

[Logging]
  Disable = ${disableLogging}
  File = "${logFile}"
  Level = "${logLevel}"
EOF
fi

if [ ${IDENTIFIER} == "mix1" ]; then
pk='-----BEGIN ED25519 PRIVATE KEY-----
LGH1LpJqJAG+Hg1wV6oMgYtQ838xZS9vRwjQ0B0kybfiv2N48Bu7OHK1yczedsPl
RA6bOXVfWEjxA4xuFOuDyw==
-----END ED25519 PRIVATE KEY-----'
pub='-----BEGIN ED25519 PUBLIC KEY-----
4r9jePAbuzhytcnM3nbD5UQOmzl1X1hI8QOMbhTrg8s=
-----END ED25519 PUBLIC KEY-----'
link='-----BEGIN X25519 PRIVATE KEY-----
Xv4PJ6gPCJEvuLVQ7dsUMT21KnASz4O82CyJhpCPoXg=
-----END X25519 PRIVATE KEY-----'
fi


if [ ${IDENTIFIER} == "mix2" ]; then
pk='-----BEGIN ED25519 PRIVATE KEY-----
ApYndNBRTyY5Peu9sO7fQiZphU04Kepp3ai+dFXDdrLXHy2rTiFj6Lzej1ifaet9
OQWPiNcgxYJaKVHt+CtbIw==
-----END ED25519 PRIVATE KEY-----'
pub='-----BEGIN ED25519 PUBLIC KEY-----
1x8tq04hY+i83o9Yn2nrfTkFj4jXIMWCWilR7fgrWyM=
-----END ED25519 PUBLIC KEY-----'
link='-----BEGIN X25519 PRIVATE KEY-----
MpTalIlC0SqMEWvgqiohvu18vfxxq6Rm5cTVZBFE9eY=
-----END X25519 PRIVATE KEY-----'
fi

if [ ${IDENTIFIER} == "mix3" ]; then
pk='-----BEGIN ED25519 PRIVATE KEY-----
olXV48urFRPjucBUbvZ1i8bDGQozzyaazDndya9OmiAtrYWexsqhTx3NCF83jOTF
JifpPstFaYWfQmC2USSVDA==
-----END ED25519 PRIVATE KEY-----'
pub='-----BEGIN ED25519 PUBLIC KEY-----
La2FnsbKoU8dzQhfN4zkxSYn6T7LRWmFn0JgtlEklQw=
-----END ED25519 PUBLIC KEY-----'
link='-----BEGIN X25519 PRIVATE KEY-----
87rDmtVXd1DW3eqh6Zfs+ICrRu0YT2Rx8CuRs9sIfCE=
-----END X25519 PRIVATE KEY-----'
fi


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
