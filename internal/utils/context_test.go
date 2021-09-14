package utils

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCustomErrorContext(t *testing.T) {
	ctx, f := NewCustomErrorContext()
	f(errors.New("a"))
	f(errors.New("b"))
	f(errors.New("c"))
	assert.EqualError(t, ctx.Err(), "a")
	select {
	case <-ctx.Done():
		assert.True(t, true)
	default:
		assert.True(t, false)
	}
}
