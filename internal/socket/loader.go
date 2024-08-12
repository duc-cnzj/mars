package socket

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/duc-cnzj/mars/api/v4/mars"
	mars2 "github.com/duc-cnzj/mars/v4/internal/util/mars"
	yaml2 "github.com/duc-cnzj/mars/v4/internal/util/yaml"
	"go.uber.org/config"
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
)

func defaultLoaders() []Loader {
	return []Loader{
		&ChartFileLoader{
			chartLoader: &defaultChartLoader{},
			fileOpener:  &defaultFileOpener{},
		},
		&SystemVariableLoader{},
		&UserConfigLoader{},
		&ElementsLoader{},
		&MergeValuesLoader{},
	}
}

type Loader interface {
	Load(*jobRunner) error
}

type helmChartLoader interface {
	LoadDir(dir string) (*chart.Chart, error)
	LoadArchive(in io.Reader) (*chart.Chart, error)
}

type defaultFileOpener struct {
	f *os.File
}

func (d *defaultFileOpener) Open(name string) (io.ReadCloser, error) {
	open, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	d.f = open
	return open, err
}

func (d *defaultFileOpener) Close() error {
	return d.f.Close()
}

type defaultChartLoader struct{}

func (d *defaultChartLoader) LoadArchive(in io.Reader) (*chart.Chart, error) {
	return loader.LoadArchive(in)
}

func (d *defaultChartLoader) LoadDir(dir string) (*chart.Chart, error) {
	return loader.LoadDir(dir)
}

type fileOpener interface {
	Open(name string) (io.ReadCloser, error)
	Close() error
}

// ChartFileLoader 下载远程 chart 文件
type ChartFileLoader struct {
	chartLoader helmChartLoader
	fileOpener  fileOpener
}

func (c *ChartFileLoader) Load(j *jobRunner) error {
	const loaderName = "[ChartFileLoader]: "
	j.Messager().SendMsg(loaderName + "加载 helm chart 文件")
	j.Messager().To(20)

	// 下载 helm charts
	split := strings.Split(j.config.LocalChartPath, "|")
	var (
		files        []string
		tmpChartsDir string
		deleteDirFn  func()
		dir          string
	)
	// 如果是这个格式意味着是远程项目, 'uid|branch|path'
	j.Messager().SendMsg(fmt.Sprintf(loaderName+"下载 helm charts path: %s", j.config.LocalChartPath))

	var (
		pid    string = fmt.Sprintf("%d", j.repo.GitProjectID)
		branch string = j.input.GitBranch
		path   string = j.config.LocalChartPath
	)
	if true {
		pid = split[0]
		branch = split[1]
		path = split[2]
		j.logger.Warning("split", pid, branch, path, j.pluginMgr.Git())
		files, _ = j.pluginMgr.Git().GetDirectoryFilesWithBranch(pid, branch, path, true)
		if len(files) < 1 {
			return errors.New("charts 文件不存在")
		}
		var err error
		tmpChartsDir, deleteDirFn, err = j.DownloadFiles(pid, branch, files)
		if err != nil {
			return err
		}

		dir = path
		j.Messager().SendMsg(fmt.Sprintf(loaderName+"识别为远程仓库 uid %v branch %s path %s", pid, branch, path))
	} else {
		var err error
		dir = j.config.LocalChartPath
		files, _ = j.pluginMgr.Git().GetDirectoryFilesWithSha(fmt.Sprintf("%d", j.repo.GitProjectID), j.input.GitCommit, j.config.LocalChartPath, true)
		tmpChartsDir, deleteDirFn, err = j.DownloadFiles(j.repo.GitProjectID, j.input.GitCommit, files)
		if err != nil {
			return err
		}
	}
	j.OnFinally(1, func(err error, sendResultToUser func()) {
		sendResultToUser()
		deleteDirFn()
	})

	loadDir, err := c.chartLoader.LoadDir(filepath.Join(tmpChartsDir, dir))
	if err != nil {
		return err
	}
	if loadDir.Metadata.Dependencies != nil && action.CheckDependencies(loadDir, loadDir.Metadata.Dependencies) != nil {
		for _, dependency := range loadDir.Metadata.Dependencies {
			if strings.HasPrefix(dependency.Repository, "file://") {
				depFiles, _ := j.pluginMgr.Git().GetDirectoryFilesWithBranch(pid, branch, filepath.Join(path, strings.TrimPrefix(dependency.Repository, "file://")), true)
				_, depDeleteFn, err := j.DownloadFilesToDir(pid, branch, depFiles, tmpChartsDir)
				if err != nil {
					return err
				}
				j.OnFinally(1, func(err error, sendResultToUser func()) {
					sendResultToUser()
					depDeleteFn()
				})
				j.Messager().SendMsg(fmt.Sprintf(loaderName+"下载本地依赖 %s", dependency.Name))
			}
		}
	}

	chartDir := filepath.Join(tmpChartsDir, dir)

	j.Messager().To(30)
	j.Messager().SendMsg(loaderName + "打包 helm charts")
	chart, err := j.helmer.PackageChart(chartDir, chartDir)
	if err != nil {
		return err
	}
	archive, err := c.fileOpener.Open(chart)
	if err != nil {
		return err
	}
	defer archive.Close()

	j.chart, err = c.chartLoader.LoadArchive(archive)

	return err
}

