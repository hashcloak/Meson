package katzenmint

import (
	"encoding/binary"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/cosmos/iavl"
	"github.com/hashcloak/Meson/katzenmint/config"
	"github.com/hashcloak/Meson/katzenmint/s11n"
	katvoting "github.com/katzenpost/authority/voting/server/config"
	"github.com/katzenpost/core/pki"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/ed25519"
	cryptoenc "github.com/tendermint/tendermint/crypto/encoding"
	"github.com/tendermint/tendermint/crypto/merkle"
	pc "github.com/tendermint/tendermint/proto/tendermint/crypto"
	dbm "github.com/tendermint/tm-db"
)

const (
	GenesisEpoch  uint64        = 1
	EpochInterval int64         = 10
	HeightPeriod  time.Duration = 1 * time.Second
	LifeCycle     int           = 3
)

var (
	errStateClosed               = errors.New("katzenmint state is closed")
	errDocInsufficientDescriptor = errors.New("insufficient descriptors uploaded")
	errDocInsufficientProvider   = errors.New("no providers uploaded")
)

type descriptor struct {
	desc *pki.MixDescriptor
	raw  []byte
}

type document struct {
	doc *pki.Document
	raw []byte
}

type KatzenmintState struct {
	sync.RWMutex

	// Staged State
	tree             *iavl.MutableTree
	appHash          []byte
	blockHeight      int64
	currentEpoch     uint64
	epochStartHeight int64
	layers           int
	minNodesPerLayer int
	parameters       *katvoting.Parameters

	// Changes to be made
	memAdded         *dbm.MemDB
	validatorUpdates []abcitypes.ValidatorUpdate

	// Cached information
	prevDocument    *pki.Document
	prevCommitError error
}

/*****************************************
 *            Load & Save State          *
 *****************************************/

func NewKatzenmintState(kConfig *config.Config, db dbm.DB) *KatzenmintState {
	tree, err := iavl.NewMutableTree(db, 100)
	if err != nil {
		panic(fmt.Errorf("error creating iavl tree"))
	}
	version, err := tree.Load()
	if err != nil {
		panic(fmt.Errorf("error loading iavl tree"))
	}
	state := &KatzenmintState{
		tree:             tree,
		appHash:          tree.Hash(),
		blockHeight:      version,
		layers:           kConfig.Layers,
		minNodesPerLayer: kConfig.MinNodesPerLayer,
		parameters:       &kConfig.Parameters,
		prevCommitError:  nil,
	}
	_, epochInfoValue := state.tree.Get([]byte(epochInfoKey))
	if version == 0 {
		state.currentEpoch = GenesisEpoch
		state.epochStartHeight = state.blockHeight
	} else if epochInfoValue == nil || len(epochInfoValue) != 16 {
		panic("error loading the current epoch number and its starting height")
	} else {
		state.currentEpoch, _ = binary.Uvarint(epochInfoValue[:8])
		state.epochStartHeight, _ = binary.Varint(epochInfoValue[8:])
	}
	keyDoc := storageKey(documentsBucket, []byte{}, state.currentEpoch-1)
	_, rawDoc := state.tree.Get(keyDoc)
	state.prevDocument, _ = s11n.VerifyAndParseDocument(rawDoc)
	return state
}

func (state *KatzenmintState) Commit() ([]byte, error) {
	state.Lock()
	defer state.Unlock()
	if state.isClosed() {
		return nil, errStateClosed
	}

	// Save descriptors/authorities persistently
	iter, _ := state.memAdded.Iterator([]byte{0x00}, []byte{0xFF})
	for ; iter.Valid(); iter.Next() {
		_ = state.tree.Set(iter.Key(), iter.Value())
	}
	iter.Close()
	state.memAdded.Close()

	// Generate and save document persistently
	var err error
	if state.newDocumentRequired() {
		var doc *document
		if doc, err = state.generateDocument(); err == nil {
			state.prevDocument = doc.doc
			key := storageKey(documentsBucket, []byte{}, state.currentEpoch)
			_ = state.tree.Set(key, doc.raw)
			state.currentEpoch++
			state.epochStartHeight = state.blockHeight + 1
			// TODO: Prune related descriptors
		}
	}

	// Save epoch info persistently
	state.blockHeight++
	epochInfoValue := make([]byte, 16)
	binary.PutUvarint(epochInfoValue[:8], state.currentEpoch)
	binary.PutVarint(epochInfoValue[8:], state.epochStartHeight)
	_ = state.tree.Set([]byte(epochInfoKey), epochInfoValue)

	// Mark a new version
	appHash, _, errSave := state.tree.SaveVersion()
	if errSave != nil {
		return nil, errSave
	}
	state.appHash = appHash

	// Mute the error if keeps occuring
	if err == state.prevCommitError {
		err = nil
	} else {
		state.prevCommitError = err
	}
	return appHash, err
}

