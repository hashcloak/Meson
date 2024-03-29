// mixkey.go - Katzenpost server mix key store.
// Copyright (C) 2017  Yawning Angel.
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

package server

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/hashcloak/Meson/server/internal/constants"
	"github.com/hashcloak/Meson/server/internal/glue"
	"github.com/hashcloak/Meson/server/internal/mixkey"
	"github.com/katzenpost/core/crypto/ecdh"
	"gopkg.in/op/go-logging.v1"
)

type mixKeys struct {
	sync.Mutex

	glue glue.Glue
	log  *logging.Logger

	keys map[uint64]*mixkey.MixKey
}

func (m *mixKeys) init() error {
	// Generate/load the initial set of keys.
	//
	// TODO: In theory this should also try to load the previous epoch's key
	// if the current time is in the clock skew grace period.  But it may not
	// matter much in practice.
	epoch, _, _, err := m.glue.PKI().Now()
	if err != nil {
		return err
	}
	if _, err := m.Generate(epoch); err != nil {
		return err
	}

	// Clean up stale mix keys hanging around the data directory.
	files, err := filepath.Glob(filepath.Join(m.glue.Config().Server.DataDir, mixkey.KeyGlob))
	if err != nil {
		m.log.Warningf("Failed to find persisted keys: %v", err)
	}
	keyFmt := filepath.Join(m.glue.Config().Server.DataDir, mixkey.KeyFmt)
	for _, f := range files {
		e := uint64(0)
		if _, err := fmt.Sscanf(f, keyFmt, &e); err != nil {
			m.log.Debugf("Failed to extract epoch from '%v': %v", f, err)
			continue
		}
		if _, ok := m.keys[e]; !ok && e < epoch {
			m.log.Debugf("Purging stale key: %v", f)
			os.Remove(f)
		}
	}

	return nil
}

func (m *mixKeys) Generate(baseEpoch uint64) (bool, error) {
	didGenerate := false

	m.Lock()
	defer m.Unlock()
	for e := baseEpoch; e < baseEpoch+constants.NumMixKeys; e++ {
		// Skip keys that we already have.
		if _, ok := m.keys[e]; ok {
			continue
		}

		didGenerate = true
		k, err := mixkey.New(m.glue.Config().Server.DataDir, e)
		if err != nil {
			// Clean up whatever keys that may have succeeded.
			for ee := baseEpoch; ee < baseEpoch+constants.NumMixKeys; ee++ {
				if kk, ok := m.keys[ee]; ok {
					kk.Deref(ee)
					delete(m.keys, ee)
				}
			}
			return false, err
		}
		k.SetUnlinkIfExpired(true)
		m.keys[e] = k
	}

	return didGenerate, nil
}

func (m *mixKeys) Prune() bool {
	epoch, _, _, err := m.glue.PKI().Now()
	if err != nil {
		m.log.Debugf("Error fetching PKI epoch")
		return false
	}
	if epoch == 0 {
		m.log.Debug("before genesis epoch")
		return false
	}
	didPrune := false

	m.Lock()
	defer m.Unlock()

	for idx, v := range m.keys {
		if idx < epoch-1 {
			m.log.Debugf("Purging expired key for epoch: %v", idx)
			v.Deref(epoch)
			delete(m.keys, idx)
			didPrune = true
		}
	}

	return didPrune
}

func (m *mixKeys) Get(epoch uint64) (*ecdh.PublicKey, bool) {
	m.Lock()
	defer m.Unlock()

	if k, ok := m.keys[epoch]; ok {
		return k.PublicKey(), true
	}
	return nil, false
}

func (m *mixKeys) Shadow(dst map[uint64]*mixkey.MixKey) {
	m.Lock()
	defer m.Unlock()

	epoch, _, _, err := m.glue.PKI().Now()
	if err != nil {
		m.log.Debugf("Error fetching PKI epoch")
		return
	}

	// Purge the keys no longer listed from dst.
	for k, v := range dst {
		if _, ok := m.keys[k]; !ok {
			v.Deref(epoch)
			delete(dst, k)
		}
	}

	// Add newly listed keys to dst and bump up the refcount.
	for k, v := range m.keys {
		if _, ok := dst[k]; !ok {
			v.Ref()
			dst[k] = v
		}
	}
}

func (m *mixKeys) Halt() {
	m.Lock()
	defer m.Unlock()

	epoch, _, _, err := m.glue.PKI().Now()
	if err != nil {
		m.log.Debugf("Error fetching PKI epoch")
		return
	}
	for k, v := range m.keys {
		v.Deref(epoch)
		delete(m.keys, k)
	}
}

func newMixKeys(glue glue.Glue) (glue.MixKeys, error) {
	m := &mixKeys{
		glue: glue,
		log:  glue.LogBackend().GetLogger("mixkeys"),
		keys: make(map[uint64]*mixkey.MixKey),
	}

	if err := m.init(); err != nil {
		return nil, err
	}

	return m, nil
}
