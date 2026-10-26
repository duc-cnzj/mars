package mcp

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/duc-cnzj/mars/api/v6/proto/mars"
	repopb "github.com/duc-cnzj/mars/api/v6/proto/repo"
	apitypes "github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/deploy"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/util/yaml"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/samber/lo"
	"google.golang.org/protobuf/proto"
)

// elementTypeNames 是 ElementType 枚举到对外名称的映射。读工具用它把 proto 的数字枚举
// 投影成类型名（避免模型把 2 误读成别的类型），写工具用它反解类型名。
var elementTypeNames = map[mars.ElementType]string{
	mars.ElementType_ElementTypeInput:        "Input",
	mars.ElementType_ElementTypeInputNumber:  "InputNumber",
	mars.ElementType_ElementTypeSelect:       "Select",
	mars.ElementType_ElementTypeRadio:        "Radio",
	mars.ElementType_ElementTypeSwitch:       "Switch",
	mars.ElementType_ElementTypeTextArea:     "TextArea",
	mars.ElementType_ElementTypeNumberSelect: "NumberSelect",
	mars.ElementType_ElementTypeNumberRadio:  "NumberRadio",
}

// repoCustomField 是自定义字段的对外投影：全字段都输出（不带 omitempty），
// 让模型一眼看到字段定义的完整形状，而不是靠「缺字段 = 零值」去猜。
type repoCustomField struct {
	Path             string   `json:"path"`
	Type             string   `json:"type"`
	Default          string   `json:"default"`
	Description      string   `json:"description"`
	SelectValues     []string `json:"select_values"`
	Order            uint32   `json:"order"`
	TextareaLanguage string   `json:"textarea_language"`
	Group            string   `json:"group"`
}

// repoGroupSetting 是分区展示配置的对外投影：分区名、排序（升序）、是否默认折叠。
type repoGroupSetting struct {
	Name      string `json:"name"`
	Order     uint32 `json:"order"`
	Collapsed bool   `json:"collapsed"`
}

// repoHibernate 是仓库定时休眠配置的对外投影。Hibernate 字段为 null 表示该仓库
// 根本没有休眠配置；enabled=false 则表示配置了但已停用——两者语义不同，不可混用。
type repoHibernate struct {
	Enabled      bool     `json:"enabled"`
	ReplicaPaths []string `json:"replica_paths"`
	WindowNames  []string `json:"window_names"`
}

// repoCustomFieldsOut 是两个自定义字段工具的统一出参：除字段定义外，还回吐仓库
// 部署相关的主要配置（values.yaml、分支白名单、分区、休眠），让模型改字段前后
// 都能看清整体形状——只给 elements 的话，模型无法判断某个字段该挂在哪个分区、
// 也不了解仓库的部署模板长什么样。config_file 等其余配置不在本投影内。
type repoCustomFieldsOut struct {
	RepoID        int32              `json:"repo_id"`
	RepoName      string             `json:"repo_name"`
	ValuesYaml    string             `json:"values_yaml"`
	Branches      []string           `json:"branches"`
	GroupSettings []repoGroupSetting `json:"group_settings"`
	Hibernate     *repoHibernate     `json:"hibernate"`
	Elements      []repoCustomField  `json:"elements"`
}

// getRepoCustomFieldsIn 是 get_repo_custom_fields 的入参。
type getRepoCustomFieldsIn struct {
	RepoID int32 `json:"repo_id" jsonschema:"仓库 ID（必填）"`
}

// repoElementIn 是 update_repo_custom_fields 里单个字段的新定义。
// type 接受短名（InputNumber）或枚举全名（ElementTypeInputNumber）；
// order 用指针区分「未传」（保留/追加）与「显式传 0」。
type repoElementIn struct {
	Path             string   `json:"path" jsonschema:"字段路径（必填），格式如 app->replicas"`
	Type             string   `json:"type" jsonschema:"字段类型（必填）：Input/InputNumber/Select/Radio/Switch/TextArea/NumberSelect/NumberRadio，也接受 ElementType 前缀全名"`
	Default          string   `json:"default,omitempty" jsonschema:"默认值；选择类必须落在 select_values 内"`
	Description      string   `json:"description,omitempty" jsonschema:"字段描述"`
	SelectValues     []string `json:"select_values,omitempty" jsonschema:"可选值列表（Select/Radio/Number* 需要，默认值必须在其中）"`
	Order            *uint32  `json:"order,omitempty" jsonschema:"排序值（可选）：更新已有字段不传保留原值，新增不传追加到末尾"`
	TextareaLanguage string   `json:"textarea_language,omitempty" jsonschema:"TextArea 编辑器语言提示（可选）"`
	Group            string   `json:"group,omitempty" jsonschema:"分区名（可选，仅前端展示）"`
}