// UserConfigLoader 把用户配置字段 merge 到 values.yaml
// 返回的应该是一个片段
// ```yaml
// conf:
//
//	key: value
//
// ```
type UserConfigLoader struct{}

func (d *UserConfigLoader) Load(j *jobRunner) error {
	const loaderName = "[UserConfigLoader]: "

	j.Messager().To(50)
	j.Messager().SendMsg(fmt.Sprintf(loaderName+"%v", "检查到用户传入的配置"))

	if j.input.Config == "" {
		j.Messager().SendMsg(fmt.Sprintf(loaderName+"%v", "未发现用户自定义配置"))
		return nil
	}

	userConfigYaml, err := mars2.ParseInputConfig(j.config, j.input.Config)
	if err != nil && !errors.Is(err, io.EOF) {
		j.logger.Error(err)
		return err
	}
	j.userConfigYaml = userConfigYaml

	return nil
}

// ElementsLoader 后台自定义配置加载, 返回一个片段
/*
```yaml
resources:
  limits:
	cpu: 100m

ingress:
	enables: true
```
*/
type ElementsLoader struct{}

func (d *ElementsLoader) Load(j *jobRunner) error {
	const loaderName = "[ElementsLoader]: "

	j.Messager().To(60)
	j.Messager().SendMsg(fmt.Sprintf(loaderName+"%v", "检查项目额外的配置"))

	if len(j.input.ExtraValues) <= 0 {
		j.Messager().SendMsg(fmt.Sprintf(loaderName+"%v", "未发现项目额外的配置"))
	}

	var validValuesMap = make(map[string]any)
	var useDefaultMap = make(map[string]bool)

	var configElementsMap = make(map[string]*mars.Element)
	for _, element := range j.config.Elements {
		configElementsMap[element.Path] = element
		defaultValue, e := d.typedValue(element, element.Default)
		if e != nil {
			return e
		}
		validValuesMap[element.Path] = defaultValue
		useDefaultMap[element.Path] = true
	}

	// validate
	for _, value := range j.input.ExtraValues {
		var fieldValid bool
		if element, ok := configElementsMap[value.Path]; ok {
			fieldValid = true
			useDefaultMap[value.Path] = false
			typeValue, err := d.typedValue(element, value.Value)
			if err != nil {
				return err
			}
			validValuesMap[value.Path] = typeValue
		}
		if !fieldValid {
			j.Messager().SendMsg(fmt.Sprintf("不允许自定义字段 %s", value.Path))
		}
	}

	j.elementValues = d.deepSetItems(validValuesMap)
	var ds []string
	for k, ok := range useDefaultMap {
		if ok {
			ds = append(ds, k)
		}
	}
	if len(ds) > 0 {
		j.Messager().SendMsg(fmt.Sprintf(loaderName+"已经为 '%s' 设置系统默认值", strings.Join(ds, ",")))
	}

	return nil
}

