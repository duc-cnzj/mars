package yaml

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	goyaml "github.com/goccy/go-yaml"
	"github.com/tidwall/gjson"
)

const Separator = "->"

var (
	ErrorInvalidSeparator = errors.New("error invalid Separator")
)

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

/*
DeepGet: get val

	a:
	  b:
	    c: d

	a->b->c => d
*/
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

// YamlDeepSetKey 把 'user->name: duc' 设置成
//
//	user:
//	  name: duc
func YamlDeepSetKey(field string, data any) ([]byte, error) {
	if strings.HasPrefix(field, Separator) || strings.HasSuffix(field, Separator) {
		return nil, fmt.Errorf("%w: %s", ErrorInvalidSeparator, field)
	}

	return PrettyMarshal(deepSet(field, data))
}

// PrettyMarshal 这里想用 LiteralStyle, 不然前端显示的时候是一坨
func PrettyMarshal(v any) ([]byte, error) {
	return goyaml.MarshalWithOptions(v, goyaml.UseLiteralStyleIfMultiline(true), goyaml.Indent(2), goyaml.IndentSequence(true))
}
