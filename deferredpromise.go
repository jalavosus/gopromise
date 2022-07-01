package promise

import (
	"sync/atomic"
)

// DeferredPromise is a Promise which isn't initialized with a Func;
// instead, a DeferredPromise begins its lifecycle when Run is called.
type DeferredPromise[T any] interface {
	Promise[T]

	// Run causes the DeferredPromise to call the passed Func,
	// thus starting the Promise lifecycle.
	// Run is multiprocess-safe: after the first call, subsequent calls
	// will do nothing.
	Run(Func[T])

	// Started returns whether or not Run has been called for
	// this DeferredPromise.
	Started() bool
}

type deferredPromise[T any] struct {
	*promise[T]
	started atomic.Bool
}

// NewDeferredPromise returns a DeferredPromise.
func NewDeferredPromise[T any]() DeferredPromise[T] {
	p := new(deferredPromise[T])
	p.promise = newPromise[T]()

	return p
}

func (p *deferredPromise[T]) Run(fn Func[T]) {
	if !p.Started() {
		p.run(fn)
		p.started.Store(true)
	}
}

func (p *deferredPromise[T]) Started() bool {
	return p.started.Load()
}

func (p *deferredPromise[T]) Fulfilled() bool {
	if p.Started() {
		return p.promise.Fulfilled()
	}

	return false
}

func (p *deferredPromise[T]) Rejected() bool {
	if p.Started() {
		return p.promise.Rejected()
	}

	return false
}
