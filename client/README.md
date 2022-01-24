# Meson-client
[![Go](https://github.com/hashcloak/Meson-client/actions/workflows/go.yml/badge.svg)](https://github.com/hashcloak/Meson-client/actions/workflows/go.yml)

A simple client for use with the Meson mixnet software

## Tests

Since this library requires to connect to an existing katzenpost mixnet one needs to run the tests inside of a docker container and connect to the mixnet docker network. You can run a mixnet by following the instructions at [https://github.com/hashcloak/Meson](https://github.com/hashcloak/Meson)

```
docker run --rm \
  -v `pwd`:/client \
  --network nonvoting_testnet_nonvoting_test_net \
  -v /tmp/gopath-pkg:/go/pkg \
  -w /client \
  golang:buster \
  /bin/bash -c "GORACE=history_size=7 go test -race"
```

The above can be de-constructed as follows:
- ```-v `pwd`:/client```: Mount the current directory inside the docker container at `/client`
- `--network nonvoting_testnet_nonvoting_test_net`: Connect to the existing docker network mixnet
- `-v /tmp/gopath-pkg:/go/pkg`: Cache the go modules that belong to this container in `/tmp/gopath-pkg`
- `-w /client`: Working directory for the docker image
- `golang:buster`: The docker image to be used
-  `/bin/bash -c "GORACE=history_size=7 go test -race"`: The command to run inside the container
