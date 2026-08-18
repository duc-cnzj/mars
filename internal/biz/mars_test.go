package biz_test

import (
	"strings"
	"testing"

	mars2 "github.com/duc-cnzj/mars/api/v6/proto/mars"
	"github.com/lithammer/dedent"

	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/stretchr/testify/assert"
)

func TestGetMarsNamespaceWithPrefix(t *testing.T) {
	ns := "dev"
	prefix := "devops-"
	expected := "devops-dev"

	result := biz.GetNamespace(ns, prefix)

	assert.Equal(t, expected, result)
}

func TestGetMarsNamespaceWithoutPrefix(t *testing.T) {
	ns := "devops-dev"
	prefix := "devops-"
	expected := "devops-dev"

	result := biz.GetNamespace(ns, prefix)

	assert.Equal(t, expected, result)
}

func TestBranchPass(t *testing.T) {
	cfg := &mars2.Config{
		Branches: []string{"master"},
	}
	assert.True(t, biz.MatchBranch(cfg.Branches, "master"))
	assert.False(t, biz.MatchBranch(cfg.Branches, "dev"))
	cfg = &mars2.Config{
		Branches: []string{"*"},
	}
	assert.True(t, biz.MatchBranch(cfg.Branches, "master"))
	cfg = &mars2.Config{
		Branches: []string{"dev-*"},
	}
	assert.True(t, biz.MatchBranch(cfg.Branches, "dev-aaa"))
	assert.False(t, biz.MatchBranch(cfg.Branches, "nodev-aaa"))
	cfg = &mars2.Config{}
	assert.True(t, biz.MatchBranch(cfg.Branches, "dev-aaa"))
	assert.True(t, biz.MatchBranch(cfg.Branches, "ccc"))
	cfg = &mars2.Config{Branches: []string{"*-dev"}}
	assert.True(t, biz.MatchBranch(cfg.Branches, "a-dev"))
	assert.True(t, biz.MatchBranch(cfg.Branches, "b-dev"))

	// regex syntax error
	cfg = &mars2.Config{Branches: []string{"[a-zA-Z]{10000,}*"}}
	assert.False(t, biz.MatchBranch(cfg.Branches, strings.Repeat("a", 100000)))
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
			res, err := biz.ParseInputConfig(&mars2.Config{
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

func TestIsRemoteLocalChartPath(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"三段+整数 uid", "12|master|charts/app", true},
		{"uid 为 int64 最大值", "9223372036854775807|master|charts/app", true},
		{"uid 溢出 int64", "9223372036854775808|master|charts/app", false},
		{"uid 非整数", "abc|master|charts/app", false},
		{"uid 为小数", "1.5|master|charts/app", false},
		{"uid 为空段", "|master|charts/app", false},
		{"仅两段", "12|master", false},
		{"四段", "12|master|charts/app|extra", false},
		{"空串", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, biz.IsRemoteLocalChartPath(tt.input))
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

	_, err := biz.ParseInputConfig(m, input)

	assert.NotNil(t, err)
}

// TestPipelinePassStatus 验证流水线通过规则的判定：
// 无规则返回原始状态；全部命中成功才 success；命中失败/运行/缺失各自按语义返回。
// 原始状态用 StatusUnknown 区分，确保结果来自规则计算而非原样透传。
func TestPipelinePassStatus(t *testing.T) {
	job := func(stage, name string, status biz.Status) biz.PipelineJob {
		return biz.PipelineJob{Name: name, Status: status, StageName: stage}
	}
	rule := func(stage, name string) *mars2.PipelinePassRule {
		return &mars2.PipelinePassRule{StageName: stage, JobName: name}
	}
	tests := []struct {
		name     string
		pipeline *biz.Pipeline
		rules    []*mars2.PipelinePassRule
		want     biz.Status
	}{
		{
			name:     "无规则返回原始状态",
			pipeline: &biz.Pipeline{Status: biz.StatusUnknown, Jobs: []biz.PipelineJob{job("stage2", "test", biz.StatusFailed)}},
			rules:    nil,
			want:     biz.StatusUnknown,
		},
		{
			name:     "全部规则命中成功",
			pipeline: &biz.Pipeline{Status: biz.StatusFailed, Jobs: []biz.PipelineJob{job("stage2", "test", biz.StatusFailed), job("stage1", "build", biz.StatusSuccess)}},
			rules:    []*mars2.PipelinePassRule{rule("stage1", "build")},
			want:     biz.StatusSuccess,
		},
		{
			name:     "命中 job 失败判 failed",
			pipeline: &biz.Pipeline{Status: biz.StatusUnknown, Jobs: []biz.PipelineJob{job("stage1", "build", biz.StatusSuccess), job("stage2", "test", biz.StatusFailed)}},
			rules:    []*mars2.PipelinePassRule{rule("stage2", "test")},
			want:     biz.StatusFailed,
		},
		{
			name:     "命中 job 在运行判 running",
			pipeline: &biz.Pipeline{Status: biz.StatusUnknown, Jobs: []biz.PipelineJob{job("stage1", "build", biz.StatusSuccess), job("stage2", "test", biz.StatusRunning)}},
			rules:    []*mars2.PipelinePassRule{rule("stage2", "test")},
			want:     biz.StatusRunning,
		},
		{
			name:     "命中 job 未开始（unknown）判 running",
			pipeline: &biz.Pipeline{Status: biz.StatusFailed, Jobs: []biz.PipelineJob{job("stage1", "build", biz.StatusUnknown)}},
			rules:    []*mars2.PipelinePassRule{rule("stage1", "build")},
			want:     biz.StatusRunning,
		},
		{
			name:     "manual 优先于未开始",
			pipeline: &biz.Pipeline{Status: biz.StatusUnknown, Jobs: []biz.PipelineJob{job("stage1", "a", biz.StatusUnknown), job("stage2", "b", biz.StatusManual)}},
			rules:    []*mars2.PipelinePassRule{rule("stage1", "a"), rule("stage2", "b")},
			want:     biz.StatusManual,
		},
		{
			name:     "规则指定 job 缺失回退原始状态",
			pipeline: &biz.Pipeline{Status: biz.StatusUnknown, Jobs: []biz.PipelineJob{job("stage1", "build", biz.StatusSuccess)}},
			rules:    []*mars2.PipelinePassRule{rule("stageX", "build")},
			want:     biz.StatusUnknown,
		},
		{
			name:     "多规则全部成功",
			pipeline: &biz.Pipeline{Status: biz.StatusFailed, Jobs: []biz.PipelineJob{job("stage1", "build", biz.StatusSuccess), job("stage2", "test", biz.StatusSuccess)}},
			rules:    []*mars2.PipelinePassRule{rule("stage1", "build"), rule("stage2", "test")},
			want:     biz.StatusSuccess,
		},
		{
			name:     "失败优先于运行",
			pipeline: &biz.Pipeline{Status: biz.StatusUnknown, Jobs: []biz.PipelineJob{job("stage1", "a", biz.StatusRunning), job("stage2", "b", biz.StatusFailed)}},
			rules:    []*mars2.PipelinePassRule{rule("stage1", "a"), rule("stage2", "b")},
			want:     biz.StatusFailed,
		},
		{
			name:     "命中 job 为 manual 判 manual",
			pipeline: &biz.Pipeline{Status: biz.StatusUnknown, Jobs: []biz.PipelineJob{job("stage1", "build", biz.StatusManual)}},
			rules:    []*mars2.PipelinePassRule{rule("stage1", "build")},
			want:     biz.StatusManual,
		},
		{
			name:     "manual 优先于运行",
			pipeline: &biz.Pipeline{Status: biz.StatusUnknown, Jobs: []biz.PipelineJob{job("stage1", "a", biz.StatusRunning), job("stage2", "b", biz.StatusManual)}},
			rules:    []*mars2.PipelinePassRule{rule("stage1", "a"), rule("stage2", "b")},
			want:     biz.StatusManual,
		},
		{
			name:     "失败优先于 manual",
			pipeline: &biz.Pipeline{Status: biz.StatusUnknown, Jobs: []biz.PipelineJob{job("stage1", "a", biz.StatusManual), job("stage2", "b", biz.StatusFailed)}},
			rules:    []*mars2.PipelinePassRule{rule("stage1", "a"), rule("stage2", "b")},
			want:     biz.StatusFailed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, biz.PipelinePassStatus(tt.pipeline, tt.rules))
		})
	}
}
