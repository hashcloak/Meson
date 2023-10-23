package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/p2p"
)

var (
	showNodeIDCmd = &cobra.Command{
		Use:   "show-node-id",
		Short: "Show p2p id of the node",
		RunE:  showNodeID,
	}
)

func showNodeID(cmd *cobra.Command, args []string) error {
	_, config, err := initConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}
	nodeKey, err := p2p.LoadNodeKey(config.NodeKeyFile())
	if err != nil {
		return err
	}
	fmt.Println(nodeKey.ID())
	return nil
}
