package yaml

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	goyaml "github.com/goccy/go-yaml"
	"github.com/tidwall/gjson"
)

// Separator 是字段路径的分隔符。
const Separator = "->"

var (
	// ErrorInvalidSeparator 表示字段路径以分隔符开头或结尾。
	ErrorInvalidSeparator = errors.New("error invalid Separator")
)

// deepSet 将 key 按分隔符展开为嵌套 map，叶子值为 data。
func deepSet(key string, data any) map[string]any {
	res := map[string]any{}

	s := strings.SplitN(key, Separator, 2)
	if len(s) > 1 {
		res[s[0]] = deepSet(s[1], data)
	} else {
		res = map[string]any{key: data}
	}

	return res
}

// IsSimpleEnv 判断 key 在 yamlData 中是否是简单标量值（非 map）；
// key 缺失时返回 (true, 错误)。
func IsSimpleEnv(key string, yamlData string) (bool, error) {
	var m map[string]any
	if err := yaml.Unmarshal([]byte(yamlData), &m); err != nil {
		return true, err
	}
	if res, got := DeepGet(key, m); got {
		switch res.(type) {
		case map[string]any:
			return false, nil
		default:
			return true, nil
		}
	}
	return true, fmt.Errorf("key '%v' not found", key)
}

// DeepGet 按 `a->b->c` 分隔路径从嵌套 map 中取值：
//
//	a:
//	  b:
//	    c: d
//
// a->b->c => d
func DeepGet(key string, data map[string]any) (res any, got bool) {
	keys := strings.Split(key, "->")

	marshal, err := json.Marshal(data)
	if err != nil {
		//mlog.Error(err)
		return nil, false
	}
	value := gjson.Get(string(marshal), strings.Join(keys, "."))
	return value.Value(), value.Exists()
}

// DeepSetKey 把 'user->name: duc' 设置成
//
//	user:
//	  name: duc
func DeepSetKey(field string, data any) ([]byte, error) {
	if strings.HasPrefix(field, Separator) || strings.HasSuffix(field, Separator) {
		return nil, fmt.Errorf("%w: %s", ErrorInvalidSeparator, field)
	}

	return PrettyMarshal(deepSet(field, data))
}

// PrettyMarshal 这里想用 LiteralStyle, 不然前端显示的时候是一坨
func PrettyMarshal(v any) ([]byte, error) {
	return goyaml.MarshalWithOptions(v, goyaml.UseLiteralStyleIfMultiline(true), goyaml.Indent(2), goyaml.IndentSequence(true))
}
