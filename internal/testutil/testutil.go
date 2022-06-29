package testutil

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/jalavosus/gopromise"
)

const ctxTimeout = 2 * time.Second

type TestCase struct {
	name    string
	wantErr bool
	resolve promise.Func[uint]
}

var testCases = []TestCase{
	{
		name:    "timeout-occurs=false",
		resolve: makePromiseResolver(1500 * time.Millisecond),
		wantErr: false,
	},
	{
		name:    "timeout-occurs=true",
		resolve: makePromiseResolver(2050 * time.Millisecond),
		wantErr: true,
	},
}

type (
	PromiseMaker      func(*testing.T, context.Context, promise.Func[uint]) promise.Promise[uint]
	PromiseTestRunner func(*testing.T, context.Context, promise.Func[uint], promise.Promise[uint])
	PromiseResolver   func(*testing.T, promise.Promise[uint]) promise.Result[uint]
)

func newContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, ctxTimeout)
}

func makePromiseResolver(fnWait time.Duration) promise.Func[uint] {
	const res uint = 42

	return func(ctx context.Context, _ ...any) (uint, error) {
		ticker := time.NewTicker(fnWait)

		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return res, ctx.Err()
			case <-ticker.C:
				ticker.Stop()
				return res, nil
			}
		}
	}
}

func TestResolve(
	t *testing.T,
	newProm PromiseMaker,
	runProm PromiseTestRunner,
	resolveAsync bool,
) {

	testCtx := context.Background()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := newContext(testCtx)
			defer cancel()

			prom := newProm(t, ctx, tc.resolve)

			var res promise.Result[uint]
			runProm(t, ctx, tc.resolve, prom)

			if resolveAsync {
			ResolveAsyncLoop:
				for {
					select {
					case val := <-prom.ResolveAsync():
						res = val
						break ResolveAsyncLoop
					}
				}
			} else {
				res = prom.Resolve()
			}

			assertPromiseResolveTests(t, tc, prom, res)
		})
	}
}

func assertPromiseResolveTests(t *testing.T, tc TestCase, prom promise.Promise[uint], res promise.Result[uint]) {
	t.Helper()

	if tc.wantErr {
		assert.True(t, prom.Rejected())
		assert.False(t, prom.Fulfilled())

		assert.Nil(t, res.Result())
		assert.Error(t, res.Err())
	} else {
		assert.True(t, prom.Fulfilled())
		assert.False(t, prom.Rejected())

		assert.NotNil(t, res.Result())
		assert.NoError(t, res.Err())
	}
}