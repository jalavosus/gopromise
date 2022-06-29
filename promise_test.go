package promise_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/jalavosus/gopromise"
)

const ctxTimeout = 5 * time.Second

type testCase struct {
	name    string
	fnWait  time.Duration
	wantErr bool
}

var testCases = []testCase{
	{
		name:    "err=nil",
		fnWait:  4500 * time.Millisecond,
		wantErr: false,
	},
	{
		name:    "err!=nil",
		fnWait:  5050 * time.Millisecond,
		wantErr: true,
	},
}

func resolveFn(fnWait time.Duration) func(context.Context) (uint, error) {
	const res uint = 42

	return func(ctx context.Context) (uint, error) {
		t := time.NewTicker(fnWait)

		for {
			select {
			case <-ctx.Done():
				t.Stop()
				return res, ctx.Err()
			case <-t.C:
				t.Stop()
				return res, nil
			}
		}
	}
}

func TestPromise_Resolve(t *testing.T) {
	testCtx := context.Background()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			prom, _, cancel := newPromise(t, testCtx, tc.fnWait)
			defer cancel()

			res := prom.Resolve()
			assertPromiseResolveTests(t, tc, prom, res)
		})
	}
}

func TestPromise_ResolveAsync(t *testing.T) {
	testCtx := context.Background()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			prom, _, cancel := newPromise(t, testCtx, tc.fnWait)
			defer cancel()

			resChan := prom.ResolveAsync()

			for {
				select {
				case res := <-resChan:
					assertPromiseResolveTests(t, tc, prom, res)
					return
				}
			}
		})
	}
}

func TestDeferredPromise_Resolve(t *testing.T) {
	testCtx := context.Background()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			prom, ctx, cancel := newDeferredPromise(t, testCtx)
			defer cancel()

			assert.False(t, prom.Started())
			prom.Run(ctx, resolveFn(tc.fnWait))
			assert.True(t, prom.Started())

			res := prom.Resolve()

			assertPromiseResolveTests(t, tc, prom, res)
		})
	}
}

func TestDeferredPromise_ResolveAsync(t *testing.T) {
	testCtx := context.Background()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			prom, ctx, cancel := newDeferredPromise(t, testCtx)
			defer cancel()

			assert.False(t, prom.Started())
			prom.Run(ctx, resolveFn(tc.fnWait))
			assert.True(t, prom.Started())

			resChan := prom.ResolveAsync()

			for {
				select {
				case res := <-resChan:
					assertPromiseResolveTests(t, tc, prom, res)
					return
				}
			}
		})
	}
}

func newPromise(t *testing.T, ctx context.Context, fnWait time.Duration) (promise.Promise[uint], context.Context, context.CancelFunc) {
	t.Helper()

	promCtx, cancel := context.WithTimeout(ctx, ctxTimeout)

	return promise.NewPromise(promCtx, resolveFn(fnWait)), promCtx, cancel
}

func newDeferredPromise(t *testing.T, ctx context.Context) (promise.DeferredPromise[uint], context.Context, context.CancelFunc) {
	t.Helper()

	promCtx, cancel := context.WithTimeout(ctx, ctxTimeout)

	return promise.NewDeferredPromise[uint](), promCtx, cancel
}

func assertPromiseResolveTests(t *testing.T, tc testCase, prom promise.Promise[uint], res promise.Result[uint]) {
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
