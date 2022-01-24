// katzenmint pkiclient interface
package pkiclient

import (
	"context"

	"github.com/katzenpost/core/crypto/eddsa"
	cpki "github.com/katzenpost/core/pki"
)

// Client is the abstract interface used for PKI interaction.
type Client interface {
	// GetEpoch returns the epoch information of PKI.
	GetEpoch(ctx context.Context) (epoch uint64, ellapsedHeight uint64, err error)

	// GetDoc returns the PKI document along with the raw serialized form for the provided epoch.
	GetDoc(ctx context.Context, epoch uint64) (*cpki.Document, []byte, error)

	// Post posts the node's descriptor to the PKI for the provided epoch.
	Post(ctx context.Context, epoch uint64, signingKey *eddsa.PrivateKey, d *cpki.MixDescriptor) error

	// Deserialize returns PKI document given the raw bytes.
	Deserialize(raw []byte) (*cpki.Document, error)

	// Shutdown the client
	Shutdown()
}
