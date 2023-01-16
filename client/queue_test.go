package client

import (
	"testing"
	"sync"

	"github.com/katzenpost/client/constants"
	"github.com/stretchr/testify/assert"
)

type foo struct {
	x string
}

func (f foo) Priority() uint64 {
	return uint64(0)
}

func TestQueue(t *testing.T) {
	assert := assert.New(t)
	q := new(Queue)
	err := q.Push(foo{"hello"})
	assert.NoError(err)
	s, err := q.Pop()
	assert.NoError(err)
	assert.Equal(s.(foo).x, "hello")
	_, err = q.Pop()
	assert.Error(err)

	for i := 0; i < constants.MaxEgressQueueSize; i++ {
		err := q.Push(foo{"hello"})
		assert.NoError(err)
	}
	err = q.Push(foo{"hello"})
	assert.Error(err)
	for i := 0; i < constants.MaxEgressQueueSize; i++ {
		s, err = q.Pop()
		assert.NoError(err)
		assert.Equal(s.(foo).x, "hello")
	}
	_, err = q.Pop()
	assert.Error(err)
}

func FuzzQueue(f *testing.F) {
	f.Fuzz(func(t *testing.T, _, s string) {
		q := new(Queue)
		t.Log("Pushing", s)
		err := q.Push(foo{s})
		if err != nil {
			t.Errorf("Push %v %v", s, err)
		}
		item, err := q.Pop()
		if err != nil {
			t.Errorf("Pop %v %v %v", s, item, err)
		}
	})
}

func FuzzQueuePushSerial(f *testing.F) {
	f.Fuzz(func(t *testing.T, i int, s string) {
		q := new(Queue)
		if i > constants.MaxEgressQueueSize {
			return
		}
		for j := 0; j < i; j++ {
			t.Log("Pushing", s)
			err := q.Push(foo{s})
			if err != nil {
				t.Errorf("Push %v %v", s, err)
			}
		}
	})
}

func FuzzQueuePushParalell(f *testing.F) {
	f.Add(1, "abcde")
	f.Add(1, "")
	f.Add(1, string([]byte{0x00}))
	f.Fuzz(func(t *testing.T, i int, s string) {
		q := new(Queue)
		if i > constants.MaxEgressQueueSize {
			return
		}
		if i < 0 {
			return
		}
		t.Logf("Pushing '%v' for %v times", s, i)
		var wg sync.WaitGroup
		for j := 0; j < i; j++ {
			t.Log("Pushing", s)
			wg.Add(1)
			go func(wg *sync.WaitGroup, s string, q *Queue, t *testing.T) {
				defer wg.Done()
				err := q.Push(foo{s})
				if err != nil {
					t.Errorf("Push %v %v", s, err)
				}
			}(&wg, s, q, t)
		}
		wg.Wait()
		if q.len != i {
			t.Errorf("Expected len: %v, actual %v", i, q.len)
		}
	})
}
