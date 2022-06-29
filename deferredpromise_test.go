package promise_test

import (
	"context"
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

	const wantVal uint = 1

	var (
		counter   uint = 0
		counterCh      = make(chan uint, 2)
	)

	var fn promise.Func[uint] = func(_ context.Context, rest ...any) (uint, error) {
		counter++

		if len(rest) == 1 {
			ch, ok := rest[0].(chan uint)
			if ok {
				ch <- counter
			}
		}

		return counter, nil
	}

	prom := promise.NewDeferredPromise[uint]()

	prom.Run(testCtx, fn, counterCh)
	<-counterCh

	assert.True(t, prom.Started())
	assert.Equal(t, wantVal, counter)

	t.Run("ensure Run() only calls fn once", func(t *testing.T) {
		assert.True(t, prom.Started())

		prom.Run(testCtx, fn, counterCh)

		var (
			counterVal = counter
			ctxErr     error
		)

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

		assert.ErrorIs(t, context.DeadlineExceeded, ctxErr)
		assert.Equal(t, wantVal, counterVal)
		assert.Equal(t, wantVal, counter)

		res := prom.Resolve()
		resVal := res.Result()
		assert.True(t, prom.Fulfilled())
		assert.NotNil(t, resVal)
		assert.Equal(t, wantVal, *resVal)
	})
}