// updateRepoCustomFieldsIn 是 update_repo_custom_fields 的入参。
type updateRepoCustomFieldsIn struct {
	RepoID      int32           `json:"repo_id" jsonschema:"仓库 ID（必填）"`
	Upsert      []repoElementIn `json:"upsert_elements,omitempty" jsonschema:"按 path 新增或整条替换自定义字段"`
	RemovePaths []string        `json:"remove_paths,omitempty" jsonschema:"要删除的字段 path 列表；path 不存在会报错"`
}

// elementPatch 是一条待应用的字段变更：element 是校验后的新定义，order 为 nil
// 时表示调用方未指定排序（更新保留原值、新增追加到末尾）。
type elementPatch struct {
	element *mars.Element
	order   *uint32
}

// maxElementGroupBytes 对齐 proto 校验规则：Element.group 非空时最长 64 字节
// （见 api/proto/mars/mars.pb.validate.go）。MCP 在进程内直调 services，gRPC 的
// ValidatorUnaryServerInterceptor 不执行，故必须在此自校验——否则会写进一条之后被
// 后台表单/其他 gRPC 更新路径拒绝的配置，仓库将无法再次保存。
const maxElementGroupBytes = 64

// getRepoCustomFields 读取仓库级自定义字段定义。仅管理员可调用：普通用户的连接
// 根本不会注册本工具（tools/list 不可见），handler 内的 requireAdmin 作为纵深防御。
func (t *toolset) getRepoCustomFields(ctx context.Context, _ *sdk.CallToolRequest, in getRepoCustomFieldsIn) (*sdk.CallToolResult, any, error) {
	if err := t.requireAdmin(ctx, "get_repo_custom_fields"); err != nil {
		return nil, nil, err
	}
	if in.RepoID <= 0 {
		return nil, nil, errs.InvalidArgument("repo_id 必须大于 0")
	}
	resp, err := t.repoSvc.Show(ctx, &repopb.ShowRequest{Id: in.RepoID})
	if err != nil {
		return nil, nil, err
	}
	item := resp.GetItem()
	if item == nil {
		return nil, nil, errs.NotFound(fmt.Sprintf("仓库 %d 不存在", in.RepoID))
	}
	return nil, projectRepoCustomFields(item), nil
}

// updateRepoCustomFields 增量新增/覆盖/删除仓库级自定义字段。仓库配置是整段替换，
// 故必须先 Show 读出当前配置、在克隆上只改 elements，再原值回填其余字段后 Update——
// 否则 values_yaml/branches/配置片段/分区/休眠配置都会被清空。
//
// 并发说明：Show 与 Update 之间没有版本校验（RepoSvc 不提供 CAS），两个管理员同时
// 修改同一仓库时后写者会覆盖先写者的改动；本工具不引入额外锁，调用方应避免并发编辑。
//
// 对存量项目的影响：本工具只改仓库模板，不触发重部署；但项目下次部署（含定时唤醒）
// 会用其已存的自定义值按新定义重新校验——改 type/收窄 select_values 会让不兼容的项目
// 部署失败，删除字段会让该值被静默丢弃并随后从项目记录中移除。这是与 Web 后台编辑
// 完全一致的行为（同一 biz.Update 链路），工具不额外提供跨项目影响预检。
func (t *toolset) updateRepoCustomFields(ctx context.Context, _ *sdk.CallToolRequest, in updateRepoCustomFieldsIn) (*sdk.CallToolResult, any, error) {
	if err := t.requireAdmin(ctx, "update_repo_custom_fields"); err != nil {
		return nil, nil, err
	}
	if in.RepoID <= 0 {
		return nil, nil, errs.InvalidArgument("repo_id 必须大于 0")
	}
	if len(in.Upsert) == 0 && len(in.RemovePaths) == 0 {
		return nil, nil, errs.InvalidArgument("upsert_elements 与 remove_paths 至少提供一个")
	}
	// 先完成纯入参校验，任何非法字段都不触达仓库读/写。
	patches, err := buildElementPatches(in.Upsert)
	if err != nil {
		return nil, nil, err
	}

	resp, err := t.repoSvc.Show(ctx, &repopb.ShowRequest{Id: in.RepoID})
	if err != nil {
		return nil, nil, err
	}
	current := resp.GetItem()
	if current == nil {
		return nil, nil, errs.NotFound(fmt.Sprintf("仓库 %d 不存在", in.RepoID))
	}
	// Repo.Update 在 need_git_repo=true 时会用 git_project_id 重新解析 Git 项目；
	// 未绑定（0）必然失败。提前给出可操作的错误，而不是让底层报「git 项目不存在」。
	if current.GetNeedGitRepo() && current.GetGitProjectId() <= 0 {
		return nil, nil, errs.InvalidArgument(fmt.Sprintf("仓库 %d 标记为需要 git 仓库但未绑定 git_project_id，无法修改自定义字段；请先在后台补齐绑定", in.RepoID))
	}
	cfg := cloneMarsConfig(current.GetMarsConfig())
	elements, err := applyElementPatches(cfg.GetElements(), patches, in.RemovePaths)
	if err != nil {
		return nil, nil, err
	}
	cfg.Elements = elements

	updated, err := t.repoSvc.Update(ctx, &repopb.UpdateRequest{
		Id:           current.GetId(),
		Name:         current.GetName(),
		NeedGitRepo:  current.GetNeedGitRepo(),
		GitProjectId: currentGitProjectID(current),
		MarsConfig:   cfg,
		Description:  current.GetDescription(),
		// AiAgent 留 nil：表示未指定，保持仓库现有 AI Agent 标记不变。
	})
	if err != nil {
		return nil, nil, err
	}
	return nil, projectRepoCustomFields(updated.GetItem()), nil
}

