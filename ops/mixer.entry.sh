#!/bin/bash
port="${PORT:=29485}"
dockerNetworkIP=$(ip addr show eth0 | grep inet | grep -v inet6 | cut -d' ' -f6 | cut -d\/ -f1)
address="${ADDRESS:=$dockerNetworkIP}"
isProvider="${IS_PROVIDER:=false}"
authorityAddress="${AUTH_ADDRESS:=172.28.1.2}"
authorityPort="${AUTH_PORT:=29483}"
authorityPubKey="${AUTH_PUBK:=o4w1Nyj/nKNwho5SWfAIfh7SMU8FRx52nMHGgYsMHqQ=}"
configFile="${CONFIG_FILE:=/conf/katzenpost.toml}"
dataDir="${DATA_DIR:=/conf/data}"
mkdir -p $dataDir
chmod -R 700 $dataDir
# Level specifies the log level out of `ERROR`, `WARNING`, `NOTICE`,
# `INFO` and `DEBUG`.
logLevel="${LOG_LEVEL:=ERROR}"
disablelogging="${DISABLE_LOGS:=true}"

cat - > $configFile <<EOF
# Katzenpost server configuration file.
[Server]
  Identifier = "${IDENTIFIER}"
  Addresses = [ "${address}:${port}"]
  DataDir = "${dataDir}"
  IsProvider = ${isProvider}

[PKI]
  [PKI.Nonvoting]
    Address = "${authorityAddress}:${authorityPort}"
    PublicKey = "${authorityPubKey}"

[Logging]
  Disable = ${disablelogging}
  File = "${dataDir}/katzenpost.log"
  Level = "${logLevel}"
EOF

if [ ${IDENTIFIER} == "mix1" ]; then
pk='-----BEGIN ED25519 PRIVATE KEY-----
LGH1LpJqJAG+Hg1wV6oMgYtQ838xZS9vRwjQ0B0kybfiv2N48Bu7OHK1yczedsPl
RA6bOXVfWEjxA4xuFOuDyw==
-----END ED25519 PRIVATE KEY-----'
pub='-----BEGIN ED25519 PUBLIC KEY-----
4r9jePAbuzhytcnM3nbD5UQOmzl1X1hI8QOMbhTrg8s=
-----END ED25519 PUBLIC KEY-----'
link='-----BEGIN X25519 PRIVATE KEY-----
Xv4PJ6gPCJEvuLVQ7dsUMT21KnASz4O82CyJhpCPoXg=
-----END X25519 PRIVATE KEY-----'
fi


if [ ${IDENTIFIER} == "mix2" ]; then
pk='-----BEGIN ED25519 PRIVATE KEY-----
ApYndNBRTyY5Peu9sO7fQiZphU04Kepp3ai+dFXDdrLXHy2rTiFj6Lzej1ifaet9
OQWPiNcgxYJaKVHt+CtbIw==
-----END ED25519 PRIVATE KEY-----'
pub='-----BEGIN ED25519 PUBLIC KEY-----
1x8tq04hY+i83o9Yn2nrfTkFj4jXIMWCWilR7fgrWyM=
-----END ED25519 PUBLIC KEY-----'
link='-----BEGIN X25519 PRIVATE KEY-----
MpTalIlC0SqMEWvgqiohvu18vfxxq6Rm5cTVZBFE9eY=
-----END X25519 PRIVATE KEY-----'
fi

if [ ${IDENTIFIER} == "mix3" ]; then
pk='-----BEGIN ED25519 PRIVATE KEY-----
olXV48urFRPjucBUbvZ1i8bDGQozzyaazDndya9OmiAtrYWexsqhTx3NCF83jOTF
JifpPstFaYWfQmC2USSVDA==
-----END ED25519 PRIVATE KEY-----'
pub='-----BEGIN ED25519 PUBLIC KEY-----
La2FnsbKoU8dzQhfN4zkxSYn6T7LRWmFn0JgtlEklQw=
-----END ED25519 PUBLIC KEY-----'
link='-----BEGIN X25519 PRIVATE KEY-----
87rDmtVXd1DW3eqh6Zfs+ICrRu0YT2Rx8CuRs9sIfCE=
-----END X25519 PRIVATE KEY-----'
fi

cat - > ${dataDir}/identity.private.pem << EOF
${pk}
EOF

cat - > ${dataDir}/identity.public.pem << EOF
${pub}
EOF

cat - > ${dataDir}/link.private.pem << EOF
${link}
EOF

exec /go/bin/server -f $configFile
