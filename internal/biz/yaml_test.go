package biz

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringYamlPrettier_PrettyYaml(t *testing.T) {
	s := StringYamlPrettier{Str: "hello"}
	assert.Equal(t, "hello", s.PrettyYaml())
}

func TestAnyYamlPrettier_PrettyYaml(t *testing.T) {
	m := AnyYamlPrettier{"name": "mars", "count": 3}
	got := m.PrettyYaml()
	// yaml.PrettyMarshal 输出带缩进的 YAML 文本，校验关键字段存在即可（map 迭代顺序无关）。
	assert.Contains(t, got, "name: mars")
	assert.Contains(t, got, "count: 3")
}
