package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/tidwall/gjson"

	"gopkg.in/yaml.v3"
)

const separator = "->"

var (
	ErrorInvalidSeparator = errors.New("error invalid separator")
)

func deepSet(key string, data any) map[string]any {
	res := map[string]any{}

	s := strings.SplitN(key, separator, 2)
	if len(s) > 1 {
		res[s[0]] = deepSet(s[1], data)
	} else {
		res = map[string]any{key: data}
	}

	return res
}

/*
deepGet: get val

	a:
	  b:
	    c: d

	a->b->c => d
*/
func deepGet(key string, data map[string]any) (res any, got bool) {
	keys := strings.Split(key, "->")

	marshal, err := json.Marshal(data)
	if err != nil {
		mlog.Error(err)
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
	if strings.HasPrefix(field, separator) || strings.HasSuffix(field, separator) {
		return nil, fmt.Errorf("%w: %s", ErrorInvalidSeparator, field)
	}

	return yaml.Marshal(deepSet(field, data))
}
