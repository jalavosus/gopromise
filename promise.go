package promise

import (
	"context"
	"sync"
	"sync/atomic"
)

// Func is any function which takes a Context and returns
// a given type + error.
type Func[T any] func(context.Context, ...any) (T, error)

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

type promise[T any] struct {
	result atomic.Pointer[promiseResult[T]]
	once   sync.Once
	wg     *sync.WaitGroup
	ch     chan Result[T]
}

// NewPromise returns a Promise. Upon instantiation, the passed Func is called in a goroutine
// which stores the returned data/error.
func NewPromise[T any](ctx context.Context, fn Func[T], fnArgs ...any) Promise[T] {
	p := newPromise[T]()
	p.run(ctx, fn, fnArgs...)

	return p
}

func newPromise[T any]() *promise[T] {
	p := new(promise[T])
	p.ch = make(chan Result[T], 1)

	var wg sync.WaitGroup
	p.wg = &wg

	return p
}

func (p *promise[T]) run(ctx context.Context, fn Func[T], fnArgs ...any) {
	p.once.Do(func() {
		p.wg.Add(1)
		go p.doFn(ctx, fn, fnArgs...)
		return
	})
}

func (p *promise[T]) doFn(ctx context.Context, fn Func[T], fnArgs ...any) {
	defer p.wg.Done()

	promRes := new(promiseResult[T])
	promRes.result, promRes.err = fn(ctx, fnArgs...)
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
