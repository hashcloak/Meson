// descriptor.go - Katzenpost authority descriptor s11n.
// Copyright (C) 2017, 2018  Yawning Angel, masala, David Stainton
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

// Package s11n implements serialization routines for the various PKI
// data structures.
package s11n

import (
	"fmt"
	"net"
	"strconv"

	"github.com/hashcloak/Meson/katzenmint/cert"
	"github.com/katzenpost/core/pki"
	"github.com/katzenpost/core/sphinx/constants"
	"github.com/ugorji/go/codec"
	"golang.org/x/net/idna"
)

const (
	nodeDescriptorVersion = "v0"
)

var (
	// CertificateExpiration is the time a descriptor certificate will be valid for.
	// 600 epoch by default
	CertificateExpiration uint64 = 600
)

type nodeDescriptor struct {
	// Version uniquely identifies the descriptor format as being for the
	// specified version so that it can be rejected if the format changes.
	Version string

	pki.MixDescriptor
}

// SignDescriptor signs and serializes the descriptor with the provided signing
// key.
// TODO: figure out a way to calculate epoch without initialize a http client
// then we can remove expiration from function
func SignDescriptor(signer cert.Signer, base *pki.MixDescriptor, expiration uint64) ([]byte, error) {
	d := new(nodeDescriptor)
	d.MixDescriptor = *base
	d.Version = nodeDescriptorVersion

	// Serialize the descriptor.
	var payload []byte
	enc := codec.NewEncoderBytes(&payload, jsonHandle)
	if err := enc.Encode(d); err != nil {
		return nil, err
	}

	// Sign the descriptor.
	signed, err := cert.Sign(signer, payload, expiration)
	if err != nil {
		return nil, err
	}
	return signed, nil
}

// GetVerifierFromDescriptor returns a verifier for the given
// mix descriptor certificate.
func GetVerifierFromDescriptor(rawDesc []byte) (cert.Verifier, error) {
	payload, err := cert.GetCertified(rawDesc)
	if err != nil {
		return nil, err
	}
	// Parse the payload.
	d := new(nodeDescriptor)
	dec := codec.NewDecoderBytes(payload, jsonHandle)
	if err = dec.Decode(d); err != nil {
		return nil, err
	}
	return d.IdentityKey, nil
}

func ParseDescriptor(b []byte, epochNow uint64) (*pki.MixDescriptor, error) {
	signedCert, err := cert.GetCertificate(b)
	if err != nil {
		return nil, err
	}
	if (epochNow + 1) > signedCert.Expiration {
		return nil, cert.ErrCertificateExpired
	}
	payload, err := cert.GetCertified(b)
	if err != nil {
		return nil, err
	}
	d := new(nodeDescriptor)
	dec := codec.NewDecoderBytes(payload, jsonHandle)
	if err = dec.Decode(d); err != nil {
		return nil, err
	}
	return &d.MixDescriptor, nil
}

// VerifyAndParseDescriptor verifies the signature and deserializes the
// descriptor.  MixDescriptors returned from this routine are guaranteed
// to have been correctly self signed by the IdentityKey listed in the
// MixDescriptor.
func VerifyAndParseDescriptor(verifier cert.Verifier, b []byte, epoch uint64, epochNow uint64) (*pki.MixDescriptor, error) {
	signatures, err := cert.GetSignatures(b)
	if err != nil {
		return nil, err
	}
	if len(signatures) != 1 {
		return nil, fmt.Errorf("expected 1 signature, got: %v", len(signatures))
	}

	// Verify that the descriptor is signed by the verifier.
	payload, err := cert.Verify(verifier, b)
	if err != nil {
		return nil, err
	}

	signedCert, err := cert.GetCertificate(b)
	if err != nil {
		return nil, err
	}

	if epochNow > signedCert.Expiration {
		return nil, cert.ErrCertificateExpired
	}

	// Parse the payload.
	d := new(nodeDescriptor)
	dec := codec.NewDecoderBytes(payload, jsonHandle)
	if err = dec.Decode(d); err != nil {
		return nil, err
	}

	// Ensure the descriptor is well formed.
	if d.Version != nodeDescriptorVersion {
		return nil, fmt.Errorf("invalid Descriptor Version: '%v'", d.Version)
	}
	if err = IsDescriptorWellFormed(&d.MixDescriptor, epoch); err != nil {
		return nil, err
	}
	return &d.MixDescriptor, nil
}

