package mixin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEmail(t *testing.T) {
	email := NewEmail(WithEmailFieldName("aa"), WithEmailFieldComment("cc")).(*Email)
	assert.Equal(t, "cc", email.comment)
	assert.Equal(t, "aa", email.name)
}
