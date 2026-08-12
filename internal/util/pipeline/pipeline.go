package pipeline

// Pipeline 将值 t 依次穿过一组中间件后交给最终处理器；零值可直接使用。
type Pipeline[T any] struct {
	t           T
	middlewares []func(func(T)) func(T)
}

// New 构造一个无中间件的空 Pipeline。
func New[T any]() *Pipeline[T] {
	return &Pipeline[T]{}
}

// Send 设置待处理的值并返回自身，便于链式调用。
func (m *Pipeline[T]) Send(t T) *Pipeline[T] {
	m.t = t
	return m
}

// Through 追加一组中间件，后追加的中间件包裹在外层。
// 每个中间件通过调用 next 放行到下一层，也可选择不放行提前终止。
func (m *Pipeline[T]) Through(middlewares ...func(t T, next func())) *Pipeline[T] {
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

// Then 执行最终处理，值依次穿过已注册的中间件后到达 f。
// 后注册的中间件包裹在外层，故从后往前逐层包住 f 再执行。
func (m *Pipeline[T]) Then(f func(T)) {
	for i := len(m.middlewares) - 1; i >= 0; i-- {
		f = m.middlewares[i](f)
	}
	f(m.t)
}
