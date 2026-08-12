package biz

import (
	"regexp"
	"strconv"
	"strings"

	yaml2 "github.com/duc-cnzj/mars/v6/internal/util/yaml"

	"github.com/duc-cnzj/mars/api/v6/proto/mars"
	"gopkg.in/yaml.v3"
)

// GetNamespace 为命名空间名称补上 prefix 前缀，已带前缀时原样返回。
// 例如 prefix `devops-`：
//
//	namespace    output
//	dev          devops-dev
//	devops-dev   devops-dev
func GetNamespace(ns, prefix string) string {
	if strings.HasPrefix(ns, prefix) {
		return ns
	}

	return prefix + ns
}

// IsRemoteLocalChartPath 判断 input 是否为"远程本地 chart 路径"：以 "|" 分隔为三段，
// 且首段 uid 可解析为十进制整数（git 项目 ID）。该格式由前端以 "git 项目 ID|分支|chart 目录"
// 拼出，data.GitRepo 与 deploy.ChartFileLoader 共用同一判定，故作为 mars 域格式收进 biz
// （util/ 只收通用抽象，域语义归领域层；对齐 util/annotation → biz/annotations.go 先例）。
func IsRemoteLocalChartPath(input string) bool {
	split := strings.Split(input, "|")
	return len(split) == 3 && intPid(split[0])
}

// intPid 判断 pid 是否为十进制整数。
func intPid(pid string) bool {
	_, err := strconv.ParseInt(pid, 10, 64)
	return err == nil
}

// MatchBranch 判断 name 是否命中 branches 中的任一规则。
// branches 为空时恒为 true；支持精确匹配与 `*` 通配符（正则展开匹配）。
func MatchBranch(branches []string, name string) bool {
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

// ParseInputConfig 将用户输入 input 解析后写入 config 的 ConfigField 字段，
// 返回深合并后的 yaml 字符串；input 为空时返回空串。
//
// 简单环境变量（IsSimpleEnv）直接以字段路径深写入；否则先解析 map，
// 再按分隔符路径合并到 ConfigField 指向的节点。
func ParseInputConfig(config *mars.Config, input string) (string, error) {
	var (
		err      error
		yamlData []byte
	)
	if input == "" {
		return "", nil
	}

	if config.IsSimpleEnv {
		if yamlData, err = yaml2.DeepSetKey(config.ConfigField, input); err != nil {
			return "", err
		}
	} else {
		var data map[string]any
		decoder := yaml.NewDecoder(strings.NewReader(input))
		if err := decoder.Decode(&data); err != nil {
			return "", err
		}

		split := strings.Split(config.ConfigField, yaml2.Separator)
		var key = config.ConfigField
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
					if err := yaml.Unmarshal([]byte(config.ValuesYaml), m); err != nil {
						return "", err
					}
					_, hasKey := yaml2.DeepGet(config.ConfigField+yaml2.Separator+key, m)
					if !hasKey {
						newData = value
					}
				}
			}
		}

		if yamlData, err = yaml2.DeepSetKey(config.ConfigField, newData); err != nil {
			return "", err
		}
	}

	return string(yamlData), nil
}
