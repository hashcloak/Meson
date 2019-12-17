#!/bin/bash
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

ln -s /dev/stdout $logFile

default="4r9jePAbuzhytcnM3nbD5UQOmzl1X1hI8QOMbhTrg8s=
1x8tq04hY+i83o9Yn2nrfTkFj4jXIMWCWilR7fgrWyM=
La2FnsbKoU8dzQhfN4zkxSYn6T7LRWmFn0JgtlEklQw="
mixes="${MIXES:=$default}"

default="provider1,2krwfNDfbakZCSTUUZYKXwdduzlEgS9Jfwm7eyZ0sCg=
provider2,imigzI26tTRXyYLXujLEPI9QrNYOEgC4DElsFdP9acQ="
providers="${PROVIDERS:=$default}"

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

for mix_id_key in $mixes; do
cat - >> $configFile <<EOF
[[Mixes]]
  IdentityKey = "${mix_id_key}"

EOF
done

for prov in $providers; do
cat - >> $configFile <<EOF
[[Providers]]
  Identifier = "$(echo $prov | cut -d',' -f1)"
  IdentityKey = "$(echo $prov | cut -d',' -f2)"

EOF
done


cat - > ${dataDir}/identity.private.pem << EOF
-----BEGIN ED25519 PRIVATE KEY-----
bI4vCmWUlQOupW2Tr/rLbDjzDmE1kL5Qb7doaSpHOWKjjDU3KP+co3CGjlJZ8Ah+
HtIxTwVHHnacwcaBiwwepA==
-----END ED25519 PRIVATE KEY-----
EOF


cat - > ${dataDir}/identity.public.pem << EOF
-----BEGIN ED25519 PUBLIC KEY-----
o4w1Nyj/nKNwho5SWfAIfh7SMU8FRx52nMHGgYsMHqQ=
-----END ED25519 PUBLIC KEY-----
EOF

exec /go/bin/nonvoting -f $configFile
