package serialize_test

import (
	"testing"

	"github.com/duc-cnzj/mars/v5/internal/util/serialize"
	"github.com/stretchr/testify/assert"
)

func TestSerializeWithIntegers(t *testing.T) {
	input := []int{1, 2, 3, 4, 5}
	expected := []int{2, 4, 6, 8, 10}

	result := serialize.Serialize(input, func(v int) int {
		return v * 2
	})

	assert.Equal(t, expected, result)
}

func TestSerializeWithEmptySlice(t *testing.T) {
	input := []int{}
	expected := []int{}

	result := serialize.Serialize(input, func(v int) int {
		return v * 2
	})

	assert.Equal(t, expected, result)
}

func TestSerializeWithStrings(t *testing.T) {
	input := []string{"a", "b", "c"}
	expected := []string{"aa", "bb", "cc"}

	result := serialize.Serialize(input, func(v string) string {
		return v + v
	})

	assert.Equal(t, expected, result)
}

func TestSerializeWithNilSlice(t *testing.T) {
	var input []int
	var expected = []int{}

	result := serialize.Serialize(input, func(v int) int {
		return v * 2
	})

	assert.Equal(t, expected, result)
}
