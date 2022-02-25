package utils

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/plugins"

	"github.com/duc-cnzj/mars-client/v3/mars"
	"gopkg.in/yaml.v2"
)

func BranchPass(mars *mars.MarsConfig, name string) bool {
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

func GetProjectMarsConfig(projectId interface{}, branch string) (*mars.MarsConfig, error) {
	var marsC mars.MarsConfig

	var gp models.GitlabProject
	pid := fmt.Sprintf("%v", projectId)
	if app.DB().Where("`gitlab_project_id` = ?", pid).First(&gp).Error == nil {
		if gp.GlobalEnabled {
			return gp.GlobalMarsConfig(), nil
		}
	}

	// 因为 protobuf 没有生成yaml的tag，所以需要通过json来转换一下
	data, err := plugins.GetGitServer().GetFileContentWithBranch(pid, branch, ".mars.yaml")
	if err != nil {
		return nil, err
	}
	decoder := yaml.NewDecoder(strings.NewReader(data))
	var m map[string]interface{}
	if err := decoder.Decode(&m); err != nil {
		return nil, err
	}
	marshal, err := json.Marshal(&m)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(marshal, &marsC)

	return &marsC, nil
}

func ParseInputConfig(mars *mars.MarsConfig, input string) (string, error) {
	var (
		err      error
		yamlData []byte
	)
	if input == "" {
		return "", nil
	}

	if mars.IsSimpleEnv {
		if yamlData, err = YamlDeepSetKey(mars.ConfigField, input); err != nil {
			return "", err
		}
	} else {
		var data map[string]interface{}
		decoder := yaml.NewDecoder(strings.NewReader(input))
		if err := decoder.Decode(&data); err != nil {
			return "", err
		}

		if yamlData, err = YamlDeepSetKey(mars.ConfigField, data); err != nil {
			return "", err
		}
	}

	return string(yamlData), nil
}

// IsRemoteConfigFile 如果是这个格式意味着是远程项目, "pid|branch|filename"
func IsRemoteConfigFile(mars *mars.MarsConfig) bool {
	split := strings.Split(mars.ConfigFile, "|")

	return len(split) == 3 && intPid(split[0])
}

func IsRemoteChart(mars *mars.MarsConfig) bool {
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
