package biz

import (
	"testing"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/stretchr/testify/assert"
)

// fullSettingsConfig 构造全字段配置：插件 args 含敏感/非敏感键、镜像凭证、
// OIDC provider，覆盖各分组枚举分支。
func fullSettingsConfig() *config.Config {
	return &config.Config{
		AppPort:         "4000",
		GrpcPort:        "50000",
		MetricsPort:     "9090",
		Debug:           true,
		ExternalIp:      "127.0.0.1",
		ExcludeServer:   "metrics,cron",
		LogChannel:      "zap",
		TracingEndpoint: "localhost:4317",
		CacheDriver:     "memory",
		DBAutoMigrate:   true,
		GitServerCached: true,
		UploadMaxSize:   "50M",
		UploadDir:       "/tmp/mars-uploads",
		GitServerPlugin: config.Plugin{Name: "gitlab", Args: map[string]any{
			"baseurl":  "https://gitlab/api/v4",
			"token":    "t0k3n",
			"insecure": true,
		}},
		WsSenderPlugin:      config.Plugin{Name: "ws_sender_memory"},
		DomainManagerPlugin: config.Plugin{Name: "default_domain_manager"},
		PicturePlugin:       config.Plugin{Name: "picture_bing", Args: map[string]any{"secret": "s3cr3t"}},
		DBDriver:            "mysql",
		DBDatabase:          "marsv5",
		DBHost:              "127.0.0.1",
		DBPort:              "3306",
		DBUsername:          "root",
		DBPassword:          "pw",
		DBSlowLogEnabled:    true,
		DBSlowLogThreshold:  200 * time.Millisecond,
		DBDebug:             true,
		KubeConfig:          "/tmp/kube",
		NsPrefix:            "devops-",
		InstallTimeout:      90 * time.Second,
		S3Enabled:           true,
		S3Endpoint:          "localhost:9000",
		S3UseSSL:            false,
		S3Bucket:            "mars",
		S3AccessKeyID:       "ak",
		S3SecretAccessKey:   "sk",
		AdminPassword:       "123456",
		ImagePullSecrets: config.DockerAuths{
			{Server: "reg.io", Username: "u", Password: "p"},
		},
		Oidc: []config.OidcSetting{
			{Name: "gitlab", Enabled: true, ProviderUrl: "https://gitlab/.well-known/openid-configuration", ClientID: "cid", ClientSecret: "cs", RedirectUrl: "https://mars/cb"},
		},
		AI: config.AIConfig{
			Model:       "deepseek-chat",
			BaseURL:     "https://api.deepseek.com/v1",
			APIKey:      "sk-0",
			IdleTimeout: 45 * time.Minute,
		},
	}
}

// findSetting 按 key 在分组内查配置项，未命中直接 FailNow。
func findSetting(t *testing.T, group *ConfigGroup, key string) *ConfigItem {
	t.Helper()
	for _, item := range group.Items {
		if item.Key == key {
			return item
		}
	}
	t.Fatalf("key %q not found in group %q", key, group.ID)
	return nil
}

