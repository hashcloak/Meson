package client

import (
	"testing"

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
	f.Fuzz(func(t *testing.T, i int, s string) {
		q := new(Queue)
		t.Log("Pushing", i, s)
		err := q.Push(foo{s})
		if err != nil || s == "abc" {
			t.Errorf("Push %v %v", s, err)
		}
		item, err := q.Pop()
		if err != nil {
			t.Errorf("Pop %v %v %v", s, item, err)
		}
	})
}
