package deploy

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

	"github.com/duc-cnzj/mars/api/v6/proto/mars"
	websocket_pb "github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"github.com/duc-cnzj/mars/v6/internal/app"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/uploader"
	"github.com/duc-cnzj/mars/v6/internal/util/rand"
	"github.com/duc-cnzj/mars/v6/internal/util/timer"
	yaml2 "github.com/duc-cnzj/mars/v6/internal/util/yaml"
	"go.uber.org/config"
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli/values"
)

// defaultLoaders 返回部署默认的加载链：chart 文件 → 系统变量 → 用户配置 → 后台元素 → 合并。
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

// Loader 是"加载阶段"的插件：把 chart/values 从远程仓库、用户请求与后台配置
// 汇总为一份可安装的 values。它只与 LoadContext 打交道，不感知 jobRunner 的其他状态。
type Loader interface {
	Load(ctx *LoadContext) error
}

// LoadContext 是一次部署"加载阶段"的显式上下文：把加载链需要的只读输入、依赖、
// 中间产物与临时资源清理集中在一起。每个 Loader 只读写它，不再直接触碰 jobRunner
// 的 40 多个字段，使数据流可见、可单测。
type LoadContext struct {
	// 只读输入：jobRunner 在 Validate 阶段准备好，加载链只读不写。
	Config           *mars.Config
	Input            *JobInput
	Project          *biz.Project
	Namespace        *biz.Namespace
	Repo             *biz.Repo
	Commit           *biz.Commit
	ImagePullSecrets []string

	// 依赖：传输层/插件层端口。
	Messager       DeployMsger
	PluginMgr      app.PluginManager
	Helmer         biz.HelmerRepo
	Logger         mlog.Logger
	AccessTokenBiz biz.AccessTokenBiz

	// 中间产物：由加载链依次写入，链结束后由 jobRunner 收口回自身字段。
	Chart            *chart.Chart
	UserConfigYaml   string
	ElementValues    []string
	FinalExtraValues []*websocket_pb.ExtraValue
	SystemValuesYaml string
	Vars             vars
	ValuesOptions    *values.Options

	// 基础设施：下载 chart 文件与写临时 values 文件时使用。
	uploader uploader.Uploader
	timer    timer.Timer

	// cleanups 是加载期产生的临时资源清理函数（下载目录、临时 values 文件），
	// 链结束后由 jobRunner 统一注册为 finally 回调，保证部署结束一定回收。
	cleanups []func()
}

// AddCleanup 登记一个临时资源清理函数，部署结束后由 jobRunner 统一执行。
func (c *LoadContext) AddCleanup(fn func()) {
	c.cleanups = append(c.cleanups, fn)
}

// DownloadFiles 把 git 仓库中的文件下载到随机临时目录，返回目录与清理函数。
func (c *LoadContext) DownloadFiles(pid any, commit string, files []string) (string, func(), error) {
	id := fmt.Sprintf("%v", pid)
	dir := fmt.Sprintf("mars_tmp_%s", rand.String(10))
	if err := c.uploader.LocalUploader().MkDir(dir, true); err != nil {
		return "", nil, err
	}

	downloadDir, deleteFn := c.DownloadFilesToDir(id, commit, files, c.uploader.LocalUploader().AbsolutePath(dir))
	return downloadDir, deleteFn, nil
}

// DownloadFilesToDir 把 git 仓库里的文件下载到 dir；错误在函数内部记录，
// 恒返回 nil error，故不设 error 返回值（调用方无需也无法处理失败）。
func (c *LoadContext) DownloadFilesToDir(pid any, commit string, files []string, dir string) (string, func()) {
	wg := &sync.WaitGroup{}
	wg.Add(len(files))
	for _, file := range files {
		go func(file string) {
			defer wg.Done()
			defer c.Logger.HandlePanic("DownloadFilesToDir")
			raw, err := c.PluginMgr.Git().GetFileContentWithSha(fmt.Sprintf("%v", pid), commit, file)
			if err != nil {
				c.Logger.Error(err)
			}
			localPath := filepath.Join(dir, file)
			if _, err := c.uploader.LocalUploader().Put(localPath, strings.NewReader(raw)); err != nil {
				c.Logger.Errorf("[DownloadFilesToDir]: err '%s'", err.Error())
			}
		}(file)
	}
	wg.Wait()

	return dir, func() {
		err := c.uploader.LocalUploader().DeleteDir(dir)
		if err != nil {
			c.Logger.Warning(err)
			return
		}
		c.Logger.Debug("remove " + dir)
	}
}

