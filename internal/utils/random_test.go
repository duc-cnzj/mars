package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandomString(t *testing.T) {
	assert.Regexp(t, `[a-zA-Z0-9]*`, RandomString(10000))
	assert.Equal(t, "", RandomString(-1))
}
