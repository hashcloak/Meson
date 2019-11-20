FROM ethereum/client-go:%%GETH_VERSION%%

# The why of this custom geth docker image is
# because piping stdout in docker compose results
# in a an error when starting the geth container
ENTRYPOINT \
  geth \
  --$CHAIN \
  --nousb \
  --datadir=/blockchain \
  --syncmode=fast \
  --rpc \
  --rpcport=9545 \
  --rpcaddr=0.0.0.0 \
  --rpcapi=eth,net,web3 \
  2> /blockchain/geth.log