func (d *ElementsLoader) typedValue(element *mars.Element, input string) (any, error) {
	switch element.Type {
	case mars.ElementType_ElementTypeSwitch:
		if input == "" {
			input = "false"
		}
		v, err := strconv.ParseBool(input)
		if err != nil {
			return nil, fmt.Errorf("%s 字段类型不正确，应该为 bool，你传入的是 %s", element.Path, input)
		}
		return v, nil
	case mars.ElementType_ElementTypeInputNumber:
		if input == "" {
			input = "0"
		}
		v, err := strconv.ParseInt(input, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("%s 字段类型不正确，应该为整数，你传入的是 %s", element.Path, input)
		}
		return v, nil
	case mars.ElementType_ElementTypeRadio,
		mars.ElementType_ElementTypeSelect,
		mars.ElementType_ElementTypeNumberSelect,
		mars.ElementType_ElementTypeNumberRadio:
		var in bool
		for _, selectValue := range element.SelectValues {
			if input == selectValue {
				in = true
				break
			}
		}
		if !in {
			return nil, fmt.Errorf("%s 必须在 '%v' 里面, 你传的是 %s", element.Path, strings.Join(element.SelectValues, ","), input)
		}
		if element.Type == mars.ElementType_ElementTypeNumberSelect ||
			element.Type == mars.ElementType_ElementTypeNumberRadio {
			if atoi, err := strconv.Atoi(input); err == nil {
				return atoi, nil
			}
			return nil, fmt.Errorf("[ElementsLoader]: '%v' 非 number 类型, 无法转换", input)
		}

		return input, nil
	default:
		return input, nil
	}
}

func (d *ElementsLoader) deepSetItems(items map[string]any) []string {
	var evs []string
	for k, v := range items {
		ysk, err := yaml2.YamlDeepSetKey(k, v)
		if err != nil {
			continue
		}
		evs = append(evs, string(ysk))
	}
	return evs
}

const (
	leftDelim  = "<"
	rightDelim = ">"

	VarImagePullSecrets       = "ImagePullSecrets"
	VarImagePullSecretsNoName = "ImagePullSecretsNoName"
	VarBranch                 = "Branch"
	VarCommit                 = "Commit"
	VarPipeline               = "Pipeline"
	VarClusterIssuer          = "ClusterIssuer"
	VarHost                   = "Host"
	VarTlsSecret              = "TlsSecret"
)

var tagRegex = regexp.MustCompile(leftDelim + `\s*(\.Branch|\.Commit|\.Pipeline)\s*` + rightDelim)

// SystemVariableLoader 系统内置变量替换， values.yaml 中的 <.var> 变量
// 例如：<.Branch> <.Commit> <.Pipeline> <.Host1> <.TlsSecret1>
type SystemVariableLoader struct {
	values vars
}

func (v *SystemVariableLoader) Add(key, value string) {
	v.values.Add(key, value)
}

func (v *SystemVariableLoader) Load(j *jobRunner) error {
	const loaderName = "[SystemVariableLoader]: "
	j.Messager().To(40)
	j.Messager().SendMsg(fmt.Sprintf(loaderName+"%v", "注入内置环境变量"))

	if j.config.ValuesYaml == "" {
		j.Messager().SendMsg(fmt.Sprintf(loaderName+"%v", "未发现可用的 values.yaml"))
		return nil
	}

	if v.values == nil {
		v.values = vars{}
	}
	type pullSecrets struct {
		ImagePullSecrets []string
	}
	t, _ := template.New("").Parse(`
{{ define "pullSecrets"}}[{{- range .ImagePullSecrets }}{name: {{ . }}}, {{ end }}]{{ end }}
{{ define "pullSecretsNoName"}}[{{- range .ImagePullSecrets }}{{ . }}, {{ end }}]{{ end }}
`)
	// ImagePullSecrets
	// [{name: secret1}, {name: secret2}, ]
	//parse, _ := template.New("").Parse("[{{- range .ImagePullSecrets }}{name: {{ . }}}, {{- end }}]")
	renderResult := &bytes.Buffer{}
	t.ExecuteTemplate(renderResult, "pullSecrets", pullSecrets{
		ImagePullSecrets: j.imagePullSecrets,
	})

	// ImagePullSecretsNoName
	// [secret1, secret2, ]
	renderResultNoName := &bytes.Buffer{}
	t.ExecuteTemplate(renderResultNoName, "pullSecretsNoName", pullSecrets{
		ImagePullSecrets: j.imagePullSecrets,
	})

	v.Add(VarImagePullSecrets, renderResult.String())
	v.Add(VarImagePullSecretsNoName, renderResultNoName.String())

	//Host1...Host10
	sub := j.projRepo.GetPreOccupiedLenByValuesYaml(j.config.ValuesYaml)
	j.logger.Debug("getPreOccupiedLenByValuesYaml: ", sub)
	for i := 1; i <= 10; i++ {
		v.Add(fmt.Sprintf("%s%d", VarHost, i), j.pluginMgr.Domain().GetDomainByIndex(j.project.Name, j.Namespace().Name, i, sub))
		v.Add(fmt.Sprintf("%s%d", VarTlsSecret, i), j.pluginMgr.Domain().GetCertSecretName(j.project.Name, i))
	}

	//{{.Branch}}{{.Commit}}{{.Pipeline}}
	var (
		pipelineID     int64
		pipelineBranch string = j.project.GitBranch
		pipelineCommit string = j.Commit().GetShortID()
	)

	if j.repo.NeedGitRepo {
		// 如果存在需要传变量的，则必须有流水线信息
		if pipeline, e := j.pluginMgr.Git().GetCommitPipeline(fmt.Sprintf("%d", j.project.GitProjectID), j.project.GitBranch, j.project.GitCommit); e == nil {
			pipelineID = pipeline.GetID()
			pipelineBranch = pipeline.GetRef()

			j.Messager().SendMsg(fmt.Sprintf(loaderName+"镜像分支 %s 镜像commit %s 镜像 pipeline_id %d", pipelineBranch, pipelineCommit, pipelineID))
		} else {
			if tagRegex.MatchString(j.config.ValuesYaml) {
				return errors.New("无法获取 Pipeline 信息")
			}
		}
	}

	v.Add(VarBranch, pipelineBranch)
	v.Add(VarCommit, pipelineCommit)
	v.Add(VarPipeline, fmt.Sprintf("%d", pipelineID))

	// ingress
	v.Add(VarClusterIssuer, j.pluginMgr.Domain().GetClusterIssuer())

	tpl, err := template.New("values_yaml").Delims(leftDelim, rightDelim).Parse(j.config.ValuesYaml)
	if err != nil {
		return err
	}
	bf := bytes.Buffer{}
	tpl.Execute(&bf, v.values)
	j.systemValuesYaml = bf.String()
	j.vars = v.values

	return nil
}

