// katzenmint pkiclient implementation

package pkiclient

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math"

	costypes "github.com/cosmos/cosmos-sdk/store/types"
	kpki "github.com/hashcloak/Meson/katzenmint"
	"github.com/hashcloak/Meson/katzenmint/s11n"
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

var (
	protocolVersion = "development"
)

type BroadcastTxError struct {
	Resp *ctypes.ResultBroadcastTxCommit
}

func (e BroadcastTxError) Error() string {
	if !e.Resp.CheckTx.IsOK() {
		return fmt.Sprintf("send transaction failed at checking tx: %v", e.Resp.CheckTx.Log)
	}
	return fmt.Sprintf("send transaction failed at delivering tx: %v", e.Resp.DeliverTx.Log)
}

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
		Version: protocolVersion,
		Epoch:   epoch,
		Command: command,
		Payload: "",
	}
	data, err := kpki.EncodeJson(query)
	if err != nil {
		return nil, fmt.Errorf("failed to encode data: %v", err)
	}
	//p.log.Debugf("Query: %v", query)

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
	p.log.Debugf("Query epoch")

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
	p.log.Debugf("Get document for epoch %d", epoch)

	// Make the query
	resp, err := p.query(ctx, epoch, kpki.GetConsensus)
	if err != nil {
		return nil, nil, err
	}
	if resp.Response.Code != 0 {
		if resp.Response.Code == kpki.ErrQueryNoDocument.Code {
			return nil, nil, cpki.ErrNoDocument
		}
		return nil, nil, fmt.Errorf(resp.Response.Log)
	}

	// Verify and parse the document
	doc, err := s11n.VerifyAndParseDocument(resp.Response.Value, epoch)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to extract doc: %v", err)
	}
	if doc.Epoch != epoch {
		p.log.Warningf("Get() returned pki document for wrong epoch: %v", doc.Epoch)
		return nil, nil, s11n.ErrInvalidEpoch
	}

	s := "Topology: [0]{"
	for l, nodes := range doc.Topology {
		for idx, v := range nodes {
			s += v.Name
			if idx != len(nodes)-1 {
				s += ","
			}
		}
		if l < len(doc.Topology)-1 {
			s += fmt.Sprintf("}, [%v]{", l+1)
		}
	}
	s += "}, Providers: "
	for _, v := range doc.Providers {
		s += v.Name + ","
	}
	p.log.Debugf("Document summary: " + s)

	return doc, resp.Response.Value, nil
}

// Post posts the node's descriptor to the PKI for the provided epoch.
func (p *PKIClient) Post(ctx context.Context, epoch uint64, signingKey *eddsa.PrivateKey, d *cpki.MixDescriptor) error {
	p.log.Debugf("Post descriptor for epoch %d: %v", epoch, d)

	// Ensure that the descriptor we are about to post is well formed.
	if err := s11n.IsDescriptorWellFormed(d, epoch); err != nil {
		return err
	}

	// Make a serialized + signed + serialized descriptor.
	signed, err := s11n.SignDescriptor(signingKey, d, epoch+s11n.CertificateExpiration)
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
		if _, ok := err.(BroadcastTxError); ok {
			return cpki.ErrInvalidPostEpoch
		}
		return err
	}
	return nil
}

// PostTx posts the transaction to the katzenmint node.
func (p *PKIClient) PostTx(ctx context.Context, tx []byte) (*ctypes.ResultBroadcastTxCommit, error) {

	// Broadcast the abci transaction
	resp, err := p.light.BroadcastTxCommit(ctx, tx)
	if err != nil {
		return nil, err
	}
	if !resp.CheckTx.IsOK() {
		return nil, BroadcastTxError{Resp: resp}
	}
	if !resp.DeliverTx.IsOK() {
		return nil, BroadcastTxError{Resp: resp}
	}
	return resp, nil
}

// Deserialize returns PKI document given the raw bytes.
func (p *PKIClient) Deserialize(raw []byte) (*cpki.Document, error) {
	// TODO: figure out a better way
	return s11n.VerifyAndParseDocument(raw, math.MaxUint64)
}

// DeserializeWithEpoch returns PKI document given the raw bytes.
func (p *PKIClient) DeserializeWithEpoch(raw []byte, epoch uint64) (*cpki.Document, error) {
	return s11n.VerifyAndParseDocument(raw, epoch)
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
	p.light.RegisterOpDecoder(costypes.ProofOpIAVLCommitment, costypes.CommitmentOpDecoder)
	return p, nil
}

// NewPKIClientFromLightClient create PKI Client from tendermint rpc light client
func NewPKIClientFromLightClient(light *lightrpc.Client, logBackend *log.Backend) (*PKIClient, error) {
	p := new(PKIClient)
	p.log = logBackend.GetLogger("pki/client")
	p.light = light
	p.light.RegisterOpDecoder(costypes.ProofOpIAVLCommitment, costypes.CommitmentOpDecoder)
	return p, nil
}

// Shutdown the client
func (p *PKIClient) Shutdown() {
	_ = p.light.Stop()
	_ = p.db.Close()
}
