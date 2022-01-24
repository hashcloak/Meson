package katzenmint

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/hashcloak/katzenmint-pki/config"
	"github.com/hashcloak/katzenmint-pki/s11n"
	"github.com/hashcloak/katzenmint-pki/testutil"
	"github.com/katzenpost/core/crypto/eddsa"
	"github.com/katzenpost/core/crypto/rand"
	"github.com/katzenpost/core/pki"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tm-db"

	// "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testEpoch = genesisEpoch

var kConfig *config.Config

func init() {
	kConfig = config.DefaultConfig()
}

func TestNewStateBasic(t *testing.T) {
	require := require.New(t)

	// create katzenmint state
	db := dbm.NewMemDB()
	defer db.Close()
	state := NewKatzenmintState(kConfig, db)

	// advance block height
	require.Equal(int64(0), state.blockHeight)
	state.BeginBlock()
	_, _ = state.Commit()
	require.Equal(int64(1), state.blockHeight)

	// test that basic state info can be rebuilt
	state = NewKatzenmintState(kConfig, db)
	require.Equal(int64(1), state.blockHeight)
	require.Equal(genesisEpoch, state.currentEpoch)
	require.Equal(int64(0), state.epochStartHeight)
}

func TestUpdateDescriptor(t *testing.T) {
	require := require.New(t)

	// create katzenmint state
	db := dbm.NewMemDB()
	defer db.Close()
	state := NewKatzenmintState(kConfig, db)

	// create test descriptor
	desc, rawDesc, _ := testutil.CreateTestDescriptor(require, 1, pki.LayerProvider, testEpoch)

	// update test descriptor
	state.BeginBlock()
	err := state.updateMixDescriptor(rawDesc, desc, testEpoch)
	if err != nil {
		t.Fatalf("Failed to update test descriptor: %+v\n", err)
	}
	_, err = state.Commit()
	if err != nil {
		t.Fatalf("Failed to commit: %v\n", err)
	}

	// test the data exists in database
	key := storageKey(descriptorsBucket, desc.IdentityKey.Bytes(), testEpoch)
	gotRaw, err := state.Get(key)
	if err != nil {
		t.Fatalf("Failed to get mix descriptor from database: %+v\n", err)
	}
	if !bytes.Equal(gotRaw, rawDesc) {
		t.Fatalf("Got a wrong descriptor from database\n")
	}
}

func TestUpdateDocument(t *testing.T) {
	require := require.New(t)

	// create katzenmint state
	db := dbm.NewMemDB()
	defer db.Close()
	state := NewKatzenmintState(kConfig, db)

	// create, validate and deserialize document
	_, sDoc := testutil.CreateTestDocument(require, testEpoch)
	dDoc, err := s11n.VerifyAndParseDocument(sDoc)
	if err != nil {
		t.Fatalf("Failed to VerifyAndParseDocument document: %+v\n", err)
	}

	// update document
	state.BeginBlock()
	err = state.updateDocument(sDoc, dDoc, testEpoch)
	if err != nil {
		t.Fatalf("Failed to update pki document: %+v\n", err)
	}
	state.currentEpoch++
	state.epochStartHeight = state.blockHeight
	_, err = state.Commit()
	if err != nil {
		t.Fatalf("Failed to commit: %v\n", err)
	}

	// test the data exists in database
	e := make([]byte, 8)
	binary.PutUvarint(e, testEpoch)
	key := storageKey(documentsBucket, e, testEpoch)
	gotRaw, err := state.Get(key)
	if err != nil {
		t.Fatalf("Failed to get pki document from database: %+v\n", err)
	}
	if !bytes.Equal(gotRaw, sDoc) {
		t.Fatalf("Got a wrong document from database\n")
	}

	// test the data exists in memory
	if state.prevDocument == nil {
		t.Fatal("Failed to get pki document from memory\n")
	}
	if !bytes.Equal(state.prevDocument.raw, sDoc) {
		t.Fatalf("Got a wrong document from memory\n")
	}

	// test the data can be reloaded into memory
	state = NewKatzenmintState(kConfig, db)
	if state.prevDocument == nil {
		t.Fatal("Failed to reload pki document into memory\n")
	}
	if !bytes.Equal(state.prevDocument.raw, sDoc) {
		t.Fatalf("Got a wrong document from reloaded memory\n")
	}
}

