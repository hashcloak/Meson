// caching pki client
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

// Package pkiclient implements a caching wrapper
package pkiclient

import (
	"container/list"
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/katzenpost/core/crypto/eddsa"
	"github.com/katzenpost/core/pki"
	"github.com/katzenpost/core/worker"
)

var (
	errNotSupported = errors.New("pkiclient: operation not supported")
	errHalted       = errors.New("pkiclient: client was halted")

	fetchBacklog          = 8
	lruMaxSize            = 8
	epochRetrieveInterval = 3 * time.Second
)

type cacheEntry struct {
	raw []byte
	doc *pki.Document
}

// Cache is a caching PKI client.
type Cache struct {
	sync.Mutex
	worker.Worker

	impl Client
	docs map[uint64]*list.Element
	lru  list.List

	timer     *time.Timer
	memEpoch  uint64
	memHeight uint64

	fetchQueue chan *fetchOp
}

type fetchOp struct {
	ctx    context.Context
	epoch  uint64
	doneCh chan interface{}
}

// Halt tears down the Client instance.
func (c *Cache) Halt() {
	c.Worker.Halt()

	// Clean out c.fetchQueue.
	for {
		select {
		case op := <-c.fetchQueue:
			op.doneCh <- errHalted
		default:
			return
		}
	}
}

// GetEpoch returns the epoch information of PKI.
func (c *Cache) GetEpoch(ctx context.Context) (epoch uint64, ellapsedHeight uint64, err error) {
	select {
	case <-c.timer.C:
		epoch, ellapsedHeight, err = c.impl.GetEpoch(ctx)
		if err == nil {
			c.memEpoch = epoch
			c.memHeight = ellapsedHeight
			c.timer.Reset(epochRetrieveInterval)
		} else {
			c.timer.Reset(0)
		}
	default:
		return c.memEpoch, c.memHeight, nil
	}
	return
}

// GetDoc returns the PKI document for the provided epoch.
func (c *Cache) GetDoc(ctx context.Context, epoch uint64) (*pki.Document, []byte, error) {
	// Fast path, cache hit.
	if d := c.cacheGet(epoch); d != nil {
		return d.doc, d.raw, nil
	}

	op := &fetchOp{
		ctx:    ctx,
		epoch:  epoch,
		doneCh: make(chan interface{}),
	}
	c.fetchQueue <- op
	v := <-op.doneCh
	switch r := v.(type) {
	case error:
		return nil, nil, r
	case *cacheEntry:
		// Worker will handle the LRU.
		return r.doc, r.raw, nil
	default:
		return nil, nil, fmt.Errorf("BUG: pkiclient: worker returned nonsensical result: %+v", r)
	}
}

// Post posts the node's descriptor to the PKI for the provided epoch.
func (c *Cache) Post(ctx context.Context, epoch uint64, signingKey *eddsa.PrivateKey, d *pki.MixDescriptor) error {
	return errNotSupported
}

// Deserialize returns PKI document given the raw bytes.
func (c *Cache) Deserialize(raw []byte) (*pki.Document, error) {
	return c.impl.Deserialize(raw) // I hope impl.Deserialize is re-entrant.
}

func (c *Cache) cacheGet(epoch uint64) *cacheEntry {
	c.Lock()
	defer c.Unlock()

	if e, ok := c.docs[epoch]; ok {
		c.lru.MoveToFront(e)
		return e.Value.(*cacheEntry)
	}
	return nil
}

func (c *Cache) insertLRU(newEntry *cacheEntry) {
	c.Lock()
	defer c.Unlock()

	e := c.lru.PushFront(newEntry)
	c.docs[newEntry.doc.Epoch] = e

	// Enforce the max size, by purging based off the LRU.
	for c.lru.Len() > lruMaxSize {
		e = c.lru.Back()
		d := e.Value.(*cacheEntry)

		delete(c.docs, d.doc.Epoch)
		c.lru.Remove(e)
	}
}

func (c *Cache) worker() {
	c.timer = time.NewTimer(0)
	defer c.timer.Stop()
	for {
		var op *fetchOp
		select {
		case <-c.HaltCh():
			return
		case op = <-c.fetchQueue:
		}

		// The fetch may have been in progress while the op was sitting in
		// queue, check again.
		if d := c.cacheGet(op.epoch); d != nil {
			op.doneCh <- d
			continue
		}

		// Slow path, have to call into the PKI client.
		//
		// TODO: This could allow concurrent fetches at some point, but for
		// most common client use cases, this shouldn't matter much.
		d, raw, err := c.impl.GetDoc(op.ctx, op.epoch)
		if err != nil {
			op.doneCh <- err
			continue
		}
		e := &cacheEntry{doc: d, raw: raw}
		c.insertLRU(e)
		op.doneCh <- e
	}
}

// New constructs a new Client backed by an existing pki.Client instance.
func NewCacheClient(impl Client) *Cache {
	c := new(Cache)
	c.impl = impl
	c.docs = make(map[uint64]*list.Element)
	c.fetchQueue = make(chan *fetchOp, fetchBacklog)

	c.Go(c.worker)
	return c
}

// Shutdown the client
func (c *Cache) Shutdown() {
	c.Halt()
}
