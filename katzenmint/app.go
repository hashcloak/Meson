package katzenmint

import (
	"crypto/ed25519"
	"fmt"
	"sort"

	dbm "github.com/cometbft/cometbft-db"
	costypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/hashcloak/Meson/katzenmint/cert"
	"github.com/hashcloak/Meson/katzenmint/config"
	"github.com/hashcloak/Meson/katzenmint/s11n"
	"github.com/katzenpost/core/pki"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmcrypto "github.com/tendermint/tendermint/proto/tendermint/crypto"
	"github.com/tendermint/tendermint/version"
	"github.com/ugorji/go/codec"
	// cryptoenc "github.com/tendermint/tendermint/crypto/encoding"
)

var (
	_               abcitypes.Application = (*KatzenmintApplication)(nil)
	protocolVersion string                = "development"
	buildTime       string                = "2021"
	appVersion      uint64                = 1
)

type katzenmintVersion struct {
	Protocol   string
	Tendermint string
	ABCI       string
	BuildTime  string
}

type KatzenmintApplication struct {
	state *KatzenmintState

	logger log.Logger
}

func NewKatzenmintApplication(kConfig *config.Config, db dbm.DB, dbCacheSize int, logger log.Logger) *KatzenmintApplication {
	state := NewKatzenmintState(kConfig, db, dbCacheSize)
	return &KatzenmintApplication{
		state:  state,
		logger: logger,
	}
}

func (app *KatzenmintApplication) Close() {
	app.state.Close()
}

func (app *KatzenmintApplication) Info(req abcitypes.RequestInfo) (res abcitypes.ResponseInfo) {
	ver := katzenmintVersion{protocolVersion, version.TMCoreSemVer, version.ABCIVersion, buildTime}
	verJSON, err := EncodeJson(ver)
	res = abcitypes.ResponseInfo{
		Data:             fmt.Sprint(app.state.currentEpoch),
		AppVersion:       appVersion,
		LastBlockHeight:  app.state.blockHeight,
		LastBlockAppHash: app.state.appHash,
	}
	if err != nil {
		return
	}
	res.Version = string(verJSON)
	return
}

func (app *KatzenmintApplication) SetOption(req abcitypes.RequestSetOption) abcitypes.ResponseSetOption {
	return abcitypes.ResponseSetOption{}
}

func (app *KatzenmintApplication) isTxValid(rawTx []byte) (
	tx *Transaction,
	payload []byte,
	desc *pki.MixDescriptor,
	doc *pki.Document,
	auth *AuthorityChecked,
	err error,
) {
	// decode raw into transcation
	tx = new(Transaction)
	dec := codec.NewDecoderBytes(rawTx, jsonHandle)
	if err = dec.Decode(tx); err != nil {
		err = ErrTxIsNotValidJSON
		return
	}

	// verify transaction signature
	if len(tx.PublicKey) != ed25519.PublicKeySize*2 {
		err = ErrTxWrongPublicKeySize
		return
	}
	if len(tx.Signature) != ed25519.SignatureSize*2 {
		err = ErrTxWrongSignatureSize
		return
	}
	if !tx.IsVerified() {
		err = ErrTxWrongSignature
		return
	}

	// command specific checks
	switch tx.Command {
	case PublishMixDescriptor:
		var verifier cert.Verifier
		payload = DecodeHex(tx.Payload)
		verifier, err = s11n.GetVerifierFromDescriptor(payload)
		if err != nil {
			err = ErrTxDescInvalidVerifier
			return
		}
		desc, err = s11n.VerifyAndParseDescriptor(verifier, payload, tx.Epoch, app.state.currentEpoch)
		if err != nil {
			err = ErrTxDescFalseVerification
			return
		}
		if !app.state.isDescriptorAuthorized(desc) {
			err = ErrTxDescNotAuthorized
			return
		}

	case AddNewAuthority:
		payload = []byte(tx.Payload)
		auth, err = VerifyAndParseAuthority(payload)
		if err != nil {
			err = ErrTxAuthorityParse
			return
		}
		if !app.state.isAuthorityNew(auth) {
			err = ErrTxAuthorityExists
			return
		}
		if !app.state.isAuthorityAuthorized(tx.Address(), auth) {
			err = ErrTxAuthorityNotAuthorized
			return
		}
	default:
		err = ErrTxCommandNotFound
	}

	return
}

