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
	sConfig "github.com/hashcloak/Meson-server/config"
	aConfig "github.com/katzenpost/authority/nonvoting/server/config"
	"github.com/katzenpost/core/crypto/ecdh"
	"github.com/katzenpost/core/crypto/eddsa"
	"github.com/katzenpost/core/crypto/rand"
	"github.com/tendermint/tendermint/light"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
)

const (
	basePort             = 30000
	nrNodes              = 6
	nrProviders          = 2
	minimumNodesPerLayer = 2
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

	authConfig      *aConfig.Config
	authIdentity    *eddsa.PrivateKey
	authPubIdentity string

	nodeConfigs []*sConfig.Config
	lastPort    uint16
	nodeIdx     int
	providerIdx int

	recipients       map[string]*ecdh.PublicKey
	voting           bool
	nrProviders      int
	nrNodes          int
	vRPCURL          string
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
	_ = os.Mkdir(filepath.Join(s.outputDir, identifier(cfg)), 0700)
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

	if s.voting {
		cfg.PKI = &sConfig.PKI{
			Voting: &sConfig.Voting{
				ChainID:            s.chainID,
				TrustOptions:       s.trustOptions,
				PrimaryAddress:     s.vRPCURL,
				WitnessesAddresses: []string{s.vRPCURL},
				DatabaseName:       fmt.Sprintf("%s-db", name),
				DatabaseDir:        filepath.Join(s.outputDir, name),
				RPCAddress:         s.vRPCURL,
			},
		}
	} else {
		cfg.PKI = new(sConfig.PKI)
		cfg.PKI.Nonvoting = new(sConfig.Nonvoting)
		cfg.PKI.Nonvoting.Address = fmt.Sprintf(s.authAddress+":%d", basePort)
		idKey := []byte(s.authPubIdentity)
		if s.authIdentity != nil {
			idKey, err = s.authIdentity.PublicKey().MarshalText()
			if err != nil {
				return nil, err
			}
		}
		cfg.PKI.Nonvoting.PublicKey = string(idKey)
	}
	// Server section.
	cfg.Server = new(sConfig.Server)
	cfg.Server.Identifier = name
	cfg.Server.Addresses = []string{fmt.Sprintf("0.0.0.0:%d", s.lastPort)}
	cfg.Server.AltAddresses = map[string][]string{
		"tcp4": []string{fmt.Sprintf(s.publicIPAddress+":%d", s.lastPort)},
	}
	cfg.Server.OnlyAdvertiseAltAddresses = true
	cfg.Server.DataDir = filepath.Join(s.baseDir, name)
	cfg.Server.IsProvider = false

	_ = os.Mkdir(cfg.Server.DataDir, 0700)

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

func (s *katzenpost) genNodeConfig(isProvider bool, isVoting bool) error {

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

	_ = os.Mkdir(filepath.Join(s.outputDir, name), 0700)

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

func (s *katzenpost) genAuthConfig() error {
	authLogFile := s.baseDir + "/" + "authority.log"
	cfg := new(aConfig.Config)

	// Authority section.
	cfg.Authority = new(aConfig.Authority)
	cfg.Authority.Addresses = []string{fmt.Sprintf("0.0.0.0:%d", basePort)}
	cfg.Authority.DataDir = filepath.Join(s.baseDir)

	// Logging section.
	cfg.Logging = new(aConfig.Logging)
	cfg.Logging.File = authLogFile
	cfg.Logging.Level = "DEBUG"

	name := "nonvoting"
	_ = os.Mkdir(filepath.Join(s.outputDir, name), 0700)
	// Generate keys
	priv := filepath.Join(s.outputDir, name, "identity.private.pem")
	public := filepath.Join(s.outputDir, name, "identity.public.pem")
	idKey, err := eddsa.Load(priv, public, rand.Reader)
	s.authIdentity = idKey
	s.authPubIdentity = idKey.PublicKey().String()
	if err != nil {
		return err
	}

	// Debug section.
	cfg.Debug = new(aConfig.Debug)
	cfg.Debug.MinNodesPerLayer = minimumNodesPerLayer
	if cfg.Debug.MinNodesPerLayer > s.nrNodes {
		return fmt.Errorf("Not enough nodes to fill up each layer")
	}
	cfg.Debug.Layers = s.nrNodes / cfg.Debug.MinNodesPerLayer
	if err := cfg.FixupAndValidate(); err != nil {
		return err
	}
	s.authConfig = cfg
	return nil
}

func (s *katzenpost) generateWhitelist() ([]*aConfig.Node, []*aConfig.Node, error) {
	mixes := []*aConfig.Node{}
	providers := []*aConfig.Node{}
	for _, nodeCfg := range s.nodeConfigs {
		if nodeCfg.Server.IsProvider {
			provider := &aConfig.Node{
				Identifier:  nodeCfg.Server.Identifier,
				IdentityKey: s.spk(nodeCfg),
			}
			providers = append(providers, provider)
			continue
		}
		mix := &aConfig.Node{
			IdentityKey: s.spk(nodeCfg),
		}
		mixes = append(mixes, mix)
	}

	return providers, mixes, nil

}

func (s *katzenpost) generateNonVotingMixnetConfigs() {
	if err := s.genAuthConfig(); err != nil {
		log.Fatalf("Failed to generate authority config: %v", err)
	}

	s.generateNodesOfMixnet()
	// The node lists.
	if providers, mixes, err := s.generateWhitelist(); err == nil {
		s.authConfig.Mixes = mixes
		s.authConfig.Providers = providers
	} else {
		log.Fatalf("Failed to generateWhitelist with %s", err)
	}

	if err := saveCfg(s.outputDir, s.authConfig); err != nil {
		log.Fatalf("Failed to saveCfg of authority with %s", err)
	}
}

func (s *katzenpost) fetchKatzenmintInfo() error {
	c, err := rpchttp.New(s.vRPCURL, "/websocket")
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
	if err := s.fetchKatzenmintInfo(); err != nil {
		log.Fatalf("fetchKatzenmintInfo failed: %s", err)
	}

	s.generateNodesOfMixnet()
	// TODO: update katzenmint config?
}

func (s *katzenpost) generateNodesOfMixnet() {
	// Generate the provider configs.
	for i := 0; i < s.nrProviders; i++ {
		if err := s.genNodeConfig(true, s.voting); err != nil {
			log.Fatalf("Failed to generate provider config: %v", err)
		}
	}

	// Generate the node configs.
	for i := 0; i < s.nrNodes; i++ {
		if err := s.genNodeConfig(false, s.voting); err != nil {
			log.Fatalf("Failed to generate node config: %v", err)
		}
	}
}

func main() {
	var err error
	nrNodes := flag.Int("n", nrNodes, "Number of mixes.")
	nrProviders := flag.Int("p", nrProviders, "Number of providers.")
	voting := flag.Bool("v", false, "Generate voting configuration.")
	vRPCURL := flag.String("vrpc", "", "Voting RPC URL of katzenmint.")
	goBinDir := flag.String("g", "/go/bin", "Path to golang bin.")
	baseDir := flag.String("b", "/conf", "Path to for DataDir in the config files.")
	outputDir := flag.String("o", "./output", "Output path of the generate config files.")
	authAddress := flag.String("a", "127.0.0.1", "Non-voting authority public ip address.")
	mixNodeConfig := flag.Bool("node", false, "Only generate a mix node config.")
	providerNodeConfig := flag.Bool("provider", false, "Only generate a provider node config.")
	publicIPAddress := flag.String("ipv4", "", "The public ipv4 address of the single node.")
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
	err = os.Mkdir(s.outputDir, 0700)
	if err != nil && err.(*os.PathError).Err.Error() != "file exists" {
		fmt.Fprintf(os.Stderr, "Failed to create output directory: %v\n", err)
		os.Exit(-1)
	}

	s.authAddress = *authAddress
	s.voting = *voting
	s.nrProviders = *nrProviders
	s.nrNodes = *nrNodes
	s.vRPCURL = *vRPCURL
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
		if s.publicIPAddress == "" {
			fmt.Fprintf(os.Stderr, "Ip address not provided to provide a name with the -name flag\n")
			os.Exit(-1)
		}
		err = s.genNodeConfig(s.onlyProviderNode, *voting)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(-1)
		}
	} else {
		if s.publicIPAddress == "" {
			s.publicIPAddress = s.authAddress
		}
		if *voting {
			s.generateVotingMixnetConfigs()
		} else {
			s.generateNonVotingMixnetConfigs()
		}
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
	case *aConfig.Config:
		return "authority.toml"
	default:
		log.Fatalf("identifier() passed unexpected type")
		return ""
	}
}

func identifier(cfg interface{}) string {
	switch cfg.(type) {
	case *sConfig.Config:
		return cfg.(*sConfig.Config).Server.Identifier
	case *aConfig.Config:
		return "nonvoting"
	default:
		log.Fatalf("identifier() passed unexpected type")
		return ""
	}
}

func saveCfg(outputDir string, cfg interface{}) error {
	_ = os.Mkdir(filepath.Join(outputDir, identifier(cfg)), 0700)

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
