package utils

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"gopkg.in/yaml.v2"
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
func deepGet(key string, data map[any]any) (any, bool) {
	var item any = data
	s := strings.Split(key, separator)
	i := 0
	for i < len(s) {
		ii, ok := item.(map[any]any)
		if ok {
			ii, ok := ii[s[i]]
			if ok {
				item = ii
				i++
				continue
			}
		}
		item = nil
		break
	}

	return item, i == len(s)
}

// YamlDeepSetKey 把 'user->name: duc' 设置成
//
//	user:
//	  name: duc
func YamlDeepSetKey(field string, data any) ([]byte, error) {
	if strings.HasPrefix(field, separator) || strings.HasSuffix(field, separator) {
		return nil, fmt.Errorf("%w: %s", ErrorInvalidSeparator, field)
	}

	bf := &bytes.Buffer{}
	encoder := yaml.NewEncoder(bf)

	if err := encoder.Encode(deepSet(field, data)); err != nil {
		return nil, err
	}

	return bf.Bytes(), nil
}
