// genconfig.go - Meson config generation.
// Copyright (C) 2020  Hashcloak.
// genconfig.go - Katzenpost self contained test network.
// Copyright (C) 2017  Yawning Angel, David Stainton.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	currencyConf "github.com/hashcloak/Meson-plugin/pkg/config"
	kConfig "github.com/hashcloak/Meson/katzenmint/config"
	sConfig "github.com/hashcloak/Meson/server/config"
	"github.com/katzenpost/core/crypto/ecdh"
	"github.com/katzenpost/core/crypto/eddsa"
	"github.com/katzenpost/core/crypto/rand"
	"github.com/tendermint/tendermint/crypto/tmhash"
	"github.com/tendermint/tendermint/light"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"

	cfg "github.com/tendermint/tendermint/config"
	tmos "github.com/tendermint/tendermint/libs/os"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"
)

const (
	basePort             = 30000
	nrNodes              = 6
	nrProviders          = 2
	minimumNodesPerLayer = 2
	nValidators          = 4
	dirPerm              = 0700
	initialHeight        = 0
	keyType              = types.ABCIPubKeyTypeEd25519
)

var currencyList = []*currencyConf.Config{
	&currencyConf.Config{
		Ticker:  "gor",
		RPCUser: "rpcuser",
		RPCPass: "rpcpassword",
		RPCURL:  "https://goerli.hashcloak.com",
	},
	&currencyConf.Config{
		Ticker:  "tbnb",
		RPCUser: "rpcuser",
		RPCPass: "rpcpassword",
		RPCURL:  "https://tbinance.hashcloak.com",
	},
	&currencyConf.Config{
		Ticker:  "rin",
		RPCUser: "rpcuser",
		RPCPass: "rpcpassword",
		RPCURL:  "https://rinkeby.hashcloak.com",
	},
}

type katzenpost struct {
	goBinDir    string
	baseDir     string
	outputDir   string
	authAddress string
	currency    int

	authIdentity    *eddsa.PrivateKey
	authPubIdentity string

	nodeConfigs []*sConfig.Config
	lastPort    uint16
	nodeIdx     int
	providerIdx int

	recipients       map[string]*ecdh.PublicKey
	nrProviders      int
	nrNodes          int
	nValidators      int
	trustOptions     light.TrustOptions
	chainID          string
	onlyMixNode      bool
	onlyProviderNode bool
	publicIPAddress  string
	nameOfSingleNode string
}