// TestSettingsBiz_Get 全字段配置聚合：七组存在、代表性项 key/value/masked 落位、
// 敏感项（密码/密钥/凭证 token）正确脱敏。
func TestSettingsBiz_Get(t *testing.T) {
	got := NewSettingsBiz(fullSettingsConfig()).Get()
	assert.Len(t, got.Groups, 7)

	server := got.Groups[0]
	assert.Equal(t, "server", server.ID)
	assert.Equal(t, "4000", findSetting(t, server, "app_port").Value)
	assert.Equal(t, "9090", findSetting(t, server, "metrics_port").Value)
	assert.Equal(t, "metrics,cron", findSetting(t, server, "exclude_server").Value)
	assert.False(t, findSetting(t, server, "debug").Masked)

	runtime := got.Groups[1]
	assert.Equal(t, "runtime", runtime.ID)
	assert.Equal(t, "50M", findSetting(t, runtime, "upload_max_size").Value)

	plugins := got.Groups[2]
	assert.Equal(t, "plugins", plugins.ID)
	assert.Equal(t, "gitlab", findSetting(t, plugins, "git_server_plugin.name").Value)
	assert.Equal(t, "https://gitlab/api/v4", findSetting(t, plugins, "git_server_plugin.baseurl").Value)
	assert.False(t, findSetting(t, plugins, "git_server_plugin.baseurl").Masked)
	token := findSetting(t, plugins, "git_server_plugin.token")
	assert.Equal(t, "t0k3n", token.Value)
	assert.True(t, token.Masked)
	assert.Equal(t, "true", findSetting(t, plugins, "git_server_plugin.insecure").Value)
	assert.Equal(t, "ws_sender_memory", findSetting(t, plugins, "ws_sender_plugin.name").Value)
	assert.True(t, findSetting(t, plugins, "picture_plugin.secret").Masked)

	database := got.Groups[3]
	assert.Equal(t, "database", database.ID)
	assert.Equal(t, "pw", findSetting(t, database, "db_password").Value)
	assert.True(t, findSetting(t, database, "db_password").Masked)
	assert.Equal(t, "200ms", findSetting(t, database, "db_slow_log_threshold").Value)
	assert.Equal(t, "true", findSetting(t, database, "db_debug").Value)

	cluster := got.Groups[4]
	assert.Equal(t, "cluster", cluster.ID)
	assert.Equal(t, "devops-", findSetting(t, cluster, "ns_prefix").Value)
	assert.Equal(t, "1m30s", findSetting(t, cluster, "install_timeout").Value)
	assert.Equal(t, "sk", findSetting(t, cluster, "s3_secret_access_key").Value)
	assert.True(t, findSetting(t, cluster, "s3_secret_access_key").Masked)

	auth := got.Groups[5]
	assert.Equal(t, "auth", auth.ID)
	assert.Equal(t, "123456", findSetting(t, auth, "admin_password").Value)
	assert.True(t, findSetting(t, auth, "admin_password").Masked)
	assert.Equal(t, "u", findSetting(t, auth, "reg.io.username").Value)
	assert.Equal(t, "p", findSetting(t, auth, "reg.io.password").Value)
	assert.True(t, findSetting(t, auth, "reg.io.password").Masked)
	assert.Equal(t, "true", findSetting(t, auth, "oidc.gitlab.enabled").Value)
	assert.Equal(t, "cid", findSetting(t, auth, "oidc.gitlab.client_id").Value)
	assert.Equal(t, "cs", findSetting(t, auth, "oidc.gitlab.client_secret").Value)
	assert.True(t, findSetting(t, auth, "oidc.gitlab.client_secret").Masked)

	ai := got.Groups[6]
	assert.Equal(t, "ai", ai.ID)
	assert.Equal(t, "deepseek-chat", findSetting(t, ai, "ai.model").Value)
	assert.Equal(t, "https://api.deepseek.com/v1", findSetting(t, ai, "ai.base_url").Value)
	assert.False(t, findSetting(t, ai, "ai.base_url").Masked)
	assert.Equal(t, "sk-0", findSetting(t, ai, "ai.api_key").Value)
	assert.True(t, findSetting(t, ai, "ai.api_key").Masked)
	assert.Equal(t, "45m0s", findSetting(t, ai, "ai.idle_timeout").Value)
}

// TestSettingsBiz_Get_Empty 空配置聚合：七组仍存在且 id 非空，插件组无 args 行。
func TestSettingsBiz_Get_Empty(t *testing.T) {
	got := NewSettingsBiz(&config.Config{}).Get()
	assert.Len(t, got.Groups, 7)
	for _, g := range got.Groups {
		assert.NotEmpty(t, g.ID)
	}
	plugins := got.Groups[2]
	assert.Equal(t, "plugins", plugins.ID)
	// 空插件 args：name 行仍存在，无任何 args 展开行。
	assert.Len(t, plugins.Items, 4)
	// 新增标量在空配置下仍落行：metrics_port/exclude_server 空串、db_debug 假值。
	server := got.Groups[0]
	assert.Equal(t, "", findSetting(t, server, "metrics_port").Value)
	assert.Equal(t, "", findSetting(t, server, "exclude_server").Value)
	database := got.Groups[3]
	assert.Equal(t, "false", findSetting(t, database, "db_debug").Value)
	// 空配置下 AI 组仍为完整四项：模型/地址/密钥空串（密钥仍掩码），阈值 0s。
	ai := got.Groups[6]
	assert.Equal(t, "ai", ai.ID)
	assert.Len(t, ai.Items, 4)
	assert.Equal(t, "", findSetting(t, ai, "ai.model").Value)
	assert.Equal(t, "", findSetting(t, ai, "ai.base_url").Value)
	assert.Equal(t, "", findSetting(t, ai, "ai.api_key").Value)
	assert.True(t, findSetting(t, ai, "ai.api_key").Masked)
	assert.Equal(t, "0s", findSetting(t, ai, "ai.idle_timeout").Value)
}

// TestSensitiveSettingKey 敏感参数键判定：token/password/secret 片段命中，其余放行。
func TestSensitiveSettingKey(t *testing.T) {
	assert.True(t, sensitiveSettingKey("token"))
	assert.True(t, sensitiveSettingKey("password"))
	assert.True(t, sensitiveSettingKey("secret"))
	assert.True(t, sensitiveSettingKey("client_token"))
	assert.False(t, sensitiveSettingKey("baseurl"))
	assert.False(t, sensitiveSettingKey("name"))
}
