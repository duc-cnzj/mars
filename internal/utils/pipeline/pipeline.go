package pipeline

type Pipeline[T any] interface {
	Send(T) Pipeline[T]
	Through(middlewares ...func(func(T)) func(T)) Pipeline[T]
	Then(func(T))
}

type pipeline[T any] struct {
	t           T
	middlewares []func(func(T)) func(T)
}

func NewPipeline[T any]() Pipeline[T] {
	return &pipeline[T]{}
}

func (m *pipeline[T]) Send(t T) Pipeline[T] {
	m.t = t
	return m
}

func (m *pipeline[T]) Through(middlewares ...func(func(T)) func(T)) Pipeline[T] {
	m.middlewares = middlewares
	return m
}

func (m *pipeline[T]) Then(f func(T)) {
	var fn func(T)
	for idx := range m.middlewares {
		fn = m.middlewares[len(m.middlewares)-1-idx](f)
		f = fn
	}
	fn(m.t)
}
