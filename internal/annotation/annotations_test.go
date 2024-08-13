package annotation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIgnoreContainerNames(t *testing.T) {
	expected := "mars.duc-cnzj.github.io/ignore-containers"
	assert.Equal(t, IgnoreContainerNames, expected)
}

func TestPodOrderIndexNames(t *testing.T) {
	expected := "mars.duc-cnzj.github.io/order-index"
	assert.Equal(t, PodOrderIndex, expected)
}
