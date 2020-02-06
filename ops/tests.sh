#!/usr/bin/env bash
set -ex

if [ -z "$ETHEREUM_PK" ]; then
  echo "Need to set the ETHEREUM_PK env var"
  exit -1
fi

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
  -k /tmp/meson-current/$2/currency.toml \
  -pk $3
}

publicIP=$(ip route get 1 | head -1 | sed 's/.*src//' | cut -f2 -d' ')
authorityPublicKey=$(cat /tmp/meson-current/nonvoting/identity.public.pem | grep -v "PUBLIC")
generateClientToml true $publicIP $authorityPublicKey

# Commit that has the integration tests 
# Can be replaced to maaster once it is merged
testsCommit=a8af29632080a7755d734825052f86ce5cb651a2
git clone https://github.com/hashcloak/Meson-client /tmp/Meson-client ||
  git --git-dir=/tmp/Meson-client/.git --work-tree=/tmp/Meson-client pull origin
cd /tmp/Meson-client
git checkout $testsCommit

runIntegrationTest gor provider-0 $ETHEREUM_PK
