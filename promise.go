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
	p := new(promise[T])
	p.ch = make(chan Result[T], 1)

	var wg sync.WaitGroup
	wg.Add(1)
	p.wg = &wg

	go p.run(ctx, fn)

	return p
}

func (p *promise[T]) run(ctx context.Context, fn Func[T]) {
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
