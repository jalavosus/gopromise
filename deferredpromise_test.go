package promise_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/jalavosus/gopromise"

	"github.com/jalavosus/gopromise/internal/testutil"
)

func newDeferredPromise(t *testing.T, _ context.Context, _ promise.Func[uint]) promise.Promise[uint] {
	t.Helper()
	return promise.NewDeferredPromise[uint]()
}

func runDeferredPromise(t *testing.T, ctx context.Context, fn promise.Func[uint], pprom promise.Promise[uint]) {
	prom := pprom.(promise.DeferredPromise[uint])

	assert.False(t, prom.Started())
	assert.False(t, prom.Fulfilled())
	assert.False(t, prom.Rejected())
	prom.Run(ctx, fn)
	assert.True(t, prom.Started())
}

func TestDeferredPromise_Resolve(t *testing.T) {
	testutil.TestResolve(t, newDeferredPromise, runDeferredPromise, false)
}

func TestDeferredPromise_ResolveAsync(t *testing.T) {
	testutil.TestResolve(t, newDeferredPromise, runDeferredPromise, true)
}

func TestDeferredPromise_Run(t *testing.T) {
	testCtx := context.Background()

	const wantVal uint32 = 1

	var (
		counter   atomic.Uint32
		counterCh = make(chan uint32, 2)
	)

	var fn promise.Func[uint32] = func(_ context.Context, rest ...any) (uint32, error) {
		counter.Add(1)
		val := counter.Load()

		if len(rest) == 1 {
			ch, ok := rest[0].(chan uint32)
			if ok {
				ch <- val
			}
		}

		return val, nil
	}

	prom := promise.NewDeferredPromise[uint32]()

	prom.Run(testCtx, fn, counterCh)
	<-counterCh

	assert.True(t, prom.Started())
	assert.Equal(t, wantVal, counter.Load())

	t.Run("fn-called-once", func(t *testing.T) {
		assert.True(t, prom.Started())

		var (
			counterVal = counter.Load()
			ctxErr     error
		)

		prom.Run(testCtx, fn, counterCh)

		ctx, cancel := context.WithTimeout(testCtx, 3*time.Second)
		defer cancel()

	WaitLoop:
		for {
			select {
			case c := <-counterCh:
				counterVal = c
				break WaitLoop
			case <-ctx.Done():
				ctxErr = ctx.Err()
				break WaitLoop
			}
		}

		assert.Error(t, ctxErr)
		assert.ErrorIs(t, context.DeadlineExceeded, ctxErr)
		assert.Equal(t, wantVal, counterVal)
		assert.Equal(t, wantVal, counter.Load())

		res := prom.Resolve()
		resVal := res.Result()
		assert.True(t, prom.Fulfilled())
		assert.NotNil(t, resVal)
		assert.Equal(t, wantVal, *resVal)
	})
}
