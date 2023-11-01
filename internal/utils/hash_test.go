package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHash(t *testing.T) {
	assert.Equal(t, "5994471abb01112afcc18159f6cc74b4f511b99806da59b3caf5a9c173cacfc5", Hash("12345"))
}
