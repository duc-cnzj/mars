package config

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/spf13/viper"
)

// defaultMaxUploadSize 是未显式配置时上传体积上限的默认值（humanize 可解析格式）。
const defaultMaxUploadSize = "50M"

// defaultRootDir 是未显式配置时本地上传目录的默认值。
const defaultRootDir = "/tmp/mars-uploads"

// Plugin 描述一个可加载插件的名称与构造参数，对应配置文件的 *_plugin 段。
type Plugin struct {
	Name string         `mapstructure:"name"`
	Args map[string]any `mapstructure:"args"`
}

// DockerAuths 是多个 Docker 仓库登录凭据的列表。
type DockerAuths []*DockerAuth

// dockerConfigJSON 是写入 docker config.json 的序列化结构。
type dockerConfigJSON struct {
	Auths map[string]dockerConfigEntry `json:"auths"`
}

// dockerConfigEntry 对应 docker config.json 中单个仓库的凭据条目。
type dockerConfigEntry struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Email    string `json:"email,omitempty"`
	Auth     string `json:"auth,omitempty"`
}

// DockerAuth 描述一个 Docker 仓库的登录凭据。
type DockerAuth struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Email    string `mapstructure:"email"`
	Server   string `mapstructure:"server"`
}

// FormatDockerCfg 把凭据列表序列化成 docker config.json 字节串，
// 每个仓库生成 base64 编码的 Auth 字段（username:password）。
func (a DockerAuths) FormatDockerCfg() []byte {
	var cfg = dockerConfigJSON{Auths: map[string]dockerConfigEntry{}}
	for _, auth := range a {
		cfg.Auths[auth.Server] = dockerConfigEntry{
			Username: auth.Username,
			Password: auth.Password,
			Email:    auth.Email,
			Auth:     base64.StdEncoding.EncodeToString([]byte(auth.Username + ":" + auth.Password)),
		}
	}

	marshal, _ := json.Marshal(cfg)
	return marshal
}

// ExcludeServerTags 是以 ',' 分隔的启动排除服务标签列表。
type ExcludeServerTags string

// List 把标签串按 ',' 拆分，逐项去空白并过滤空项。
func (est ExcludeServerTags) List() (res []string) {
	for _, s := range strings.Split(string(est), ",") {
		trims := strings.TrimSpace(s)
		if trims != "" {
			res = append(res, trims)
		}
	}
	return
}

// Config 汇总服务运行所需的全部配置，字段与配置文件 key 一一对应。
type Config struct {
	AppPort         string `mapstructure:"app_port"`
	GrpcPort        string `mapstructure:"grpc_port"`
	Debug           bool   `mapstructure:"debug"`
	LogChannel      string `mapstructure:"log_channel"`
	GitServerCached bool   `mapstructure:"git_server_cached"`
	CacheDriver     string `mapstructure:"cache_driver"`
	// 启动时排除这些服务，用 ',' 隔开
	ExcludeServer ExcludeServerTags `mapstructure:"exclude_server"`

	MetricsPort string `mapstructure:"metrics_port"`

	AdminPassword string `mapstructure:"admin_password"`
	PrivateKey    string `mapstructure:"private_key"  json:"-"`

	DomainManagerPlugin Plugin `mapstructure:"domain_manager_plugin"`
	WsSenderPlugin      Plugin `mapstructure:"ws_sender_plugin"`
	PicturePlugin       Plugin `mapstructure:"picture_plugin"`
	GitServerPlugin     Plugin `mapstructure:"git_server_plugin"`

	UploadDir     string `mapstructure:"upload_dir"`
	UploadMaxSize string `mapstructure:"upload_max_size"`

	S3Enabled         bool   `mapstructure:"s3_enabled"`
	S3Endpoint        string `mapstructure:"s3_endpoint"`
	S3AccessKeyID     string `mapstructure:"s3_access_key_id"`
	S3SecretAccessKey string `mapstructure:"s3_secret_access_key"`
	S3Bucket          string `mapstructure:"s3_bucket"`
	S3UseSSL          bool   `mapstructure:"s3_use_ssl"`

	KubeConfig string `mapstructure:"kubeconfig"`
	NsPrefix   string `mapstructure:"ns_prefix"`
	ExternalIp string `mapstructure:"external_ip"`

	TracingEndpoint string `mapstructure:"tracing_endpoint"`

	// mysql
	DBDriver           string        `mapstructure:"db_driver"`
	DBHost             string        `mapstructure:"db_host"`
	DBPort             string        `mapstructure:"db_port"`
	DBUsername         string        `mapstructure:"db_username"`
	DBPassword         string        `mapstructure:"db_password"`
	DBDatabase         string        `mapstructure:"db_database"`
	DBSlowLogEnabled   bool          `mapstructure:"db_slow_log_enabled"`
	DBSlowLogThreshold time.Duration `mapstructure:"db_slow_log_threshold"`
	DBDebug            bool          `mapstructure:"db_debug"`
	DBAutoMigrate      bool          `mapstructure:"db_auto_migrate"`

	ImagePullSecrets DockerAuths `mapstructure:"imagepullsecrets"`

	InstallTimeout time.Duration `mapstructure:"install_timeout" json:"-"`
	Oidc           []OidcSetting `mapstructure:"oidc"`
}

