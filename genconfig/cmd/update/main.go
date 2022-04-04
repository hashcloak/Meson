package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	sConfig "github.com/hashcloak/Meson/server/config"
	"github.com/tendermint/tendermint/light"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
)

type KatzenmintInfo struct {
	ChainID      string
	TrustOptions light.TrustOptions
}

func fetchKatzenmintInfo(vRPCURL string) (*KatzenmintInfo, error) {
	c, err := rpchttp.New(vRPCURL, "/websocket")
	if err != nil {
		return nil, err
	}
	info, err := c.ABCIInfo(context.Background())
	if err != nil {
		return nil, err
	}
	genesis, err := c.Genesis(context.Background())
	if err != nil {
		return nil, err
	}
	blockHeight := info.Response.LastBlockHeight
	block, err := c.Block(context.Background(), &blockHeight)
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, fmt.Errorf("couldn't find block: %d", blockHeight)
	}
	kInfo := new(KatzenmintInfo)
	kInfo.ChainID = genesis.Genesis.ChainID
	kInfo.TrustOptions = light.TrustOptions{
		Period: 10 * time.Minute,
		Height: blockHeight,
		Hash:   block.BlockID.Hash,
	}
	return kInfo, nil
}

func saveConfig(fileName string, cfg interface{}) error {
	log.Printf("saveCfg of %s", fileName)
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := toml.NewEncoder(f)
	return enc.Encode(cfg)
}

func main() {
	configFile := flag.String("f", "katzenpost.toml", "Config name.")
	flag.Parse()

	if _, err := os.Stat(*configFile); err != nil {
		panic(err)
	}
	cfg, err := sConfig.LoadFile(*configFile)
	if err != nil {
		panic(err)
	}
	if cfg == nil || cfg.PKI == nil || cfg.PKI.Voting == nil {
		panic("failed to parse config file")
	}
	log.Println("Loaded config, start to fetch katzenmint info")
	info, err := fetchKatzenmintInfo(cfg.PKI.Voting.RPCAddress)
	if err != nil {
		panic(err)
	}

	if info.ChainID != cfg.PKI.Voting.ChainID {
		panic("chain id was not the same")
	}
	log.Printf("Latest katzenmint info: %+v", info)
	cfg.PKI.Voting.TrustOptions = info.TrustOptions
	err = saveConfig(*configFile, cfg)
	if err != nil {
		panic(err)
	}
}
