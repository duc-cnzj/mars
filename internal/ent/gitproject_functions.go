package ent

import (
	"sort"

	"github.com/duc-cnzj/mars/api/v4/mars"
	"gopkg.in/yaml.v3"
)

func (g *GitProject) PrettyYaml() string {
	cfg := g.GlobalConfig
	var v map[string]any
	yaml.Unmarshal([]byte(cfg.ValuesYaml), &v)
	var data = struct {
		ConfigFile       string          `yaml:"config_file"`
		ConfigFileValues string          `yaml:"config_file_values"`
		ConfigField      string          `yaml:"config_field"`
		IsSimpleEnv      bool            `yaml:"is_simple_env"`
		ConfigFileType   string          `yaml:"config_file_type"`
		LocalChartPath   string          `yaml:"local_chart_path"`
		Branches         []string        `yaml:"branches"`
		ValuesYaml       map[string]any  `yaml:"values_yaml"`
		Elements         []*mars.Element `yaml:"elements"`
		DisplayName      string          `yaml:"display_name"`
	}{
		ConfigFile:       cfg.ConfigFile,
		ConfigFileValues: cfg.ConfigFileValues,
		ConfigField:      cfg.ConfigField,
		IsSimpleEnv:      cfg.IsSimpleEnv,
		ConfigFileType:   cfg.ConfigFileType,
		LocalChartPath:   cfg.LocalChartPath,
		Branches:         cfg.Branches,
		ValuesYaml:       v,
		Elements:         cfg.Elements,
		DisplayName:      cfg.DisplayName,
	}

	out, _ := yaml.Marshal(data)
	return string(out)
}

type sortedElements []*mars.Element

func (s sortedElements) Len() int {
	return len(s)
}

func (s sortedElements) Less(i, j int) bool {
	return s[i].Order < s[j].Order
}

func (s sortedElements) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (g *GitProject) GlobalMarsConfig() *mars.Config {
	if g.GlobalConfig == nil {
		return &mars.Config{}
	}

	sort.Sort(sortedElements(g.GlobalConfig.Elements))
	return g.GlobalConfig
}