// IsK8sEnv 判断是否运行在 Kubernetes 环境：显式配置了 kubeconfig，
// 或容器内注入了 KUBERNETES_SERVICE_HOST/PORT 环境变量。
func (c *Config) IsK8sEnv() bool {
	return c.KubeConfig != "" || (os.Getenv("KUBERNETES_SERVICE_HOST") != "" && os.Getenv("KUBERNETES_SERVICE_PORT") != "")
}

// DSN 拼出 MySQL 连接串。
func (c *Config) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%v)/%s?charset=utf8mb4&parseTime=True&loc=Local", c.DBUsername, c.DBPassword, c.DBHost, c.DBPort, c.DBDatabase)
}

// OidcSetting 描述一个 OIDC 提供方的配置。
type OidcSetting struct {
	Name         string `mapstructure:"name"`
	Enabled      bool   `mapstructure:"enabled"`
	ProviderUrl  string `mapstructure:"provider_url"`
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	RedirectUrl  string `mapstructure:"redirect_url"`
}

// Init 依据配置文件构造 Config：路径缺省时读取当前目录下的 config.yaml，
// 读取或解码失败即 panic（fail-fast）。随后补齐插件、端口、上传等默认值。
func Init(cfgFile string) *Config {
	if cfgFile != "" {
		if !filepath.IsAbs(cfgFile) {
			viper.AddConfigPath(".")
			abs, err := filepath.Abs(cfgFile)
			if err != nil {
				panic(err)
			}
			viper.SetConfigFile(abs)
		} else {
			viper.SetConfigFile(cfgFile)
		}
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	viper.SetDefault("log_channel", "zap")
	viper.SetDefault("cache_driver", "db")
	viper.SetDefault("git_server_cached", true)
	viper.SetDefault("domain_manager_plugin", map[string]any{
		"name": "default_domain_manager",
		"args": nil,
	})

	viper.SetDefault("ws_sender_plugin", map[string]any{
		"name": "ws_sender_memory",
		"args": nil,
	})

	viper.SetDefault("picture_plugin", map[string]any{
		"name": "picture_bing",
		"args": nil,
	})

	cfg := &Config{NsPrefix: "devops-"}

	if err := viper.Unmarshal(&cfg); err != nil {
		panic(err)
	}
	for _, s := range cfg.ImagePullSecrets {
		if s.Server == "" {
			s.Server = "https://index.docker.io/v1/"
		}
	}
	if cfg.GrpcPort == "" {
		port, err := GetFreePort()
		if err != nil {
			// 与 Init 其余报错路径一致的 fail-fast：绝不静默返回 nil
			//（cmd 调用方不判 nil，静默 nil 会直接 nil-deref）。
			panic(err)
		}
		cfg.GrpcPort = fmt.Sprintf("%d", port)
	}

	if cfg.UploadMaxSize == "" {
		cfg.UploadMaxSize = defaultMaxUploadSize
	}

	if cfg.UploadDir == "" {
		cfg.UploadDir = defaultRootDir
	}

	return cfg
}

// MaxUploadSize 把 UploadMaxSize 解析为字节数，解析失败回退到默认上限。
func (c *Config) MaxUploadSize() uint64 {
	bytes, err := humanize.ParseBytes(c.UploadMaxSize)
	if err != nil {
		parseBytes, _ := humanize.ParseBytes(defaultMaxUploadSize)
		return parseBytes
	}
	return bytes
}

// listenTCP 是 GetFreePort 分配端口用的监听探针：测试可整体替换它以
// 确定性覆盖 net.Listen 失败的错误分支（真实网络层无法稳定触发）。
var listenTCP = net.Listen

// GetFreePort 监听回环地址的临时端口并返回内核分配的端口号。
func GetFreePort() (int, error) {
	ln, err := listenTCP("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer ln.Close()

	return ln.Addr().(*net.TCPAddr).Port, nil
}
