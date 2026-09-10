package rand_test

import (
	"testing"

	"github.com/duc-cnzj/mars/v6/internal/util/rand"
	"github.com/stretchr/testify/assert"
)

func TestIntnReturnsValueWithinRange(t *testing.T) {
	t.Parallel()
	n := 100
	result := rand.Intn(n)
	assert.True(t, result >= 0 && result < n)
}

func TestIntnReturnsZeroForZeroInput(t *testing.T) {
	t.Parallel()
	n := 0
	result := rand.Intn(n)
	assert.Equal(t, 0, result)
}

func TestIntnReturnsZeroForNegativeInput(t *testing.T) {
	t.Parallel()
	result := rand.Intn(-1)
	assert.Equal(t, 0, result)
}

func TestStringReturnsEmptyStringForZeroLength(t *testing.T) {
	t.Parallel()
	length := 0
	result := rand.String(length)
	assert.Equal(t, "", result)
}

func TestStringReturnsEmptyStringForNegativeLength(t *testing.T) {
	t.Parallel()
	result := rand.String(-1)
	assert.Equal(t, "", result)
}

func TestStringReturnsStringOfCorrectLength(t *testing.T) {
	t.Parallel()
	length := 10
	result := rand.String(length)
	assert.Equal(t, length, len(result))
}
