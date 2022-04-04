# Test

To test locally, one has to setup katzenmint PKI and configure Meson server accordingly. The prerequisite is to build both katzenmint PKI docker and Meson server.

Before running katzenmint PKI docker, one has to make a cleanup to update the genesis blocktime.
```BASH
$ cd katzenmint-pki/docker
$ sh cleanup.sh
$ docker-compose up
```

The sample toml file in this directory is compatible with katzenmint PKI docker, except that the updated genesis file has to be taken into account. The script contains a hacky way to do so.
```BASH
$ cd Meson-server/cmd/server
$ sh reset.sh
$ ./meson-server -f katzenpost.toml.sample
```

