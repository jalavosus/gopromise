package promise

func (p *promise[T]) resolveFunc(ch chan<- *promiseResult[T]) ResolveFunc[T] {
	return func(result T) {
		ch <- &promiseResult[T]{
			result:    result,
			fulfilled: true,
			rejected:  false,
		}
	}
}

func (p *promise[T]) rejectFunc(ch chan<- *promiseResult[T]) RejectFunc {
	return func(err error) {
		ch <- &promiseResult[T]{
			err:       err,
			fulfilled: false,
			rejected:  true,
		}
	}
}
