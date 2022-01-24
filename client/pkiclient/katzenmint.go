// katzenmint pkiclient implementation

package pkiclient

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/cosmos/iavl"
	kpki "github.com/hashcloak/katzenmint-pki"
	"github.com/hashcloak/katzenmint-pki/s11n"
	"github.com/katzenpost/core/crypto/eddsa"
	"github.com/katzenpost/core/log"
	cpki "github.com/katzenpost/core/pki"
	"github.com/tendermint/tendermint/crypto/merkle"
	"github.com/tendermint/tendermint/light"
	lightrpc "github.com/tendermint/tendermint/light/rpc"
	dbs "github.com/tendermint/tendermint/light/store/db"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	"github.com/tendermint/tendermint/rpc/client/http"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	dbm "github.com/tendermint/tm-db"
	"gopkg.in/op/go-logging.v1"
)

type PKIClientConfig struct {
	LogBackend         *log.Backend
	ChainID            string
	TrustOptions       light.TrustOptions
	PrimaryAddress     string
	WitnessesAddresses []string
	DatabaseName       string
	DatabaseDir        string
	RPCAddress         string
}

type PKIClient struct {
	// TODO: do we need katzenpost pki client interface?
	// cpki.Client
	light *lightrpc.Client
	log   *logging.Logger

	// TODO: should care about cache client?
	db dbm.DB
}

func (p *PKIClient) query(ctx context.Context, epoch uint64, command kpki.Command) (*ctypes.ResultABCIQuery, error) {
	// Form the abci query
	query := kpki.Query{
		Version: kpki.ProtocolVersion,
		Epoch:   epoch,
		Command: command,
		Payload: "",
	}
	data, err := kpki.EncodeJson(query)
	if err != nil {
		return nil, fmt.Errorf("failed to encode data: %v", err)
	}
	p.log.Debugf("Query: %v", query)

	// Make the abci query
	opts := rpcclient.ABCIQueryOptions{Prove: true}
	resp, err := p.light.ABCIQueryWithOptions(ctx, "", data, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to query katzenmint pki: %v", err)
	}
	return resp, nil
}

// GetEpoch returns the epoch information of PKI.
func (p *PKIClient) GetEpoch(ctx context.Context) (epoch uint64, ellapsedHeight uint64, err error) {
	resp, err := p.query(ctx, 0, kpki.GetEpoch)
	if err != nil {
		return
	}
	if resp.Response.Code != 0 {
		err = errors.New(resp.Response.Log)
		return
	}
	if len(resp.Response.Value) != 16 {
		err = fmt.Errorf("retrieved epoch information has incorrect format")
		return
	}
	epoch, _ = binary.Uvarint(resp.Response.Value[:8])
	startingHeight, _ := binary.Varint(resp.Response.Value[8:16])
	if startingHeight > resp.Response.Height {
		err = fmt.Errorf("retrieved starting height is more than the corresponding block height")
		return
	}
	ellapsedHeight = uint64(resp.Response.Height - startingHeight)
	return
}

// GetDoc returns the PKI document along with the raw serialized form for the provided epoch.
func (p *PKIClient) GetDoc(ctx context.Context, epoch uint64) (*cpki.Document, []byte, error) {
	p.log.Debugf("Get(ctx, %d)", epoch)

	// Make the query
	resp, err := p.query(ctx, epoch, kpki.GetConsensus)
	if err != nil {
		return nil, nil, err
	}
	if resp.Response.Code != 0 {
		if resp.Response.Code == 0x5 { //! Warning: MAGIC error code
			return nil, nil, cpki.ErrNoDocument
		}
		return nil, nil, fmt.Errorf(resp.Response.Log)
	}

	// Verify and parse the document
	doc, err := s11n.VerifyAndParseDocument(resp.Response.Value)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to extract doc: %v", err)
	}
	if doc.Epoch != epoch {
		p.log.Warningf("Get() returned pki document for wrong epoch: %v", doc.Epoch)
		return nil, nil, s11n.ErrInvalidEpoch
	}
	p.log.Debugf("Document: %v", doc)

	return doc, resp.Response.Value, nil
}