// projectRepoCustomFields 把仓库模型投影成工具出参：字段定义 + 部署相关的主要配置。
func projectRepoCustomFields(item *apitypes.RepoModel) repoCustomFieldsOut {
	cfg := item.GetMarsConfig()
	return repoCustomFieldsOut{
		RepoID:        item.GetId(),
		RepoName:      item.GetName(),
		ValuesYaml:    cfg.GetValuesYaml(),
		Branches:      append([]string{}, cfg.GetBranches()...),
		GroupSettings: projectGroupSettings(cfg.GetGroupSettings()),
		Hibernate:     projectHibernate(cfg.GetHibernate()),
		Elements:      projectElements(cfg.GetElements()),
	}
}

// projectGroupSettings 把分区配置投影成固定形状：空配置回吐空数组而非 null。
func projectGroupSettings(settings []*mars.GroupSetting) []repoGroupSetting {
	out := make([]repoGroupSetting, 0, len(settings))
	for _, s := range settings {
		out = append(out, repoGroupSetting{
			Name:      s.GetName(),
			Order:     s.GetOrder(),
			Collapsed: s.GetCollapsed(),
		})
	}
	return out
}

// projectHibernate 把休眠配置投影成固定形状；nil 保持 null，表示未配置休眠。
func projectHibernate(h *mars.Hibernate) *repoHibernate {
	if h == nil {
		return nil
	}
	return &repoHibernate{
		Enabled:      h.GetEnabled(),
		ReplicaPaths: append([]string{}, h.GetReplicaPaths()...),
		WindowNames:  append([]string{}, h.GetWindowNames()...),
	}
}

// projectElements 把 proto Element 投影成类型名友好的出参。
func projectElements(elements []*mars.Element) []repoCustomField {
	out := make([]repoCustomField, 0, len(elements))
	for _, el := range elements {
		out = append(out, repoCustomField{
			Path:             el.GetPath(),
			Type:             elementTypeName(el.GetType()),
			Default:          el.GetDefault(),
			Description:      el.GetDescription(),
			SelectValues:     append([]string{}, el.GetSelectValues()...),
			Order:            el.GetOrder(),
			TextareaLanguage: el.GetTextareaLanguage(),
			Group:            el.GetGroup(),
		})
	}
	return out
}

// elementTypeName 返回枚举的对外名称；未知类型返回 "Unknown"（便于发现枚举漂移）。
func elementTypeName(t mars.ElementType) string {
	if name, ok := elementTypeNames[t]; ok {
		return name
	}
	return "Unknown"
}

// parseElementType 反解类型名：同时接受短名与 ElementType 前缀全名。
func parseElementType(name string) (mars.ElementType, bool) {
	short := strings.TrimPrefix(strings.TrimSpace(name), "ElementType")
	for t, n := range elementTypeNames {
		if n == short {
			return t, true
		}
	}
	return mars.ElementType_ElementTypeUnknown, false
}