func TestUpdateAuthority(t *testing.T) {
	require := require.New(t)

	// create katzenmint state
	db := dbm.NewMemDB()
	defer db.Close()
	state := NewKatzenmintState(kConfig, db)

	// create authority
	k, err := eddsa.NewKeypair(rand.Reader)
	require.NoError(err, "eddsa.NewKeypair()")
	authority := &Authority{
		Auth:    "katzenmint",
		Power:   1,
		PubKey:  k.PublicKey().Bytes(),
		KeyType: "",
	}
	rawAuth, err := EncodeJson(authority)
	if err != nil {
		t.Fatalf("Failed to marshal authority: %+v\n", err)
	}

	// update authority
	state.BeginBlock()
	validator := abcitypes.UpdateValidator(authority.PubKey, authority.Power, authority.KeyType)
	err = state.updateAuthority(rawAuth, validator)
	if err != nil {
		fmt.Printf("Failed to update authority: %+v\n", err)
		return
	}
	_, err = state.Commit()
	if err != nil {
		t.Fatalf("Failed to commit: %v\n", err)
	}

	// test the data exists in database
	protoPubKey, err := validator.PubKey.Marshal()
	if err != nil {
		t.Fatalf("Failed to encode public with protobuf: %v\n", err)
	}
	key := storageKey(authoritiesBucket, protoPubKey, 0)
	_, err = state.Get(key)
	if err != nil {
		t.Fatalf("Failed to get authority from database: %+v\n", err)
	}
	// TODO: check that the value is correct
	if len(state.validatorUpdates) != 1 {
		t.Fatal("Failed to update authority\n")
	}
}

func TestDocumentGenerationUponCommit(t *testing.T) {
	require := require.New(t)

	// create katzenmint state
	db := dbm.NewMemDB()
	defer db.Close()
	state := NewKatzenmintState(kConfig, db)
	epoch := state.currentEpoch
	e := make([]byte, 8)
	binary.PutUvarint(e, epoch)
	key := storageKey(documentsBucket, e, epoch)

	// create descriptorosts of providers
	providers := make([]descriptor, 0)
	for i := 0; i < state.minNodesPerLayer; i++ {
		desc, rawDesc, _ := testutil.CreateTestDescriptor(require, i, pki.LayerProvider, epoch)
		providers = append(providers, descriptor{desc: desc, raw: rawDesc})
	}

	// create descriptors of mixs
	mixs := make([]descriptor, 0)
	for layer := 0; layer < state.layers; layer++ {
		for i := 0; i < state.minNodesPerLayer; i++ {
			desc, rawDesc, _ := testutil.CreateTestDescriptor(require, i, 0, epoch)
			mixs = append(mixs, descriptor{desc: desc, raw: rawDesc})
		}
	}

	// update part of the descriptors
	state.BeginBlock()
	for _, p := range providers {
		err := state.updateMixDescriptor(p.raw, p.desc, epoch)
		if err != nil {
			t.Fatalf("Failed to update provider descriptor: %+v\n", err)
		}
	}
	for i, m := range mixs {
		if i == 0 {
			// skip one of the mix descriptors
			continue
		}
		err := state.updateMixDescriptor(m.raw, m.desc, epoch)
		if err != nil {
			t.Fatalf("Failed to update mix descriptor: %+v\n", err)
		}
	}
	_, err := state.Commit()
	if err != nil {
		t.Fatalf("Failed to commit: %v\n", err)
	}

	// proceed with enough block commits to enter the next epoch
	for i := 0; i < int(epochInterval)-1; i++ {
		state.BeginBlock()
		_, err = state.Commit()
		if err != nil {
			t.Fatalf("Failed to commit: %v\n", err)
		}
	}
	state.BeginBlock()
	_, err = state.Commit()
	if err == nil {
		t.Fatal("Commit should report an error as a side effect because threshold of document creation is not achieved")
	}

	// test the non-existence of the document
	_, err = state.Get(key)
	if state.prevDocument != nil || err == nil {
		t.Fatalf("The pki document should not be generated at this moment because there is not enough mix descriptors\n")
	}

	// update the remaining descriptors up to the required threshold
	state.BeginBlock()
	err = state.updateMixDescriptor(mixs[0].raw, mixs[0].desc, epoch)
	if err != nil {
		t.Fatalf("Failed to update mix descriptor: %+v\n", err)
	}
	_, err = state.Commit()
	if err != nil {
		t.Fatalf("Failed to commit: %v\n", err)
	}

	// test the existence of the document
	_, err = state.Get(key)
	if state.prevDocument == nil || err != nil {
		t.Fatalf("The pki document should be generated automatically\n")
	}

	// one more height to proceed the epoch
	_, err = state.Commit()
	if err != nil {
		t.Fatalf("Failed to commit: %v\n", err)
	}

	// test the document can be reloaded
	newState := NewKatzenmintState(kConfig, db)
	if newState.prevDocument == nil {
		t.Fatalf("The pki document should be reloaded\n")
	}
	if !bytes.Equal(newState.prevDocument.raw, state.prevDocument.raw) {
		t.Fatalf("Reloaded doc inconsistent with the generated doc\n")
	}
}
