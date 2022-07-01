package promise

type (
	// ResolveFunc is a function passed to functions called by Promise,
	// to be called by those functions if they successfully return a result.
	ResolveFunc[T any] func(T)

	// RejectFunc is a function passed to functions called by Promise,
	// to be called if they return an error.
	RejectFunc func(error)

	// Func represents a wrapper function, to be wrapped around functions which
	// are being called by a Promise.
	Func[T any] func(ResolveFunc[T], RejectFunc)
)
