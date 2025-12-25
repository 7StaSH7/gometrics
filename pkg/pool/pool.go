package pool

import "sync"

// Resetter defines the interface for types that can be reset.
type Resetter interface {
	Reset()
}

// Pool is a generic wrapper around sync.Pool.
type Pool[T Resetter] struct {
	pool sync.Pool
}

// New creates a new Pool.
// newFn is responsible for creating a new instance of T.
func New[T Resetter](newFn func() T) *Pool[T] {
	return &Pool[T]{
		pool: sync.Pool{
			New: func() any {
				return newFn()
			},
		},
	}
}

// Get retrieves an item from the pool.
func (p *Pool[T]) Get() T {
	return p.pool.Get().(T)
}

// Put puts an item to the pool.
// It calls Reset() on the item before putting it back.
func (p *Pool[T]) Put(v T) {
	v.Reset()
	p.pool.Put(v)
}
