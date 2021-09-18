package contracts

type PluginInterface interface {
	Name() string
	Initialize() error
	Destroy() error
}
