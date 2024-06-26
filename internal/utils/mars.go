package utils

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/duc-cnzj/mars/api/v4/mars"
	app "github.com/duc-cnzj/mars/v4/internal/app/helper"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/models"
	"github.com/duc-cnzj/mars/v4/internal/plugins"

	goyaml "github.com/ghodss/yaml"
	"gopkg.in/yaml.v3"
)

func BranchPass(mars *mars.Config, name string) bool {
	if len(mars.Branches) < 1 {
		return true
	}

	for _, branch := range mars.Branches {
		if branch == "*" || branch == name {
			return true
		}

		if strings.Contains(branch, "*") {
			branch = strings.ReplaceAll(branch, "*", ".*")
			compile, err := regexp.Compile(branch)
			if err != nil {
				continue
			}

			return compile.FindString(name) == name
		}
	}

	return false
}

func GetProjectMarsConfig(projectId any, branch string) (*mars.Config, error) {
	var marsC mars.Config

	var gp models.GitProject
	pid := fmt.Sprintf("%v", projectId)
	if err := app.DB().Where("`git_project_id` = ?", pid).First(&gp).Error; err == nil {
		if gp.GlobalEnabled {
			return gp.GlobalMarsConfig(), nil
		}
	}
	// 因为 protobuf 没有生成yaml的tag，所以需要通过json来转换一下
	data, err := plugins.GetGitServer().GetFileContentWithBranch(pid, branch, ".mars.yaml")
	if err != nil {
		return nil, err
	}
	toJSON, _ := goyaml.YAMLToJSON([]byte(data))
	json.Unmarshal(toJSON, &marsC)

	return &marsC, nil
}

/*
	values.yaml:
	 ```
	 command:
	   command:
	     - sleep 3600
	 ```

	 config_field: command

	 input:
	 ```
	 command:
	  - app
	 ```

	 wants:
	 ```
	  command:
	    command:
	    - app
	 ```

------------------------------------

	   values.yaml:
	   command:
	   - sleep 3600

	   config_field: command

	   input:
	   ```
	   command:
		- app
	   ```

	   wants:
	   ```
		command:
	    - app
	   ```
*/
func ParseInputConfig(mars *mars.Config, input string) (string, error) {
	var (
		err      error
		yamlData []byte
	)
	if input == "" {
		return "", nil
	}

	if mars.IsSimpleEnv {
		if yamlData, err = YamlDeepSetKey(mars.ConfigField, input); err != nil {
			mlog.Error(err, mars.ConfigField, input)
			return "", err
		}
	} else {
		var data map[string]any
		decoder := yaml.NewDecoder(strings.NewReader(input))
		if err := decoder.Decode(&data); err != nil {
			return "", err
		}

		split := strings.Split(mars.ConfigField, separator)
		var key = mars.ConfigField
		if len(split) > 0 {
			key = split[len(split)-1]
		}
		var newData any = data
		if len(data) == 1 {
			cdata, ok := data[key]
			if ok {
				value, ok := cdata.([]any)
				if ok {
					m := make(map[string]any)
					if err := yaml.Unmarshal([]byte(mars.ValuesYaml), m); err != nil {
						return "", err
					}
					_, hasKey := deepGet(mars.ConfigField+separator+key, m)
					if !hasKey {
						newData = value
					}
				}
			}
		}

		if yamlData, err = YamlDeepSetKey(mars.ConfigField, newData); err != nil {
			return "", err
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

func GetProjectName[T ~int | ~int64 | ~string](projectID T, marsC *mars.Config) string {
	if marsC.DisplayName != "" {
		return marsC.DisplayName
	}
	gitProject, _ := plugins.GetGitServer().GetProject(fmt.Sprintf("%v", projectID))
	return gitProject.GetName()
}
