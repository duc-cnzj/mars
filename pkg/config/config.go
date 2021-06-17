package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type DockerAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Server   string `json:"server"`

	// Server eg: registry.cn-hangzhou.aliyuncs.com
}

type Config struct {
	AppPort        string
	Debug          bool
	LogChannel     string
	ProfileEnabled bool

	KubeConfig string

	WildcardDomain string
	ClusterIssuer  string
	ClusterIp      string

	// mysql
	DBHost     string
	DBPort     string
	DBUsername string
	DBPassword string
	DBDatabase string

	ImagePullSecrets []DockerAuth

	GitlabToken   string
	GitlabBaseURL string
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

	cfg := &Config{
		AppPort:        viper.GetString("app_port"),
		Debug:          viper.GetBool("debug"),
		LogChannel:     viper.GetString("log_channel"),
		ProfileEnabled: viper.GetBool("profile_enabled"),
		KubeConfig:     viper.GetString("kubeconfig"),
		WildcardDomain: viper.GetString("wildcard_domain"),
		ClusterIssuer:  viper.GetString("cluster_issuer"),
		ClusterIp:      viper.GetString("cluster_ip"),
		DBHost:         viper.GetString("db_host"),
		DBPort:         viper.GetString("db_port"),
		DBUsername:     viper.GetString("db_username"),
		DBPassword:     viper.GetString("db_password"),
		DBDatabase:     viper.GetString("db_database"),
		GitlabToken:    viper.GetString("gitlab_token"),
		GitlabBaseURL:  viper.GetString("gitlab_baseurl"),
	}

	dockerAuths := viper.Get("imagepullsecrets")
	if dockerAuths != nil {
		if m, ok := dockerAuths.([]interface{}); ok {
			for _, interf := range m {
				if m2, ok := interf.(map[interface{}]interface{}); ok {
					username, usernameOk := m2["username"]
					password, passwordOk := m2["password"]
					email, emailOk := m2["email"]
					server, serverOk := m2["server"]
					if usernameOk && passwordOk && serverOk {
						var em string
						if emailOk {
							em = email.(string)
						}
						cfg.ImagePullSecrets = append(cfg.ImagePullSecrets, DockerAuth{
							Username: username.(string),
							Password: password.(string),
							Email:    em,
							Server:   server.(string),
						})
					}
				}
			}
		}
	}

	return cfg
}

func (c *Config) HasWildcardDomain() bool {
	if c.WildcardDomain != "" {
		return strings.HasPrefix(c.WildcardDomain, "*.")
	}

	return false
}

func (c *Config) GetDomain(sub string) string {
	if !c.HasWildcardDomain() {
		return ""
	}

	return fmt.Sprintf("%s.%s", sub, strings.TrimLeft(c.WildcardDomain, "*."))
}
