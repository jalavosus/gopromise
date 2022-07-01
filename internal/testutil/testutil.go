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
	fnWait  time.Duration
}

var testCases = []TestCase{
	{
		name:    "timeout-occurs=false",
		fnWait:  1500 * time.Millisecond,
		wantErr: false,
	},
	{
		name:    "timeout-occurs=true",
		fnWait:  2050 * time.Millisecond,
		wantErr: true,
	},
}

type (
	PromiseMaker      func(*testing.T, promise.Func[uint]) promise.Promise[uint]
	PromiseTestRunner func(*testing.T, promise.Func[uint], promise.Promise[uint])
	PromiseResolver   func(*testing.T, promise.Promise[uint]) promise.Result[uint]
)

func newContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, ctxTimeout)
}

func makePromiseResolver(ctx context.Context, fnWait time.Duration) promise.Func[uint] {
	const res uint = 42

	return func(resolve promise.ResolveFunc[uint], reject promise.RejectFunc) {
		ticker := time.NewTicker(fnWait)

		defer func() {
			ticker.Stop()
		}()

		for {
			select {
			case <-ctx.Done():
				if err := ctx.Err(); err != nil {
					reject(err)
				} else {
					resolve(res)
				}

				return
			case <-ticker.C:
				resolve(res)
				return
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

			var res promise.Result[uint]

			promFn := makePromiseResolver(ctx, tc.fnWait)
			prom := newProm(t, promFn)
			runProm(t, promFn, prom)

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
