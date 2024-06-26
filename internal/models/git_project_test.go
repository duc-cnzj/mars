package models

import (
	"encoding/json"
	"sort"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/api/v4/mars"
	"github.com/duc-cnzj/mars/api/v4/types"
	"github.com/duc-cnzj/mars/v4/internal/utils/date"
	"github.com/stretchr/testify/assert"

	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

func TestGitProject_GlobalMarsConfig(t *testing.T) {
	marsCfg := mars.Config{
		ConfigFile:       "cfgfile",
		ConfigFileValues: "values",
		ConfigField:      "conf",
		IsSimpleEnv:      true,
		ConfigFileType:   "yaml",
		LocalChartPath:   "./charts",
		Branches:         []string{"master", "dev"},
		ValuesYaml:       "name: duc\nage: 27",
		Elements: []*mars.Element{
			{
				Path:         "conf->env",
				Type:         mars.ElementType_ElementTypeSelect,
				Default:      "dev",
				Description:  "environment",
				SelectValues: []string{"dev", "master", "*"},
			},
		},
	}
	marshal, _ := json.Marshal(&marsCfg)
	m := GitProject{
		ID:           1,
		GlobalConfig: string(marshal),
	}
	assert.Equal(t, &marsCfg, m.GlobalMarsConfig())

	m2 := GitProject{}
	assert.Equal(t, (&mars.Config{}).String(), m2.GlobalMarsConfig().String())

	// sortedElements
	cfg := mars.Config{
		Elements: []*mars.Element{
			{
				Path:  "1",
				Order: 1,
			},
			{
				Path:  "3",
				Order: 3,
			},
			{
				Path:  "2",
				Order: 2,
			},
		},
	}
	bytes, _ := json.Marshal(&cfg)
	m3 := GitProject{
		GlobalConfig: string(bytes),
	}
	config := m3.GlobalMarsConfig()
	assert.Equal(t, []*mars.Element{
		{
			Path:  "1",
			Order: 1,
		},
		{
			Path:  "2",
			Order: 2,
		},
		{
			Path:  "3",
			Order: 3,
		},
	}, config.Elements)
}

func TestGitProject_PrettyYaml(t *testing.T) {
	marsCfg := mars.Config{
		ConfigFile:       "cfgfile",
		ConfigFileValues: "values",
		ConfigField:      "conf",
		IsSimpleEnv:      true,
		ConfigFileType:   "yaml",
		LocalChartPath:   "./charts",
		Branches:         []string{"master", "dev"},
		ValuesYaml:       "name: duc\nage: 27",
		Elements: []*mars.Element{
			{
				Path:         "conf->env",
				Type:         mars.ElementType_ElementTypeSelect,
				Default:      "dev",
				Description:  "environment",
				SelectValues: []string{"dev", "master", "*"},
			},
		},
		DisplayName: "app",
	}
	marshal, _ := json.Marshal(&marsCfg)
	m := GitProject{
		ID:           1,
		GlobalConfig: string(marshal),
	}

	assert.Equal(t, `config_file: cfgfile
config_file_values: values
config_field: conf
is_simple_env: true
config_file_type: yaml
local_chart_path: ./charts
branches:
    - master
    - dev
values_yaml:
    age: 27
    name: duc
elements:
    - path: conf->env
      type: 3
      default: dev
      description: environment
      selectvalues:
        - dev
        - master
        - '*'
      order: 0
display_name: app
`, m.PrettyYaml())
}

// 确保 mars config 和 GitProject global_config 保存的项目字段数量是一致的，避免在增加或者删除字段时导致两边不一致
func TestGitProject_PrettyYaml_SameAsMarsConfig(t *testing.T) {
	marsCfg := mars.Config{
		ConfigFile:       "cfgfile",
		ConfigFileValues: "values",
		ConfigField:      "conf",
		IsSimpleEnv:      true,
		ConfigFileType:   "yaml",
		LocalChartPath:   "./charts",
		Branches:         []string{"master", "dev"},
		ValuesYaml:       "name: duc\nage: 27",
		Elements: []*mars.Element{
			{
				Path:         "conf->env",
				Type:         mars.ElementType_ElementTypeSelect,
				Default:      "dev",
				Description:  "environment",
				SelectValues: []string{"dev", "master", "*"},
			},
		},
		DisplayName: "app",
	}
	marshal, _ := json.Marshal(&marsCfg)
	m := GitProject{
		ID:           1,
		GlobalConfig: string(marshal),
	}

	var yamlMap = make(map[string]any)
	var jsonMap = make(map[string]any)
	assert.Nil(t, yaml.Unmarshal([]byte(m.PrettyYaml()), &yamlMap))
	assert.Nil(t, json.Unmarshal(marshal, &jsonMap))
	assert.Equal(t, len(yamlMap), len(jsonMap))
}

func TestGitProject_ProtoTransform(t *testing.T) {
	m := GitProject{
		ID:            1,
		DefaultBranch: "dev",
		Name:          "mars",
		GitProjectId:  100,
		Enabled:       true,
		GlobalEnabled: true,
		GlobalConfig:  "xxx",
		CreatedAt:     time.Now().Add(15 * time.Minute),
		UpdatedAt:     time.Now().Add(30 * time.Minute),
		DeletedAt: gorm.DeletedAt{
			Time:  time.Now().Add(-10 * time.Second),
			Valid: true,
		},
	}
	assert.Equal(t, &types.GitProjectModel{
		Id:            int64(m.ID),
		DefaultBranch: m.DefaultBranch,
		Name:          m.Name,
		GitProjectId:  int64(m.GitProjectId),
		Enabled:       m.Enabled,
		GlobalEnabled: m.GlobalEnabled,
		GlobalConfig:  m.GlobalConfig,
		CreatedAt:     date.ToRFC3339DatetimeString(&m.CreatedAt),
		UpdatedAt:     date.ToRFC3339DatetimeString(&m.UpdatedAt),
		DeletedAt:     date.ToRFC3339DatetimeString(&m.DeletedAt.Time),
	}, m.ProtoTransform())
}

func Test_sortedElements(t *testing.T) {
	var eles = sortedElements{
		{
			Path:  "1",
			Order: 1,
		},
		{
			Path:  "4",
			Order: 4,
		},
		{
			Path:  "3",
			Order: 3,
		},
		{
			Path:  "2",
			Order: 2,
		},
	}
	sort.Sort(eles)
	assert.Len(t, eles, 4)
	assert.Equal(t, sortedElements{
		{
			Path:  "1",
			Order: 1,
		},
		{
			Path:  "2",
			Order: 2,
		},
		{
			Path:  "3",
			Order: 3,
		},
		{
			Path:  "4",
			Order: 4,
		},
	}, eles)
}