func (state *KatzenmintState) Close() {
	state.Lock()
	defer state.Unlock()
	state.tree = nil
	if state.memAdded != nil {
		_ = state.memAdded.Close()
	}
}

func (state *KatzenmintState) isClosed() bool {
	return state.tree == nil
}

/*****************************************
 *              Modify State             *
 *****************************************/

func (state *KatzenmintState) BeginBlock() {
	state.memAdded = dbm.NewMemDB()
	state.validatorUpdates = make([]abcitypes.ValidatorUpdate, 0)
}

// Epoch has to be in [current epoch, current epoch + LifeCycle].
func (state *KatzenmintState) updateMixDescriptor(rawDesc []byte, desc *pki.MixDescriptor, epoch uint64) (err error) {
	key := storageKey(descriptorsBucket, desc.IdentityKey.Bytes(), epoch)

	// Check for redundant uploads.
	if _, err := state.get(key); err == nil {
		return fmt.Errorf("duplicated descriptor with key %s for epoch %d", EncodeHex(desc.IdentityKey.Bytes()), epoch)
	}

	// Check for epoch
	if epoch < state.currentEpoch {
		return fmt.Errorf("late descriptor upload with key %s for epoch %d", EncodeHex(desc.IdentityKey.Bytes()), epoch)
	}
	if epoch >= state.currentEpoch+uint64(LifeCycle) {
		return fmt.Errorf("early descriptor upload with key %s for epoch %d", EncodeHex(desc.IdentityKey.Bytes()), epoch)
	}

	// Save it to memory db.
	return state.set(key, rawDesc)
}

func (state *KatzenmintState) updateAuthority(rawAuth []byte, v abcitypes.ValidatorUpdate) error {
	// TODO: make sure the voting power not exceed 1/3

	pubkey, err := cryptoenc.PubKeyFromProto(v.PubKey)
	if err != nil {
		return fmt.Errorf("can't decode public key: %w", err)
	}
	key := storageKey(authoritiesBucket, pubkey.Address(), 0)

	if rawAuth == nil {
		rawAuth, err = EncodeJson(Authority{
			Auth:    "katzenmint",
			PubKey:  v.PubKey.GetEd25519(),
			KeyType: ed25519.KeyType,
			Power:   v.Power,
		})
		if err != nil {
			return err
		}
	}

	err = state.set(key, rawAuth)
	if err != nil {
		return err
	}
	state.validatorUpdates = append(state.validatorUpdates, v)
	return nil
}

/*****************************************
 *        Criteria of State Change       *
 *****************************************/

func (state *KatzenmintState) newDocumentRequired() bool {
	// TODO: determine when to finish the current epoch
	return state.blockHeight >= state.epochStartHeight+EpochInterval-1
}

func (state *KatzenmintState) isDescriptorAuthorized(desc *pki.MixDescriptor) bool {
	// TODO: determine the criteria to prevent sybil attacks
	return true
}

func (state *KatzenmintState) isAuthorityAuthorized(addr string, auth *AuthorityChecked) bool {
	// TODO: determine the criteria to prevent sybil attacks
	return auth.Val.Power <= 1
}

func (state *KatzenmintState) isAuthorityNew(auth *AuthorityChecked) bool {
	pubkey, err := cryptoenc.PubKeyFromProto(auth.Val.PubKey)
	if err != nil {
		return false
	}
	key := storageKey(authoritiesBucket, pubkey.Address(), 0)
	val, _ := state.get(key)
	return val == nil
}

/*****************************************
 *           Internal Getter             *
 *****************************************/

func (state *KatzenmintState) get(key []byte) (val []byte, err error) {
	state.Lock()
	defer state.Unlock()
	if state.isClosed() {
		return nil, errStateClosed
	}

	has, err := state.memAdded.Has(key)
	if err != nil {
		return nil, err
	}
	if has {
		val, _ = state.memAdded.Get(key)
	} else {
		_, val = state.tree.Get(key)
		if val == nil {
			return nil, fmt.Errorf("key '%v' does not exist", key)
		}
	}
	ret := make([]byte, len(val))
	copy(ret, val)
	return ret, nil
}

func (state *KatzenmintState) getProof(key []byte, height int64) ([]byte, *iavl.RangeProof, error) {
	state.Lock()
	defer state.Unlock()
	if state.isClosed() {
		return nil, nil, errStateClosed
	}
	return state.tree.GetVersionedWithProof(key, height)
}

func (state *KatzenmintState) set(key []byte, value []byte) error {
	state.Lock()
	defer state.Unlock()
	if state.isClosed() {
		return errStateClosed
	}
	return state.memAdded.Set(key, value)
}

/*****************************************
 *           External Getter             *
 *****************************************/

