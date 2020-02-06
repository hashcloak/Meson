#!/usr/bin/env bash
set -ex

if [ -z "$ETHEREUM_PK" ]; then
  echo "Need to set the ETHEREUM_PK env var"
  exit -1
fi

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

# Commit that has the integration tests 
# Can be replaced to master once it is merged
testsCommit=${TEST_COMMIT:-a8af29632080a7755d734825052f86ce5cb651a2}
git clone https://github.com/hashcloak/Meson-client /tmp/Meson-client || true
cd /tmp/Meson-client && git fetch && git checkout $testsCommit
runIntegrationTest gor provider-0 $ETHEREUM_PK
