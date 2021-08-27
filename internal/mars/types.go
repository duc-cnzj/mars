package mars

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/duc-cnzj/mars/internal/mlog"

	"github.com/duc-cnzj/mars/internal/utils"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v2"
)

// Config mars 配置文件，默认取的是当前 branch 的最新的 .mars.yaml
type Config struct {
	// ConfigFile 指定项目下的默认配置文件, 也可以是别的项目的文件，格式为 "pid|branch|filename"
	ConfigFile string `json:"config_file" yaml:"config_file"`
	// ConfigFileValues 全局配置文件，如果没有 ConfigFile 则使用这个
	ConfigFileValues string `json:"config_file_values" yaml:"config_file_values"`

	// ConfigFileType 配置文件类型，php/env/yaml...
	ConfigFileType   string `json:"config_file_type" yaml:"config_file_type"`
	DockerRepository string `json:"docker_repository" yaml:"docker_repository"`

	// DockerTagFormat 可用变量 {{.Branch}} {{.Commit}} {{.Pipeline}}
	DockerTagFormat string `json:"docker_tag_format" yaml:"docker_tag_format"`

	// LocalChartPath helm charts 目录, charts 文件在项目中存放的目录(必填), 也可以是别的项目的文件，格式为 "pid|branch|path"
	LocalChartPath string `json:"local_chart_path" yaml:"local_chart_path"`
	ConfigField    string `json:"config_field" yaml:"config_field"`
	IsSimpleEnv    bool   `json:"is_simple_env" yaml:"is_simple_env"`

	// DefaultValues 默认的配置，和 values.yaml 一样写就行了
	DefaultValues map[string]interface{} `json:"default_values" yaml:"default_values"`

	// Branches 启用的分支
	Branches []string `json:"branches" yaml:"branches"`

	// 如果默认的ingress 规则不符合，你可以通过这个重写
	// 可用变量 {{Host1}} {{TlsSecret1}} {{Host2}} {{TlsSecret2}} {{Host3}} {{TlsSecret3}} ... {{Host10}} {{TlsSecret10}}
	IngressOverwriteValues []string `json:"ingress_overwrite_values" yaml:"ingress_overwrite_values"`

	// Atomic 既 helm 中的 --atomic
	Atomic bool `json:"atomic" yaml:"atomic"`

	ImagePullSecrets []string `json:"image_pull_secrets"`
}

func (mars *Config) BranchPass(name string) bool {
	if len(mars.Branches) < 1 {
		return true
	}

	for _, branch := range mars.Branches {
		if branch == "*" || branch == name {
			return true
		}

		if strings.Contains(branch, ".*?") {
			compile, err := regexp.Compile(branch)
			if err != nil {
				continue
			}

			return compile.FindString(name) == name
		}
	}

	return false
}

// GenerateDefaultValuesYamlFile 支持变量 "$imagePullSecrets"
func (mars *Config) GenerateDefaultValuesYamlFile() (string, func(), error) {
	if len(mars.DefaultValues) < 1 {
		return "", func() {}, nil
	}
	bf := &bytes.Buffer{}
	encoder := yaml.NewEncoder(bf)
	if err := encoder.Encode(mars.DefaultValues); err != nil {
		return "", nil, err
	}

	b := strings.Replace(bf.String(), `$imagePullSecrets`, `[{{- range .ImagePullSecrets }}{name: {{ . }}}, {{- end }}]`, -1)

	parse, e := template.New("").Parse(b)
	if e != nil {
		return "", nil, e
	}

	renderResult := &bytes.Buffer{}
	if err := parse.Execute(renderResult, struct {
		ImagePullSecrets []string
	}{
		ImagePullSecrets: mars.ImagePullSecrets,
	}); err != nil {
		return "", nil, err
	}

	res := renderResult.Bytes()
	mlog.Warning("GenerateDefaultValuesYamlFile", string(res))

	return utils.WriteConfigYamlToTmpFile(res)

}

func (mars *Config) GenerateConfigYamlFileByInput(input string) (string, func(), error) {
	var (
		err      error
		yamlData []byte
	)
	if mars.IsSimpleEnv {
		if yamlData, err = utils.EncodeConfigToYaml(mars.ConfigField, input); err != nil {
			return "", nil, err
		}
	} else {
		switch mars.ConfigFileType {
		case "yaml":
			var data map[string]interface{}
			decoder := yaml.NewDecoder(strings.NewReader(input))
			if err := decoder.Decode(&data); err != nil {
				return "", nil, err
			}

			if yamlData, err = EncodeConfigToYaml(mars.ConfigField, data); err != nil {
				return "", nil, err
			}
		case "json":
			var data map[string]interface{}
			if err := json.Unmarshal([]byte(input), &data); err != nil {
				return "", nil, err
			}

			if yamlData, err = EncodeConfigToYaml(mars.ConfigField, data); err != nil {
				return "", nil, err
			}
		case "env", "dotenv", ".env":
			parse, err := godotenv.Parse(strings.NewReader(input))
			if err != nil {
				return "", nil, err
			}

			if yamlData, err = EncodeConfigToYaml(mars.ConfigField, parse); err != nil {
				return "", nil, err
			}
		default:
			mlog.Error("unsupport type: " + mars.ConfigFileType)
			return "", func() {}, nil
		}
	}

	return utils.WriteConfigYamlToTmpFile(yamlData)
}

// IsRemoteConfigFile 如果是这个格式意味着是远程项目, "pid|branch|filename"
func (mars *Config) IsRemoteConfigFile() bool {
	split := strings.Split(mars.ConfigFile, "|")

	return len(split) == 3 && intPid(split[0])
}

func (mars *Config) IsRemoteChart() bool {
	split := strings.Split(mars.LocalChartPath, "|")
	// 如果是这个格式意味着是远程项目, 'uid|branch|path'

	return len(split) == 3 && intPid(split[0])
}

func intPid(pid string) bool {
	if _, err := strconv.ParseInt(pid, 10, 64); err == nil {
		return true
	}
	return false
}

func EncodeConfigToYaml(field string, data interface{}) ([]byte, error) {
	bf := &bytes.Buffer{}
	encoder := yaml.NewEncoder(bf)
	if err := encoder.Encode(map[string]interface{}{
		field: data,
	}); err != nil {
		return nil, err
	}

	return bf.Bytes(), nil
}
