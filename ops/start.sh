#!/usr/bin/env bash
set -ex

if [ -z "$ETHEREUM_PK" ]; then
  echo "Need to set the ETHEREUM_PK env var"
  exit -1
fi


tempDir=$(mktemp -d /tmp/meson-conf.XXXX)
rm -f /tmp/meson-current
ln -s $tempDir /tmp/meson-current 
numberNodes=${NUMBER_NODES:-2}
publicIP=$(ip route get 1)
publicIP=$(echo $publicIP | head -1 | sed 's/.*src//' | cut -f2 -d' ')
genconfig -o /tmp/meson-current -n $numberNodes -a $publicIP

echo "Binance"
# This is temporary until a binance node is more permanent
sed -i 's|RPCURL =.*|RPCURL = "'$publicIP':27147"|g' /tmp/meson-current/provider-1/currency.toml
docker service create --name binance -d -e "NODETYPE=lightnode" -p 27147:27147 hashcloak/binance:latest

echo "Authority"
docker service create --name authority -d \
  -p 30000:30000 \
  --mount type=bind,source=/tmp/meson-current/nonvoting,destination=/conf \
  $KATZEN_AUTH

for i in $(seq 0 1); do
  port=$(($i+1))
  echo "Provider $i with port 3000$port"
  docker service create --name provider-$i -d \
    -p 3000$port:3000$port \
    -p 4000$port:4000$port \
    --mount type=bind,source=/tmp/meson-current/provider-$i,destination=/conf \
    $MESON_IMAGE
done

for i in $(seq 0 $(($numberNodes-1))); do
  port=$(($i+3))
  echo "Node $i with port 3000$port"
  docker service create --name node-$i -d \
    -p 3000$port:3000$port \
    --mount type=bind,source=/tmp/meson-current/node-$i,destination=/conf \
    $MESON_IMAGE
done

if [ ! -z "$CI" ]; then
 sudo chown ${USER} -R /tmp
fi

authorityPublicKey=$(cat /tmp/meson-current/nonvoting/identity.public.pem | grep -v "PUBLIC")
cat - > /tmp/meson-current/client.toml <<EOF
[Logging]
  Disable = false
  Level = "DEBUG"
  File = ""

[UpstreamProxy]
  Type = "none"

[Debug]
  DisableDecoyTraffic = true
  CaseSensitiveUserIdentifiers = false
  PollingInterval = 1

[NonvotingAuthority]
    Address = "${publicIP}:30000"
    PublicKey = "${authorityPublicKey}"
EOF

sleep 10
go run /home/sebas/repos/hashcloack/Meson-client/integration/tests.go \
  -c /tmp/meson-current/client.toml \
  -t gor -s gor \
  -pk $ETHEREUM_PK
