package config

import (
	"context"
	"fmt"
	"time"

	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
)

func (c *Config) UpdateTrust() error {
	client, err := rpchttp.New(c.Katzenmint.PrimaryAddress, "/websocket")
	if err != nil {
		return err
	}
	info, err := client.ABCIInfo(context.Background())
	if err != nil {
		return err
	}
	genesis, err := client.Genesis(context.Background())
	if err != nil {
		return err
	}
	blockHeight := info.Response.LastBlockHeight
	block, err := client.Block(context.Background(), &blockHeight)
	if err != nil {
		return err
	}
	if block == nil {
		return fmt.Errorf("couldn't find block: %d", blockHeight)
	}
	if genesis.Genesis.ChainID != c.Katzenmint.ChainID {
		return fmt.Errorf("wrong chain ID")
	}
	c.Katzenmint.TrustOptions.Period = 10 * time.Minute
	c.Katzenmint.TrustOptions.Height = blockHeight
	c.Katzenmint.TrustOptions.Hash = block.BlockID.Hash
	return nil
}