func (state *KatzenmintState) GetAuthority(addr string) (*pc.PublicKey, error) {
	key := storageKey(authoritiesBucket, []byte(addr), 0)
	val, err := state.get(key)
	if err != nil {
		return nil, err
	}
	auth, err := VerifyAndParseAuthority(val)
	if err != nil {
		return nil, err
	}
	return &auth.Val.PubKey, nil
}

func (state *KatzenmintState) GetEpoch(height int64) ([]byte, merkle.ProofOperator, error) {
	key := []byte(epochInfoKey)
	val, proof, err := state.getProof(key, height)
	if err != nil {
		return nil, nil, err
	}
	if len(val) != 16 {
		return nil, nil, fmt.Errorf("error fetching latest epoch for height %v", height)
	}
	valueOp := iavl.NewValueOp(key, proof)
	return val, valueOp, nil
}

func (state *KatzenmintState) GetDocument(epoch uint64, height int64) ([]byte, merkle.ProofOperator, error) {
	// TODO: postpone the document for some blocks?
	// var postponDeadline = 10

	if epoch == 0 {
		return nil, nil, ErrQueryDocumentUnknown
	}
	key := storageKey(documentsBucket, []byte{}, epoch)
	doc, proof, err := state.getProof(key, height)
	if err != nil {
		return nil, nil, err
	}
	if doc == nil {
		if epoch < state.currentEpoch {
			return nil, nil, ErrQueryNoDocument
		}
		if epoch == state.currentEpoch {
			return nil, nil, ErrQueryDocumentNotReady
		}
		return nil, nil, fmt.Errorf("requesting document for a too future epoch %d", epoch)
	}
	valueOp := iavl.NewValueOp(key, proof)
	return doc, valueOp, nil
}

/*****************************************
 *               Document                *
 *****************************************/

func (s *KatzenmintState) generateDocument() (*document, error) {
	// Cannot lock here

	// Load descriptors (providers and nodesDesc).
	var providersDesc, nodesDesc []*descriptor
	begin := storageKey(descriptorsBucket, []byte{}, s.currentEpoch)
	end := storageKey(descriptorsBucket, []byte{}, s.currentEpoch+1)
	_ = s.tree.IterateRange(begin, end, true, func(key, value []byte) (ret bool) {
		desc, _ := s11n.ParseDescriptorWithoutVerify(value)
		v := &descriptor{desc: desc, raw: value}
		if v.desc.Layer == pki.LayerProvider {
			providersDesc = append(providersDesc, v)
		} else {
			nodesDesc = append(nodesDesc, v)
		}
		return false
	})

	// Assign nodes to layers. # No randomness yet.
	var topology [][][]byte
	if len(nodesDesc) < s.layers*s.minNodesPerLayer {
		return nil, errDocInsufficientDescriptor
	}
	sortNodesByPublicKey(nodesDesc)
	if s.prevDocument != nil {
		topology = generateTopology(nodesDesc, s.prevDocument, s.layers)
	} else {
		topology = generateRandomTopology(nodesDesc, s.layers)
	}

	// Sort the providers
	var providers [][]byte
	if len(providersDesc) == 0 {
		return nil, errDocInsufficientProvider
	}
	sortNodesByPublicKey(providersDesc)
	for _, v := range providersDesc {
		providers = append(providers, v.raw)
	}

	// Build the Document.
	doc := &s11n.Document{
		Epoch:             s.currentEpoch,
		GenesisEpoch:      GenesisEpoch,
		SendRatePerMinute: s.parameters.SendRatePerMinute,
		Mu:                s.parameters.Mu,
		MuMaxDelay:        s.parameters.MuMaxDelay,
		LambdaP:           s.parameters.LambdaP,
		LambdaPMaxDelay:   s.parameters.LambdaPMaxDelay,
		LambdaL:           s.parameters.LambdaL,
		LambdaLMaxDelay:   s.parameters.LambdaLMaxDelay,
		LambdaD:           s.parameters.LambdaD,
		LambdaDMaxDelay:   s.parameters.LambdaDMaxDelay,
		LambdaM:           s.parameters.LambdaM,
		LambdaMMaxDelay:   s.parameters.LambdaMMaxDelay,
		Topology:          topology,
		Providers:         providers,
	}

	// TODO: what to do with shared random value?

	// Serialize the Document.
	serialized, err := s11n.SerializeDocument(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize document: %v", err)
	}

	// Ensure the document is sane.
	pDoc, err := s11n.VerifyAndParseDocument(serialized)
	if err != nil {
		return nil, fmt.Errorf("signed document failed validation: %v", err)
	}
	if pDoc.Epoch != s.currentEpoch {
		return nil, fmt.Errorf("signed document has invalid epoch: %v", pDoc.Epoch)
	}
	ret := &document{
		doc: pDoc,
		raw: serialized,
	}
	return ret, nil
}
