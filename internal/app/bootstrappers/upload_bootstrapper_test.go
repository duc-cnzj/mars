package bootstrappers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUploadBootstrapper_Tags(t *testing.T) {
	assert.Equal(t, []string{}, (&UploadBootstrapper{}).Tags())
}
