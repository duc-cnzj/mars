package biz

// settings.go 定义平台配置只读聚合用例（SettingsBiz）：把已加载的 config.Config
// 枚举成管理员后台可展示的分组视图。只读不写——本包不提供任何配置修改能力，
// 敏感项（密码/密钥/凭证）标 Masked 由前端掩码展示。biz 直接引用叶子包 config
// （零内部依赖），无分层违规。

import (
	"fmt"
	"sort"
	"strings"

	"github.com/duc-cnzj/mars/v6/internal/config"
)

// ConfigItem 是单条平台配置项（只读聚合视图）：Key 为配置点路径（如
// "git_server_plugin.token"），Value 为格式化后的展示值，Masked 标记敏感项。
type ConfigItem struct {
	Key    string
	Value  string
	Masked bool
}

// ConfigGroup 是平台配置分组：ID 为分组标识（前端按组渲染卡片）。
type ConfigGroup struct {
	ID    string
	Items []*ConfigItem
}

// Settings 是平台配置聚合结果。
type Settings struct {
	Groups []*ConfigGroup
}

// SettingsBiz 是平台配置只读聚合用例：从已加载配置枚举展示项，无任何写能力。
type SettingsBiz interface {
	// Get 返回平台配置分组视图（纯内存读取，无 I/O，无需 ctx）。
	Get() *Settings
}

var _ SettingsBiz = (*settingsBiz)(nil)

// settingsBiz 是 SettingsBiz 的生产实现：持有已加载配置，按分组枚举展示项。
type settingsBiz struct {
	cfg *config.Config
}

// NewSettingsBiz 构造配置聚合用例：cfg 为已加载配置，由 wire 注入。
func NewSettingsBiz(cfg *config.Config) SettingsBiz {
	return &settingsBiz{cfg: cfg}
}

// Get 返回平台配置分组视图：服务 / 运行与优化 / 插件 / 数据库 / 集群与存储 /
// 认证与安全 六组，全部来自只读枚举，敏感字段标 Masked。
func (s *settingsBiz) Get() *Settings {
	return &Settings{Groups: []*ConfigGroup{
		settingsServerGroup(s.cfg),
		settingsRuntimeGroup(s.cfg),
		settingsPluginGroup(s.cfg),
		settingsDatabaseGroup(s.cfg),
		settingsClusterGroup(s.cfg),
		settingsAuthGroup(s.cfg),
	}}
}

// setting 构造单条配置项：value 统一格式化为字符串，masked 标记敏感项。
func setting(key string, value any, masked bool) *ConfigItem {
	return &ConfigItem{Key: key, Value: fmt.Sprintf("%v", value), Masked: masked}
}

// settingsServerGroup 枚举服务组配置：端口/调试开关/对外地址。
func settingsServerGroup(cfg *config.Config) *ConfigGroup {
	return &ConfigGroup{ID: "server", Items: []*ConfigItem{
		setting("app_port", cfg.AppPort, false),
		setting("grpc_port", cfg.GrpcPort, false),
		setting("debug", cfg.Debug, false),
		setting("external_ip", cfg.ExternalIp, false),
	}}
}

// settingsRuntimeGroup 枚举运行与优化组配置：日志/链路/缓存/迁移/上传上限。
func settingsRuntimeGroup(cfg *config.Config) *ConfigGroup {
	return &ConfigGroup{ID: "runtime", Items: []*ConfigItem{
		setting("log_channel", cfg.LogChannel, false),
		setting("tracing_endpoint", cfg.TracingEndpoint, false),
		setting("cache_driver", cfg.CacheDriver, false),
		setting("db_auto_migrate", cfg.DBAutoMigrate, false),
		setting("git_server_cached", cfg.GitServerCached, false),
		setting("upload_max_size", cfg.UploadMaxSize, false),
		setting("upload_dir", cfg.UploadDir, false),
	}}
}

