package utils

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"

	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/pkg/mars"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v2"
)

func BranchPass(mars *mars.Config, name string) bool {
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

func ParseInputConfigToMap(mars *mars.Config, input string) (map[string]interface{}, error) {
	data, err := ParseInputConfig(mars, input)
	if err != nil {
		return nil, err
	}
	v := map[string]interface{}{}
	if err := yaml.Unmarshal([]byte(data), &v); err != nil {
		return nil, err
	}
	return v, nil
}

func ParseInputConfig(mars *mars.Config, input string) (string, error) {
	var (
		err      error
		yamlData []byte
	)
	if mars.IsSimpleEnv {
		if yamlData, err = YamlDeepSetKey(mars.ConfigField, input); err != nil {
			return "", err
		}
	} else {
		switch mars.ConfigFileType {
		case "":
			return "", nil
		case "yaml":
			var data map[string]interface{}
			decoder := yaml.NewDecoder(strings.NewReader(input))
			if err := decoder.Decode(&data); err != nil {
				return "", err
			}

			if yamlData, err = YamlDeepSetKey(mars.ConfigField, data); err != nil {
				return "", err
			}
		case "json":
			var data map[string]interface{}
			if err := json.Unmarshal([]byte(input), &data); err != nil {
				return "", err
			}

			if yamlData, err = YamlDeepSetKey(mars.ConfigField, data); err != nil {
				return "", err
			}
		case "env", "dotenv", ".env":
			parse, err := godotenv.Parse(strings.NewReader(input))
			if err != nil {
				return "", err
			}

			if yamlData, err = YamlDeepSetKey(mars.ConfigField, parse); err != nil {
				return "", err
			}
		default:
			mlog.Error("unsupport type: " + mars.ConfigFileType)
			return "", nil
		}
	}

	return string(yamlData), nil
}

// IsRemoteConfigFile 如果是这个格式意味着是远程项目, "pid|branch|filename"
func IsRemoteConfigFile(mars *mars.Config) bool {
	split := strings.Split(mars.ConfigFile, "|")

	return len(split) == 3 && intPid(split[0])
}

func IsRemoteChart(mars *mars.Config) bool {
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