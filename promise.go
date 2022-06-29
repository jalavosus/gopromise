package promise

import (
	"context"
	"sync"
	"sync/atomic"
)

// Promise provides an interface for types
// which fetch data or calculate results asyncronously,
// and can then wait for the result to return.
type Promise[T any] interface {
	// Resolve blocks until this Promise's execution finishes and
	// a Result is returned.
	Resolve() Result[T]

	// ResolveAsync returns a chan(Result) with a buffer size of 1, which
	// can be utilized in any way desired.
	ResolveAsync() <-chan Result[T]

	// Fulfilled returns true if this Promise resolved successfully.
	// More specifically, it returns true if the result of Err is nil;
	// the value returned by Result _can_ still be nil even if Fulfilled returns true.
	Fulfilled() bool

	// Rejected returns true if this Promise did not resolve successfully - specifically,
	// if the return result of Err is not nil.
	// Rejected will return false if the result of both Err and Result are nil.
	Rejected() bool
}

// DeferredPromise is a Promise which isn't initialized with a Func;
// instead, a DeferredPromise begins its lifecycle when Run is called.
type DeferredPromise[T any] interface {
	Promise[T]

	// Run causes the DeferredPromise to call the passed Func,
	// thus starting the Promise lifecycle.
	// Run is multiprocess-safe: after the first call, subsequent calls
	// will do nothing.
	Run(context.Context, Func[T])

	// Started returns whether or not Run has been called for
	// this DeferredPromise.
	Started() bool
}

// Result is the result data returned by a function
// called by a Promise.
type Result[T any] interface {
	// Result returns the result of a Promise function call,
	// or nil if Err != nil
	Result() *T

	// Err returns any resultant error of a Promise function call.
	Err() error
}

// Func is any function which takes a Context and returns
// a given type + error.
type Func[T any] func(context.Context) (T, error)

// WrapFunc wraps a function which doesn't take a Context
// inside of a Func.
func WrapFunc[T any](fn func() (T, error)) Func[T] {
	return func(_ context.Context) (T, error) {
		return fn()
	}
}

type promise[T any] struct {
	result atomic.Pointer[promiseResult[T]]
	once   sync.Once
	wg     *sync.WaitGroup
	ch     chan Result[T]
}

// NewPromise returns a Promise. Upon instantiation, the passed Func is called in a goroutine
// which stores the returned data/error.
func NewPromise[T any](ctx context.Context, fn Func[T]) Promise[T] {
	p := newPromise[T]()
	p.run(ctx, fn)

	return p
}

func newPromise[T any]() *promise[T] {
	p := new(promise[T])
	p.ch = make(chan Result[T], 1)

	var wg sync.WaitGroup
	p.wg = &wg

	return p
}

func (p *promise[T]) run(ctx context.Context, fn Func[T]) {
	p.once.Do(func() {
		p.wg.Add(1)
		go p.doFn(ctx, fn)
	})
}

func (p *promise[T]) doFn(ctx context.Context, fn Func[T]) {
	defer p.wg.Done()

	promRes := new(promiseResult[T])
	promRes.result, promRes.err = fn(ctx)
	if promRes.err == nil {
		promRes.fulfilled = true
	}

	p.setResult(promRes)
}

func (p *promise[T]) setResult(res *promiseResult[T]) {
	p.result.Store(res)
	p.ch <- res
}

func (p *promise[T]) Resolve() Result[T] {
	p.wg.Wait()
	return <-p.ch
}

func (p *promise[T]) loadResult() *promiseResult[T] {
	return p.result.Load()
}

func (p *promise[T]) ResolveAsync() <-chan Result[T] {
	return p.ch
}

func (p *promise[T]) Fulfilled() bool {
	if res := p.loadResult(); res != nil {
		return res.Result() != nil && res.Err() == nil
	}

	return false
}

func (p *promise[T]) Rejected() bool {
	if res := p.loadResult(); res != nil {
		return res.Err() != nil
	}

	return false
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

func (p *deferredPromise[T]) Run(ctx context.Context, fn Func[T]) {
	if !p.Started() {
		p.run(ctx, fn)
		p.started.Store(true)
	}
}

func (p *deferredPromise[T]) Started() bool {
	return p.started.Load()
}

func (p *deferredPromise[T]) Resolve() Result[T] {
	return p.promise.Resolve()
}

func (p *deferredPromise[T]) ResolveAsync() <-chan Result[T] {
	return p.promise.ResolveAsync()
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

type promiseResult[T any] struct {
	result    T
	err       error
	fulfilled bool
}

func (r *promiseResult[T]) Result() *T {
	if r.fulfilled {
		return &r.result
	}

	return nil
}

func (r *promiseResult[T]) Err() error {
	return r.err
}
