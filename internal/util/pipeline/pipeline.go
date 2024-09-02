package pipeline

type Pipeline[T any] interface {
	Send(T) Pipeline[T]
	Through(middlewares ...func(t T, next func())) Pipeline[T]
	Then(func(T))
}

type pipeline[T any] struct {
	t           T
	middlewares []func(func(T)) func(T)
}

func New[T any]() Pipeline[T] {
	return &pipeline[T]{}
}

func (m *pipeline[T]) Send(t T) Pipeline[T] {
	m.t = t
	return m
}

func (m *pipeline[T]) Through(middlewares ...func(t T, next func())) Pipeline[T] {
	for idx := range middlewares {
		middleware := middlewares[idx]
		m.middlewares = append(m.middlewares, func(next func(t T)) func(t T) {
			return func(t T) {
				middleware(t, func() { next(t) })
			}
		})
	}
	return m
}

func (m *pipeline[T]) Then(f func(T)) {
	var fn func(T) = f
	for idx := range m.middlewares {
		fn = m.middlewares[len(m.middlewares)-1-idx](f)
		f = fn
	}
	fn(m.t)
}
