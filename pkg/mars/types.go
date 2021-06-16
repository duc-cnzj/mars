package mars

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"

	"github.com/DuC-cnZj/mars/pkg/utils"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v2"
)

type Config struct {
	ConfigFile       string   `json:"config_file" yaml:"config_file"`
	ConfigFileType   string   `json:"config_file_type" yaml:"config_file_type"`
	DockerRepository string   `json:"docker_repository" yaml:"docker_repository"`
	DockerTagFormat  string   `json:"docker_tag_format" yaml:"docker_tag_format"`
	LocalChartPath   string   `json:"local_chart_path" yaml:"local_chart_path"`
	ConfigField      string   `json:"config_field" yaml:"config_field"`
	IsSimpleEnv      bool     `json:"is_simple_env" yaml:"is_simple_env"`
	DefaultValues    []string `json:"default_values" yaml:"default_values"`
	// TODO Branches 我还没限制
	Branches         []string `json:"branches" yaml:"branches"`
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
			return "", nil, errors.New("unsupport type: " + mars.ConfigFileType)
		}
	}

	return utils.WriteConfigYamlToTmpFile(yamlData)
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

//config_file: .env.example # 必填
//config_file_type: env # 必填 env / yaml
//repository: duc/sso # 必填 仓库名称
//tag_format: "$branch-$commit" # 必填 版本规则 自动注入 $branch, $commit
//local_chart: my_chart.tgz # .tgz 结尾, 如果和 chart 同时配置了，那么优先使用这个配置
//chart: laravel-charts/laravel-workflow # local_chart 未配置时，使用这个配置，如果配置了，那么 helm_repo_name 和 helm_repo_url 应该也是搭配一起配置的
//helm_repo_name: laravel-charts # chart name 默认 空, 可选
//helm_repo_url: https://github.com/DuC-cnZj/laravel-charts # chart url 默认 空, 可选
//chart_version: '0.2.1' # chart_version, 可选
//config_field: envFile # values 中应用配置对应的字段，默认 'env', 可选
//is_simple_env: true # env 是一个 configmap，需要整体配置时, 默认 true, 可选
//default_values: # 默认的 chart values 配置 默认 空, 可选
//- 'redis.enabled=true'
//- 'redis.cluster.slaveCount=0'
//- 'redis.usePassword=false'
//- 'service.type=NodePort'
//branches: # 配置的分支 默认 *, 可选
//- "*"
