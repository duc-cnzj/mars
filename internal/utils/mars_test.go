package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/duc-cnzj/mars-client/v4/mars"
	"github.com/duc-cnzj/mars/v4/internal/app/instance"
	"github.com/duc-cnzj/mars/v4/internal/mock"
	"github.com/duc-cnzj/mars/v4/internal/models"
	"github.com/duc-cnzj/mars/v4/internal/testutil"

	"github.com/golang/mock/gomock"
	"github.com/lithammer/dedent"
	"github.com/stretchr/testify/assert"
)

func TestBranchPass(t *testing.T) {
	cfg := &mars.Config{
		Branches: []string{"master"},
	}
	assert.True(t, BranchPass(cfg, "master"))
	assert.False(t, BranchPass(cfg, "dev"))
	cfg = &mars.Config{
		Branches: []string{"*"},
	}
	assert.True(t, BranchPass(cfg, "master"))
	cfg = &mars.Config{
		Branches: []string{"dev-*"},
	}
	assert.True(t, BranchPass(cfg, "dev-aaa"))
	assert.False(t, BranchPass(cfg, "nodev-aaa"))
	cfg = &mars.Config{}
	assert.True(t, BranchPass(cfg, "dev-aaa"))
	assert.True(t, BranchPass(cfg, "ccc"))
	cfg = &mars.Config{Branches: []string{"*-dev"}}
	assert.True(t, BranchPass(cfg, "a-dev"))
	assert.True(t, BranchPass(cfg, "b-dev"))

	// regex syntax error
	cfg = &mars.Config{Branches: []string{"[a-zA-Z]{10000,}*"}}
	assert.False(t, BranchPass(cfg, strings.Repeat("a", 100000)))
}

func TestGetProjectMarsConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	app := mock.NewMockApplicationInterface(ctrl)
	defer ctrl.Finish()
	instance.SetInstance(app)
	db, closeFn := testutil.SetGormDB(ctrl, app)
	defer closeFn()
	db.AutoMigrate(&models.GitProject{})
	mc := mars.Config{
		ConfigFile:       "cf",
		ConfigFileValues: "vv",
		ConfigField:      "f",
		Elements: []*mars.Element{
			{
				Path:         "image->tag",
				Type:         mars.ElementType_ElementTypeSelect,
				Default:      "v1",
				Description:  "tag",
				SelectValues: []string{"v1", "v2"},
				Order:        0,
			},
		},
	}
	marshal, _ := json.Marshal(&mc)
	db.Create(&models.GitProject{
		GitProjectId:  99,
		GlobalEnabled: true,
		GlobalConfig:  string(marshal),
	})
	cfg, _ := GetProjectMarsConfig(99, "dev")
	assert.Equal(t, &mc, cfg)
	db.Create(&models.GitProject{
		GitProjectId:  199,
		GlobalEnabled: false,
	})
	gs := testutil.MockGitServer(ctrl, app)
	pid := 199
	gs.EXPECT().GetFileContentWithBranch(fmt.Sprintf("%v", pid), "dev", ".mars.yaml").Return("", errors.New("xxx"))
	_, err := GetProjectMarsConfig(pid, "dev")
	assert.Equal(t, "xxx", err.Error())
	gs.EXPECT().GetFileContentWithBranch(fmt.Sprintf("%v", pid), "dev", ".mars.yaml").Return(string(marshal), nil)
	cfg, _ = GetProjectMarsConfig(pid, "dev")
	assert.Equal(t, &mc, cfg)
}

func TestIsRemoteChart(t *testing.T) {
	assert.False(t, IsRemoteChart(&mars.Config{LocalChartPath: "abc|branch|path"}))
	assert.True(t, IsRemoteChart(&mars.Config{LocalChartPath: "1|branch|path"}))
	assert.False(t, IsRemoteChart(&mars.Config{LocalChartPath: "pid"}))
}

