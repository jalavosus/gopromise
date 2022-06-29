package promise

// Result is the result data returned by a function
// called by a Promise.
type Result[T any] interface {
	// Result returns the result of a Promise function call,
	// or nil if Err != nil
	Result() *T

	// Err returns any resultant error of a Promise function call.
	Err() error
}

type promiseResult[T any] struct {
	result    T
	err       error
	fulfilled bool
}

func (r *promiseResult[T]) Result() *T {
	if r.fulfilled {
		return &r.result
	}

	return nil
}

func (r *promiseResult[T]) Err() error {
	return r.err
}
