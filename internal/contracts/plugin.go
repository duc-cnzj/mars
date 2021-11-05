package contracts

type PluginInterface interface {
	Name() string
	Initialize(args map[string]interface{}) error
	Destroy() error
}
