package katzenmint

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"sync"

	"github.com/cosmos/iavl"
	"github.com/hashcloak/katzenmint-pki/config"
	"github.com/hashcloak/katzenmint-pki/s11n"
	katvoting "github.com/katzenpost/authority/voting/server/config"
	"github.com/katzenpost/core/pki"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	cryptoenc "github.com/tendermint/tendermint/crypto/encoding"
	"github.com/tendermint/tendermint/crypto/merkle"
	pc "github.com/tendermint/tendermint/proto/tendermint/crypto"
	dbm "github.com/tendermint/tm-db"
)

const genesisEpoch uint64 = 1
const epochInterval int64 = 5

var (
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
	appHash          []byte
	blockHeight      int64
	currentEpoch     uint64
	epochStartHeight int64

	tree *iavl.MutableTree

	layers           int
	minNodesPerLayer int
	parameters       *katvoting.Parameters
	validators       map[string]pc.PublicKey
	validatorUpdates []abcitypes.ValidatorUpdate

	prevDocument    *document
	prevCommitError error
}

func NewKatzenmintState(kConfig *config.Config, db dbm.DB) *KatzenmintState {
	// TODO: should load the current state from database
	tree, err := iavl.NewMutableTree(db, 100)
	if err != nil {
		panic(fmt.Errorf("error creating iavl tree"))
	}
	version, err := tree.Load()
	if err != nil {
		panic(fmt.Errorf("error loading iavl tree"))
	}
	state := &KatzenmintState{
		appHash:          tree.Hash(),
		blockHeight:      version,
		tree:             tree,
		layers:           kConfig.Layers,
		minNodesPerLayer: kConfig.MinNodesPerLayer,
		parameters:       &kConfig.Parameters,
		validators:       make(map[string]pc.PublicKey),
		validatorUpdates: make([]abcitypes.ValidatorUpdate, 0),
		prevDocument:     nil,
		prevCommitError:  nil,
	}

	// Load current epoch and its start height
	epochInfoValue, err := state.Get([]byte(epochInfoKey))
	if version == 0 {
		state.currentEpoch = genesisEpoch
		state.epochStartHeight = 0
	} else if err != nil || epochInfoValue == nil || len(epochInfoValue) != 16 {
		panic("error loading the current epoch number and its starting height")
	} else {
		state.currentEpoch, _ = binary.Uvarint(epochInfoValue[:8])
		state.epochStartHeight, _ = binary.Varint(epochInfoValue[8:])
	}

	// Load previous document
	e := make([]byte, 8)
	binary.PutUvarint(e, state.currentEpoch-1)
	key := storageKey(documentsBucket, e, state.currentEpoch-1)
	if val, err := state.Get(key); err == nil {
		if doc, err := s11n.VerifyAndParseDocument(val); err == nil {
			state.prevDocument = &document{doc: doc, raw: val}
		}
	}

	// Load validators
	end := make([]byte, len(authoritiesBucket))
	copy(end, []byte(authoritiesBucket))
	end = append(end, 0xff)
	_ = tree.IterateRange([]byte(authoritiesBucket), end, true, func(key, value []byte) bool {
		id, _ := unpackStorageKey(key)
		if id == nil {
			// panic(fmt.Errorf("unable to unpack storage key %v", key))
			return true
		}
		auth, err := VerifyAndParseAuthority(value)
		if err != nil {
			// panic(fmt.Errorf("error parsing authority: %v", err))
			return true
		}
		var protopk pc.PublicKey
		err = protopk.Unmarshal(id)
		if err != nil {
			// panic(fmt.Errorf("error unmarshal proto: %v", err))
			return true
		}
		pk, err := cryptoenc.PubKeyFromProto(protopk)
		if err != nil {
			panic(fmt.Errorf("error extraction from proto: %v", err))
			// return true
		}
		state.validators[string(pk.Address())] = protopk
		if !bytes.Equal(auth.PubKey, protopk.GetEd25519()) {
			panic(fmt.Errorf("storage key id %v has another authority id %v", id, auth.PubKey))
			// return false
		}
		return false
	})

	return state
}

