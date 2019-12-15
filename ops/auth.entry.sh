#!/bin/bash
port="${PORT:=29483}"
dockerNetworkIP=$(ip addr show eth0 | grep inet | grep -v inet6 | cut -d' ' -f6 | cut -d\/ -f1)
address="${ADDRESS:=$dockerNetworkIP}"
config_file="${CONFIG_FILE:=/conf/authority.toml}"

# Level specifies the log level out of `ERROR`, `WARNING`, `NOTICE`,
# `INFO` and `DEBUG`.
log_level="${LOG_LEVEL:=ERROR}"
disable_logging="${DISABLE_LOGS:=true}"

default="4r9jePAbuzhytcnM3nbD5UQOmzl1X1hI8QOMbhTrg8s=
1x8tq04hY+i83o9Yn2nrfTkFj4jXIMWCWilR7fgrWyM=
La2FnsbKoU8dzQhfN4zkxSYn6T7LRWmFn0JgtlEklQw="
mixes="${MIXES:=$default}"

default="provider1,2krwfNDfbakZCSTUUZYKXwdduzlEgS9Jfwm7eyZ0sCg=
provider2,imigzI26tTRXyYLXujLEPI9QrNYOEgC4DElsFdP9acQ="
providers="${PROVIDERS:=$default}"

cat - > $config_file <<EOF
# Katzenpost non-voting authority configuration file.
[Authority]
  Addresses = [ "${address}:${port}" ]
  DataDir = "/conf/authority_data"

[Logging]
  Disable = ${disable_logging}
  File = "/conf/authority_data/katzenpost.log"
  Level = "${log_level}"

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
cat - >> $config_file <<EOF
[[Mixes]]
  IdentityKey = "${mix_id_key}"

EOF
done

for prov in $providers; do
cat - >> $config_file <<EOF
[[Providers]]
  Identifier = "$(echo $prov | cut -d',' -f1)"
  IdentityKey = "$(echo $prov | cut -d',' -f2)"

EOF
done

exec /go/bin/nonvoting -f $config_file
