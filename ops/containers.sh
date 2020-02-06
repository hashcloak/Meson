#!/usr/bin/env bash
set -e
tempDir=$(mktemp -d /tmp/meson-conf.XXXX)
rm -f /tmp/meson-current
ln -s $tempDir /tmp/meson-current 
numberNodes=${NUMBER_NODES:-2}
publicIP=$(ip route get 1)
publicIP=$(echo $publicIP | head -1 | sed 's/.*src//' | cut -f2 -d' ')
genconfig -o /tmp/meson-current -n $numberNodes -a $publicIP

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
    $MESON_IMAGE &
done

for i in $(seq 0 $(($numberNodes-1))); do
  port=$(($i+3))
  echo "Node $i with port 3000$port"
  docker service create --name node-$i -d \
    -p 3000$port:3000$port \
    --mount type=bind,source=/tmp/meson-current/node-$i,destination=/conf \
    $MESON_IMAGE &
done
