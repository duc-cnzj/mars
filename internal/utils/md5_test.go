package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMd5(t *testing.T) {
	assert.Equal(t, "827ccb0eea8a706c4c34a16891f84e7b", Md5("12345"))
}
