package config_test

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// TestMaxUploadSizeWithValidSize 验证可解析的上传上限被正确换算成字节。
func TestMaxUploadSizeWithValidSize(t *testing.T) {
	cfg := &config.Config{UploadMaxSize: "100Mib"}
	assert.Equal(t, uint64(100*1024*1024), cfg.MaxUploadSize())
}

// TestMaxUploadSizeWithInvalidSize 验证解析失败时回退到默认上限。
func TestMaxUploadSizeWithInvalidSize(t *testing.T) {
	cfg := &config.Config{UploadMaxSize: "invalid"}
	assert.Equal(t, uint64(50*1000*1000), cfg.MaxUploadSize())
}

// TestGetFreePort 验证能分配一个合法范围内的临时端口。
func TestGetFreePort(t *testing.T) {
	port, err := config.GetFreePort()
	assert.NoError(t, err)
	assert.Greater(t, port, 0)
	assert.Less(t, port, 65536)
}

// TestInitWithMissingConfigFilePanics 验证配置文件不存在时 Init fail-fast panic。
func TestInitWithMissingConfigFilePanics(t *testing.T) {
	assert.Panics(t, func() { config.Init("not_exist_config.yaml") })
}

// TestConfigDSN 验证 MySQL DSN 拼接格式。
func TestConfigDSN(t *testing.T) {
	cfg := &config.Config{
		DBUsername: "user",
		DBPassword: "pass",
		DBHost:     "localhost",
		DBPort:     "3306",
		DBDatabase: "testdb",
	}
	expectedDSN := "user:pass@tcp(localhost:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local"
	assert.Equal(t, expectedDSN, cfg.DSN())
}

// TestConfigFormatDockerCfg 验证 docker config.json 序列化，含 base64 编码的 Auth 字段。
func TestConfigFormatDockerCfg(t *testing.T) {
	auths := config.DockerAuths{
		&config.DockerAuth{Username: "user", Password: "pass", Email: "email", Server: "server"},
	}
	raw := auths.FormatDockerCfg()

	var decoded struct {
		Auths map[string]struct {
			Username string `json:"username"`
			Password string `json:"password"`
			Email    string `json:"email"`
			Auth     string `json:"auth"`
		} `json:"auths"`
	}
	assert.NoError(t, json.Unmarshal(raw, &decoded))
	entry, ok := decoded.Auths["server"]
	assert.True(t, ok)
	assert.Equal(t, "user", entry.Username)
	assert.Equal(t, "pass", entry.Password)
	assert.Equal(t, "email", entry.Email)
	assert.Equal(t, base64.StdEncoding.EncodeToString([]byte("user:pass")), entry.Auth)
}

// TestConfigListExcludeServerTags 验证标签串按 ',' 拆分、逐项去空白、过滤空项。
func TestConfigListExcludeServerTags(t *testing.T) {
	tags := config.ExcludeServerTags("tag1, tag2, ,tag3, ")
	assert.Equal(t, []string{"tag1", "tag2", "tag3"}, tags.List())
}

// TestInitWithConfigFileAndValidValues 验证绝对路径配置文件加载 + 默认值补齐。
func TestInitWithConfigFileAndValidValues(t *testing.T) {
	dir, err := os.Getwd()
	assert.NoError(t, err)
	cfg := config.Init(filepath.Join(dir, "testdata/config_minimal.yaml"))
	assert.NotNil(t, cfg)
	assert.Equal(t, "zap", cfg.LogChannel)
	assert.Equal(t, "db", cfg.CacheDriver)
	assert.Equal(t, true, cfg.GitServerCached)
	assert.Equal(t, "default_domain_manager", cfg.DomainManagerPlugin.Name)
	assert.Equal(t, "ws_sender_memory", cfg.WsSenderPlugin.Name)
	assert.Equal(t, "picture_bing", cfg.PicturePlugin.Name)
	// 最小配置未声明 grpc_port，Init 应自动分配一个端口。
	assert.NotEmpty(t, cfg.GrpcPort)
	// 未声明的上传配置使用默认值。
	assert.Equal(t, "50M", cfg.UploadMaxSize)
	assert.Equal(t, "/tmp/mars-uploads", cfg.UploadDir)
}

// TestIsK8sEnv 验证 kubeconfig 与 K8S 环境变量两种判据。
func TestIsK8sEnv(t *testing.T) {
	t.Run("kubeconfig set", func(t *testing.T) {
		cfg := &config.Config{KubeConfig: "/tmp/kubeconfig"}
		assert.True(t, cfg.IsK8sEnv())
	})

	t.Run("env host and port set", func(t *testing.T) {
		t.Setenv("KUBERNETES_SERVICE_HOST", "10.0.0.1")
		t.Setenv("KUBERNETES_SERVICE_PORT", "443")
		cfg := &config.Config{}
		assert.True(t, cfg.IsK8sEnv())
	})

	t.Run("env cleared", func(t *testing.T) {
		// 显式清空，避免测试机恰好运行在 K8s 容器内导致 False 断言漂移。
		t.Setenv("KUBERNETES_SERVICE_HOST", "")
		t.Setenv("KUBERNETES_SERVICE_PORT", "")
		cfg := &config.Config{}
		assert.False(t, cfg.IsK8sEnv())
	})

	t.Run("only host set", func(t *testing.T) {
		t.Setenv("KUBERNETES_SERVICE_HOST", "10.0.0.1")
		t.Setenv("KUBERNETES_SERVICE_PORT", "")
		cfg := &config.Config{}
		assert.False(t, cfg.IsK8sEnv())
	})
}

// TestInitWithNoConfigFileUsesDefaults 验证未传配置路径时读取当前目录 config.yaml。
func TestInitWithNoConfigFileUsesDefaults(t *testing.T) {
	orig, err := os.Getwd()
	assert.NoError(t, err)
	dir := t.TempDir()
	assert.NoError(t, os.WriteFile(filepath.Join(dir, "config.yaml"), []byte("log_channel: test_channel\n"), 0600))
	assert.NoError(t, os.Chdir(dir))
	defer func() { _ = os.Chdir(orig) }()

	// 清掉其它测试通过 SetConfigFile 残留的全局状态，强制 Init("") 走 SetConfigName 分支。
	viper.SetConfigFile("")

	cfg := config.Init("")
	assert.NotNil(t, cfg)
	assert.Equal(t, "test_channel", cfg.LogChannel)
}

// TestInitWithTypeMismatchConfigPanics 验证配置解码失败时 Init panic。
func TestInitWithTypeMismatchConfigPanics(t *testing.T) {
	assert.Panics(t, func() { config.Init("testdata/config_badtype.yaml") })
}

// TestInitWithUnreachableCwdPanics 覆盖 Init 里 filepath.Abs 失败分支：
// 深目录使 getcwd 超 PATH_MAX（ENAMETOOLONG），相对路径转绝对必然失败。
func TestInitWithUnreachableCwdPanics(t *testing.T) {
	orig, err := os.Getwd()
	assert.NoError(t, err)
	tmp, err := os.MkdirTemp("", "deep-cwd")
	assert.NoError(t, err)
	defer os.RemoveAll(tmp)
	assert.NoError(t, os.Chdir(tmp))
	defer func() { _ = os.Chdir(orig) }()

	for i := 0; i < 800; i++ {
		assert.NoError(t, os.Mkdir("d", 0755))
		assert.NoError(t, os.Chdir("d"))
	}

	assert.Panics(t, func() { config.Init("relative.yaml") })
}
