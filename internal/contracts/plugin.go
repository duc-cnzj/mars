package contracts

type PluginInterface interface {
	Name() string
	Initialize(args map[string]any) error
	Destroy() error
}
