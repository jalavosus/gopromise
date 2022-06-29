package promise_test

import (
	"context"
	"sync"
	"testing"

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

	var (
		counter uint = 0
		wg      sync.WaitGroup
	)

	var fn promise.Func[uint] = func(_ context.Context, rest ...any) (uint, error) {
		fnWg := rest[0].(*sync.WaitGroup)
		if fnWg != nil {
			defer fnWg.Done()
		}

		counter++
		return counter, nil
	}

	prom := promise.NewDeferredPromise[uint]()

	wg.Add(1)
	prom.Run(testCtx, fn, &wg)
	wg.Wait()

	assert.True(t, prom.Started())
	assert.Equal(t, counter, uint(1))

	t.Run("ensure Run() only calls fn once", func(t *testing.T) {
		assert.True(t, prom.Started())

		prom.Run(testCtx, fn, nil)

		res := prom.Resolve()
		resVal := res.Result()

		assert.True(t, prom.Fulfilled())
		assert.NotNil(t, resVal)
		assert.Equal(t, *resVal, uint(1))
		assert.Equal(t, counter, uint(1))
	})
}
