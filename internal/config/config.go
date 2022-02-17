package config

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/spf13/viper"
)

type Plugin struct {
	Name string                 `mapstructure:"name"`
	Args map[string]interface{} `mapstructure:"args"`
}

func (p Plugin) GetArgs() map[string]interface{} {
	if p.Args == nil {
		return map[string]interface{}{}
	}

	return p.Args
}

type DockerAuth struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Email    string `mapstructure:"email"`
	Server   string `mapstructure:"server"`
}

type Config struct {
	AppPort        string `mapstructure:"app_port"`
	GrpcPort       string `mapstructure:"grpc_port"`
	Debug          bool   `mapstructure:"debug"`
	LogChannel     string `mapstructure:"log_channel"`
	ProfileEnabled bool   `mapstructure:"profile_enabled"`

	AdminPassword string `mapstructure:"admin_password"`
	PrivateKey    string `mapstructure:"private_key"`

	DomainManagerPlugin Plugin `mapstructure:"domain_manager_plugin"`
	WsSenderPlugin      Plugin `mapstructure:"ws_sender_plugin"`
	PicturePlugin       Plugin `mapstructure:"picture_plugin"`
	GitServerPlugin     Plugin `mapstructure:"git_server_plugin"`

	UploadDir     string `mapstructure:"upload_dir"`
	UploadMaxSize string `mapstructure:"upload_max_size"`

	KubeConfig string `mapstructure:"kubeconfig"`
	NsPrefix   string `mapstructure:"ns_prefix"`
	ExternalIp string `mapstructure:"external_ip"`

	JaegerUser          string `mapstructure:"jaeger_user"`
	JaegerPassword      string `mapstructure:"jaeger_password"`
	JaegerAgentHostPort string `mapstructure:"jaeger_agent_host_port"`

	// mysql
	DBDriver   string `mapstructure:"db_driver"`
	DBHost     string `mapstructure:"db_host"`
	DBPort     string `mapstructure:"db_port"`
	DBUsername string `mapstructure:"db_username"`
	DBPassword string `mapstructure:"db_password"`
	DBDatabase string `mapstructure:"db_database"`

	ImagePullSecrets []DockerAuth `mapstructure:"imagepullsecrets"`

	GitlabToken   string `mapstructure:"gitlab_token"`
	GitlabBaseURL string `mapstructure:"gitlab_baseurl"`

	InstallTimeout time.Duration `mapstructure:"install_timeout"`
	Oidc           []OidcSetting `mapstructure:"oidc"`
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

	viper.SetDefault("domain_manager_plugin", map[string]interface{}{
		"name": "default_domain_manager",
		"args": nil,
	})

	viper.SetDefault("ws_sender_plugin", map[string]interface{}{
		"name": "ws_sender_memory",
		"args": nil,
	})

	viper.SetDefault("picture_plugin", map[string]interface{}{
		"name": "picture_bing",
		"args": nil,
	})

	cfg := &Config{NsPrefix: "devops-"}

	viper.Unmarshal(&cfg)
	if cfg.GrpcPort == "" {
		port, err := GetFreePort()
		if err != nil {
			return nil
		}
		cfg.GrpcPort = fmt.Sprintf("%d", port)
	}

	return cfg
}

func (c *Config) MaxUploadSize() uint64 {
	bytes, err := humanize.ParseBytes(c.UploadMaxSize)
	if err != nil {
		return 50 << 20
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