type MergeValuesLoader struct{}

// Load
// imagePullSecrets 会自动注入到 imagePullSecrets 中
func (m *MergeValuesLoader) Load(j *jobRunner) error {
	const loaderName = "[MergeValuesLoader]: "
	j.Messager().To(70)
	j.Messager().SendMsg(fmt.Sprintf(loaderName+"%v", "合并配置文件到 values.yaml"))

	// 自动注入 imagePullSecrets
	var imagePullSecrets = make([]map[string]any, len(j.imagePullSecrets))
	for i, s := range j.imagePullSecrets {
		imagePullSecrets[i] = map[string]any{"name": s}
	}
	var yamlImagePullSecrets []byte
	if len(imagePullSecrets) > 0 {
		yamlImagePullSecrets, _ = yaml.Marshal(map[string]any{
			"imagePullSecrets": imagePullSecrets,
		})
	}

	var opts []config.YAMLOption
	if j.systemValuesYaml != "" {
		opts = append(opts, config.Source(strings.NewReader(j.systemValuesYaml)))
	}
	if j.userConfigYaml != "" {
		opts = append(opts, config.Source(strings.NewReader(j.userConfigYaml)))
	}
	if len(yamlImagePullSecrets) != 0 {
		opts = append(opts, config.Source(bytes.NewReader(yamlImagePullSecrets)))
	}

	for _, value := range j.elementValues {
		opts = append(opts, config.Source(strings.NewReader(value)))
	}

	if len(opts) < 1 {
		return nil
	}

	// 5. 用用户传入的yaml配置去合并 `default_values`
	provider, err := config.NewYAML(opts...)
	if err != nil {
		j.logger.Error(loaderName, err, j.systemValuesYaml, j.userConfigYaml)

		return err
	}
	var mergedDefaultAndConfigYamlValues map[string]any
	if err := provider.Get("").Populate(&mergedDefaultAndConfigYamlValues); err != nil {
		j.logger.Error(loaderName, mergedDefaultAndConfigYamlValues, err)
		return err
	}

	var fileData []byte
	if fileData, err = yaml.Marshal(&mergedDefaultAndConfigYamlValues); err != nil {
		return err
	}
	//mlog.Debug("fileData", fileData)
	mergedFile, closer, err := j.WriteConfigYamlToTmpFile(fileData)
	if err != nil {
		return err
	}
	j.OnFinally(1, func(err error, sendResultToUser func()) {
		sendResultToUser()
		closer.Close()
	})
	j.valuesOptions.ValueFiles = append(j.valuesOptions.ValueFiles, mergedFile)

	return nil
}
