package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	dbm "github.com/cosmos/cosmos-db"
	katzenmint "github.com/hashcloak/Meson/katzenmint"
	kcfg "github.com/hashcloak/Meson/katzenmint/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	abci "github.com/tendermint/tendermint/abci/types"
	cfg "github.com/tendermint/tendermint/config"
	tmflags "github.com/tendermint/tendermint/libs/cli/flags"
	"github.com/tendermint/tendermint/libs/log"
	nm "github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/proxy"
)

var (
	rootCmd = &cobra.Command{
		Use: "katzenmint",
	}
	runCmd = &cobra.Command{
		Use:   "run",
		Short: "Run katzenmint PKI node",
		RunE:  runNode,
	}
	configFile  string
	dbCacheSize int
)

func readTendermintConfig(tConfigFile string) (config *cfg.Config, err error) {
	config = cfg.DefaultConfig()
	config.SetRoot(filepath.Dir(filepath.Dir(tConfigFile)))
	viper.SetConfigFile(tConfigFile)
	if err = viper.ReadInConfig(); err != nil {
		err = fmt.Errorf("viper failed to read config file: %w", err)
		return
	}
	if err = viper.Unmarshal(config); err != nil {
		err = fmt.Errorf("viper failed to unmarshal config: %w", err)
		return
	}
	if err = config.ValidateBasic(); err != nil {
		err = fmt.Errorf("config is invalid: %w", err)
		return
	}
	return
}

func getRpcAddresses(config *cfg.Config) []string {
	var rpcAddr []string
	peers := append(
		strings.Split(config.P2P.PersistentPeers, ","),
		strings.Split(config.P2P.Seeds, ",")...,
	)
	for _, element := range peers {
		element = strings.Trim(element, " ")
		if element == "" {
			continue
		}
		parsed := strings.Split(element, "@")
		ipAddr := strings.Split(parsed[len(parsed)-1], ":")[0]
		rpcAddr = append(rpcAddr, "tcp://"+ipAddr+":26657")
	}
	return rpcAddr
}

func newTendermint(app abci.Application, config *cfg.Config, logger log.Logger) (node *nm.Node, err error) {

	// read private validator
	pv := privval.LoadFilePV(
		config.PrivValidatorKeyFile(),
		config.PrivValidatorStateFile(),
	)

	// read node key
	var nodeKey *p2p.NodeKey
	if nodeKey, err = p2p.LoadNodeKey(config.NodeKeyFile()); err != nil {
		return nil, fmt.Errorf("failed to load node's key: %w", err)
	}

	// create node
	node, err = nm.NewNode(
		config,
		pv,
		nodeKey,
		proxy.NewLocalClientCreator(app),
		nm.DefaultGenesisDocProviderFunc(config),
		nm.DefaultDBProvider,
		nm.DefaultMetricsProvider(config.Instrumentation),
		logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create new Tendermint node: %w", err)
	}

	return node, nil
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "katzenmint.toml", "Path to katzenmint.toml")
	runCmd.Flags().IntVar(&dbCacheSize, "dbcachesize", 100, "Cache size for katzenmint db")
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(registerValidatorCmd)
}

func initConfig() (kConfig *kcfg.Config, config *cfg.Config, err error) {
	kConfig, err = kcfg.LoadFile(configFile)
	if err != nil {
		return
	}
	config, err = readTendermintConfig(kConfig.TendermintConfigPath)
	return
}

func runNode(cmd *cobra.Command, args []string) error {
	kConfig, config, err := initConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}
	db, err := dbm.NewDB("katzenmint_db", dbm.GoLevelDBBackend, kConfig.DBPath)
	if err != nil {
		return fmt.Errorf("failed to open badger db: %v\ntry running with -tags badgerdb", err)
	}
	defer db.Close()

	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	if logger, err = tmflags.ParseLogLevel(config.LogLevel, logger, cfg.DefaultLogLevel); err != nil {
		return fmt.Errorf("failed to parse log level: %+v", err)
	}

	app := katzenmint.NewKatzenmintApplication(kConfig, db, dbCacheSize, logger)
	defer app.Close()

	node, err := newTendermint(app, config, logger)
	if err != nil {
		return err
	}
	if err = node.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(2)
	}
	defer func() {
		_ = node.Stop()
		node.Wait()
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	return nil
}
