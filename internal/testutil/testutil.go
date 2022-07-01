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
	name       string
	fnWait     time.Duration
	assertions TestAssertions
}

type TestAssertions struct {
	AssertFulfilled assert.BoolAssertionFunc
	AssertRejected  assert.BoolAssertionFunc
	AssertResult    assert.ValueAssertionFunc
	AssertErr       assert.ErrorAssertionFunc
}

func NewTestAssertions(wantFulfilled, wantRes, wantErr bool) TestAssertions {
	return TestAssertions{
		AssertFulfilled: BoolAssertion(wantFulfilled),
		AssertRejected:  BoolAssertion(!wantFulfilled),
		AssertResult:    NilAssertion(!wantRes),
		AssertErr:       ErrorAssertion(wantErr),
	}
}

func (ta TestAssertions) AssertAll(t *testing.T, prom promise.Promise[uint], res promise.Result[uint]) {
	t.Helper()

	ta.AssertFulfilled(t, prom.Fulfilled())
	ta.AssertRejected(t, prom.Rejected())

	ta.AssertResult(t, res.Result())
	ta.AssertErr(t, res.Err())
}

var testCases = []TestCase{
	{
		name:       "timeout-occurs=false",
		fnWait:     1500 * time.Millisecond,
		assertions: NewTestAssertions(true, true, false),
	},
	{
		name:       "timeout-occurs=true",
		fnWait:     2050 * time.Millisecond,
		assertions: NewTestAssertions(false, false, true),
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

	t.Helper()

	testCtx := context.Background()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Helper()

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

			tc.assertions.AssertAll(t, prom, res)
		})
	}
}
