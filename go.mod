module github.com/hashcloak/Meson

go 1.16

require (
	git.schwanenlied.me/yawning/aez.git v0.0.0-20180408160647-ec7426b44926
	git.schwanenlied.me/yawning/avl.git v0.0.0-20180224045358-04c7c776e391
	git.schwanenlied.me/yawning/bloom.git v0.0.0-20181019144233-44d6c5c71ed1
	github.com/BurntSushi/toml v1.2.1
	github.com/confio/ics23/go v0.9.0
	github.com/cosmos/iavl v0.20.0
	github.com/fxamacker/cbor/v2 v2.2.0
	github.com/hashcloak/Meson-plugin v0.0.0-20200627021923-d4745a3c9e02
	github.com/jackc/pgx v3.6.2+incompatible
	github.com/katzenpost/authority v0.0.14
	github.com/katzenpost/client v0.0.3
	github.com/katzenpost/core v0.0.12
	github.com/katzenpost/registration_client v0.0.1
	github.com/prometheus/client_golang v1.14.0
	github.com/spf13/cobra v1.6.1
	github.com/spf13/viper v1.14.0
	github.com/stretchr/testify v1.8.2
	github.com/tendermint/tendermint v0.37.0-rc2
	github.com/tendermint/tm-db v0.6.7
	github.com/ugorji/go/codec v1.2.7
	go.etcd.io/bbolt v1.3.7
	golang.org/x/net v0.9.0
	golang.org/x/text v0.9.0
	gopkg.in/eapache/channels.v1 v1.1.0
	gopkg.in/op/go-logging.v1 v1.0.0-20160211212156-b2cb9fa56473
)

require (
	github.com/DataDog/zstd v1.5.0 // indirect
	github.com/OneOfOne/xxhash v1.2.5 // indirect
	github.com/cosmos/cosmos-db v0.0.0-20221226095112-f3c38ecb5e32
	github.com/cosmos/cosmos-sdk v0.47.3
	github.com/katzenpost/noise v0.0.2 // indirect
	github.com/katzenpost/server v0.0.12
	github.com/shopspring/decimal v1.2.0 // indirect
)

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