// WriteConfigYamlToTmpFile 将合并后的 values 写入临时文件，返回路径与清理器。
func (c *LoadContext) WriteConfigYamlToTmpFile(data []byte) (string, io.Closer, error) {
	file := fmt.Sprintf("mars-%s-%s.yaml", c.timer.Now().Format("2006-01-02"), rand.String(20))
	info, err := c.uploader.LocalUploader().Put(file, bytes.NewReader(data))
	if err != nil {
		return "", nil, err
	}
	path := info.Path()

	return path, NewCloser(func() error {
		c.Logger.Debug("delete file: " + path)
		if err := c.uploader.LocalUploader().Delete(path); err != nil {
			c.Logger.Error("WriteConfigYamlToTmpFile error: ", err)
			return err
		}

		return nil
	}), nil
}

type helmChartLoader interface {
	LoadDir(dir string) (*chart.Chart, error)
	LoadArchive(in io.Reader) (*chart.Chart, error)
}

type defaultFileOpener struct{}

// Open 打开本地文件，句柄由调用方（ChartFileLoader）负责关闭。
func (d *defaultFileOpener) Open(name string) (io.ReadCloser, error) {
	return os.Open(name)
}

type defaultChartLoader struct{}

// LoadArchive 从 Reader 加载 helm chart 归档。
func (d *defaultChartLoader) LoadArchive(in io.Reader) (*chart.Chart, error) {
	return loader.LoadArchive(in)
}

// LoadDir 从本地目录加载 helm chart。
func (d *defaultChartLoader) LoadDir(dir string) (*chart.Chart, error) {
	return loader.LoadDir(dir)
}

type fileOpener interface {
	Open(name string) (io.ReadCloser, error)
}

// ChartFileLoader 下载远程 chart 文件
type ChartFileLoader struct {
	chartLoader helmChartLoader
	fileOpener  fileOpener
}

