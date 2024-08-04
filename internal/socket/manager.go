package socket

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/duc-cnzj/mars/v4/internal/application"
	"github.com/duc-cnzj/mars/v4/internal/auth"
	"github.com/duc-cnzj/mars/v4/internal/ent"
	"github.com/duc-cnzj/mars/v4/internal/locker"
	"github.com/duc-cnzj/mars/v4/internal/repo"
	"github.com/duc-cnzj/mars/v4/internal/transformer"
	"github.com/duc-cnzj/mars/v4/internal/uploader"
	"github.com/duc-cnzj/mars/v4/internal/util"
	"github.com/duc-cnzj/mars/v4/internal/util/closeable"
	mars2 "github.com/duc-cnzj/mars/v4/internal/util/mars"
	"github.com/duc-cnzj/mars/v4/internal/util/rand"
	yaml2 "github.com/duc-cnzj/mars/v4/internal/util/yaml"

	"github.com/duc-cnzj/mars/api/v4/mars"
	"github.com/duc-cnzj/mars/api/v4/types"
	websocket_pb "github.com/duc-cnzj/mars/api/v4/websocket"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	mysort "github.com/duc-cnzj/mars/v4/internal/util/xsort"
	"go.uber.org/config"
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli/values"
)

const (
	ResultError             = websocket_pb.ResultType_Error
	ResultSuccess           = websocket_pb.ResultType_Success
	ResultDeployed          = websocket_pb.ResultType_Deployed
	ResultDeployFailed      = websocket_pb.ResultType_DeployedFailed
	ResultDeployCanceled    = websocket_pb.ResultType_DeployedCanceled
	ResultLogWithContainers = websocket_pb.ResultType_LogWithContainers

	WsSetUid             = websocket_pb.Type_SetUid
	WsReloadProjects     = websocket_pb.Type_ReloadProjects
	WsCancel             = websocket_pb.Type_CancelProject
	WsCreateProject      = websocket_pb.Type_CreateProject
	WsUpdateProject      = websocket_pb.Type_UpdateProject
	WsProcessPercent     = websocket_pb.Type_ProcessPercent
	WsClusterInfoSync    = websocket_pb.Type_ClusterInfoSync
	WsInternalError      = websocket_pb.Type_InternalError
	WsHandleExecShell    = websocket_pb.Type_HandleExecShell
	WsHandleExecShellMsg = websocket_pb.Type_HandleExecShellMsg
	WsHandleCloseShell   = websocket_pb.Type_HandleCloseShell
	WsAuthorize          = websocket_pb.Type_HandleAuthorize
	ProjectPodEvent      = websocket_pb.Type_ProjectPodEvent

	// Maximum message size allowed from peer.
	maxMessageSize = 1024 * 1024 * 20 // 20MB
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second
	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second
	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 8) / 10
)

var (
	ErrorVersionNotMatched = errors.New("当前版本和最新版本存在差异，请刷新重试")
)

func reloadProjectsMessage[T int64 | int](nsID T) *websocket_pb.WsReloadProjectsResponse {
	return &websocket_pb.WsReloadProjectsResponse{
		Metadata:    &websocket_pb.Metadata{Type: WsReloadProjects},
		NamespaceId: int32(nsID),
	}
}

type WsResponse = websocket_pb.WsMetadataResponse

type safeWriteMessageCh struct {
	logger    mlog.Logger
	closeable closeable.Closeable

	chMu sync.Mutex
	ch   chan contracts.MessageItem
}

func (s *safeWriteMessageCh) Close() {
	s.logger.Debug("safeWriteMessageCh closed")
	if s.closeable.Close() {
		close(s.ch)
	}
}

func (s *safeWriteMessageCh) Chan() <-chan contracts.MessageItem {
	return s.ch
}

func (s *safeWriteMessageCh) Send(m contracts.MessageItem) {
	if s.closeable.IsClosed() {
		s.logger.Debugf("[Websocket]: Drop %s type %s", m.Msg, m.Type)
		return
	}
	// 存在并发写
	s.chMu.Lock()
	defer s.chMu.Unlock()
	s.ch <- m
}

type vars map[string]string

