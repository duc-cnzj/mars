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
		data  any
	}
	tests := []struct {
		name string
		args args
		want any
		err  error
	}{
		{
			name: "ok",
			args: args{
				field: "name->duc",
				data:  "duc",
			},
			want: map[string]any{
				"name": map[string]any{
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
			want: map[string]any{
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
			want: map[string]any{
				"name": map[string]any{
					"duc": map[string]any{
						"a": map[string]any{
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

func Test_deepGet(t *testing.T) {
	var tests = []struct {
		input   map[any]any
		key     string
		wants   bool
		wantRes any
	}{
		{
			input: map[any]any{
				"a": map[any]any{
					"b": map[any]any{
						"c": map[any]any{
							"d": "d",
						},
					},
				},
			},
			key:   "a->b->c",
			wants: true,
			wantRes: map[any]any{
				"d": "d",
			},
		},
		{
			input: map[any]any{
				"a": map[any]any{
					"b": map[any]any{},
				},
			},
			key:     "a->b->c",
			wants:   false,
			wantRes: nil,
		},
		{
			input: map[any]any{
				"a": map[any]any{
					"b": map[any]any{},
				},
			},
			key:     "",
			wants:   false,
			wantRes: nil,
		},
	}
	for _, test := range tests {
		tt := test
		t.Run(tt.key, func(t *testing.T) {
			res, b := deepGet(tt.key, tt.input)
			assert.Equal(t, tt.wants, b)
			assert.Equal(t, tt.wantRes, res)
		})
	}
}
