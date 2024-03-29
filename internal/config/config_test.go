package config

import (
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfig_MaxUploadSize(t *testing.T) {
	// ParseBytes("42 MB") -> 42000000, nil
	// ParseBytes("42 mib") -> 44040192, nil
	cfg := Config{
		UploadMaxSize: "invalid",
	}
	assert.Equal(t, uint64(50*1000*1000), cfg.MaxUploadSize())
	cfg.UploadMaxSize = "50m" // 50,000,000
	assert.Equal(t, uint64(50*1000*1000), cfg.MaxUploadSize())
	cfg.UploadMaxSize = "50mib" // 50 * 1024 * 1024
	assert.Equal(t, uint64(50<<20), cfg.MaxUploadSize())
	cfg.UploadMaxSize = "1Gi" // 1 * 1024 * 1024 * 1024
	assert.Equal(t, uint64(1<<30), cfg.MaxUploadSize())
	cfg.UploadMaxSize = "1G" // 1000,000,000
	assert.Equal(t, uint64(1*math.Pow10(9)), cfg.MaxUploadSize())
	cfg.UploadMaxSize = "1kib" // 1 * 1024
	assert.Equal(t, uint64(1<<10), cfg.MaxUploadSize())
	cfg.UploadMaxSize = "1k" // 1000
	assert.Equal(t, uint64(1*math.Pow10(3)), cfg.MaxUploadSize())
}

func TestPlugin_GetArgs(t *testing.T) {
	plugin := Plugin{
		Name: "",
		Args: nil,
	}
	assert.Equal(t, map[string]any{}, plugin.GetArgs())
	plugin.Args = map[string]any{"name": "duc", "age": 17}
	assert.Equal(t, map[string]any{"name": "duc", "age": 17}, plugin.GetArgs())
}

func TestGetFreePort(t *testing.T) {
	port, err := GetFreePort()
	assert.Nil(t, err)
	assert.Greater(t, port, 0)
}

var pwd, _ = os.Getwd()

func TestInit(t *testing.T) {
	cfg := Init(filepath.Join(pwd, "../../config_example.yaml"))
	assert.Greater(t, cfg.GrpcPort, "0")
	assert.Equal(t, "devops-", cfg.NsPrefix)
	assert.Len(t, cfg.ImagePullSecrets, 2)
	assert.Equal(t, cfg.ImagePullSecrets[0].Server, "https://index.docker.io/v1/")
	assert.Equal(t, cfg.ImagePullSecrets[1].Server, "registry.cn-hangzhou.aliyuncs.com")
	assert.Equal(t, true, cfg.DBSlowLogEnabled)
	assert.Equal(t, 200*time.Millisecond, cfg.DBSlowLogThreshold)
}

func TestPlugin_String(t *testing.T) {
	p := Plugin{
		Name: "test",
		Args: map[string]any{"k": "v"},
	}
	assert.Equal(t, "test k=v", p.String())
}

func TestDockerAuths_String(t *testing.T) {
	auths := DockerAuths{
		{
			Username: "u",
			Password: "p",
			Email:    "e",
			Server:   "s",
		},
	}
	assert.Equal(t, "[username='u' password='p' email='e' server='s']", auths.String())
}

func TestExcludeServerTags_List(t *testing.T) {
	var tests = []struct {
		wants []string
		input string
	}{
		{
			wants: []string{"a", "b", "c"},
			input: "a,b,c",
		},
		{
			wants: []string{"a", "b", "c"},
			input: "a,b, c",
		},
		{
			wants: []string{"a", "b", "c"},
			input: " a, b, c",
		},
		{
			wants: []string{"a"},
			input: " a ",
		},
		{
			wants: []string{"a"},
			input: " ,a ",
		},
	}
	for _, test := range tests {
		tt := test
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.wants, ExcludeServerTags(tt.input).List())
		})
	}
}

func TestDockerAuths_FormatDockerCfg(t *testing.T) {
	var au = DockerAuths{}
	assert.Equal(t, `{"auths":{}}`, string(au.FormatDockerCfg()))

	au = DockerAuths{
		{
			Username: "duc",
			Password: "123",
			Email:    "1@q.c",
			Server:   "https://index.docker.io/v1/",
		},
		{
			Username: "abc",
			Password: "456",
			Email:    "1@q.c",
			Server:   "https://index.reg.io/",
		},
	}
	assert.Equal(t,
		`{"auths":{"https://index.docker.io/v1/":{"username":"duc","password":"123","email":"1@q.c","auth":"ZHVjOjEyMw=="},"https://index.reg.io/":{"username":"abc","password":"456","email":"1@q.c","auth":"YWJjOjQ1Ng=="}}}`,
		string(au.FormatDockerCfg()),
	)
}
