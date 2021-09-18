package plugins

import (
	"sync"

	"github.com/duc-cnzj/mars/internal/contracts"
)

var (
	mu        sync.RWMutex
	pluginSet = make(map[string]contracts.PluginInterface)
)

func GetPlugins() map[string]contracts.PluginInterface {
	mu.RLock()
	defer mu.RLock()
	return pluginSet
}

func RegisterPlugin(name string, pluginInterface contracts.PluginInterface) {
	mu.Lock()
	defer mu.Unlock()
	pluginSet[name] = pluginInterface
}
