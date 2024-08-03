package serialize

func Serialize[T any, V any](v []T, fn func(v T) V) []V {
	items := make([]V, 0, len(v))
	for _, t := range v {
		items = append(items, fn(t))
	}
	return items
}
