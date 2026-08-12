package biz

import "github.com/duc-cnzj/mars/v6/internal/util/yaml"

// StringYamlPrettier 将字符串包装为 YamlPrettier。
type StringYamlPrettier struct {
	Str string
}

// PrettyYaml 原样返回包装的字符串。
func (s *StringYamlPrettier) PrettyYaml() string { return s.Str }

// AnyYamlPrettier 将任意 map 包装为 YamlPrettier。
type AnyYamlPrettier map[string]any

// PrettyYaml 将 map 序列化为格式化 YAML 文本。
func (s AnyYamlPrettier) PrettyYaml() string {
	marshal, _ := yaml.PrettyMarshal(s)
	return string(marshal)
}
