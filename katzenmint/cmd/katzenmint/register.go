package main

import (
	"context"
	"encoding/binary"
	"fmt"

	katzenmint "github.com/hashcloak/Meson/katzenmint"
	"github.com/katzenpost/core/crypto/eddsa"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/rpc/client/http"
)

var (
	registerValidatorCmd = &cobra.Command{
		Use:   "register",
		Short: "Register validator for katzenmint PKI",
		RunE:  registerValidator,
	}
)

func joinNetwork(config *cfg.Config) error {
	// connect to peers
	var rpc *http.HTTP
	var err error
	for _, addr := range getRpcAddresses(config) {
		rpc, err = http.New(addr, "/websocket")
		if err != nil {
			fmt.Printf("%v\n", err)
		} else {
			break
		}
	}
	if rpc == nil {
		return fmt.Errorf("cannot connect and broadcast to peers")
	}

	// load keys
	pv := privval.LoadFilePV(
		config.PrivValidatorKeyFile(),
		config.PrivValidatorStateFile(),
	)
	privKey := new(eddsa.PrivateKey)
	err = privKey.FromBytes(pv.Key.PrivKey.Bytes())
	if err != nil {
		return err
	}

	// prepare AddAuthority transaction
	raw, err := katzenmint.EncodeJson(katzenmint.Authority{
		Auth:    config.Moniker,
		Power:   1,
		PubKey:  pv.Key.PubKey.Bytes(),
		KeyType: pv.Key.PubKey.Type(),
	})
	if err != nil {
		return err
	}
	info, err := rpc.ABCIInfo(context.Background())
	if err != nil {
		return err
	}
	epoch, _ := binary.Uvarint(katzenmint.DecodeHex(info.Response.Data))
	tx, err := katzenmint.FormTransaction(katzenmint.AddNewAuthority, epoch, string(raw), privKey)
	if err != nil {
		return err
	}

	// post transaction
	resp, err := rpc.BroadcastTxSync(context.Background(), tx)
	if err != nil {
		return err
	}
	if resp.Code != abci.CodeTypeOK {
		return fmt.Errorf("broadcast tx error: %v", resp.Log)
	}
	fmt.Printf("transaction sent: %+v", resp)
	return nil
}

func registerValidator(cmd *cobra.Command, args []string) error {
	err := joinNetwork(config)
	if err != nil {
		return err
	}
	return nil
}
