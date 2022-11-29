package katzenmint

import (
	"bytes"
	"fmt"
	"testing"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/hashcloak/Meson/katzenmint/config"
	"github.com/hashcloak/Meson/katzenmint/s11n"
	"github.com/hashcloak/Meson/katzenmint/testutil"
	"github.com/katzenpost/core/crypto/eddsa"
	"github.com/katzenpost/core/crypto/rand"
	"github.com/katzenpost/core/pki"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	cryptoenc "github.com/tendermint/tendermint/crypto/encoding"

	// "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testEpoch = GenesisEpoch
const testDBCacheSize = 100

var kConfig *config.Config

func init() {
	kConfig = config.DefaultConfig()
}

func TestNewStateBasic(t *testing.T) {
	require := require.New(t)

	// create katzenmint state
	db := dbm.NewMemDB()
	defer db.Close()
	state := NewKatzenmintState(kConfig, db, testDBCacheSize)

	// advance block height
	require.Equal(int64(0), state.blockHeight)
	state.BeginBlock()
	_, _ = state.Commit()
	require.Equal(int64(1), state.blockHeight)

	// test that basic state info can be rebuilt
	state = NewKatzenmintState(kConfig, db, testDBCacheSize)
	require.Equal(int64(1), state.blockHeight)
	require.Equal(GenesisEpoch, state.currentEpoch)
	require.Equal(int64(0), state.epochStartHeight)
}

func TestUpdateDescriptor(t *testing.T) {
	require := require.New(t)

	// create katzenmint state
	db := dbm.NewMemDB()
	defer db.Close()
	state := NewKatzenmintState(kConfig, db, testDBCacheSize)

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
	gotRaw, err := state.get(key)
	if err != nil {
		t.Fatalf("Failed to get mix descriptor from database: %+v\n", err)
	}
	if !bytes.Equal(gotRaw, rawDesc) {
		t.Fatalf("Got a wrong descriptor from database\n")
	}
}

func TestUpdateAuthority(t *testing.T) {
	require := require.New(t)

	// create katzenmint state
	db := dbm.NewMemDB()
	defer db.Close()
	state := NewKatzenmintState(kConfig, db, testDBCacheSize)

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
	pubkey, err := cryptoenc.PubKeyFromProto(validator.PubKey)
	if err != nil {
		t.Fatalf("Failed to decode public key: %v\n", err)
	}
	key := storageKey(authoritiesBucket, pubkey.Address(), 0)
	_, err = state.get(key)
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
	state := NewKatzenmintState(kConfig, db, testDBCacheSize)
	epoch := state.currentEpoch

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

	// proceed with enough block commits to enter the next epoch
	for i := 0; i < int(EpochInterval-1); i++ {
		_, err := state.Commit()
		if err != nil {
			t.Fatalf("Failed to commit: %v\n", err)
		}
		state.BeginBlock()
	}
	_, err := state.Commit()
	if err == nil {
		t.Fatal("Commit should report an error as a side effect because threshold of document creation is not achieved")
	}

	// test the non-existence of the document
	key := storageKey(documentsBucket, []byte{}, epoch)
	_, err = state.get(key)
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
	_, err = state.get(key)
	if state.prevDocument == nil || err != nil {
		t.Fatalf("The pki document should be generated automatically\n")
	}

	// one more height to proceed the epoch
	_, err = state.Commit()
	if err != nil {
		t.Fatalf("Failed to commit: %v\n", err)
	}

	// test the document can be reloaded
	newState := NewKatzenmintState(kConfig, db, testDBCacheSize)
	if newState.prevDocument == nil {
		t.Fatalf("The pki document should be reloaded\n")
	}
	if newState.prevDocument.String() != state.prevDocument.String() {
		t.Fatalf("Reloaded doc inconsistent with the generated doc\n")
	}

	// test the document can be queried
	loaded, _, err := state.GetDocument(testEpoch, state.blockHeight-1)
	require.Nil(err, "Failed to get pki document from state: %+v\n", err)
	require.NotNil(loaded, "Failed to get pki document from state: wrong key")
	_, err = s11n.VerifyAndParseDocument(loaded)
	require.Nil(err, "Failed to parse pki document: %+v\n", err)
}
