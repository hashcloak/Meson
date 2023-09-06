module github.com/hashcloak/Meson

go 1.16

require (
	git.schwanenlied.me/yawning/aez.git v0.0.0-20180408160647-ec7426b44926
	git.schwanenlied.me/yawning/avl.git v0.0.0-20180224045358-04c7c776e391
	git.schwanenlied.me/yawning/bloom.git v0.0.0-20181019144233-44d6c5c71ed1
	github.com/BurntSushi/toml v1.2.1
	github.com/confio/ics23/go v0.7.0
	github.com/cosmos/iavl v1.0.0-beta.2
	github.com/fxamacker/cbor/v2 v2.2.0
	github.com/hashcloak/Meson-plugin v0.0.0-20200627021923-d4745a3c9e02
	github.com/jackc/pgx v3.6.2+incompatible
	github.com/katzenpost/authority v0.0.14
	github.com/katzenpost/client v0.0.3
	github.com/katzenpost/core v0.0.12
	github.com/katzenpost/registration_client v0.0.1
	github.com/prometheus/client_golang v1.16.0
	github.com/spf13/cobra v1.7.0
	github.com/spf13/viper v1.16.0
	github.com/stretchr/testify v1.8.4
	github.com/tendermint/tendermint v0.34.22
	github.com/tendermint/tm-db v0.6.7
	github.com/ugorji/go/codec v1.2.7
	go.etcd.io/bbolt v1.3.7
	golang.org/x/net v0.12.0
	golang.org/x/text v0.11.0
	gopkg.in/eapache/channels.v1 v1.1.0
	gopkg.in/op/go-logging.v1 v1.0.0-20160211212156-b2cb9fa56473
)

require github.com/dgraph-io/ristretto v0.1.1 // indirect

require (
	cosmossdk.io/log v1.1.1-0.20230704160919-88f2c830b0ca // indirect
	github.com/Azure/go-ansiterm v0.0.0-20230124172434-306776ec8161 // indirect
	github.com/DataDog/zstd v1.5.5 // indirect
	github.com/Microsoft/go-winio v0.6.0 // indirect
	github.com/OneOfOne/xxhash v1.2.5 // indirect
	github.com/cloudflare/circl v1.3.1 // indirect
	github.com/cockroachdb/errors v1.10.0 // indirect
	github.com/cockroachdb/pebble v0.0.0-20230711190327-88bbab59ff4f // indirect
	github.com/cosmos/cosmos-db v1.0.0
	github.com/cosmos/go-bip39 v1.0.0 // indirect
	github.com/cosmos/gogoproto v1.4.10 // indirect
	github.com/dgryski/go-farm v0.0.0-20200201041132-a6ae2369ad13 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/emicklei/dot v1.5.0 // indirect
	github.com/getsentry/sentry-go v0.22.0 // indirect
	github.com/go-logfmt/logfmt v0.6.0 // indirect
	github.com/gofrs/uuid v4.4.0+incompatible // indirect
	github.com/katzenpost/noise v0.0.2 // indirect
	github.com/katzenpost/server v0.0.12
	github.com/klauspost/compress v1.16.7 // indirect
	github.com/lib/pq v1.10.7 // indirect
	github.com/linxGnu/grocksdb v1.8.0 // indirect
	github.com/opencontainers/image-spec v1.1.0-rc2 // indirect
	github.com/petermattis/goid v0.0.0-20230518223814-80aa455d8761 // indirect
	github.com/prometheus/common v0.44.0 // indirect
	github.com/prometheus/procfs v0.11.0 // indirect
	github.com/rogpeppe/go-internal v1.11.0 // indirect
	github.com/rs/cors v1.8.3 // indirect
	github.com/shopspring/decimal v1.2.0 // indirect
	github.com/syndtr/goleveldb v1.0.1-0.20220721030215-126854af5e6d // indirect
	golang.org/x/exp v0.0.0-20230711153332-06a737ee72cb // indirect
	golang.org/x/sync v0.3.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230711160842-782d3b101e98 // indirect
)

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