func (v vars) ToKeyValue() (res []*types.KeyValue) {
	for k, va := range v {
		res = append(res, &types.KeyValue{
			Key:   k,
			Value: va,
		})
	}
	return
}
func (v vars) MustGetString(key string) string {
	if value, ok := v[key]; ok {
		return value
	}

	return ""
}

func (v vars) Add(key, value string) {
	v[key] = value
}

type jobRunner struct {
	logger       mlog.Logger
	nsRepo       repo.NamespaceRepo
	projRepo     repo.ProjectRepo
	gitServer    application.GitServer
	domainServer application.DomainManager
	helmer       repo.HelmerRepo
	locker       locker.Locker
	k8sRepo      repo.K8sRepo
	eventRepo    repo.EventRepo
	toolRepo     repo.ToolRepo
	uploader     uploader.Uploader

	err error

	deployResult DeployResult

	loaders []Loader

	dryRun    bool
	manifests []string

	input *JobInput

	finallyCallback mysort.PrioritySort[func(err error, next func())]
	errorCallback   mysort.PrioritySort[func(err error, next func())]
	successCallback mysort.PrioritySort[func(err error, next func())]

	imagePullSecrets  []string
	vars              vars
	dynamicConfigYaml string
	extraValues       []string
	valuesYaml        string
	chart             *chart.Chart
	valuesOptions     *values.Options
	installer         contracts.ReleaseInstaller
	commit            application.Commit

	messageCh contracts.SafeWriteMessageChInterface
	stopCtx   context.Context
	stopFn    func(error)

	config *mars.Config

	isNew       bool
	ns          *ent.Namespace
	project     *ent.Project
	prevProject *ent.Project

	percenter contracts.Percentable
	messager  contracts.DeployMsger

	user           *auth.UserInfo
	timeoutSeconds int64
}

type Option func(*jobRunner)

func WithDryRun(dryRun bool) Option {
	return func(j *jobRunner) {
		j.dryRun = dryRun
	}
}

func (j *jobRunner) SetError(err error) *jobRunner {
	j.err = err
	return j
}

func (j *jobRunner) HasError() bool {
	return j.err != nil
}

func (j *jobRunner) IsNew() bool {
	return j.isNew
}

func (j *jobRunner) IsDryRun() bool {
	return j.dryRun
}

func (j *jobRunner) Commit() application.Commit {
	return j.commit
}

func (j *jobRunner) User() *auth.UserInfo {
	return j.user
}

func (j *jobRunner) ProjectModel() *types.ProjectModel {
	if j.project == nil {
		return nil
	}
	return transformer.FromProject(j.project)
}

func (j *jobRunner) Project() *ent.Project {
	return j.project
}

func (j *jobRunner) Namespace() *ent.Namespace {
	return j.ns
}

func (j *jobRunner) IsStopped() bool {
	select {
	case <-j.stopCtx.Done():
		return true
	default:
	}

	return false
}

type DeployResult struct {
	sync.RWMutex
	result websocket_pb.ResultType
	msg    string
	model  *types.ProjectModel
	set    bool
}

func (d *DeployResult) IsSet() bool {
	d.RLock()
	defer d.RUnlock()
	return d.set
}

func (d *DeployResult) Msg() string {
	d.RLock()
	defer d.RUnlock()
	return d.msg
}

func (d *DeployResult) Model() *types.ProjectModel {
	d.RLock()
	defer d.RUnlock()
	return d.model
}

func (d *DeployResult) ResultType() websocket_pb.ResultType {
	d.RLock()
	defer d.RUnlock()
	return d.result
}

func (d *DeployResult) Set(t websocket_pb.ResultType, msg string, model *types.ProjectModel) {
	d.Lock()
	defer d.Unlock()
	d.result = t
	d.msg = msg
	d.model = model
	d.set = true
}

