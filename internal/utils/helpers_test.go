package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilter(t *testing.T) {
	res := Filter[string]([]string{"a", "b"}, func(item string, index int) bool {
		return item == "a"
	})
	assert.Equal(t, []string{"a"}, res)
	res2 := Filter[int]([]int{1, 2, 3}, func(item int, index int) bool {
		return item > 1
	})
	assert.Equal(t, []int{2, 3}, res2)
}
