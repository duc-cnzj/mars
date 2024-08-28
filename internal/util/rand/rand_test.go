package rand_test

import (
	"testing"

	"github.com/duc-cnzj/mars/v5/internal/util/rand"
	"github.com/stretchr/testify/assert"
)

func TestIntnReturnsValueWithinRange(t *testing.T) {
	n := 100
	result := rand.Intn(n)
	assert.True(t, result >= 0 && result < n)
}

func TestIntnReturnsZeroForZeroInput(t *testing.T) {
	n := 0
	result := rand.Intn(n)
	assert.Equal(t, 0, result)
}

func TestStringReturnsEmptyStringForZeroLength(t *testing.T) {
	length := 0
	result := rand.String(length)
	assert.Equal(t, "", result)
}

func TestStringReturnsStringOfCorrectLength(t *testing.T) {
	length := 10
	result := rand.String(length)
	assert.Equal(t, length, len(result))
}