func (j *jobRunner) HandleMessage(ctx context.Context) {
	defer j.logger.Debug("HandleMessage exit")
	ch := j.messageCh.Chan()
	for {
		select {
		case <-ctx.Done():
			return
		case s, ok := <-ch:
			if !ok {
				return
			}
			switch s.Type {
			case contracts.MessageText:
				j.Messager().SendMsgWithContainerLog(s.Msg, s.Containers)
			case contracts.MessageError:
				select {
				case <-j.stopCtx.Done():
					j.SetDeployResult(ResultDeployCanceled, j.stopCtx.Err().Error(), transformer.FromProject(j.project))
				default:
					j.SetDeployResult(ResultDeployFailed, s.Msg, transformer.FromProject(j.project))
				}
				return
			case contracts.MessageSuccess:
				j.SetDeployResult(ResultDeployed, s.Msg, transformer.FromProject(j.project))
				return
			}
		}
	}
}

type userConfig struct {
	Config           string              `yaml:"config"`
	Branch           string              `yaml:"branch"`
	Commit           string              `yaml:"commit"`
	Atomic           bool                `yaml:"atomic"`
	WebUrl           string              `yaml:"web_url"`
	Title            string              `yaml:"title"`
	ExtraValues      []*types.ExtraValue `yaml:"extra_values"`
	FinalExtraValues mergeYamlString     `yaml:"final_extra_values"`
	EnvValues        vars                `yaml:"env_values"`
}

func newUserConfig(p *ent.Project) *userConfig {
	var v = vars{}
	for _, value := range p.EnvValues {
		v.Add(value.Key, value.Value)
	}
	return &userConfig{
		Config:           p.Config,
		Branch:           p.GitBranch,
		Commit:           p.GitCommit,
		Atomic:           p.Atomic,
		ExtraValues:      p.ExtraValues,
		FinalExtraValues: p.FinalExtraValues,
		EnvValues:        v,
		WebUrl:           p.GitCommitWebURL,
		Title:            p.GitCommitTitle,
	}
}

type mergeYamlString []string

func (s mergeYamlString) MarshalYAML() (any, error) {
	var opts []config.YAMLOption
	for _, item := range s {
		opts = append(opts, config.Source(strings.NewReader(item)))
	}
	if len(opts) == 0 {
		return "", nil
	}
	provider, _ := config.NewYAML(opts...)
	var merged map[string]any
	provider.Get("").Populate(&merged)

	out, _ := yaml2.PrettyMarshal(&merged)

	return string(out), nil
}

func (u userConfig) PrettyYaml() string {
	sort.Sort(sortableExtraItem(u.ExtraValues))
	out, _ := yaml2.PrettyMarshal(&u)

	return string(out)
}

type sortableExtraItem []*types.ExtraValue

func (s sortableExtraItem) Len() int {
	return len(s)
}

func (s sortableExtraItem) Less(i, j int) bool {
	return s[i].Path < s[j].Path
}

func (s sortableExtraItem) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (j *jobRunner) SetDeployResult(t websocket_pb.ResultType, msg string, model *types.ProjectModel) {
	j.deployResult.Set(t, msg, model)
}

func (j *jobRunner) GetStoppedErrorIfHas() error {
	if j.IsStopped() {
		return j.stopCtx.Err()
	}
	return nil
}

func (j *jobRunner) ReleaseInstaller() contracts.ReleaseInstaller {
	return j.installer
}

func (j *jobRunner) Messager() contracts.DeployMsger {
	return j.messager
}

func (j *jobRunner) PubSub() application.PubSub {
	return j.input.PubSub
}

func (j *jobRunner) Percenter() contracts.Percentable {
	return j.percenter
}