// elementDefaultRequired 判断该类型的默认值是否必填，与后台仓库表单的
// DEFAULT_REQUIRED_TYPES 保持同一口径；Input/TextArea 允许空默认值。
func elementDefaultRequired(t mars.ElementType) bool {
	switch t {
	case mars.ElementType_ElementTypeInputNumber,
		mars.ElementType_ElementTypeSwitch,
		mars.ElementType_ElementTypeSelect,
		mars.ElementType_ElementTypeRadio,
		mars.ElementType_ElementTypeNumberSelect,
		mars.ElementType_ElementTypeNumberRadio:
		return true
	default:
		return false
	}
}

// buildElementPatches 校验 upsert 入参并转换为待应用变更；path 在本次入参内必须唯一。
func buildElementPatches(in []repoElementIn) ([]elementPatch, error) {
	patches := make([]elementPatch, 0, len(in))
	seen := make(map[string]struct{}, len(in))
	for _, raw := range in {
		element, err := toRepoElement(raw)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[element.GetPath()]; ok {
			return nil, errs.InvalidArgument(fmt.Sprintf("upsert_elements 中 path 重复: %s", element.GetPath()))
		}
		seen[element.GetPath()] = struct{}{}
		patches = append(patches, elementPatch{element: element, order: raw.Order})
	}
	return patches, nil
}

// validateElementPath 校验字段路径能被部署链路正确展开：按 `->` 拆开后每段都不能为空、
// 也不能带首尾空白。部署侧 yaml.DeepSetKey 只拒绝首尾分隔符，中段为空或段内带空白都会
// 生成错位的 YAML 键——既不报错也不生效（deepSetItems 对 DeepSetKey 的错误还会静默跳过）。
// 若不在这里拦住，就会写进一条「保存成功、部署不生效」的字段，极难排查。
func validateElementPath(path string) error {
	for _, segment := range strings.Split(path, yaml.Separator) {
		if segment == "" {
			return errs.InvalidArgument(fmt.Sprintf("自定义字段 path 非法（存在空路径段）: %q", path))
		}
		if segment != strings.TrimSpace(segment) {
			return errs.InvalidArgument(fmt.Sprintf("自定义字段 path 非法（路径段首尾不能有空白）: %q", path))
		}
	}
	return nil
}

// toRepoElement 把入参转换为 proto Element，并复用部署链路的默认值校验
// （deploy.ValidateElementDefault），避免在 MCP 侧复制一套类型转换规则。
func toRepoElement(in repoElementIn) (*mars.Element, error) {
	path := strings.TrimSpace(in.Path)
	if path == "" {
		return nil, errs.InvalidArgument("自定义字段 path 不能为空")
	}
	if err := validateElementPath(path); err != nil {
		return nil, err
	}
	elementType, ok := parseElementType(in.Type)
	if !ok {
		return nil, errs.InvalidArgument(fmt.Sprintf("未知的自定义字段类型: %q", in.Type))
	}
	if len(in.Group) > maxElementGroupBytes {
		return nil, errs.InvalidArgument(fmt.Sprintf("自定义字段 %s 的分区名过长（最多 %d 字节，当前 %d）", path, maxElementGroupBytes, len(in.Group)))
	}
	// 描述与默认值的必填口径对齐后台仓库表单（mars-ui RepoFormModal.validateElements）：
	// MCP 若允许空值落库，管理员之后打开后台表单会因「描述必填 / 默认值必填」无法保存该仓库。
	if strings.TrimSpace(in.Description) == "" {
		return nil, errs.InvalidArgument(fmt.Sprintf("自定义字段 %s 的描述不能为空", path))
	}
	if elementDefaultRequired(elementType) && in.Default == "" {
		return nil, errs.InvalidArgument(fmt.Sprintf("自定义字段 %s 的默认值不能为空", path))
	}
	element := &mars.Element{
		Path:             path,
		Type:             elementType,
		Default:          in.Default,
		Description:      in.Description,
		SelectValues:     append([]string(nil), in.SelectValues...),
		TextareaLanguage: in.TextareaLanguage,
		Group:            in.Group,
	}
	if err := deploy.ValidateElementDefault(element); err != nil {
		return nil, err
	}
	return element, nil
}