func (app *KatzenmintApplication) executeTx(
	tx *Transaction, payload []byte,
	desc *pki.MixDescriptor,
	doc *pki.Document,
	auth *AuthorityChecked,
) error {
	// check for the epoch relative to the current epoch
	if tx.Epoch < app.state.currentEpoch || tx.Epoch >= app.state.currentEpoch+LifeCycle {
		return ErrTxWrongEpoch
	}
	switch tx.Command {
	case PublishMixDescriptor:
		err := app.state.updateMixDescriptor(payload, desc, tx.Epoch)
		if err != nil {
			app.logger.Error("failed to publish descriptor", "epoch", tx.Epoch, "error", err)
			return ErrTxUpdateDesc
		}
	case AddNewAuthority:
		err := app.state.updateAuthority(payload, *auth.Val)
		if err != nil {
			app.logger.Error("failed to add new authority", "epoch", app.state.currentEpoch, "error", err)
			return ErrTxUpdateAuth
		}
	default:
		return ErrTxCommandNotFound
	}
	// Unreached
	return nil
}

func (app *KatzenmintApplication) DeliverTx(req abcitypes.RequestDeliverTx) abcitypes.ResponseDeliverTx {
	tx, payload, desc, doc, auth, err := app.isTxValid(req.Tx)
	if err != nil {
		return abcitypes.ResponseDeliverTx{Code: err.(KatzenmintError).Code, Log: err.(KatzenmintError).Msg}
	}
	err = app.executeTx(tx, payload, desc, doc, auth)
	if err != nil {
		return abcitypes.ResponseDeliverTx{Code: err.(KatzenmintError).Code, Log: err.(KatzenmintError).Msg}
	}
	return abcitypes.ResponseDeliverTx{Code: abcitypes.CodeTypeOK}
}

// TODO: gas formula
func (app *KatzenmintApplication) CheckTx(req abcitypes.RequestCheckTx) abcitypes.ResponseCheckTx {
	_, _, _, _, _, err := app.isTxValid(req.Tx)
	if err != nil {
		return abcitypes.ResponseCheckTx{Code: err.(KatzenmintError).Code, Log: err.(KatzenmintError).Msg, GasWanted: 1}
	}
	return abcitypes.ResponseCheckTx{Code: abcitypes.CodeTypeOK, GasWanted: 1}
}

// TODO: should update the validators map after commit
func (app *KatzenmintApplication) Commit() abcitypes.ResponseCommit {
	appHash, err := app.state.Commit()
	if err != nil {
		// swallow insufficient descriptor / provider error
		if err == errDocInsufficientDescriptor || err == errDocInsufficientProvider {
			app.logger.Debug("insufficient error", "epoch", app.state.currentEpoch, "height", app.state.blockHeight, "error", err)
		} else {
			app.logger.Error("commit failed", "epoch", app.state.currentEpoch, "height", app.state.blockHeight, "error", err)
		}
	}
	return abcitypes.ResponseCommit{Data: appHash}
}

