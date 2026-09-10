package slice_test

import (
	"testing"

	"github.com/duc-cnzj/mars/v6/internal/util/slice"
	"github.com/stretchr/testify/assert"
)

func TestMapWithIntegers(t *testing.T) {
	input := []int{1, 2, 3, 4, 5}
	expected := []int{2, 4, 6, 8, 10}

	result := slice.Map(input, func(v int) int {
		return v * 2
	})

	assert.Equal(t, expected, result)
}

func TestMapWithEmptySlice(t *testing.T) {
	input := []int{}
	expected := []int{}

	result := slice.Map(input, func(v int) int {
		return v * 2
	})

	assert.Equal(t, expected, result)
}

func TestMapWithStrings(t *testing.T) {
	input := []string{"a", "b", "c"}
	expected := []string{"aa", "bb", "cc"}

	result := slice.Map(input, func(v string) string {
		return v + v
	})

	assert.Equal(t, expected, result)
}

func TestMapWithNilSlice(t *testing.T) {
	var input []int
	var expected = []int{}

	result := slice.Map(input, func(v int) int {
		return v * 2
	})

	assert.Equal(t, expected, result)
}
