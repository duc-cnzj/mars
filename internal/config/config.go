package config

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/spf13/viper"
)

const DefaultMaxUploadSize = "50M"

type Plugin struct {
	Name string         `mapstructure:"name"`
	Args map[string]any `mapstructure:"args"`
}

func (p Plugin) String() string {
	var args = ""
	for k, v := range p.Args {
		args += fmt.Sprintf("%s=%v", k, v)
	}
	return fmt.Sprintf("%s %s", p.Name, args)
}

func (p Plugin) GetArgs() map[string]any {
	if p.Args == nil {
		return map[string]any{}
	}

	return p.Args
}

type DockerAuths []*DockerAuth

func (a DockerAuths) String() string {
	var strs []string
	for _, auth := range a {
		strs = append(strs, fmt.Sprintf("[%v]", auth))
	}
	return strings.Join(strs, " ")
}

func (a DockerAuths) FormatDockerCfg() []byte {
	var cfg = DockerConfigJSON{Auths: map[string]DockerConfigEntry{}}
	for _, auth := range a {
		cfg.Auths[auth.Server] = DockerConfigEntry{
			Username: auth.Username,
			Password: auth.Password,
			Email:    auth.Email,
			Auth:     base64.StdEncoding.EncodeToString([]byte(auth.Username + ":" + auth.Password)),
		}
	}

	marshal, _ := json.Marshal(cfg)
	return marshal
}

type DockerConfigJSON struct {
	Auths map[string]DockerConfigEntry `json:"auths"`
}

type DockerConfigEntry struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Email    string `json:"email,omitempty"`
	Auth     string `json:"auth,omitempty"`
}

type DockerAuth struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Email    string `mapstructure:"email"`
	Server   string `mapstructure:"server"`
}

func (a DockerAuth) String() string {
	return fmt.Sprintf("username='%s' password='%s' email='%s' server='%s'", a.Username, a.Password, a.Email, a.Server)
}

type ExcludeServerTags string

func (est ExcludeServerTags) List() (res []string) {
	for _, s := range strings.Split(string(est), ",") {
		trims := strings.TrimSpace(s)
		if trims != "" {
			res = append(res, trims)
		}
	}
	return
}

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

	ImagePullSecrets DockerAuths `mapstructure:"imagepullsecrets"`

	InstallTimeout time.Duration `mapstructure:"install_timeout" json:"-"`
	Oidc           []OidcSetting `mapstructure:"oidc"`
}

func (c *Config) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%v)/%s?charset=utf8mb4&parseTime=True&loc=Local", c.DBUsername, c.DBPassword, c.DBHost, c.DBPort, c.DBDatabase)
}

type OidcSetting struct {
	Name         string `mapstructure:"name"`
	Enabled      bool   `mapstructure:"enabled"`
	ProviderUrl  string `mapstructure:"provider_url"`
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	RedirectUrl  string `mapstructure:"redirect_url"`
}

func Init(cfgFile string) *Config {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		dir, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}

		viper.AddConfigPath(dir)
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}

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

	viper.Unmarshal(&cfg)
	for _, s := range cfg.ImagePullSecrets {
		if s.Server == "" {
			s.Server = "https://index.docker.io/v1/"
		}
	}
	if cfg.GrpcPort == "" {
		port, err := GetFreePort()
		if err != nil {
			return nil
		}
		cfg.GrpcPort = fmt.Sprintf("%d", port)
	}

	if cfg.UploadMaxSize == "" {
		cfg.UploadMaxSize = DefaultMaxUploadSize
	}

	if cfg.UploadDir == "" {
		cfg.UploadDir = DefaultRootDir
	}

	return cfg
}

const DefaultRootDir = "/tmp/mars-uploads"

func (c *Config) MaxUploadSize() uint64 {
	bytes, err := humanize.ParseBytes(c.UploadMaxSize)
	if err != nil {
		parseBytes, _ := humanize.ParseBytes(DefaultMaxUploadSize)
		return parseBytes
	}
	return bytes
}

func GetFreePort() (int, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer ln.Close()

	// get port
	tcpAddr, ok := ln.Addr().(*net.TCPAddr)
	if !ok {
		return 0, fmt.Errorf("invalid listen address: %q", ln.Addr().String())
	}

	return tcpAddr.Port, nil
}
