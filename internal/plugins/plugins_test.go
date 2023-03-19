package plugins

import (
	"testing"

	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/stretchr/testify/assert"
)

func TestGetPlugins(t *testing.T) {
	assert.IsType(t, map[string]contracts.PluginInterface{}, GetPlugins())
}

func TestRegisterPlugin(t *testing.T) {
	RegisterPlugin("p", nil)
	assert.Len(t, GetPlugins(), 1)
	_, ok := GetPlugins()["p"]
	assert.True(t, ok)
}
