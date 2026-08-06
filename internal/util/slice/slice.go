package slice

// Map 用 fn 将切片 v 逐元素映射为新切片。
func Map[T any, V any](v []T, fn func(v T) V) []V {
	items := make([]V, 0, len(v))
	for _, t := range v {
		items = append(items, fn(t))
	}
	return items
}
