package config

import (
	"crypto/rsa"
	"log"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"

	"github.com/spf13/viper"
)

type Plugin struct {
	Name string                 `mapstructure:"name"`
	Args map[string]interface{} `mapstructure:"args"`
}

type DockerAuth struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Email    string `mapstructure:"email"`
	Server   string `mapstructure:"server"`
}

type Config struct {
	AppPort        string `mapstructure:"app_port"`
	Debug          bool   `mapstructure:"debug"`
	LogChannel     string `mapstructure:"log_channel"`
	ProfileEnabled bool   `mapstructure:"profile_enabled"`

	AdminPassword string `mapstructure:"admin_password"`
	PrivateKey    string `mapstructure:"private_key"`
	prikey        *rsa.PrivateKey
	pubkey        *rsa.PublicKey

	DockerPlugin         Plugin `mapstructure:"docker_plugin"`
	DomainResolverPlugin Plugin `mapstructure:"domain_resolver_plugin"`
	WsSenderPlugin       Plugin `mapstructure:"ws_sender_plugin"`

	KubeConfig     string `mapstructure:"kubeconfig"`
	NsPrefix       string `mapstructure:"ns_prefix"`
	WildcardDomain string `mapstructure:"wildcard_domain"`
	ClusterIssuer  string `mapstructure:"cluster_issuer"`
	ExternalIp     string `mapstructure:"external_ip"`

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

	OidcEnabled  bool   `mapstructure:"oidc_enabled"`
	ProviderUrl  string `mapstructure:"provider_url"`
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	RedirectUrl  string `mapstructure:"redirect_url"`
}

func (c *Config) Prikey() *rsa.PrivateKey {
	return c.prikey
}

func (c *Config) Pubkey() *rsa.PublicKey {
	return c.pubkey
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

	viper.SetDefault("docker_plugin", map[string]interface{}{
		"name": "docker_default",
		"args": nil,
	})
	viper.SetDefault("domain_resolver_plugin", map[string]interface{}{
		"name": "domain_resolver_default",
		"args": nil,
	})
	viper.SetDefault("ws_sender_plugin", map[string]interface{}{
		"name": "ws_sender_memory",
		"args": nil,
	})

	cfg := &Config{NsPrefix: "devops-"}

	viper.Unmarshal(&cfg)

	pem, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(cfg.PrivateKey))
	if err != nil {
		log.Fatal(err)
	}
	cfg.prikey = pem
	cfg.pubkey = pem.Public().(*rsa.PublicKey)

	return cfg
}

func (c *Config) HasWildcardDomain() bool {
	if c.WildcardDomain != "" {
		return strings.HasPrefix(c.WildcardDomain, "*.")
	}

	return false
}