func (app *KatzenmintApplication) Query(rquery abcitypes.RequestQuery) (resQuery abcitypes.ResponseQuery) {

	kquery := new(Query)
	dec := codec.NewDecoderBytes(rquery.Data, jsonHandle)
	if err := dec.Decode(kquery); err != nil {
		parseErrorResponse(ErrQueryInvalidFormat, &resQuery)
		return
	}

	switch kquery.Command {
	default:
		parseErrorResponse(ErrQueryUnsupported, &resQuery)
		return
	case GetEpoch:
		resQuery.Height = app.state.blockHeight - 1
		val, proof, err := app.state.GetEpoch(resQuery.Height)
		if err != nil {
			if err == errStateClosed {
				parseErrorResponse(ErrQueryAppClosed, &resQuery)
			} else {
				app.logger.Error("peer: failed to retrieve epoch for height", "height", resQuery.Height, "error", err)
				parseErrorResponse(ErrQueryEpochFailed, &resQuery)
			}
			return
		}
		existProof := proof.GetExist()
		if existProof == nil {
			parseErrorResponse(ErrQueryEpochFailed, &resQuery)
			return
		}
		op := costypes.NewIavlCommitmentOp(existProof.Key, proof)
		resQuery.Key = existProof.Key
		resQuery.Value = val
		resQuery.ProofOps = &tmcrypto.ProofOps{
			Ops: []tmcrypto.ProofOp{op.ProofOp()},
		}

	case GetConsensus:
		resQuery.Height = app.state.blockHeight - 1
		doc, proof, err := app.state.GetDocument(kquery.Epoch, resQuery.Height)
		if err != nil {
			if err == errStateClosed {
				parseErrorResponse(ErrQueryAppClosed, &resQuery)
			} else if err == ErrQueryNoDocument {
				app.logger.Error("warn: detected a skipped document", "miss", kquery.Epoch, "now", app.state.currentEpoch)
				parseErrorResponse(err.(KatzenmintError), &resQuery)
			} else if err == ErrQueryDocumentNotReady {
				parseErrorResponse(err.(KatzenmintError), &resQuery)
			} else {
				parseErrorResponse(ErrQueryDocumentUnknown, &resQuery)
			}
			return
		}
		existProof := proof.GetExist()
		if existProof == nil {
			parseErrorResponse(ErrQueryDocumentUnknown, &resQuery)
			return
		}
		op := costypes.NewIavlCommitmentOp(existProof.Key, proof)
		resQuery.Key = existProof.Key
		resQuery.Value = doc
		resQuery.ProofOps = &tmcrypto.ProofOps{
			Ops: []tmcrypto.ProofOp{op.ProofOp()},
		}
	}
	return
}

func (app *KatzenmintApplication) InitChain(req abcitypes.RequestInitChain) abcitypes.ResponseInitChain {
	if app.state.currentEpoch != GenesisEpoch ||
		app.state.blockHeight != app.state.epochStartHeight {
		panic("state is already initialized")
	}
	app.state.BeginBlock()
	sort.Sort(abcitypes.ValidatorUpdates(req.Validators))
	for _, v := range req.Validators {
		err := app.state.updateAuthority(nil, v)
		if err != nil {
			app.logger.Error("failed to update validators", "error", err)
		}
	}
	return abcitypes.ResponseInitChain{
		Validators: req.Validators,
	}
}

// Track the block hash and header information
func (app *KatzenmintApplication) BeginBlock(req abcitypes.RequestBeginBlock) abcitypes.ResponseBeginBlock {
	app.state.BeginBlock()

	// Punish validators who committed equivocation.
	for _, ev := range req.ByzantineValidators {
		if ev.Type == abcitypes.EvidenceType_DUPLICATE_VOTE {
			addr := string(ev.Validator.Address)
			if ev.Validator.Power <= 0 {
				app.logger.Error("non positive val power", "address", addr, "power", ev.Validator.Power)
			} else if pubKey, err := app.state.GetAuthority(addr); err == nil {
				_ = app.state.updateAuthority(nil, abcitypes.ValidatorUpdate{
					PubKey: *pubKey,
					Power:  ev.Validator.Power - 1,
				})
				app.logger.Error("decreased val power by 1 because of the equivocation", "address", addr)
			} else {
				app.logger.Error("wanted to punish val, but can't find it", "address", addr)
			}
		}
	}
	return abcitypes.ResponseBeginBlock{}
}

// Update validators
func (app *KatzenmintApplication) EndBlock(req abcitypes.RequestEndBlock) abcitypes.ResponseEndBlock {
	return abcitypes.ResponseEndBlock{
		ValidatorUpdates: app.state.validatorUpdates,
	}
}

// TODO: state sync connection
func (app *KatzenmintApplication) ListSnapshots(req abcitypes.RequestListSnapshots) (res abcitypes.ResponseListSnapshots) {
	return
}

func (app *KatzenmintApplication) OfferSnapshot(req abcitypes.RequestOfferSnapshot) (res abcitypes.ResponseOfferSnapshot) {
	res.Result = abcitypes.ResponseOfferSnapshot_ABORT
	return
}

func (app *KatzenmintApplication) LoadSnapshotChunk(req abcitypes.RequestLoadSnapshotChunk) (res abcitypes.ResponseLoadSnapshotChunk) {
	return
}

func (app *KatzenmintApplication) ApplySnapshotChunk(req abcitypes.RequestApplySnapshotChunk) (res abcitypes.ResponseApplySnapshotChunk) {
	res.Result = abcitypes.ResponseApplySnapshotChunk_ABORT
	return
}
