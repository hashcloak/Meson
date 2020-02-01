#!/usr/bin/env bash
set -ex
tempDir=$(mktemp -d /tmp/meson-conf.XXXX)
rm -f /tmp/meson-current
ln -s $tempDir /tmp/meson-current 
touch $tempDir/$tempDir
numberNodes=3
pubIP=$(ip route get 1 | cut -d' ' -f3)
genconfig -o $tempDir -n $numberNodes -a $pubIP
# This is temporary until a binance node is more permanent
sed -i 's|RPCURL =.*|RPCURL = "'$pubIP':27147"|g' $tempDir/provider-1/currency.toml

echo "Authority"
docker service create --name authority -d \
  -p 30000:30000 \
  --mount type=bind,source=$tempDir/nonvoting,destination=/conf \
  $KATZEN_AUTH

for i in $(seq 0 1); do
  port=$(($i+1))
  echo "Provider $1 with port 3000$port"
  docker service create --name provider-$i -d \
    -p 3000$port:3000$port \
    -p 4000$port:4000$port \
    --mount type=bind,source=$tempDir/provider-$i,destination=/conf \
    $MESON_IMAGE
done

for i in $(seq 0 $(($numberNodes-1))); do
  port=$(($i+3))
  echo "Node $1 with port 3000$port"
  docker service create --name node-$i -d \
    -p 3000$port:3000$port \
    --mount type=bind,source=$tempDir/node-$i,destination=/conf \
    $MESON_IMAGE
done

echo "Binance"
docker service create --name binance -d -e "NODETYPE=lightnode" -p 27147:27147 hashcloak/binance:latest

echo "Config files can be found in: $tempDir"