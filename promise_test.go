package promise_test

import (
	"context"
	"testing"

	"github.com/jalavosus/gopromise"
	"github.com/jalavosus/gopromise/internal/testutil"
)

func newPromise(t *testing.T, ctx context.Context, fn promise.Func[uint]) promise.Promise[uint] {
	t.Helper()
	return promise.NewPromise(ctx, fn)
}

func runPromise(*testing.T, context.Context, promise.Func[uint], promise.Promise[uint]) {}

func TestPromise_Resolve(t *testing.T) {
	testutil.TestResolve(t, newPromise, runPromise, false)
}

func TestPromise_ResolveAsync(t *testing.T) {
	testutil.TestResolve(t, newPromise, runPromise, true)
}