func defaultLoaders() []Loader {
	return []Loader{
		&ChartFileLoader{
			chartLoader: &defaultChartLoader{},
			fileOpener:  &defaultFileOpener{},
		},
		&VariableLoader{},
		&DynamicLoader{},
		&ExtraValuesLoader{},
		&MergeValuesLoader{},
		&ReleaseInstallerLoader{},
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

type ChartFileLoader struct {
	chartLoader helmChartLoader
	fileOpener  fileOpener
}

func (c *ChartFileLoader) Load(j *jobRunner) error {
	const loaderName = "[ChartFileLoader]: "
	j.Messager().SendMsg(loaderName + "加载 helm chart 文件")
	j.Percenter().To(20)

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
		pid    string = fmt.Sprintf("%d", j.input.GitProjectId)
		branch string = j.input.GitBranch
		path   string = j.config.LocalChartPath
	)
	if mars2.IsRemoteChart(j.config) {
		pid = split[0]
		branch = split[1]
		path = split[2]
		files, _ = j.gitServer.GetDirectoryFilesWithBranch(pid, branch, path, true)
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
		files, _ = j.gitServer.GetDirectoryFilesWithSha(fmt.Sprintf("%d", j.input.GitProjectId), j.input.GitCommit, j.config.LocalChartPath, true)
		tmpChartsDir, deleteDirFn, err = j.DownloadFiles(j.input.GitProjectId, j.input.GitCommit, files)
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
				depFiles, _ := j.gitServer.GetDirectoryFilesWithBranch(pid, branch, filepath.Join(path, strings.TrimPrefix(dependency.Repository, "file://")), true)
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

	j.Percenter().To(30)
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

type DynamicLoader struct{}

func (d *DynamicLoader) Load(j *jobRunner) error {
	const loaderName = "[DynamicLoader]: "

	j.Percenter().To(50)
	j.Messager().SendMsg(fmt.Sprintf(loaderName+"%v", "检查到用户传入的配置"))

	if j.input.Config == "" {
		j.Messager().SendMsg(fmt.Sprintf(loaderName+"%v", "未发现用户自定义配置"))
		return nil
	}

	dynamicConfigYaml, err := mars2.ParseInputConfig(j.config, j.input.Config)
	if err != nil && !errors.Is(err, io.EOF) {
		j.logger.Error(err)
		return err
	}
	j.dynamicConfigYaml = dynamicConfigYaml

	return nil
}

type ExtraValuesLoader struct{}

func (d *ExtraValuesLoader) Load(j *jobRunner) error {
	const loaderName = "[ExtraValuesLoader]: "

	j.Percenter().To(60)
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

	j.extraValues = d.deepSetItems(validValuesMap)
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

func (d *ExtraValuesLoader) typedValue(element *mars.Element, input string) (any, error) {
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
			return nil, fmt.Errorf("[ExtraValuesLoader]: '%v' 非 number 类型, 无法转换", input)
		}

		return input, nil
	default:
		return input, nil
	}
}

func (d *ExtraValuesLoader) deepSetItems(items map[string]any) []string {
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

type VariableLoader struct {
	values vars
}

func (v *VariableLoader) Add(key, value string) {
	v.values.Add(key, value)
}

func (v *VariableLoader) Load(j *jobRunner) error {
	const loaderName = "[VariableLoader]: "
	j.Percenter().To(40)
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
	sub := j.toolRepo.GetPreOccupiedLenByValuesYaml(j.config.ValuesYaml)
	j.logger.Debug("getPreOccupiedLenByValuesYaml: ", sub)
	for i := 1; i <= 10; i++ {
		v.Add(fmt.Sprintf("%s%d", VarHost, i), j.domainServer.GetDomainByIndex(j.project.Name, j.Namespace().Name, i, sub))
		v.Add(fmt.Sprintf("%s%d", VarTlsSecret, i), j.domainServer.GetCertSecretName(j.project.Name, i))
	}

	//{{.Branch}}{{.Commit}}{{.Pipeline}}
	var (
		pipelineID     int64
		pipelineBranch string = j.project.GitBranch
		pipelineCommit string = j.Commit().GetShortID()
	)

	// 如果存在需要传变量的，则必须有流水线信息
	if pipeline, e := j.gitServer.GetCommitPipeline(fmt.Sprintf("%d", j.project.GitProjectID), j.project.GitBranch, j.project.GitCommit); e == nil {
		pipelineID = pipeline.GetID()
		pipelineBranch = pipeline.GetRef()

		j.Messager().SendMsg(fmt.Sprintf(loaderName+"镜像分支 %s 镜像commit %s 镜像 pipeline_id %d", pipelineBranch, pipelineCommit, pipelineID))
	} else {
		if tagRegex.MatchString(j.config.ValuesYaml) {
			return errors.New("无法获取 Pipeline 信息")
		}
	}

	v.Add(VarBranch, pipelineBranch)
	v.Add(VarCommit, pipelineCommit)
	v.Add(VarPipeline, fmt.Sprintf("%d", pipelineID))

	// ingress
	v.Add(VarClusterIssuer, j.domainServer.GetClusterIssuer())

	tpl, err := template.New("values_yaml").Delims(leftDelim, rightDelim).Parse(j.config.ValuesYaml)
	if err != nil {
		return err
	}
	bf := bytes.Buffer{}
	tpl.Execute(&bf, v.values)
	j.valuesYaml = bf.String()
	j.vars = v.values

	return nil
}

type MergeValuesLoader struct{}

// Load
// imagePullSecrets 会自动注入到 imagePullSecrets 中
func (m *MergeValuesLoader) Load(j *jobRunner) error {
	const loaderName = "[MergeValuesLoader]: "
	j.Percenter().To(70)
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
	if j.valuesYaml != "" {
		opts = append(opts, config.Source(strings.NewReader(j.valuesYaml)))
	}
	if j.dynamicConfigYaml != "" {
		opts = append(opts, config.Source(strings.NewReader(j.dynamicConfigYaml)))
	}
	if len(yamlImagePullSecrets) != 0 {
		opts = append(opts, config.Source(bytes.NewReader(yamlImagePullSecrets)))
	}

	for _, value := range j.extraValues {
		opts = append(opts, config.Source(strings.NewReader(value)))
	}

	if len(opts) < 1 {
		return nil
	}

	// 5. 用用户传入的yaml配置去合并 `default_values`
	provider, err := config.NewYAML(opts...)
	if err != nil {
		j.logger.Error(loaderName, err, j.valuesYaml, j.dynamicConfigYaml)

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

func (j *jobRunner) WriteConfigYamlToTmpFile(data []byte) (string, io.Closer, error) {
	file := fmt.Sprintf("mars-%s-%s.yaml", time.Now().Format("2006-01-02"), rand.String(20))
	info, err := j.uploader.LocalUploader().Put(file, bytes.NewReader(data))
	if err != nil {
		return "", nil, err
	}
	path := info.Path()

	return path, util.NewCloser(func() error {
		j.logger.Debug("delete file: " + path)
		if err := j.uploader.LocalUploader().Delete(path); err != nil {
			j.logger.Error("WriteConfigYamlToTmpFile error: ", err)
			return err
		}

		return nil
	}), nil
}

type ReleaseInstallerLoader struct{}

func (r *ReleaseInstallerLoader) Load(j *jobRunner) error {
	const loaderName = "[ReleaseInstallerLoader]: "
	j.Messager().SendMsg(loaderName + "worker 已就绪, 准备安装")
	j.Percenter().To(80)
	j.installer = newReleaseInstaller(j.logger, j.helmer, j.project.Name, j.Namespace().Name, j.chart, j.valuesOptions, j.input.Atomic, j.timeoutSeconds, j.dryRun)
	return nil
}

func (j *jobRunner) DownloadFiles(pid any, commit string, files []string) (string, func(), error) {
	id := fmt.Sprintf("%v", pid)
	dir := fmt.Sprintf("mars_tmp_%s", rand.String(10))
	if err := j.uploader.LocalUploader().MkDir(dir, false); err != nil {
		return "", nil, err
	}

	return j.DownloadFilesToDir(id, commit, files, j.uploader.LocalUploader().AbsolutePath(dir))
}

func (j *jobRunner) DownloadFilesToDir(pid any, commit string, files []string, dir string) (string, func(), error) {
	wg := &sync.WaitGroup{}
	wg.Add(len(files))
	for _, file := range files {
		go func(file string) {
			defer wg.Done()
			defer j.logger.HandlePanic("DownloadFilesToDir")
			raw, err := j.gitServer.GetFileContentWithSha(fmt.Sprintf("%v", pid), commit, file)
			if err != nil {
				j.logger.Error(err)
			}
			localPath := filepath.Join(dir, file)
			if _, err := j.uploader.LocalUploader().Put(localPath, strings.NewReader(raw)); err != nil {
				j.logger.Errorf("[DownloadFilesToDir]: err '%s'", err.Error())
			}
		}(file)
	}
	wg.Wait()

	return dir, func() {
		err := j.uploader.LocalUploader().DeleteDir(dir)
		if err != nil {
			j.logger.Warning(err)
			return
		}
		j.logger.Debug("remove " + dir)
	}, nil
}