func (s *katzenpost) genProviderConfig(name string) (cfg *sConfig.Config, err error) {
	cfg, err = s.genMixNodeConfig(name)

	cfg.Server.Identifier = name
	cfg.Server.IsProvider = true

	cfg.Management = new(sConfig.Management)
	cfg.Management.Enable = true

	cfg.Provider = new(sConfig.Provider)
	loopCfg := new(sConfig.Kaetzchen)
	loopCfg.Capability = "loop"
	loopCfg.Endpoint = "+loop"
	cfg.Provider.Kaetzchen = append(cfg.Provider.Kaetzchen, loopCfg)

	/*
		keysvrCfg := new(sConfig.Kaetzchen)
		keysvrCfg.Capability = "keyserver"
		keysvrCfg.Endpoint = "+keyserver"
		cfg.Provider.Kaetzchen = append(cfg.Provider.Kaetzchen, keysvrCfg)
	*/

	cfg.Provider.EnableUserRegistrationHTTP = true
	userRegistrationPort := 10000 + s.lastPort
	cfg.Provider.UserRegistrationHTTPAddresses = []string{fmt.Sprintf("0.0.0.0:%d", userRegistrationPort)}
	cfg.Provider.AdvertiseUserRegistrationHTTPAddresses = []string{fmt.Sprintf("http://%s:%d", s.publicIPAddress, userRegistrationPort)}

	// Plugin configs
	// echo server
	pluginConf := make(map[string]interface{})
	pluginConf["log_dir"] = filepath.Join(s.baseDir, name)
	pluginConf["log_level"] = cfg.Logging.Level
	echoPlugin := sConfig.CBORPluginKaetzchen{
		Disable:        false,
		Capability:     "echo",
		Endpoint:       "+echo",
		Command:        s.goBinDir + "/echo_server",
		MaxConcurrency: 1,
		Config:         pluginConf,
	}
	cfg.Provider.CBORPluginKaetzchen = append(cfg.Provider.CBORPluginKaetzchen, &echoPlugin)

	// panda serever
	pluginConf = make(map[string]interface{})
	pluginConf["log_dir"] = filepath.Join(s.baseDir, name)
	pluginConf["log_level"] = cfg.Logging.Level
	pluginConf["fileStore"] = filepath.Join(s.baseDir, name, "/panda.storage")
	pandaPlugin := sConfig.CBORPluginKaetzchen{
		Disable:        false,
		Capability:     "panda",
		Endpoint:       "+panda",
		Command:        s.goBinDir + "/panda_server",
		MaxConcurrency: 1,
		Config:         pluginConf,
	}
	cfg.Provider.CBORPluginKaetzchen = append(cfg.Provider.CBORPluginKaetzchen, &pandaPlugin)

	// memspool
	pluginConf = make(map[string]interface{})
	// leaving this one out until it can be proven that it won't crash the spool plugin
	pluginConf["log_dir"] = filepath.Join(s.baseDir, name)
	pluginConf["data_store"] = filepath.Join(s.baseDir, name, "/memspool.storage")
	spoolPlugin := sConfig.CBORPluginKaetzchen{
		Disable:        false,
		Capability:     "spool",
		Endpoint:       "+spool",
		Command:        s.goBinDir + "/memspool",
		MaxConcurrency: 1,
		Config:         pluginConf,
	}
	cfg.Provider.CBORPluginKaetzchen = append(cfg.Provider.CBORPluginKaetzchen, &spoolPlugin)

	// Meson
	curConf := currencyList[s.currency]
	s.currency++
	if s.currency >= len(currencyList) {
		s.currency = 0
	}

	curConf.LogDir = filepath.Join(s.baseDir, name)
	curConf.LogLevel = cfg.Logging.Level
	pluginConf = make(map[string]interface{})
	pluginConf["f"] = filepath.Join(s.baseDir, name, "/currency.toml")
	pluginConf["log_dir"] = filepath.Join(s.baseDir, name)
	pluginConf["log_level"] = cfg.Logging.Level
	mesonPlugin := sConfig.CBORPluginKaetzchen{
		Disable:        false,
		Capability:     curConf.Ticker,
		Endpoint:       "+" + curConf.Ticker,
		Command:        s.goBinDir + "/Meson",
		MaxConcurrency: 1,
		Config:         pluginConf,
	}
	cfg.Provider.CBORPluginKaetzchen = append(cfg.Provider.CBORPluginKaetzchen, &mesonPlugin)

	// generate currency.toml
	_ = os.Mkdir(filepath.Join(s.outputDir, identifier(cfg)), dirPerm)
	fileName := filepath.Join(
		s.outputDir, identifier(cfg), "currency.toml",
	)
	f, err := os.Create(fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	// Serialize the descriptor.
	enc := toml.NewEncoder(f)
	err = enc.Encode(curConf)
	if err != nil {
		return nil, err
	}

	return cfg, cfg.FixupAndValidate()
}

func (s *katzenpost) genMixNodeConfig(name string) (cfg *sConfig.Config, err error) {

	priv := filepath.Join(s.outputDir, name, "identity.private.pem")
	public := filepath.Join(s.outputDir, name, "identity.public.pem")
	_, err = eddsa.Load(priv, public, rand.Reader)
	if err != nil {
		return nil, err
	}

	cfg = new(sConfig.Config)
	// PKI section
	cfg.PKI = &sConfig.PKI{
		Voting: &sConfig.Voting{
			ChainID:            s.chainID,
			TrustOptions:       s.trustOptions,
			PrimaryAddress:     s.authAddress,
			WitnessesAddresses: []string{s.authAddress},
			DatabaseName:       fmt.Sprintf("%s-db", name),
			DatabaseDir:        filepath.Join(s.outputDir, name),
			RPCAddress:         s.authAddress,
		},
	}
	// Server section.
	cfg.Server = new(sConfig.Server)
	cfg.Server.Identifier = name
	cfg.Server.Addresses = []string{
		fmt.Sprintf("0.0.0.0:%d", s.lastPort),
		// fmt.Sprintf(s.publicIPAddress+":%d", s.lastPort)
	}
	cfg.Server.AltAddresses = map[string][]string{
		"tcp4": {
			fmt.Sprintf(s.publicIPAddress+":%d", s.lastPort),
		},
	}
	cfg.Server.OnlyAdvertiseAltAddresses = true
	cfg.Server.DataDir = filepath.Join(s.baseDir, name)
	cfg.Server.IsProvider = false

	_ = os.Mkdir(cfg.Server.DataDir, dirPerm)

	// Debug section.
	cfg.Debug = new(sConfig.Debug)
	cfg.Debug.DisableRateLimit = true
	cfg.Debug.SendDecoyTraffic = false
	// This needs to be 1 because during runtime the default
	// value of this is the number of cores of the computer.
	// In docker this number is 1 but if one runs this script
	// on a personal computer the value becomes much more than
	// what docker is capable of handling. Take a look at:
	// https://github.com/katzenpost/server/blob/master/config/config.go#L261
	cfg.Debug.NumSphinxWorkers = 1

	cfg.Debug.ConnectTimeout = 120000 // 2 mins
	cfg.Debug.HandshakeTimeout = 600000

	// Logging section.
	cfg.Logging = new(sConfig.Logging)
	cfg.Logging.File = "katzenpost.log"
	cfg.Logging.Level = "DEBUG"

	return cfg, cfg.FixupAndValidate()
}

func (s *katzenpost) genNodeConfig(isProvider bool) error {

	var err error
	var cfg *sConfig.Config
	var name string

	if isProvider {
		name = fmt.Sprintf("provider-%d", s.providerIdx)
	} else {
		name = fmt.Sprintf("node-%d", s.nodeIdx)
	}

	if s.nameOfSingleNode != "" {
		name = s.nameOfSingleNode
	}

	_ = os.Mkdir(filepath.Join(s.outputDir, name), dirPerm)

	if isProvider {
		cfg, err = s.genProviderConfig(name)
		if err != nil {
			return err
		}
		s.providerIdx++
	} else {
		cfg, err = s.genMixNodeConfig(name)
		if err != nil {
			return err
		}
		s.nodeIdx++
	}

	s.nodeConfigs = append(s.nodeConfigs, cfg)
	s.lastPort++
	return nil
}

func (s *katzenpost) fetchKatzenmintInfo() error {
	c, err := rpchttp.New(s.authAddress, "/websocket")
	if err != nil {
		return err
	}
	info, err := c.ABCIInfo(context.Background())
	if err != nil {
		return err
	}
	genesis, err := c.Genesis(context.Background())
	if err != nil {
		return err
	}
	blockHeight := info.Response.LastBlockHeight
	block, err := c.Block(context.Background(), &blockHeight)
	if err != nil {
		return err
	}
	if block == nil {
		return fmt.Errorf("couldn't find block: %d", blockHeight)
	}
	s.chainID = genesis.Genesis.ChainID
	s.trustOptions = light.TrustOptions{
		Period: 10 * time.Minute,
		Height: blockHeight,
		Hash:   block.BlockID.Hash,
	}
	return nil
}

func (s *katzenpost) generateVotingMixnetConfigs() {
	if s.authAddress != "" {
		if err := s.fetchKatzenmintInfo(); err != nil {
			log.Fatalf("fetchKatzenmintInfo failed: %s", err)
		}
	}

	s.generateNodesOfMixnet()
}

func (s *katzenpost) generateNodesOfMixnet() {
	// Generate the katzenmint configs.
	if s.authAddress == "" {
		for i := 0; i < s.nValidators; i++ {
			if err := s.genValidatorConfig(i); err != nil {
				log.Fatalf("Failed to generate katzenmint config: %v", err)
			}
		}
	}

	// Generate the provider configs.
	for i := 0; i < s.nrProviders; i++ {
		if err := s.genNodeConfig(true); err != nil {
			log.Fatalf("Failed to generate provider config: %v", err)
		}
	}

	// Generate the node configs.
	for i := 0; i < s.nrNodes; i++ {
		if err := s.genNodeConfig(false); err != nil {
			log.Fatalf("Failed to generate node config: %v", err)
		}
	}
}

func (s *katzenpost) genValidatorConfig(index int) error {
	tmConfig := cfg.DefaultConfig()

	genVals := make([]types.GenesisValidator, s.nValidators)
	for i := 0; i < s.nValidators; i++ {
		nodeDirName := fmt.Sprintf("auth-%d", i)
		nodeDir := filepath.Join(s.outputDir, nodeDirName)
		tmConfig.SetRoot(nodeDir)

		err := os.MkdirAll(filepath.Join(nodeDir, "config"), dirPerm)
		if err != nil {
			_ = os.RemoveAll(nodeDir)
			return err
		}
		err = os.MkdirAll(filepath.Join(nodeDir, "data"), dirPerm)
		if err != nil {
			_ = os.RemoveAll(nodeDir)
			return err
		}

		if err := initFilesWithConfig(tmConfig); err != nil {
			return err
		}

		pvKeyFile := filepath.Join(nodeDir, tmConfig.BaseConfig.PrivValidatorKey)
		pvStateFile := filepath.Join(nodeDir, tmConfig.BaseConfig.PrivValidatorState)
		pv := privval.LoadFilePV(pvKeyFile, pvStateFile)

		pubKey, err := pv.GetPubKey()
		if err != nil {
			return fmt.Errorf("can't get pubkey: %w", err)
		}
		genVals[i] = types.GenesisValidator{
			Address: pubKey.Address(),
			PubKey:  pubKey,
			Power:   1,
			Name:    nodeDirName,
		}
	}

	// Generate genesis doc from generated validators
	chainID := "chain-" + tmrand.Str(6)
	genDoc := &types.GenesisDoc{
		ChainID:         chainID,
		GenesisTime:     tmtime.Now(),
		InitialHeight:   initialHeight,
		Validators:      genVals,
		ConsensusParams: types.DefaultConsensusParams(),
	}

	// Write genesis file.
	for i := 0; i < nValidators; i++ {
		nodeDirName := fmt.Sprintf("auth-%d", i)
		nodeDir := filepath.Join(s.outputDir, nodeDirName)
		if err := genDoc.SaveAs(filepath.Join(nodeDir, tmConfig.BaseConfig.Genesis)); err != nil {
			_ = os.RemoveAll(nodeDir)
			return err
		}
	}

	// Overwrite default config.
	for i := 0; i < nValidators; i++ {
		nodeDirName := fmt.Sprintf("auth-%d", i)
		nodeDir := filepath.Join(s.outputDir, nodeDirName)
		tmConfig.SetRoot(nodeDir)
		tmConfig.P2P.AllowDuplicateIP = true
		tmConfig.Moniker = nodeDirName
		cfgPath := filepath.Join(nodeDir, "config", "config.toml")
		cfg.WriteConfigFile(cfgPath, tmConfig)

		katConfig := kConfig.DefaultConfig()
		katConfig.DBPath = filepath.Join(nodeDir, "kdata")
		katConfig.TendermintConfigPath = cfgPath
		kConfig.SaveFile(filepath.Join(s.outputDir, nodeDirName, "katzenmint.toml"), katConfig)
	}

	log.Printf("initialized katzenmint of %v", index)
	s.chainID = chainID
	s.trustOptions = light.TrustOptions{
		Period: 1,
		Height: 1,
		Hash:   make([]byte, tmhash.Size),
	}
	s.authAddress = tmConfig.RPC.ListenAddress
	return nil
}

func initFilesWithConfig(config *cfg.Config) error {
	// private validator
	privValKeyFile := config.PrivValidatorKeyFile()
	privValStateFile := config.PrivValidatorStateFile()
	var pv *privval.FilePV
	if !tmos.FileExists(privValKeyFile) {
		pv = privval.GenFilePV(privValKeyFile, privValStateFile)
		pv.Save()
	}

	nodeKeyFile := config.NodeKeyFile()
	if !tmos.FileExists(nodeKeyFile) {
		if _, err := p2p.LoadOrGenNodeKey(nodeKeyFile); err != nil {
			return err
		}
	}

	// genesis file
	genFile := config.GenesisFile()
	if !tmos.FileExists(genFile) {
		genDoc := types.GenesisDoc{
			ChainID:         fmt.Sprintf("test-chain-%v", tmrand.Str(6)),
			GenesisTime:     tmtime.Now(),
			ConsensusParams: types.DefaultConsensusParams(),
		}
		pubKey, err := pv.GetPubKey()
		if err != nil {
			return fmt.Errorf("can't get pubkey: %w", err)
		}
		genDoc.Validators = []types.GenesisValidator{{
			Address: pubKey.Address(),
			PubKey:  pubKey,
			Power:   10,
		}}

		if err := genDoc.SaveAs(genFile); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	var err error
	nrNodes := flag.Int("n", nrNodes, "Number of mixes.")
	nrProviders := flag.Int("p", nrProviders, "Number of providers.")
	nValidators := flag.Int("v", nValidators, "Number of katzenmint validators.")
	authAddress := flag.String("a", "", "Katzenmint JSONRPC URI (when set this value, application will fetch PKI info from remote server).")
	goBinDir := flag.String("g", "/go/bin", "Path to golang bin.")
	baseDir := flag.String("b", "/conf", "Path to for DataDir in the config files.")
	outputDir := flag.String("o", "./output", "Output path of the generate config files.")
	mixNodeConfig := flag.Bool("node", false, "Only generate a mix node config.")
	providerNodeConfig := flag.Bool("provider", false, "Only generate a provider node config.")
	publicIPAddress := flag.String("ipv4", "127.0.0.1", "The public ipv4 address of the single node.")
	name := flag.String("name", "", "The name of the node.")
	authPubIdentity := flag.String("authID", "", "Authority public ID.")
	flag.Parse()
	s := &katzenpost{
		lastPort:   basePort + 1,
		recipients: make(map[string]*ecdh.PublicKey),
	}

	outDir, err := filepath.Abs(*outputDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(-1)
	}
	s.outputDir = outDir
	s.goBinDir = *goBinDir
	s.baseDir = *baseDir
	err = os.Mkdir(s.outputDir, dirPerm)
	if err != nil && err.(*os.PathError).Err.Error() != "file exists" {
		fmt.Fprintf(os.Stderr, "Failed to create output directory: %v\n", err)
		os.Exit(-1)
	}

	s.authAddress = *authAddress
	s.nrProviders = *nrProviders
	s.nrNodes = *nrNodes
	s.nValidators = *nValidators
	s.onlyMixNode = *mixNodeConfig
	s.onlyProviderNode = *providerNodeConfig
	s.publicIPAddress = *publicIPAddress
	s.nameOfSingleNode = *name
	s.authPubIdentity = *authPubIdentity

	if s.onlyMixNode || s.onlyProviderNode {
		if s.onlyMixNode && s.onlyProviderNode {
			fmt.Fprintf(os.Stderr, "Please specify one of either -node or -provider config\n")
			os.Exit(-1)
		}
		if s.nameOfSingleNode == "" {
			fmt.Fprintf(os.Stderr, "Name not provided to provide a name with the -name flag\n")
			os.Exit(-1)
		}
		if s.authPubIdentity == "" {
			fmt.Fprintf(os.Stderr, "Need to provide an authority identity\n")
			os.Exit(-1)
		}
		if s.authAddress == "" {
			fmt.Fprintf(os.Stderr, "Need to provide an katzenmint server\n")
			os.Exit(-1)
		}
		if err := s.fetchKatzenmintInfo(); err != nil {
			log.Fatalf("fetchKatzenmintInfo failed: %s", err)
		}
		err = s.genNodeConfig(s.onlyProviderNode)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(-1)
		}
	} else {
		s.generateVotingMixnetConfigs()
	}

	for _, v := range s.nodeConfigs {
		if err := saveCfg(outDir, v); err != nil {
			log.Fatalf("%s", err)
		}

	}
}

func configName(cfg interface{}) string {
	switch cfg.(type) {
	case *sConfig.Config:
		return "katzenpost.toml"
	default:
		log.Fatalf("identifier() passed unexpected type")
		return ""
	}
}

func identifier(cfg interface{}) string {
	switch cfg.(type) {
	case *sConfig.Config:
		return cfg.(*sConfig.Config).Server.Identifier
	default:
		log.Fatalf("identifier() passed unexpected type")
		return ""
	}
}

func saveCfg(outputDir string, cfg interface{}) error {
	_ = os.Mkdir(filepath.Join(outputDir, identifier(cfg)), dirPerm)

	fileName := filepath.Join(
		outputDir, identifier(cfg), configName(cfg),
	)
	log.Printf("saveCfg of %s", fileName)
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer f.Close()
	// Serialize the descriptor.
	enc := toml.NewEncoder(f)
	return enc.Encode(cfg)
}

// links between mix and providers
func (s *katzenpost) spk(a *sConfig.Config) *eddsa.PublicKey {
	priv := filepath.Join(s.outputDir, a.Server.Identifier, "identity.private.pem")
	public := filepath.Join(s.outputDir, a.Server.Identifier, "identity.public.pem")
	idKey, err := eddsa.Load(priv, public, rand.Reader)
	if err != nil {
		panic(err)
	}
	return idKey.PublicKey()
}