// settingsPluginGroup 枚举插件组配置：name + args 逐行展开（键排序保证稳定序），
// token/password/secret 类敏感参数标 Masked。
func settingsPluginGroup(cfg *config.Config) *ConfigGroup {
	items := make([]*ConfigItem, 0)
	for _, p := range []struct {
		name string
		cfg  config.Plugin
	}{
		{"git_server_plugin", cfg.GitServerPlugin},
		{"ws_sender_plugin", cfg.WsSenderPlugin},
		{"domain_manager_plugin", cfg.DomainManagerPlugin},
		{"picture_plugin", cfg.PicturePlugin},
	} {
		items = append(items, setting(p.name+".name", p.cfg.Name, false))
		keys := make([]string, 0, len(p.cfg.Args))
		for k := range p.cfg.Args {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			items = append(items, setting(p.name+"."+k, p.cfg.Args[k], sensitiveSettingKey(k)))
		}
	}
	return &ConfigGroup{ID: "plugins", Items: items}
}

// settingsDatabaseGroup 枚举数据库组配置：连接串字段，密码标 Masked。
func settingsDatabaseGroup(cfg *config.Config) *ConfigGroup {
	return &ConfigGroup{ID: "database", Items: []*ConfigItem{
		setting("db_driver", cfg.DBDriver, false),
		setting("db_database", cfg.DBDatabase, false),
		setting("db_host", cfg.DBHost, false),
		setting("db_port", cfg.DBPort, false),
		setting("db_username", cfg.DBUsername, false),
		setting("db_password", cfg.DBPassword, true),
		setting("db_slow_log_enabled", cfg.DBSlowLogEnabled, false),
		setting("db_slow_log_threshold", cfg.DBSlowLogThreshold, false),
	}}
}

// settingsClusterGroup 枚举集群与存储组配置：k8s 连接/命名空间前缀/安装超时/S3，
// S3 访问密钥标 Masked。
func settingsClusterGroup(cfg *config.Config) *ConfigGroup {
	return &ConfigGroup{ID: "cluster", Items: []*ConfigItem{
		setting("kubeconfig", cfg.KubeConfig, false),
		setting("ns_prefix", cfg.NsPrefix, false),
		setting("install_timeout", cfg.InstallTimeout, false),
		setting("s3_enabled", cfg.S3Enabled, false),
		setting("s3_endpoint", cfg.S3Endpoint, false),
		setting("s3_use_ssl", cfg.S3UseSSL, false),
		setting("s3_bucket", cfg.S3Bucket, false),
		setting("s3_access_key_id", cfg.S3AccessKeyID, true),
		setting("s3_secret_access_key", cfg.S3SecretAccessKey, true),
	}}
}

// settingsAuthGroup 枚举认证与安全组配置：内置 admin 密码、镜像仓库凭证
// （server 常显、password 掩码）、OIDC provider（client_secret 掩码）。
func settingsAuthGroup(cfg *config.Config) *ConfigGroup {
	items := []*ConfigItem{setting("admin_password", cfg.AdminPassword, true)}
	for _, auth := range cfg.ImagePullSecrets {
		items = append(items,
			setting(auth.Server+".username", auth.Username, false),
			setting(auth.Server+".password", auth.Password, true),
		)
	}
	for _, oidc := range cfg.Oidc {
		items = append(items,
			setting("oidc."+oidc.Name+".enabled", oidc.Enabled, false),
			setting("oidc."+oidc.Name+".provider_url", oidc.ProviderUrl, false),
			setting("oidc."+oidc.Name+".client_id", oidc.ClientID, false),
			setting("oidc."+oidc.Name+".redirect_url", oidc.RedirectUrl, false),
			setting("oidc."+oidc.Name+".client_secret", oidc.ClientSecret, true),
		)
	}
	return &ConfigGroup{ID: "auth", Items: items}
}

// sensitiveSettingKey 判定插件参数键是否敏感（含 token/password/secret 片段），
// 敏感项由前端掩码展示，避免密钥类参数明文暴露。
func sensitiveSettingKey(key string) bool {
	lower := strings.ToLower(key)
	return strings.Contains(lower, "token") ||
		strings.Contains(lower, "password") ||
		strings.Contains(lower, "secret")
}
