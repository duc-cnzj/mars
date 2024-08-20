package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/duc-cnzj/mars/v4/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestMaxUploadSizeWithValidSize(t *testing.T) {
	cfg := &config.Config{UploadMaxSize: "100Mib"}
	assert.Equal(t, uint64(100*1024*1024), cfg.MaxUploadSize())
}

func TestMaxUploadSizeWithInvalidSize(t *testing.T) {
	cfg := &config.Config{UploadMaxSize: "invalid"}
	assert.Equal(t, uint64(50*1000*1000), cfg.MaxUploadSize())
}

func TestGetFreePort(t *testing.T) {
	port, err := config.GetFreePort()
	assert.Nil(t, err)
	assert.True(t, port > 0)
}

func TestInitWithValidConfigFile(t *testing.T) {
	assert.Panics(t, func() {
		config.Init("invalid_config.yaml")
	})
}

func TestInitWithInvalidConfigFile(t *testing.T) {
	assert.Panics(t, func() { config.Init("invalid_config.yaml") })
}

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

func TestConfigDefaultUploadDir(t *testing.T) {
	cfg := &config.Config{}
	cfg.UploadDir = config.DefaultRootDir
	assert.Equal(t, "/tmp/mars-uploads", cfg.UploadDir)
}

func TestConfigGetArgsWhenArgsAreNil(t *testing.T) {
	plugin := &config.Plugin{Name: "test", Args: nil}
	expected := map[string]any{}
	assert.Equal(t, expected, plugin.GetArgs())
}

func TestConfigGetArgsWhenArgsNotNil(t *testing.T) {
	plugin := &config.Plugin{Name: "test", Args: map[string]any{"arg1": "value1", "arg2": "value2"}}
	expected := map[string]any{"arg1": "value1", "arg2": "value2"}
	assert.Equal(t, expected, plugin.GetArgs())
}

func TestConfigStringRepresentationOfDockerAuths(t *testing.T) {
	auths := config.DockerAuths{
		&config.DockerAuth{Username: "user1", Password: "pass1", Email: "email1", Server: "server1"},
		&config.DockerAuth{Username: "user2", Password: "pass2", Email: "email2", Server: "server2"},
	}
	expected := "[username='user1' password='pass1' email='email1' server='server1'] [username='user2' password='pass2' email='email2' server='server2']"
	assert.Equal(t, expected, auths.String())
}

func TestConfigFormatDockerCfg(t *testing.T) {
	auths := config.DockerAuths{
		&config.DockerAuth{Username: "user", Password: "pass", Email: "email", Server: "server"},
	}
	cfg := auths.FormatDockerCfg()
	assert.NotNil(t, cfg)
}

func TestConfigListExcludeServerTags(t *testing.T) {
	tags := config.ExcludeServerTags("tag1, tag2, tag3")
	expected := []string{"tag1", "tag2", "tag3"}
	assert.Equal(t, expected, tags.List())
}

func TestInitWithValidConfigFileAndValidValues(t *testing.T) {
	dir, _ := os.Getwd()
	cfg := config.Init(filepath.Join(dir, "./testdata/config_minimal.yaml"))
	assert.NotNil(t, cfg)
	assert.Equal(t, "zap", cfg.LogChannel)
	assert.Equal(t, "db", cfg.CacheDriver)
	assert.Equal(t, true, cfg.GitServerCached)
	assert.Equal(t, "default_domain_manager", cfg.DomainManagerPlugin.Name)
	assert.Equal(t, "ws_sender_memory", cfg.WsSenderPlugin.Name)
	assert.Equal(t, "picture_bing", cfg.PicturePlugin.Name)
}
