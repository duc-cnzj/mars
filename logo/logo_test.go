package logo

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogo(t *testing.T) {
	assert.NotEmpty(t, Logo())
	assert.False(t, strings.Contains(Logo(), "created by duc@2023."))
}

func TestWithAuthor(t *testing.T) {
	assert.True(t, strings.Contains(WithAuthor(), "created by duc@2023."))
}
