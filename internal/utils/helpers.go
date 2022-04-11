package utils

func Filter[T any](items []T, fn func(item T, index int) bool) []T {
	var res = make([]T, 0)
	for idx, item := range items {
		if fn(item, idx) {
			res = append(res, item)
		}
	}

	return res
}
