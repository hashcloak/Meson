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

	"github.com/hashcloak/Meson/katzenmint"
	"github.com/katzenpost/core/crypto/eddsa"
	"github.com/katzenpost/core/pki"
	"github.com/katzenpost/core/worker"
)

var (
	errHalted = errors.New("pkiclient: client was halted")

	fetchBacklog = 8
	lruMaxSize   = 8
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
	c.timer.Stop()

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
	c.Lock()
	defer c.Unlock()
	return c.memEpoch, c.memHeight, nil
}

// GetDoc returns the PKI document for the provided epoch.
func (c *Cache) GetDoc(ctx context.Context, epoch uint64) (*pki.Document, []byte, error) {
	// Fast path, cache hit.
	if d := c.cacheGet(epoch); d != nil {
		return d.doc, d.raw, nil
	}

	// Exit upon halt
	select {
	case <-c.HaltCh():
		return nil, nil, fmt.Errorf("pki client is halted, cannot get new document")
	default:
	}

	// Slow path
	op := &fetchOp{
		ctx:    ctx,
		epoch:  epoch,
		doneCh: make(chan interface{}),
	}
	c.fetchQueue <- op
	switch r := (<-op.doneCh).(type) {
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
	return c.impl.Post(ctx, epoch, signingKey, d)
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
	const retryTime = time.Second / 2

	var epoch, height uint64
	var ctx context.Context
	var err error
	for {
		var op *fetchOp
		select {
		case <-c.HaltCh():
			return
		case <-c.timer.C:
			c.Lock()
			if c.memHeight < uint64(katzenmint.EpochInterval) {
				c.memHeight++
				c.Unlock()
				c.timer.Reset(katzenmint.HeightPeriod)
				continue
			}
			ctx = context.Background()
			epoch, height, err = c.impl.GetEpoch(context.Background())
			if epoch == c.memEpoch && height < uint64(katzenmint.EpochInterval) {
				c.memHeight = height
			}
			if err != nil || epoch == c.memEpoch {
				c.timer.Reset(retryTime)
				c.Unlock()
				continue
			}
			c.memEpoch = epoch
			c.memHeight = height
			c.Unlock()
			c.timer.Reset(katzenmint.HeightPeriod)
		case op = <-c.fetchQueue:
			ctx = op.ctx
			epoch = op.epoch
		}

		// The fetch may have been in progress while the op was sitting in
		// queue, check again.
		if d := c.cacheGet(epoch); d != nil {
			if op != nil {
				op.doneCh <- d
			}
			continue
		}

		// Slow path, have to call into the PKI client.
		//
		// TODO: This could allow concurrent fetches at some point, but for
		// most common client use cases, this shouldn't matter much.
		d, raw, err := c.impl.GetDoc(ctx, epoch)
		if err != nil {
			if op != nil {
				op.doneCh <- err
			}
			continue
		}
		e := &cacheEntry{doc: d, raw: raw}
		c.insertLRU(e)
		if op != nil {
			op.doneCh <- e
		}
	}
}

// New constructs a new Client backed by an existing pki.Client instance.
func NewCacheClient(impl Client) (*Cache, error) {
	var err error
	c := new(Cache)
	c.impl = impl
	c.docs = make(map[uint64]*list.Element)
	c.fetchQueue = make(chan *fetchOp, fetchBacklog)
	c.memEpoch, c.memHeight, err = c.impl.GetEpoch(context.Background())
	if err != nil {
		return nil, err
	}
	c.timer = time.NewTimer(katzenmint.HeightPeriod)

	c.Go(c.worker)
	return c, nil
}

// Shutdown the client
func (c *Cache) Shutdown() {
	c.Halt()
	c.impl.Shutdown()
}
