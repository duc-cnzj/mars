package mars

import (
	"regexp"
	"strconv"
	"strings"

	yaml2 "github.com/duc-cnzj/mars/v4/internal/util/yaml"

	"github.com/duc-cnzj/mars/api/v4/mars"
	"gopkg.in/yaml.v3"
)

// GetMarsNamespace
// prefix `devops-`
// namespace    output
// dev          devops-dev
// devops-dev   devops-dev
func GetMarsNamespace(ns, prefix string) string {
	if strings.HasPrefix(ns, prefix) {
		return ns
	}

	return prefix + ns
}

func BranchPass(branches []string, name string) bool {
	if len(branches) < 1 {
		return true
	}

	for _, branch := range branches {
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

func GetProjectConfig(projectId any, branch string) (*mars.Config, error) {
	var marsC mars.Config

	//var gp models.GitProject
	//pid := fmt.Sprintf("%v", projectId)
	//if err := app.DB().Where("`git_project_id` = ?", pid).First(&gp).Error; err == nil {
	//	if gp.GlobalEnabled {
	//		return gp.GlobalMarsConfig(), nil
	//	}
	//}
	//// 因为 protobuf 没有生成yaml的tag，所以需要通过json来转换一下
	//data, err := plugins.GetGitServer().GetFileContentWithBranch(pid, branch, ".mars.yaml")
	//if err != nil {
	//	return nil, err
	//}
	//toJSON, _ := goyaml.YAMLToJSON([]byte(data))
	//json.Unmarshal(toJSON, &marsC)

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
		if yamlData, err = yaml2.YamlDeepSetKey(mars.ConfigField, input); err != nil {
			//mlog.Error(err, mars.ConfigField, input)
			return "", err
		}
	} else {
		var data map[string]any
		decoder := yaml.NewDecoder(strings.NewReader(input))
		if err := decoder.Decode(&data); err != nil {
			return "", err
		}

		split := strings.Split(mars.ConfigField, yaml2.Separator)
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
					_, hasKey := yaml2.DeepGet(mars.ConfigField+yaml2.Separator+key, m)
					if !hasKey {
						newData = value
					}
				}
			}
		}

		if yamlData, err = yaml2.YamlDeepSetKey(mars.ConfigField, newData); err != nil {
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

func IsRemoteLocalChartPath(input string) bool {
	split := strings.Split(input, "|")

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

func GetProjectName[T ~int | ~int32 | ~int64 | ~string](projectID T, marsC *mars.Config) string {
	return ""
	//if marsC.DisplayName != "" {
	//	return marsC.DisplayName
	//}
	//gitProject, _ := plugins.GetGitServer().GetProject(fmt.Sprintf("%v", projectID))
	//return gitProject.GetName()
}
