package promise_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/jalavosus/gopromise"
	"github.com/jalavosus/gopromise/internal/testutil"
)

func newPromise(t *testing.T, fn promise.Func[uint]) promise.Promise[uint] {
	t.Helper()
	return promise.NewPromise(fn)
}

func runPromise(*testing.T, promise.Func[uint], promise.Promise[uint]) {}

func TestPromise_Resolve(t *testing.T) {
	testutil.TestResolve(t, newPromise, runPromise, false)
}

func TestPromise_ResolveAsync(t *testing.T) {
	testutil.TestResolve(t, newPromise, runPromise, true)
}

func TestPromiseResolve_ResolveSameVal(t *testing.T) {
	const n = 42

	var fn promise.Func[int] = func(resolve promise.ResolveFunc[int], _ promise.RejectFunc) { resolve(n) }

	prom := promise.NewPromise(fn)

	res := prom.Resolve()
	assert.True(t, prom.Fulfilled())

	assert.NoError(t, res.Err())
	assert.NotNil(t, res.Result())

	t.Run("ResolveAsync after Resolve should not block", func(t *testing.T) {
		var res2 promise.Result[int]

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

	WaitLoop:
		for {
			select {
			case <-ctx.Done():
				assert.NoError(t, ctx.Err())
				break WaitLoop
			case r := <-prom.ResolveAsync():
				res2 = r
				break WaitLoop
			}
		}
		
		assert.NoError(t, res.Err())
		assert.NotNil(t, res.Result())

		assert.Equal(t, res, res2)
		assert.Equal(t, n, *res.Result())
		assert.Equal(t, n, *res2.Result())
	})
}