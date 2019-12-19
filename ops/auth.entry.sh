#!/bin/bash
set -e
port="${PORT:=29483}"
dockerNetworkIP=$(ip addr show eth0 | grep inet | grep -v inet6 | cut -d' ' -f6 | cut -d\/ -f1)
address="${ADDRESS:=$dockerNetworkIP}"
configFile="${CONFIG_FILE:=/conf/authority.toml}"
dataDir="${DATA_DIR:=/conf/data}"
mkdir -p $dataDir
chmod -R 700 $dataDir
# Level specifies the log level out of `ERROR`, `WARNING`, `NOTICE`,
# `INFO` and `DEBUG`.
logLevel="${LOG_LEVEL:=ERROR}"
disableLogging="${DISABLE_LOGS:=true}"
logFile=$dataDir/katzenpost.log
rm -f $logFile
ln -s /dev/stdout $logFile

function generateConfig {
  if [[ -z $MIX_KEYS ]]; then
    echo "ERROR: Value MIX_KEYS is not set."
    echo "Please set this value with the public keys of the mix nodes spaced with a comma"
    echo "example: key1,key2,key3...keyN"
    exit 1
  fi
  if [ -z $PROVIDERS ]; then
    echo "ERROR: Value PROVIDERS is not set."
    echo "Please set this value with the public keys of the provider nodes with their identifiers and keys"
    echo "example: identifier1:key1,identifier2:key2...identifierN:keyN"
    exit 1
  fi
  cat - > $configFile <<EOF
# Katzenpost non-voting authority configuration file.
[Authority]
  Addresses = [ "${address}:${port}" ]
  DataDir = "${dataDir}"

[Logging]
  Disable = ${disableLogging}
  File = "${logFile}"
  Level = "${logLevel}"

[Debug]
  MinNodesPerLayer = 1

[Parameters]
  SendRatePerMinute = 0
  Mu = 0.001
  MuMaxDelay = 90000
  LambdaP = 0.0001234
  LambdaPMaxDelay = 30000
  LambdaL = 0.0001234
  LambdaLMaxDelay = 30000
  LambdaD = 0.0001234
  LambdaDMaxDelay = 30000
EOF

  IFS=,
  for prov in $PROVIDERS; do
    cat - >> $configFile <<EOF
[[Providers]]
  Identifier = "$(echo $prov | cut -d':' -f1)"
  IdentityKey = "$(echo $prov | cut -d':' -f2)"
EOF
  done

  for mix_id_key in $MIX_KEYS; do
    cat - >> $configFile <<EOF
[[Mixes]]
  IdentityKey = "${mix_id_key}"
EOF
  done
}

if [[ ! -f $configFile ]]; then
  echo "Generating config file..."
  generateConfig
else
  echo "Using exsiting config file at: $configFile"
fi

printf '\n\n\n\n'
echo "The public key of this node is:"
echo $(cat ${dataDir}/identity.public.pem | grep -v PUBLIC)
printf '\n\n\n\n'

exec /go/bin/nonvoting -f $configFile
