package contracts

//go:generate mockgen -destination ../mock/mock_plugin.go -package mock github.com/duc-cnzj/mars/v4/internal/contracts PluginInterface

type PluginInterface interface {
	// Name plugin name.
	Name() string
	// Initialize init plugin.
	Initialize(args map[string]any) error
	// Destroy plugin.
	Destroy() error
}
