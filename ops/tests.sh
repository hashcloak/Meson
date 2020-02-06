#!/usr/bin/env bash
set -ex

# $1 DisableDecoyTraffic
# $2 Authority's public ipv4 address
# $3 Authority's Public key
function generateClientToml() {
  cat - > /tmp/meson-current/client.toml <<EOF
[Logging]
  Disable = false
  Level = "DEBUG"
  File = ""

[UpstreamProxy]
  Type = "none"

[Debug]
  DisableDecoyTraffic = $1
  CaseSensitiveUserIdentifiers = false
  PollingInterval = 1

[NonvotingAuthority]
    Address = "$2:30000"
    PublicKey = "$3"
EOF
}

# $1 Service to use
# $2 Provider name
# $3 private key
function runIntegrationTest() {
go run /tmp/Meson-client/integration/tests.go \
  -c /tmp/meson-current/client.toml \
  -t $1 -s $1 \
  -k /tmp/meson-current/$2/currency.toml
  -pk $3
}

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

if [ ! -z "$CI" ]; then
 sudo chown ${USER} -R /tmp
fi

authorityPublicKey=$(cat /tmp/meson-current/nonvoting/identity.public.pem | grep -v "PUBLIC")
generateClientToml true $publicIP $authorityPublicKey

# Commit that has the integration tests 
# Can be replaced to maaster once it is merged
testsCommit=5adb4b6aa9bb1eab7a59acab0f0d9e5839369908
git clone https://github.com/hashcloak/Meson-client /tmp/Meson-client
cd /tmp/Meson-client
git checkout $testsCommit

sleep 90
runIntegrationTest gor provider-0 $ETHEREUM_PK
