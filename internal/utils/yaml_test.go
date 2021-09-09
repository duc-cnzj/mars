package utils

import (
	"bytes"
	"reflect"
	"testing"

	"gopkg.in/yaml.v2"
)

func TestYamlDeepSetKey(t *testing.T) {
	type args struct {
		field string
		data  interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
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
			wantErr: false,
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
			wantErr: false,
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
			wantErr: false,
		},
		{
			name: "fail",
			args: args{
				field: "name->duc->aaaa->",
				data:  "duc",
			},
			want: map[string]interface{}{},
			wantErr: true,
		},
		{
			name: "fail",
			args: args{
				field: "->name->duc->aaaa",
				data:  "duc",
			},
			want: map[string]interface{}{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := bytes.Buffer{}
			yaml.NewEncoder(&bf).Encode(tt.want)
			want := bf.String()
			got, err := YamlDeepSetKey(tt.args.field, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("YamlDeepSetKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if reflect.DeepEqual(got, want) {
				t.Errorf("YamlDeepSetKey() got = %q, want %q", string(got), string(want))
			}
		})
	}
}