// Post posts the node's descriptor to the PKI for the provided epoch.
func (p *PKIClient) Post(ctx context.Context, epoch uint64, signingKey *eddsa.PrivateKey, d *cpki.MixDescriptor) error {
	p.log.Debugf("Post(ctx, %d, %v, %+v)", epoch, signingKey.PublicKey(), d)

	// Ensure that the descriptor we are about to post is well formed.
	if err := s11n.IsDescriptorWellFormed(d, epoch); err != nil {
		return err
	}

	// Make a serialized + signed + serialized descriptor.
	signed, err := s11n.SignDescriptor(signingKey, d)
	if err != nil {
		return err
	}

	// Form the abci transaction
	tx, err := kpki.FormTransaction(kpki.PublishMixDescriptor, epoch, kpki.EncodeHex(signed), signingKey)
	if err != nil {
		return err
	}

	// Post the abci transaction
	_, err = p.PostTx(ctx, tx)
	if err != nil {
		return err
	}
	return nil

	// TODO: Make sure to subscribe for events
	// Parse the post_descriptor_status command.
	/*
		r, ok := resp.(*commands.PostDescriptorStatus)
		if !ok {
			return fmt.Errorf("nonvoting/client: Post() unexpected reply: %T", resp)
		}
		switch r.ErrorCode {
		case commands.DescriptorOk:
			return nil
		case commands.DescriptorConflict:
			return cpki.ErrInvalidPostEpoch
		default:
			return fmt.Errorf("nonvoting/client: Post() rejected by authority: %v", postErrorToString(r.ErrorCode))
		}
	*/
}

// PostTx posts the transaction to the katzenmint node.
func (p *PKIClient) PostTx(ctx context.Context, tx []byte) (*ctypes.ResultBroadcastTxCommit, error) {

	// Broadcast the abci transaction
	resp, err := p.light.BroadcastTxCommit(ctx, tx)
	if err != nil {
		return nil, err
	}
	if !resp.CheckTx.IsOK() {
		return nil, fmt.Errorf("send transaction failed at checking tx: %v", resp.CheckTx.Log)
	}
	if !resp.DeliverTx.IsOK() {
		return nil, fmt.Errorf("send transaction failed at delivering tx: %v", resp.DeliverTx.Log)
	}
	return resp, nil
}

// Deserialize returns PKI document given the raw bytes.
func (p *PKIClient) Deserialize(raw []byte) (*cpki.Document, error) {
	return s11n.VerifyAndParseDocument(raw)
}

// NewPKIClient create PKI Client from PKI config
func NewPKIClient(cfg *PKIClientConfig) (*PKIClient, error) {
	p := new(PKIClient)
	p.log = cfg.LogBackend.GetLogger("pki/client")

	db, err := dbm.NewDB(cfg.DatabaseName, dbm.GoLevelDBBackend, cfg.DatabaseDir)
	if err != nil {
		return nil, fmt.Errorf("error opening katzenmint-pki database: %v", err)
	}
	p.db = db
	lightclient, err := light.NewHTTPClient(
		context.Background(),
		cfg.ChainID,
		cfg.TrustOptions,
		cfg.PrimaryAddress,
		cfg.WitnessesAddresses,
		dbs.New(db, "katzenmint"),
	)
	if err != nil {
		return nil, fmt.Errorf("error initialization of katzenmint-pki light client: %v", err)
	}
	provider, err := http.New(cfg.RPCAddress, "/websocket")
	if err != nil {
		return nil, fmt.Errorf("error connection to katzenmint-pki full node: %v", err)
	}
	kpFunc := lightrpc.KeyPathFn(func(_ string, key []byte) (merkle.KeyPath, error) {
		kp := merkle.KeyPath{}
		kp = kp.AppendKey(key, merkle.KeyEncodingURL)
		return kp, nil
	})
	p.light = lightrpc.NewClient(provider, lightclient, kpFunc)
	p.light.RegisterOpDecoder(iavl.ProofOpIAVLValue, iavl.ValueOpDecoder)
	return p, nil
}

// NewPKIClientFromLightClient create PKI Client from tendermint rpc light client
func NewPKIClientFromLightClient(light *lightrpc.Client, logBackend *log.Backend) (*PKIClient, error) {
	p := new(PKIClient)
	p.log = logBackend.GetLogger("pki/client")
	p.light = light
	p.light.RegisterOpDecoder(iavl.ProofOpIAVLValue, iavl.ValueOpDecoder)
	return p, nil
}

// Shutdown the client
func (p *PKIClient) Shutdown() {
	_ = p.light.Stop()
	_ = p.db.Close()
}
