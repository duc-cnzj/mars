package plugins

import (
	"sync"

	"github.com/duc-cnzj/mars/v4/internal/contracts"
)

var (
	mu        sync.RWMutex
	pluginSet = make(map[string]contracts.PluginInterface)
)

// GetPlugins get all registered plugins.
func GetPlugins() map[string]contracts.PluginInterface {
	mu.RLock()
	defer mu.RUnlock()
	return pluginSet
}

// RegisterPlugin register plugin.
func RegisterPlugin(name string, pluginInterface contracts.PluginInterface) {
	mu.Lock()
	defer mu.Unlock()
	pluginSet[name] = pluginInterface
}