func TestIsRemoteConfigFile(t *testing.T) {
	assert.False(t, IsRemoteConfigFile(&mars.Config{ConfigFile: "abc|branch|path"}))
	assert.True(t, IsRemoteConfigFile(&mars.Config{ConfigFile: "1|branch|path"}))
	assert.False(t, IsRemoteConfigFile(&mars.Config{ConfigFile: "pid"}))
}

func TestParseInputConfig(t *testing.T) {
	var tests = []struct {
		IsSimpleEnv bool
		ConfigField string
		input       string
		wants       string
		ValuesYaml  string
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
			wants: dedent.Dedent(`
				conf:
				  config: |-
				    name: duc
				    age: 18
				`),
		},
		{
			IsSimpleEnv: false,
			ConfigField: "command",
			input: dedent.Dedent(`
					command:
					- sh
					- -c
					- "sleep 3600;exit"
				`),
			wants: dedent.Dedent(`
					command:
					- sh
					- -c
					- sleep 3600;exit
				`),
		},
		{
			IsSimpleEnv: false,
			ConfigField: "command",
			input:       `command: ["sh", "-c", "sleep 3600;exit"]`,
			wants: dedent.Dedent(`
					command:
					- sh
					- -c
					- sleep 3600;exit
				`),
		},
		{
			IsSimpleEnv: false,
			ConfigField: "conf->command",
			input:       `command: ["sh", "-c", "sleep 3600;exit"]`,
			wants: dedent.Dedent(`
					conf:
					  command:
					  - sh
					  - -c
					  - sleep 3600;exit
				`),
		},
		{
			IsSimpleEnv: false,
			ConfigField: "",
			input:       `command: ["sh", "-c", "sleep 3600;exit"]`,
			wants: dedent.Dedent(`
					"":
					  command:
					  - sh
					  - -c
					  - sleep 3600;exit
				`),
		},
		{
			IsSimpleEnv: false,
			ConfigField: "command",
			input: dedent.Dedent(`
					command:
					  a: b
				`),
			wants: dedent.Dedent(`
					command:
					  command:
					    a: b
				`),
		},
		{
			IsSimpleEnv: false,
			ConfigField: "command",
			ValuesYaml: dedent.Dedent(`
					command:
					  command: []
				`),
			input: dedent.Dedent(`
				   command:
				   - a
				   - b
				`),
			wants: dedent.Dedent(`
					command:
					  command:
					  - a
					  - b
				`),
		},
		{
			IsSimpleEnv: false,
			ConfigField: "command",
			ValuesYaml: dedent.Dedent(`
					command: []
				`),
			input: dedent.Dedent(`
				   command:
				   - a
				   - b
				`),
			wants: dedent.Dedent(`
					command:
					- a
					- b
				`),
		},
	}

	for _, test := range tests {
		tt := test
		t.Run(tt.ConfigField, func(t *testing.T) {
			t.Parallel()
			res, err := ParseInputConfig(&mars.Config{
				IsSimpleEnv: tt.IsSimpleEnv,
				ValuesYaml:  tt.ValuesYaml,
				ConfigField: tt.ConfigField,
			}, strings.Trim(tt.input, "\n"))
			assert.Nil(t, err)
			assert.Equal(t, strings.Trim(tt.wants, "\n"), strings.Trim(res, "\n"))
		})
	}
}

func Test_intPid(t *testing.T) {
	assert.True(t, intPid("1"))
	assert.True(t, intPid("-1"))
	assert.True(t, intPid("10"))
	assert.False(t, intPid("abc"))
	assert.False(t, intPid("1_a"))
}

func TestGetProjectName(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	gitS := testutil.MockGitServer(m, app)
	p := mock.NewMockProjectInterface(m)
	gitS.EXPECT().GetProject("1").Return(p, nil).AnyTimes()
	p.EXPECT().GetName().Return("app").AnyTimes()
	assert.Equal(t, "app", GetProjectName("1", &mars.Config{}))
	assert.Equal(t, "app-2", GetProjectName("1", &mars.Config{DisplayName: "app-2"}))
}
