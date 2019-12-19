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
# logLevelValues can be: `ERROR`, `WARNING`, `NOTICE`, `INFO` and `DEBUG`.
logLevel="${LOG_LEVEL:=ERROR}"
disableLogging=${DISABLE_LOGS:=true}
logFile=$dataDir/katzenpost.log
disableRateLimit=${DISABLE_RATE_LIM:=true}
enableUserRegistrationHTTP=${ALLOW_HTTP_USER_REGISTRATION:=true}
management=${ALLOW_MANAGEMENT:=false}
service="${SERVICE_TICKER:=gor}"
chainID="${CHAIN_ID:=5}"
plugins="${PLUGINS:=echo panda memspool}"


function generateProviderConfig {
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
    Address = "${AUTHORITY_ADDRESS}"
    PublicKey = "${AUTHORITY_KEY}"

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

}

function generateMixConfig {
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
    Address = "${AUTHORITY_ADDRESS}"
    PublicKey = "${AUTHORITY_KEY}"
[Logging]
  Disable = ${disableLogging}
  File = "${logFile}"
  Level = "${logLevel}"
EOF
}


rm -f $logFile
ln -s /dev/stdout $logFile

if [[ ! -f "$configFile" ]]; then

  if [[ -z $AUTHORITY_ADDRESS ]]; then
    echo "ERROR: Value AUTHORITY_ADDRESS is not set."
    echo "Please set this value with the public ipv4 or ipv6 address with the port included"
    echo "example: key1,key2,key3...keyN"
    exit 1
  fi

  if [[ -z $AUTHORITY_KEY ]]; then
    echo "ERROR: Value AUTHORITY_KEY is not set."
    echo "Please set this value with the public key of the authority"
    exit 1
  fi

  echo "Generating config file..."

  if $isProvider; then
    if [[ -z $RPC_URL ]]; then
      echo "ERROR: Value RPC_URL is not set."
      echo "Please set this value to a blockchain node that Meson is capable of connecting to"
      exit 1
    fi
    generateProviderConfig
  else
    generateMixConfig
  fi
else
  echo "Using existing config file at: $configFile"
fi


printf '\n\n\n\n'
echo "The public key of this node is:"
echo $(cat ${dataDir}/identity.public.pem | grep -v PUBLIC)
printf '\n\n\n\n'

exec /go/bin/server -f $configFile