func (state *KatzenmintState) BeginBlock() {
	state.Lock()
	defer state.Unlock()
	state.validatorUpdates = make([]abcitypes.ValidatorUpdate, 0)
}

func (state *KatzenmintState) Commit() ([]byte, error) {
	state.Lock()
	defer state.Unlock()

	var err error
	state.blockHeight++
	if state.newDocumentRequired() {
		var doc *document
		if doc, err = state.generateDocument(); err == nil {
			err = state.updateDocument(doc.raw, doc.doc, state.currentEpoch)
			if err == nil {
				state.currentEpoch++
				state.epochStartHeight = state.blockHeight
				// TODO: Prune related descriptors
			}
		}
	}
	epochInfoValue := make([]byte, 16)
	binary.PutUvarint(epochInfoValue[:8], state.currentEpoch)
	binary.PutVarint(epochInfoValue[8:], state.epochStartHeight)
	_ = state.Set([]byte(epochInfoKey), epochInfoValue)
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

func (state *KatzenmintState) newDocumentRequired() bool {
	// TODO: determine when to finish the current epoch
	return state.blockHeight > state.epochStartHeight+epochInterval
}

func (s *KatzenmintState) generateDocument() (*document, error) {
	// Cannot lock here

	// // Load descriptors (providers and nodesDesc).
	var providersDesc, nodesDesc []*descriptor
	begin := storageKey(descriptorsBucket, []byte{}, s.currentEpoch)
	end := make([]byte, len(begin))
	copy(end, begin)
	end[len(end)-1] = 0xff
	_ = s.tree.IterateRange(begin, end, true, func(key, value []byte) (ret bool) {
		ret = false
		id, _ := unpackStorageKey(key)
		if id == nil {
			return
		}
		desc, err := s11n.ParseDescriptorWithoutVerify(value)
		if err != nil {
			return
		}
		v := &descriptor{desc: desc, raw: value}
		if v.desc.Layer == pki.LayerProvider {
			providersDesc = append(providersDesc, v)
		} else {
			nodesDesc = append(nodesDesc, v)
		}
		return
	})

	// Assign nodes to layers. # No randomness yet.
	var topology [][][]byte
	if len(nodesDesc) < s.layers*s.minNodesPerLayer {
		return nil, errDocInsufficientDescriptor
	}
	sortNodesByPublicKey(nodesDesc)
	if s.prevDocument != nil {
		topology = generateTopology(nodesDesc, s.prevDocument.doc, s.layers)
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
		GenesisEpoch:      genesisEpoch,
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

func (state *KatzenmintState) latestEpoch(height int64) ([]byte, merkle.ProofOperator, error) {
	key := []byte(epochInfoKey)
	val, proof, err := state.tree.GetVersionedWithProof(key, height)
	if err != nil {
		return nil, nil, err
	}
	if len(val) != 16 {
		return nil, nil, fmt.Errorf("error fetching latest epoch for height %v", height)
	}
	valueOp := iavl.NewValueOp(key, proof)
	return val, valueOp, nil

}

func (state *KatzenmintState) documentForEpoch(epoch uint64, height int64) ([]byte, merkle.ProofOperator, error) {
	// TODO: postpone the document for some blocks?
	// var postponDeadline = 10

	e := make([]byte, 8)
	binary.PutUvarint(e, epoch)
	key := storageKey(documentsBucket, e, epoch)
	doc, proof, err := state.tree.GetVersionedWithProof(key, height)
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

func (state *KatzenmintState) isAuthorityAuthorized(addr string) bool {
	return true
}

func (state *KatzenmintState) GetAuthorized(addr string) (pc.PublicKey, bool) {
	pubkey, ok := state.validators[addr]
	return pubkey, ok
}

func (state *KatzenmintState) isDescriptorAuthorized(desc *pki.MixDescriptor) bool {
	return true
}

func (state *KatzenmintState) Set(key []byte, value []byte) error {
	state.tree.Set(key, value)
	return nil
}

func (state *KatzenmintState) Delete(key []byte) error {
	_, success := state.tree.Remove(key)
	if !success {
		return fmt.Errorf("remove from database failed")
	}
	return nil
}

func (state *KatzenmintState) Get(key []byte) ([]byte, error) {
	_, val := state.tree.Get(key)
	if val == nil {
		return nil, fmt.Errorf("key '%v' does not exist", key)
	}
	ret := make([]byte, len(val))
	copy(ret, val)
	return ret, nil
}

// Note: Caller ensures that the epoch is the current epoch +- 1.
func (state *KatzenmintState) updateMixDescriptor(rawDesc []byte, desc *pki.MixDescriptor, epoch uint64) (err error) {
	state.Lock()
	defer state.Unlock()

	key := storageKey(descriptorsBucket, desc.IdentityKey.Bytes(), epoch)

	// Check for redundant uploads.
	if existing, err := state.Get(key); err == nil {
		if existing == nil {
			return fmt.Errorf("state: Wtf, raw field of descriptor for epoch %v is nil", epoch)
		}
		// If the descriptor changes, then it will be rejected to prevent
		// nodes from reneging on uploads.
		if !bytes.Equal(existing, rawDesc) {
			return fmt.Errorf("state: Node %v: Conflicting descriptor for epoch %v", desc.IdentityKey, epoch)
		}

		// Redundant uploads that don't change are harmless.
		return nil
	}

	// Ok, this is a new descriptor.
	if epoch < state.currentEpoch {
		// If there is a document already, the descriptor is late, and will
		// never appear in a document, so reject it.
		return fmt.Errorf("state: Node %v: Late descriptor upload for for epoch %v", desc.IdentityKey, epoch)
	}

	// Persist the raw descriptor to disk.
	if err := state.Set(key, rawDesc); err != nil {
		return err
	}
	return
}

// Note: Caller ensures that the epoch is the current epoch +- 1.
func (state *KatzenmintState) updateDocument(rawDoc []byte, doc *pki.Document, epoch uint64) (err error) {
	// Cannot lock here

	e := make([]byte, 8)
	binary.PutUvarint(e, epoch)
	key := storageKey(documentsBucket, e, epoch)

	//  Check for duplicates
	if existing, err := state.Get(key); err == nil {
		if !bytes.Equal(existing, rawDoc) {
			return fmt.Errorf("state: Conflicting document for epoch %v", epoch)
		}
		// Redundant uploads that don't change are harmless.
		return nil
	}

	// Persist the raw descriptor to disk.
	if err := state.Set(key, rawDoc); err != nil {
		return err
	}

	state.prevDocument = &document{doc: doc, raw: rawDoc}
	return
}

func (state *KatzenmintState) updateAuthority(rawAuth []byte, v abcitypes.ValidatorUpdate) error {
	pubkey, err := cryptoenc.PubKeyFromProto(v.PubKey)
	if err != nil {
		return fmt.Errorf("can't decode public key: %w", err)
	}
	if _, ok := state.validators[string(pubkey.Address())]; ok {
		return fmt.Errorf("authority had been added")
	}
	protoPubKey, err := v.PubKey.Marshal()
	if err != nil {
		return err
	}
	key := storageKey(authoritiesBucket, protoPubKey, 0)

	if v.Power == 0 {
		// remove validator
		auth, err := state.Get(key)
		if err != nil {
			return err
		}
		if auth != nil {
			return fmt.Errorf("cannot remove non-existent validator %s", pubkey.Address())
		}
		if err = state.Delete(key); err != nil {
			return err
		}
		delete(state.validators, string(pubkey.Address()))
	} else {
		// TODO: make sure the voting power not exceed 1/3
		// add or update validator
		if rawAuth == nil && v.Power > 0 {
			rawAuth, err = EncodeJson(Authority{
				Auth:    "katzenmint",
				PubKey:  v.PubKey.GetEd25519(),
				KeyType: "",
				Power:   v.Power,
			})
			if err != nil {
				return err
			}
		}
		if err := state.Set([]byte(key), rawAuth); err != nil {
			return err
		}
		state.validators[string(pubkey.Address())] = v.PubKey
	}

	state.validatorUpdates = append(state.validatorUpdates, v)

	return nil
}