// Load 从远程 git 仓库下载 chart 文件、打包并载入，同时处理 file:// 本地依赖。
func (c *ChartFileLoader) Load(ctx *LoadContext) error {
	const loaderName = "[ChartFileLoader]: "
	ctx.Messager.SendMsg(loaderName + "加载 helm chart 文件")
	ctx.Messager.To(20)

	if !biz.IsRemoteLocalChartPath(ctx.Config.LocalChartPath) {
		return errors.New("LocalChartPath 格式不正确: " + ctx.Config.LocalChartPath)
	}

	// 下载 helm charts
	split := strings.Split(ctx.Config.LocalChartPath, "|")
	var (
		files        []string
		tmpChartsDir string
		deleteDirFn  func()
		dir          string
	)
	// 如果是这个格式意味着是远程项目, 'uid|branch|path'
	ctx.Messager.SendMsg(fmt.Sprintf(loaderName+"下载 helm charts path: %s", ctx.Config.LocalChartPath))

	var (
		pid    = split[0]
		branch = split[1]
		path   = split[2]
	)

	files, err := ctx.PluginMgr.Git().GetDirectoryFilesWithBranch(pid, branch, path, true)
	if err != nil {
		// git 故障如实上报：吞掉会误报成"charts 文件不存在"，排障误导向。
		// 走 errs.Wrap 自动归类（git 未识别类型落 500），不散落裸 wrap。
		return errs.Wrap(err, "获取远程 charts 文件")
	}
	if len(files) < 1 {
		return fmt.Errorf("charts 文件不存在: %s", ctx.Config.LocalChartPath)
	}
	tmpChartsDir, deleteDirFn, err = ctx.DownloadFiles(pid, branch, files)
	if err != nil {
		return err
	}

	dir = path
	ctx.Messager.SendMsg(fmt.Sprintf(loaderName+"识别为远程仓库 uid %v branch %s path %s", pid, branch, path))

	// 下载的临时目录由 jobRunner 在部署结束后统一回收。
	ctx.AddCleanup(deleteDirFn)

	loadDir, err := c.chartLoader.LoadDir(filepath.Join(tmpChartsDir, dir))
	if err != nil {
		return err
	}
	if loadDir.Metadata.Dependencies != nil && action.CheckDependencies(loadDir, loadDir.Metadata.Dependencies) != nil {
		for _, dependency := range loadDir.Metadata.Dependencies {
			if strings.HasPrefix(dependency.Repository, "file://") {
				depFiles, _ := ctx.PluginMgr.Git().GetDirectoryFilesWithBranch(pid, branch, filepath.Join(path, strings.TrimPrefix(dependency.Repository, "file://")), true)
				_, depDeleteFn := ctx.DownloadFilesToDir(pid, branch, depFiles, tmpChartsDir)
				ctx.AddCleanup(depDeleteFn)
				ctx.Messager.SendMsg(fmt.Sprintf(loaderName+"下载本地依赖 %s", dependency.Name))
			}
		}
	}

	chartDir := filepath.Join(tmpChartsDir, dir)

	ctx.Messager.To(30)
	ctx.Messager.SendMsg(loaderName + "打包 helm charts")
	chart, err := ctx.Helmer.PackageChart(chartDir, chartDir)
	if err != nil {
		return err
	}
	archive, err := c.fileOpener.Open(chart)
	if err != nil {
		return err
	}
	defer archive.Close()

	ctx.Chart, err = c.chartLoader.LoadArchive(archive)

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

// Load 解析用户请求传入的配置片段并暂存到 UserConfigYaml 供合并阶段使用。
func (d *UserConfigLoader) Load(ctx *LoadContext) error {
	const loaderName = "[UserConfigLoader]: "

	ctx.Messager.To(50)
	ctx.Messager.SendMsg(fmt.Sprintf(loaderName+"%v", "检查到用户传入的配置"))

	if ctx.Input.Config == "" {
		ctx.Messager.SendMsg(fmt.Sprintf(loaderName+"%v", "未发现用户自定义配置"))
		return nil
	}

	userConfigYaml, err := biz.ParseInputConfig(ctx.Config, ctx.Input.Config)
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	ctx.UserConfigYaml = userConfigYaml

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

// Load 按后台配置的元素定义校验用户传入的 ExtraValues，
// 生成合法的 values 片段与最终元素值列表。
func (d *ElementsLoader) Load(ctx *LoadContext) error {
	const loaderName = "[ElementsLoader]: "

	ctx.Messager.To(60)
	ctx.Messager.SendMsg(fmt.Sprintf(loaderName+"%v", "检查项目额外的配置"))

	if len(ctx.Input.ExtraValues) <= 0 {
		ctx.Messager.SendMsg(fmt.Sprintf(loaderName+"%v", "未发现项目额外的配置"))
	}

	var validValuesMap = make(map[string]any)
	var useDefaultMap = make(map[string]bool)

	// 在副本上排序：ctx.Config.Elements 是共享的 repo.MarsConfig.Elements，
	// 原地排序会让同 repo 的并发部署互相踩踏切片。
	elements := append([]*mars.Element(nil), ctx.Config.Elements...)
	sort.Slice(elements, func(x, y int) bool {
		return elements[x].Order < elements[y].Order
	})
	var configElementsMap = make(map[string]*mars.Element)
	// declaredPaths 按元素声明顺序（已按 Order 排序）收集允许的字段名，仅供拒绝时的提示文案使用。
	declaredPaths := make([]string, 0, len(elements))
	for _, element := range elements {
		configElementsMap[element.Path] = element
		declaredPaths = append(declaredPaths, element.Path)
		defaultValue, e := d.typedValue(element, element.Default)
		if e != nil {
			return e
		}
		validValuesMap[element.Path] = defaultValue
		useDefaultMap[element.Path] = true
	}

	// validate
	for _, value := range ctx.Input.ExtraValues {
		element, ok := configElementsMap[value.Path]
		if !ok {
			// 声明外字段只在「历史残留」时容忍：仓库配置此后删掉了该元素，但项目行里仍留着
			// 它、前端会照搬 finalExtraValues 回传——丢弃即可，不能因此卡死老项目。
			// 真正的凭空字段一律拒绝：静默丢弃会让调用方以为设上了，实际没生效。
			if isLegacyExtraValue(ctx.Project, value.Path) {
				continue
			}
			return errs.InvalidArgument(
				fmt.Sprintf("不允许自定义字段 %s, 允许的字段为 '%s'", value.Path, strings.Join(declaredPaths, ",")))
		}
		useDefaultMap[value.Path] = false
		typeValue, err := d.typedValue(element, value.Value)
		if err != nil {
			return err
		}
		validValuesMap[value.Path] = typeValue
	}

	ctx.ElementValues = d.deepSetItems(validValuesMap)
	var finalValues []*websocket_pb.ExtraValue
	for _, element := range ctx.Config.Elements {
		finalValues = append(finalValues, &websocket_pb.ExtraValue{
			Path:  element.Path,
			Value: fmt.Sprintf("%v", validValuesMap[element.Path]),
		})
	}

	ctx.FinalExtraValues = finalValues
	var ds []string
	for k, ok := range useDefaultMap {
		if ok {
			ds = append(ds, k)
		}
	}
	if len(ds) > 0 {
		ctx.Messager.SendMsg(fmt.Sprintf(loaderName+"已经为 '%s' 设置系统默认值", strings.Join(ds, ",")))
	}

	return nil
}

// typedValue 按元素声明的类型校验并转换输入值（bool/整数/枚举/透传）。
// 校验失败一律用 errs.InvalidArgument 构造：这是调用方的入参错误（400），
// 不是服务端故障，裸 error 会被协议层归成 Unknown 让调用方无从判断。
func (d *ElementsLoader) typedValue(element *mars.Element, input string) (any, error) {
	switch element.Type {
	case mars.ElementType_ElementTypeSwitch:
		if input == "" {
			input = "false"
		}
		v, err := strconv.ParseBool(input)
		if err != nil {
			return nil, errs.InvalidArgument(fmt.Sprintf("%s 字段类型不正确，应该为 bool，你传入的是 %s", element.Path, input))
		}
		return v, nil
	case mars.ElementType_ElementTypeInputNumber:
		if input == "" {
			input = "0"
		}
		v, err := strconv.ParseInt(input, 10, 64)
		if err != nil {
			return nil, errs.InvalidArgument(fmt.Sprintf("%s 字段类型不正确，应该为整数，你传入的是 %s", element.Path, input))
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
			return nil, errs.InvalidArgument(
				fmt.Sprintf("%s 必须在 '%v' 里面, 你传的是 %s", element.Path, strings.Join(element.SelectValues, ","), input))
		}
		if element.Type == mars.ElementType_ElementTypeNumberSelect ||
			element.Type == mars.ElementType_ElementTypeNumberRadio {
			if atoi, err := strconv.Atoi(input); err == nil {
				return atoi, nil
			}
			return nil, errs.InvalidArgument(fmt.Sprintf("[ElementsLoader]: '%v' 非 number 类型, 无法转换", input))
		}

		return input, nil
	default:
		return input, nil
	}
}

// isLegacyExtraValue 判断该字段名是不是「项目已落库的自定义元素值」的历史残留：仓库配置此后
// 删除了该元素，但项目行里仍留着它，前端会按 finalExtraValues 原样回传。只有这种情况才容忍
// 丢弃——项目的当前值就是权威依据，凭它可区分「老数据尾随」与「调用方凭空捏造」。
func isLegacyExtraValue(project *biz.Project, path string) bool {
	if project == nil {
		return false
	}
	for _, v := range project.FinalExtraValues {
		if v.GetPath() == path {
			return true
		}
	}
	return false
}

// deepSetItems 将 map 展开为按路径深设的 yaml 片段列表。
func (d *ElementsLoader) deepSetItems(items map[string]any) []string {
	var evs []string
	for k, v := range items {
		ysk, err := yaml2.DeepSetKey(k, v)
		if err != nil {
			continue
		}
		evs = append(evs, string(ysk))
	}
	return evs
}

// 系统变量名常量与模板分隔符。
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
	VarNamespace              = "Namespace"
	VarLongCommit             = "LongCommit"
	VarUserToken              = "UserToken"
)

// UserTokenTTL 是部署时签发给部署用户的临时令牌有效期。导出供 agent 沙箱
// 唤醒刷新令牌时复用同一时长。
//
// 注意：令牌签发后不续期。AI 沙箱持续运行超过 12h 会因令牌过期而无法回连 mars；
// 默认空闲 30m 会先缩容、唤醒时轮换令牌，只有空闲关闭被禁用或调大到 12h 以上时
// 才会触发该边界。
const UserTokenTTL = 12 * time.Hour

var (
	tagRegex          = regexp.MustCompile(leftDelim + `\s*(\.Branch|\.Commit|\.Pipeline)\s*` + rightDelim)
	userTokenTagRegex = regexp.MustCompile(leftDelim + `\s*\.` + VarUserToken + `\s*` + rightDelim)
)

// HasUserTokenTemplate 判断 values.yaml 是否引用了 <.UserToken>。上层（AI Agent
// 沙箱的启用入口）用它做前置校验：未被引用时部署不会签发令牌，沙箱起来后访问
// mars 会全部 401——这是没有报错的软失败，必须在使用前拦住。
func HasUserTokenTemplate(valuesYaml string) bool {
	return userTokenTagRegex.MatchString(valuesYaml)
}

// SystemVariableLoader 系统内置变量替换， values.yaml 中的 <.var> 变量
// 例如：<.Branch> <.Commit> <.Pipeline> <.Host1> <.TlsSecret1>
type SystemVariableLoader struct {
	values vars
}

// Add 注入一个系统变量。
func (v *SystemVariableLoader) Add(key, value string) {
	v.values.Add(key, value)
}

// Load 注入内置系统变量（镜像仓库、域名、分支/commit/pipeline 等），
// 并用它们渲染 values.yaml 中的 <Var> 占位符。
func (v *SystemVariableLoader) Load(ctx *LoadContext) error {
	const loaderName = "[SystemVariableLoader]: "
	ctx.Messager.To(40)
	ctx.Messager.SendMsg(fmt.Sprintf(loaderName+"%v", "注入内置环境变量"))

	if ctx.Config.ValuesYaml == "" {
		ctx.Messager.SendMsg(fmt.Sprintf(loaderName+"%v", "未发现可用的 values.yaml"))
		return nil
	}

	if v.values == nil {
		v.values = vars{}
	}

	// 部署用户临时令牌：每次部署重新签发，并撤销该项目上一份令牌。
	if err := v.injectUserToken(ctx); err != nil {
		return err
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
	renderResult := &bytes.Buffer{}
	t.ExecuteTemplate(renderResult, "pullSecrets", pullSecrets{
		ImagePullSecrets: ctx.ImagePullSecrets,
	})

	// ImagePullSecretsNoName
	// [secret1, secret2, ]
	renderResultNoName := &bytes.Buffer{}
	t.ExecuteTemplate(renderResultNoName, "pullSecretsNoName", pullSecrets{
		ImagePullSecrets: ctx.ImagePullSecrets,
	})

	v.Add(VarImagePullSecrets, renderResult.String())
	v.Add(VarImagePullSecretsNoName, renderResultNoName.String())
	v.Add(VarNamespace, ctx.Namespace.Name)

	//Host1...Host10
	sub := biz.GetPreOccupiedLenByValuesYaml(ctx.Config.ValuesYaml)
	ctx.Logger.Debug("getPreOccupiedLenByValuesYaml: ", sub)
	dm := ctx.PluginMgr.Domain()
	for i := 1; i <= 10; i++ {
		v.Add(fmt.Sprintf("%s%d", VarHost, i), dm.GetDomainByIndex(ctx.Project.Name, ctx.Namespace.Name, i, sub))
		v.Add(fmt.Sprintf("%s%d", VarTlsSecret, i), dm.GetCertSecretName(ctx.Project.Name, i))
	}

	//{{.Branch}}{{.Commit}}{{.Pipeline}}
	var (
		pipelineID          int64
		pipelineBranch      = ctx.Input.GitBranch
		pipelineShortCommit = ctx.Commit.ShortID
		pipelineLongCommit  = ctx.Input.GitCommit
	)

	v.Add(VarLongCommit, pipelineLongCommit)

	if ctx.Repo.NeedGitRepo {
		// 如果存在需要传变量的，则必须有流水线信息
		if pipeline, e := ctx.PluginMgr.Git().GetCommitPipeline(fmt.Sprintf("%d", ctx.Repo.GitProjectID), pipelineBranch, pipelineLongCommit); e == nil {
			pipelineID = pipeline.ID
			pipelineBranch = pipeline.Ref

			ctx.Messager.SendMsg(fmt.Sprintf(loaderName+"镜像分支 %s 镜像commit %s 镜像 pipeline_id %d", pipelineBranch, pipelineShortCommit, pipelineID))
		} else {
			if tagRegex.MatchString(ctx.Config.ValuesYaml) {
				return fmt.Errorf("无法获取 Pipeline 信息, branch: %s, commit: %s", pipelineBranch, pipelineShortCommit)
			}
		}
	}

	v.Add(VarBranch, pipelineBranch)
	v.Add(VarCommit, pipelineShortCommit)
	v.Add(VarPipeline, fmt.Sprintf("%d", pipelineID))

	// ingress
	v.Add(VarClusterIssuer, dm.GetClusterIssuer())

	tpl, err := template.New("values_yaml").Delims(leftDelim, rightDelim).Parse(ctx.Config.ValuesYaml)
	if err != nil {
		return err
	}
	bf := bytes.Buffer{}
	tpl.Execute(&bf, v.values)
	ctx.SystemValuesYaml = bf.String()
	ctx.Vars = v.values

	return nil
}

// UserTokenUsage 返回项目部署临时令牌的 usage 标记：按项目维度轮换，仅保留最新
// 一份有效。导出供 agent 沙箱唤醒刷新令牌时复用同一轮换维度。
func UserTokenUsage(projectID int) string {
	return fmt.Sprintf("deploy-token:%d", projectID)
}

// injectUserToken 为本次部署签发 12h 临时 access token，写入 <.UserToken>。
// dry-run 与无用户上下文的部署不签发（变量渲染为空串）。
// 旧令牌不在加载阶段撤销：必须等新 release 安装成功后再撤销（见 jobRunner.Run），
// 否则安装失败/回滚时线上旧实例的令牌会提前失效。
func (v *SystemVariableLoader) injectUserToken(ctx *LoadContext) error {
	v.Add(VarUserToken, "")
	if !ctx.Input.IsNotDryRun() || ctx.Input.User == nil {
		return nil
	}
	// values.yaml 未引用该变量时不签发，避免为用不到令牌的项目持续产生令牌记录。
	if !HasUserTokenTemplate(ctx.Config.ValuesYaml) {
		return nil
	}

	token, err := ctx.AccessTokenBiz.Grant(context.TODO(), &biz.GrantAccessTokenInput{
		ExpireSeconds: int32(UserTokenTTL / time.Second),
		Usage:         UserTokenUsage(ctx.Project.ID),
		User:          tokenUser(ctx),
	})
	if err != nil {
		return err
	}

	v.Add(VarUserToken, token.Token)
	return nil
}

// tokenUser 返回部署令牌应绑定的身份：AI Agent 沙箱项目恒为项目创建者，其余项目
// 为本次部署发起人。
//
// 沙箱里的 agent 用这份令牌回连 mars（配置/MCP/LLM 代理），它代表的是「沙箱归谁」，
// 不是「谁点了这次部署」。管理员进入别人的 agent 项目重新部署时，若沿用发起人身份，
// 沙箱就会以管理员令牌运行——越权面被放大。创建者本人触发时沿用其完整 UserInfo
// （保留角色与展示名）；跨用户触发时只带创建者邮箱，角色按最小权限留空。
//
// 调用前置条件由 injectUserToken 保证：User 为 nil 时已提前返回，Project 为 nil 时
// 取 Usage 会先 panic，故此处无需再判空。
func tokenUser(ctx *LoadContext) *biz.UserInfo {
	requester := ctx.Input.User
	if requester.Email != ctx.Project.Creator && biz.IsAgentProjectOf(ctx.Project.Name, ctx.Project.Creator) {
		return &biz.UserInfo{Email: ctx.Project.Creator}
	}
	return requester
}

// MergeValuesLoader 合并系统变量/用户配置/后台元素与镜像拉取密钥为一份 values.yaml。
type MergeValuesLoader struct{}

// Load 合并各片段为最终 values，写入临时文件并登记清理；imagePullSecrets 会自动注入。
func (m *MergeValuesLoader) Load(ctx *LoadContext) error {
	const loaderName = "[MergeValuesLoader]: "
	ctx.Messager.To(70)
	ctx.Messager.SendMsg(fmt.Sprintf(loaderName+"%v", "合并配置文件到 values.yaml"))

	// 自动注入 imagePullSecrets
	var imagePullSecrets = make([]map[string]any, len(ctx.ImagePullSecrets))
	for i, s := range ctx.ImagePullSecrets {
		imagePullSecrets[i] = map[string]any{"name": s}
	}
	var yamlImagePullSecrets []byte
	if len(imagePullSecrets) > 0 {
		yamlImagePullSecrets, _ = yaml.Marshal(map[string]any{
			"imagePullSecrets": imagePullSecrets,
		})
	}

	var opts []config.YAMLOption
	if ctx.SystemValuesYaml != "" {
		opts = append(opts, config.Source(strings.NewReader(ctx.SystemValuesYaml)))
	}
	if ctx.UserConfigYaml != "" {
		opts = append(opts, config.Source(strings.NewReader(ctx.UserConfigYaml)))
	}
	if len(yamlImagePullSecrets) != 0 {
		opts = append(opts, config.Source(bytes.NewReader(yamlImagePullSecrets)))
	}

	for _, value := range ctx.ElementValues {
		opts = append(opts, config.Source(strings.NewReader(value)))
	}

	if len(opts) < 1 {
		return nil
	}

	// 5. 用用户传入的yaml配置去合并 `default_values`
	provider, err := config.NewYAML(opts...)
	if err != nil {
		return err
	}
	var mergedDefaultAndConfigYamlValues map[string]any
	// NewYAML 已在构造阶段解析并合并全部 source；Populate 与 yaml.Marshal
	// 只作用于已解析的 YAML 原生值，不会失败，故不检查其错误（避免不可达分支）。
	_ = provider.Get("").Populate(&mergedDefaultAndConfigYamlValues)

	var fileData []byte
	fileData, _ = yaml.Marshal(&mergedDefaultAndConfigYamlValues)
	mergedFile, closer, err := ctx.WriteConfigYamlToTmpFile(fileData)
	if err != nil {
		return err
	}
	// 临时 values 文件由 jobRunner 在部署结束后统一删除。
	ctx.AddCleanup(func() { _ = closer.Close() })
	ctx.ValuesOptions.ValueFiles = append(ctx.ValuesOptions.ValueFiles, mergedFile)

	return nil
}
