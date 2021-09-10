package utils

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"gopkg.in/yaml.v2"
)

func TestYamlDeepSetKey(t *testing.T) {
	type args struct {
		field string
		data  interface{}
	}
	tests := []struct {
		name string
		args args
		want interface{}
		err  error
	}{
		{
			name: "ok",
			args: args{
				field: "name->duc",
				data:  "duc",
			},
			want: map[string]interface{}{
				"name": map[string]interface{}{
					"duc": "duc",
				},
			},
			err: nil,
		},
		{
			name: "ok2",
			args: args{
				field: "name",
				data:  "duc",
			},
			want: map[string]interface{}{
				"name": "duc",
			},
			err: nil,
		},
		{
			name: "ok2",
			args: args{
				field: "name->duc->a->b",
				data:  "duc",
			},
			want: map[string]interface{}{
				"name": map[string]interface{}{
					"duc": map[string]interface{}{
						"a": map[string]interface{}{
							"b": "duc",
						},
					},
				},
			},
			err: nil,
		},
		{
			name: "fail",
			args: args{
				field: "name->duc->aaaa->",
				data:  "duc",
			},
			err: ErrorInvalidSeparator,
		},
		{
			name: "fail",
			args: args{
				field: "->name->duc->aaaa",
				data:  "duc",
			},
			err: ErrorInvalidSeparator,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Helper()
			bf := bytes.Buffer{}
			yaml.NewEncoder(&bf).Encode(tt.want)
			want := bf.Bytes()
			got, err := YamlDeepSetKey(tt.args.field, tt.args.data)
			assert.ErrorIs(t, err, tt.err)
			if err == nil {
				assert.Equal(t, got, want)
			}
		})
	}
}