// IsDescriptorWellFormed validates the descriptor and returns a descriptive
// error iff there are any problems that would make it unusable as part of
// a PKI Document.
func IsDescriptorWellFormed(d *pki.MixDescriptor, epoch uint64) error {
	if d.Name == "" {
		return fmt.Errorf("descriptor missing Name")
	}
	if len(d.Name) > constants.NodeIDLength {
		return fmt.Errorf("descriptor Name '%v' exceeds max length", d.Name)
	}
	if d.LinkKey == nil {
		return fmt.Errorf("descriptor missing LinkKey")
	}
	if d.IdentityKey == nil {
		return fmt.Errorf("descriptor missing IdentityKey")
	}
	if d.MixKeys[epoch] == nil {
		return fmt.Errorf("descriptor missing MixKey[%v]", epoch)
	}
	for e := range d.MixKeys {
		// TODO: Should this check that the epochs in MixKey are sequential?
		if e < epoch || e >= epoch+3 {
			return fmt.Errorf("descriptor contains MixKey for invalid epoch: %v", d)
		}
	}
	if len(d.Addresses) == 0 {
		return fmt.Errorf("descriptor missing Addresses")
	}
	for transport, addrs := range d.Addresses {
		if len(addrs) == 0 {
			return fmt.Errorf("descriptor contains empty Address list for transport '%v'", transport)
		}

		var expectedIPVer int
		switch transport {
		case pki.TransportInvalid:
			return fmt.Errorf("descriptor contains invalid Transport")
		case pki.TransportTCPv4:
			expectedIPVer = 4
		case pki.TransportTCPv6:
			expectedIPVer = 6
		default:
			// Unknown transports are only supported between the client and
			// provider.
			if d.Layer != pki.LayerProvider {
				return fmt.Errorf("non-provider published Transport '%v'", transport)
			}
			if transport != pki.TransportTCP {
				// Ignore transports that don't have validation logic.
				continue
			}
		}

		// Validate all addresses belonging to the TCP variants.
		for _, v := range addrs {
			h, p, err := net.SplitHostPort(v)
			if err != nil {
				return fmt.Errorf("descriptor contains invalid address ['%v']'%v': %v", transport, v, err)
			}
			if len(h) == 0 {
				return fmt.Errorf("descriptor contains invalid address ['%v']'%v'", transport, v)
			}
			if port, err := strconv.ParseUint(p, 10, 16); err != nil {
				return fmt.Errorf("descriptor contains invalid address ['%v']'%v': %v", transport, v, err)
			} else if port == 0 {
				return fmt.Errorf("descriptor contains invalid address ['%v']'%v': port is 0", transport, v)
			}
			switch expectedIPVer {
			case 4, 6:
				if ver, err := getIPVer(h); err != nil {
					return fmt.Errorf("descriptor contains invalid address ['%v']'%v': %v", transport, v, err)
				} else if ver != expectedIPVer {
					return fmt.Errorf("descriptor contains invalid address ['%v']'%v': IP version mismatch", transport, v)
				}
			default:
				// This must be TransportTCP or something else that supports
				// "sensible" DNS style hostnames.  Validate that they are
				// at least somewhat well formed.
				if _, err := idna.Lookup.ToASCII(h); err != nil {
					return fmt.Errorf("descriptor contains invalid address ['%v']'%v': %v", transport, v, err)
				}
			}
		}
	}
	if len(d.Addresses[pki.TransportTCPv4]) == 0 {
		return fmt.Errorf("descriptor contains no TCPv4 addresses")
	}
	switch d.Layer {
	case 0:
		if d.Kaetzchen != nil {
			return fmt.Errorf("descriptor contains Kaetzchen when a mix")
		}
	case pki.LayerProvider:
		if err := validateKaetzchen(d.Kaetzchen); err != nil {
			return fmt.Errorf("descriptor contains invalid Kaetzchen block: %v", err)
		}
	default: // it's ok
	}
	return nil
}

func validateKaetzchen(m map[string]map[string]interface{}) error {
	const keyEndpoint = "endpoint"

	if m == nil {
		return nil
	}

	for capa, params := range m {
		if len(capa) == 0 {
			return fmt.Errorf("capability lenght out of bounds")
		}
		if params == nil {
			return fmt.Errorf("capability '%v' has no parameters", capa)
		}

		// Ensure that an endpoint is specified.
		var ep string
		if v, ok := params[keyEndpoint]; !ok {
			return fmt.Errorf("capaiblity '%v' provided no endpoint", capa)
		} else if ep, ok = v.(string); !ok {
			return fmt.Errorf("capability '%v' invalid endpoint type: %T", capa, v)
		}
		// XXX: Should this enforce formating?
		if len(ep) == 0 || len(ep) > constants.RecipientIDLength {
			return fmt.Errorf("capability '%v' invalid endpoint, length out of bounds", capa)
		}

		// Note: This explicitly does not enforce endpoint uniqueness, because
		// it is conceivable that a single endpoint can service multiple
		// request types.
	}

	return nil
}

func getIPVer(h string) (int, error) {
	ip := net.ParseIP(h)
	if ip != nil {
		switch {
		case ip.To4() != nil:
			return 4, nil
		case ip.To16() != nil:
			return 6, nil
		default:
		}
	}
	return 0, fmt.Errorf("address is not an IP")
}
