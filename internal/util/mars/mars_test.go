package mars_test

import (
	"strings"
	"testing"

	mars2 "github.com/duc-cnzj/mars/api/v5/mars"
	"github.com/lithammer/dedent"

	"github.com/duc-cnzj/mars/v5/internal/util/mars"
	"github.com/stretchr/testify/assert"
)

func TestGetMarsNamespaceWithPrefix(t *testing.T) {
	ns := "dev"
	prefix := "devops-"
	expected := "devops-dev"

	result := mars.GetMarsNamespace(ns, prefix)

	assert.Equal(t, expected, result)
}

func TestGetMarsNamespaceWithoutPrefix(t *testing.T) {
	ns := "devops-dev"
	prefix := "devops-"
	expected := "devops-dev"

	result := mars.GetMarsNamespace(ns, prefix)

	assert.Equal(t, expected, result)
}

func TestBranchPass(t *testing.T) {
	cfg := &mars2.Config{
		Branches: []string{"master"},
	}
	assert.True(t, mars.BranchPass(cfg.Branches, "master"))
	assert.False(t, mars.BranchPass(cfg.Branches, "dev"))
	cfg = &mars2.Config{
		Branches: []string{"*"},
	}
	assert.True(t, mars.BranchPass(cfg.Branches, "master"))
	cfg = &mars2.Config{
		Branches: []string{"dev-*"},
	}
	assert.True(t, mars.BranchPass(cfg.Branches, "dev-aaa"))
	assert.False(t, mars.BranchPass(cfg.Branches, "nodev-aaa"))
	cfg = &mars2.Config{}
	assert.True(t, mars.BranchPass(cfg.Branches, "dev-aaa"))
	assert.True(t, mars.BranchPass(cfg.Branches, "ccc"))
	cfg = &mars2.Config{Branches: []string{"*-dev"}}
	assert.True(t, mars.BranchPass(cfg.Branches, "a-dev"))
	assert.True(t, mars.BranchPass(cfg.Branches, "b-dev"))

	// regex syntax error
	cfg = &mars2.Config{Branches: []string{"[a-zA-Z]{10000,}*"}}
	assert.False(t, mars.BranchPass(cfg.Branches, strings.Repeat("a", 100000)))
}

func TestParseInputConfig(t *testing.T) {
	var tests = []struct {
		IsSimpleEnv bool
		ConfigField string
		input       string
		wants       string
		ValuesYaml  string
		wantsError  bool
	}{
		{
			IsSimpleEnv: false,
			ConfigField: "conf->config",
			input:       `{"name": "duc", "age": 18}`,
			wants: dedent.Dedent(`
                conf:
                  config:
                    age: 18
                    name: duc
				`),
		},
		{
			IsSimpleEnv: true,
			ConfigField: "conf->config",
			input:       "name: duc\nage: 18",
			// 这里缩进有问题
			wants: `
conf:
  config: |-
    name: duc
    age: 18
`,
		},
		{
			IsSimpleEnv: false,
			ConfigField: "command",
			input: `
command:
  - sh
  - -c
  - "sleep 3600;exit"
`,
			wants: `
command:
  - sh
  - -c
  - sleep 3600;exit
`,
		},
		{
			IsSimpleEnv: false,
			ConfigField: "command",
			input:       `command: ["sh", "-c", "sleep 3600;exit"]`,
			wants: `
command:
  - sh
  - -c
  - sleep 3600;exit
`,
		},
		{
			IsSimpleEnv: false,
			ConfigField: "conf->command",
			input:       `command: ["sh", "-c", "sleep 3600;exit"]`,
			wants: `
conf:
  command:
    - sh
    - -c
    - sleep 3600;exit
`,
		},
		{
			IsSimpleEnv: false,
			ConfigField: "",
			input:       `command: ["sh", "-c", "sleep 3600;exit"]`,
			wants: `
"":
  command:
    - sh
    - -c
    - sleep 3600;exit
`,
		},
		{
			IsSimpleEnv: false,
			ConfigField: "command",
			input: `
command:
  a: b
`,
			wants: `
command:
  command:
    a: b
`,
		},
		{
			IsSimpleEnv: false,
			ConfigField: "command",
			ValuesYaml: `
command:
  command: []
`,
			input: `
command:
  - a
  - b
`,
			wants: `
command:
  command:
    - a
    - b
`,
		},
		{
			IsSimpleEnv: false,
			ConfigField: "command",
			ValuesYaml: `
command: []
`,
			input: `
command:
  - a
  - b
`,
			wants: `
command:
  - a
  - b
`,
		},
		{
			input: "",
			wants: "",
		},
		{
			IsSimpleEnv: true,
			ConfigField: "->command",
			input:       "xxx",
			wants:       "",
			wantsError:  true,
		},
		{
			IsSimpleEnv: true,
			ConfigField: "command->",
			input:       "xxx",
			wants:       "",
			wantsError:  true,
		},
		{
			IsSimpleEnv: false,
			ConfigField: "->command",
			input:       "xxx",
			wants:       "",
			wantsError:  true,
		},
		{
			IsSimpleEnv: false,
			ConfigField: "->command",
			input: `
command:
  - a
  - b
`,
			wants:      "",
			wantsError: true,
		},
	}

	for _, test := range tests {
		tt := test
		t.Run(tt.ConfigField, func(t *testing.T) {
			t.Parallel()
			res, err := mars.ParseInputConfig(&mars2.Config{
				IsSimpleEnv: tt.IsSimpleEnv,
				ValuesYaml:  tt.ValuesYaml,
				ConfigField: tt.ConfigField,
			}, strings.Trim(tt.input, "\n"))
			if tt.wantsError {
				assert.Error(t, err)
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, strings.Trim(tt.wants, "\n"), strings.Trim(res, "\n"))
		})
	}
}

func TestParseInputConfigWithInvalidYaml(t *testing.T) {
	m := &mars2.Config{
		ValuesYaml:  "command: [\"sh\", \"-c\", \"sleep 3600;exit\"]",
		ConfigField: "command",
	}
	input := "command: [\"sh\", \"-c\", \"sleep 3600;exit\"]"
	m.ValuesYaml = "invalid yaml"

	_, err := mars.ParseInputConfig(m, input)

	assert.NotNil(t, err)
}
