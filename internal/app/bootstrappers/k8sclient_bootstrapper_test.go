package bootstrappers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestK8sClientBootstrapper_Bootstrap(t *testing.T) {}

func TestK8sClientBootstrapper_Tags(t *testing.T) {
	assert.Equal(t, []string{}, (&K8sClientBootstrapper{}).Tags())
}
