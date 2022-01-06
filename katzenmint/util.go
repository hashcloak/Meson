// state.go - Katzenpost non-voting authority server state.
// Copyright (C) 2017  Yawning Angel.

package katzenmint

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"sort"

	"github.com/katzenpost/core/pki"
	"github.com/katzenpost/core/sphinx/constants"
)

const (
	descriptorsBucket = "k_descriptors"
	documentsBucket   = "k_documents"
	authoritiesBucket = "k_authorities"
	epochInfoKey      = "k_epoch"
)

func storageKey(keyPrefix string, keyID []byte, epoch uint64) (key []byte) {
	epochHex := make([]byte, 8)
	binary.PutUvarint(epochHex, epoch)
	epochHex = []byte(EncodeHex(epochHex))
	IDHex := []byte(EncodeHex(keyID))

	key = make([]byte, len(keyPrefix))
	copy(key, keyPrefix)
	key = append(key[:], []byte(":")...)
	key = append(key[:], epochHex[:]...)
	key = append(key[:], []byte(":")...)
	key = append(key[:], IDHex[:]...)
	return
}

func unpackStorageKey(key []byte) (keyID []byte, epoch uint64) {
	pre := bytes.Index(key, []byte(":"))
	post := bytes.LastIndex(key, []byte(":"))
	if pre < 0 || post <= pre {
		return nil, 0
	}
	epoch, read := binary.Uvarint(DecodeHex(string(key[pre+1 : post])))
	keyID = DecodeHex(string(key[post+1:]))
	if read <= 0 {
		return nil, 0
	}
	return
}

func sortNodesByPublicKey(nodes []*descriptor) {
	dTos := func(d *descriptor) string {
		pk := d.desc.IdentityKey.ByteArray()
		return string(pk[:])
	}
	sort.Slice(nodes, func(i, j int) bool { return dTos(nodes[i]) < dTos(nodes[j]) })
}

func generateTopology(nodeList []*descriptor, doc *pki.Document, layers int) [][][]byte {
	nodeMap := make(map[[constants.NodeIDLength]byte]*descriptor)
	for _, v := range nodeList {
		id := v.desc.IdentityKey.ByteArray()
		nodeMap[id] = v
	}

	// Since there is an existing network topology, use that as the basis for
	// generating the mix topology such that the number of nodes per layer is
	// approximately equal, and as many nodes as possible retain their existing
	// layer assignment to minimise network churn.

	rng := rand.New(rand.NewSource(0))
	targetNodesPerLayer := len(nodeList) / layers
	topology := make([][][]byte, layers)

	// Assign nodes that still exist up to the target size.
	for layer, nodes := range doc.Topology {
		// The existing nodes are examined in random order to make it hard
		// to predict which nodes will be shifted around.
		nodeIndexes := rng.Perm(len(nodes))
		for _, idx := range nodeIndexes {
			if len(topology[layer]) >= targetNodesPerLayer {
				break
			}

			id := nodes[idx].IdentityKey.ByteArray()
			if n, ok := nodeMap[id]; ok {
				// There is a new descriptor with the same identity key,
				// as an existing descriptor in the previous document,
				// so preserve the layering.
				topology[layer] = append(topology[layer], n.raw)
				delete(nodeMap, id)
			}
		}
	}

	// Flatten the map containing the nodes pending assignment.
	toAssign := make([]*descriptor, 0, len(nodeMap))
	for _, n := range nodeMap {
		toAssign = append(toAssign, n)
	}
	assignIndexes := rng.Perm(len(toAssign))

	// Fill out any layers that are under the target size, by
	// randomly assigning from the pending list.
	idx := 0
	for layer := range doc.Topology {
		for len(topology[layer]) < targetNodesPerLayer {
			n := toAssign[assignIndexes[idx]]
			topology[layer] = append(topology[layer], n.raw)
			idx++
		}
	}

	// Assign the remaining nodes.
	for layer := 0; idx < len(assignIndexes); idx++ {
		n := toAssign[assignIndexes[idx]]
		topology[layer] = append(topology[layer], n.raw)
		layer++
		layer = layer % len(topology)
	}

	return topology
}

func generateRandomTopology(nodes []*descriptor, layers int) [][][]byte {
	// If there is no node history in the form of a previous consensus,
	// then the simplest thing to do is to randomly assign nodes to the
	// various layers.

	rng := rand.New(rand.NewSource(0))
	nodeIndexes := rng.Perm(len(nodes))
	topology := make([][][]byte, layers)
	for idx, layer := 0, 0; idx < len(nodes); idx++ {
		n := nodes[nodeIndexes[idx]]
		topology[layer] = append(topology[layer], n.raw)
		layer++
		layer = layer % len(topology)
	}

	return topology
}
