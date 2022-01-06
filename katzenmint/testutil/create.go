package testutil

import (
	"fmt"

	"github.com/hashcloak/katzenmint-pki/s11n"
	"github.com/katzenpost/core/crypto/ecdh"
	"github.com/katzenpost/core/crypto/eddsa"
	"github.com/katzenpost/core/crypto/rand"
	"github.com/katzenpost/core/pki"
	"github.com/stretchr/testify/require"
)

// create test descriptor
func CreateTestDescriptor(require *require.Assertions, idx int, layer int, epoch uint64) (*pki.MixDescriptor, []byte, eddsa.PrivateKey) {
	desc := new(pki.MixDescriptor)
	desc.Name = fmt.Sprintf("katzenmint%d.example.net", idx)
	desc.Addresses = map[pki.Transport][]string{
		pki.TransportTCPv4: []string{fmt.Sprintf("192.0.2.%d:4242", idx)},
		pki.TransportTCPv6: []string{"[2001:DB8::1]:8901"},
		// pki.Transport("torv2"): []string{"thisisanoldonion.onion:2323"},
		// pki.TransportTCP: []string{"example.com:4242"},
	}
	desc.Layer = uint8(layer)
	desc.LoadWeight = 23
	identityPriv, err := eddsa.NewKeypair(rand.Reader)
	require.NoError(err, "eddsa.NewKeypair()")
	desc.IdentityKey = identityPriv.PublicKey()
	linkPriv, err := ecdh.NewKeypair(rand.Reader)
	require.NoError(err, "ecdh.NewKeypair()")
	desc.LinkKey = linkPriv.PublicKey()
	desc.MixKeys = make(map[uint64]*ecdh.PublicKey)
	for e := epoch; e < epoch+3; e++ {
		mPriv, err := ecdh.NewKeypair(rand.Reader)
		require.NoError(err, "[%d]: ecdh.NewKeypair()", e)
		desc.MixKeys[uint64(e)] = mPriv.PublicKey()
	}
	if layer == pki.LayerProvider {
		desc.Kaetzchen = make(map[string]map[string]interface{})
		desc.Kaetzchen["miau"] = map[string]interface{}{
			"endpoint":  "+miau",
			"miauCount": idx,
		}
	}
	err = s11n.IsDescriptorWellFormed(desc, epoch)
	require.NoError(err, "IsDescriptorWellFormed(good)")

	// Sign the descriptor.
	signed, err := s11n.SignDescriptor(identityPriv, desc)
	require.NoError(err, "SignDescriptor()")
	return desc, signed, *identityPriv
}

// create test document
func CreateTestDocument(require *require.Assertions, epoch uint64) (*s11n.Document, []byte) {
	doc := &s11n.Document{
		Epoch:             epoch,
		GenesisEpoch:      1,
		SendRatePerMinute: 3,
		Topology:          make([][][]byte, 3),
		Mu:                0.42,
		MuMaxDelay:        23,
		LambdaP:           0.69,
		LambdaPMaxDelay:   17,
	}
	idx := 1
	for l := 0; l < 3; l++ {
		for i := 0; i < 5; i++ {
			_, rawDesc, _ := CreateTestDescriptor(require, idx, 0, epoch)
			doc.Topology[l] = append(doc.Topology[l], rawDesc)
			idx++
		}
	}
	for i := 0; i < 3; i++ {
		_, rawDesc, _ := CreateTestDescriptor(require, idx, pki.LayerProvider, epoch)
		doc.Providers = append(doc.Providers, rawDesc)
		idx++
	}
	serialized, err := s11n.SerializeDocument(doc)
	require.NoError(err, "SerializeDocument()")
	return doc, serialized
}