// applyElementPatches 在现有字段列表上应用增量变更：已存在的 path 原位整条替换，
// 不存在的 path 追加到末尾；remove_paths 必须命中现有字段且不得与 upsert 冲突。
func applyElementPatches(current []*mars.Element, patches []elementPatch, removePaths []string) ([]*mars.Element, error) {
	// 以「去首尾空白后的 path」为索引：历史配置里可能存进过带空白的 path，用原始值
	// 建索引会让这些字段既改不掉也删不了。撞车（多个字段仅空白不同）时明确报错，
	// 不静默合并；但不相关 path 的修改仍应放行，避免一个脏字段锁死整个仓库。
	pathCounts := make(map[string]int, len(current))
	for _, el := range current {
		pathCounts[strings.TrimSpace(el.GetPath())]++
	}
	patchByPath := make(map[string]elementPatch, len(patches))
	for _, p := range patches {
		patchByPath[p.element.GetPath()] = p
	}
	remove := make(map[string]struct{}, len(removePaths))
	for _, raw := range removePaths {
		path := strings.TrimSpace(raw)
		if path == "" {
			return nil, errs.InvalidArgument("remove_paths 不能包含空 path")
		}
		switch pathCounts[path] {
		case 0:
			return nil, errs.InvalidArgument(fmt.Sprintf("要删除的自定义字段不存在: %s", path))
		case 1:
		default:
			return nil, errs.InvalidArgument(fmt.Sprintf("仓库中存在多个仅空白不同的自定义字段 %q，无法安全删除；请先在后台清理重复字段", path))
		}
		if _, ok := remove[path]; ok {
			return nil, errs.InvalidArgument(fmt.Sprintf("remove_paths 中 path 重复: %s", path))
		}
		if _, ok := patchByPath[path]; ok {
			return nil, errs.InvalidArgument(fmt.Sprintf("path 同时出现在 upsert 与 remove 中: %s", path))
		}
		remove[path] = struct{}{}
	}
	// upsert 命中的存量 path 若撞车，无法判断该覆盖哪一条，前置报错。
	for path := range patchByPath {
		if pathCounts[path] > 1 {
			return nil, errs.InvalidArgument(fmt.Sprintf("仓库中存在多个仅空白不同的自定义字段 %q，无法安全覆盖；请先在后台清理重复字段", path))
		}
	}

	var maxOrder uint32
	out := make([]*mars.Element, 0, len(current)+len(patches))
	for _, el := range current {
		path := strings.TrimSpace(el.GetPath())
		if _, ok := remove[path]; ok {
			continue
		}
		patch, ok := patchByPath[path]
		if !ok {
			if el.GetOrder() > maxOrder {
				maxOrder = el.GetOrder()
			}
			out = append(out, el)
			continue
		}
		applyElementOrder(patch.element, patch.order, el.GetOrder())
		// 显式 order 也算进最大值，后续「追加到末尾」的新字段才会排在它之后。
		if patch.element.GetOrder() > maxOrder {
			maxOrder = patch.element.GetOrder()
		}
		out = append(out, patch.element)
		delete(patchByPath, path)
	}
	// 第二个循环按入参顺序追加新增字段；上面已替换的条目已从 map 删除，
	// 仍在 map 里的就是新增项，直接用 p 本身应用排序（map 只用来判断是否已消费）。
	for _, p := range patches {
		if _, ok := patchByPath[p.element.GetPath()]; !ok {
			continue
		}
		if p.order == nil {
			if maxOrder == math.MaxUint32 {
				return nil, errs.InvalidArgument(fmt.Sprintf("自定义字段 order 已达上限 %d，无法自动追加；请显式指定 order 或先调整现有字段顺序", maxOrder))
			}
			maxOrder++
		}
		applyElementOrder(p.element, p.order, maxOrder)
		if p.element.GetOrder() > maxOrder {
			maxOrder = p.element.GetOrder()
		}
		out = append(out, p.element)
		delete(patchByPath, p.element.GetPath())
	}
	return out, nil
}

// applyElementOrder 写入排序值：调用方显式指定时以入参为准，否则用 fallback
// （已有字段保留原 order，新增字段用当前最大值 +1）。
func applyElementOrder(element *mars.Element, order *uint32, fallback uint32) {
	if order != nil {
		element.Order = *order
		return
	}
	element.Order = fallback
}

// cloneMarsConfig 深拷当前配置：Update 会在传入配置上回写派生字段（IsSimpleEnv），
// 不应改动 Show 返回的共享对象。
func cloneMarsConfig(cfg *mars.Config) *mars.Config {
	if cfg == nil {
		return &mars.Config{}
	}
	return proto.Clone(cfg).(*mars.Config)
}

// currentGitProjectID 回填仓库当前绑定的 GitLab 项目。Repo.Update 在 need_git_repo=true
// 时必定用该值重新解析项目，调用方已提前拦掉未绑定（id<=0）的仓库，故这里只在
// 不需要 git 仓库时返回 nil（此时 data 层会清空绑定）。
func currentGitProjectID(item *apitypes.RepoModel) *int32 {
	if !item.GetNeedGitRepo() {
		return nil
	}
	return lo.ToPtr(item.GetGitProjectId())
}
