package aichat

// skills.go 承载服务端对话的 Skills 能力：把 SKILL.md（YAML frontmatter + 正文）
// 内嵌进二进制，system prompt 只下发「名称 + 描述」目录，正文由本地工具
// load_skill 按需加载——与 opencode 的渐进披露机制一致，避免每轮都把全部手册
// 塞进上下文。
//
// 同步约定：skills/ 下的文件来自 github.com/duc-cnzj/mars-ai-agent 的 skills/ 目录。
// 上游变更时需手工同步；同步后要按服务端对话口径改写三处差异——无沙箱 shell、
// 身份是当前登录用户（而非项目创建者）、写工具必须经人工确认才会执行。

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// skillToolName 是按需加载 skill 正文的本地工具名。
// 该工具不经过 MCP：skills 是 mars 自己的知识库，与服务端工具集无关。
const skillToolName = "load_skill"

// skillsRoot 是内嵌文件系统中的 skills 根目录名。
const skillsRoot = "skills"

// skillManifestName 是每个 skill 目录下的清单文件名（与 opencode 约定一致）。
const skillManifestName = "SKILL.md"

//go:embed skills
var skillsFS embed.FS

// skill 是一份已解析的排查手册：frontmatter 提供目录信息，Body 是完整正文。
type skill struct {
	// Name 是 skill 的唯一标识，也是 load_skill 的入参值。
	Name string `yaml:"name"`
	// Description 是目录里展示给模型的一句话说明，决定模型何时想起它。
	Description string `yaml:"description"`
	// Body 是 frontmatter 之后的完整 Markdown 正文。
	Body string `yaml:"-"`
}

// skillRegistry 是全部内嵌 skill 的索引：order 保证目录输出稳定，
// byName 提供 O(1) 的按名加载。
type skillRegistry struct {
	order  []string
	byName map[string]skill
}

// newSkillRegistry 从任意文件系统加载 skills：root 下每个 <name>/SKILL.md 视为一份
// skill。抽成 fs.FS 入参是为了让「frontmatter 损坏 / 名称重复」等错误分支可单测，
// 生产调用固定传内嵌的 skillsFS。
func newSkillRegistry(fsys fs.FS, root string) (*skillRegistry, error) {
	reg := &skillRegistry{byName: map[string]skill{}}
	err := fs.WalkDir(fsys, root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || d.Name() != skillManifestName {
			return nil
		}
		raw, err := fs.ReadFile(fsys, p)
		if err != nil {
			return fmt.Errorf("读取 %s 失败: %w", p, err)
		}
		sk, err := parseSkill(raw)
		if err != nil {
			return fmt.Errorf("解析 %s 失败: %w", p, err)
		}
		if sk.Name == "" {
			return fmt.Errorf("%s 缺少 frontmatter name", p)
		}
		if _, exists := reg.byName[sk.Name]; exists {
			return fmt.Errorf("%s 的 skill 名称重复: %s", p, sk.Name)
		}
		reg.byName[sk.Name] = sk
		return nil
	})
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(reg.byName))
	for name := range reg.byName {
		names = append(names, name)
	}
	sort.Strings(names)
	reg.order = names
	return reg, nil
}

// mustLoadEmbeddedSkills 加载随二进制内嵌的 skills。
func mustLoadEmbeddedSkills() *skillRegistry {
	return mustRegistry(skillsFS, skillsRoot)
}

// mustRegistry 加载 skill 目录。内嵌内容是编译期固定的静态文件，解析失败属于
// 构建/同步错误，启动时直接 panic 比「静默少一个 skill」更容易被发现。
func mustRegistry(fsys fs.FS, root string) *skillRegistry {
	reg, err := newSkillRegistry(fsys, root)
	if err != nil {
		panic(fmt.Sprintf("加载内嵌 skills 失败: %v", err))
	}
	return reg
}

// parseSkill 解析 YAML frontmatter 与正文。格式：
//
//	---
//	name: xxx
//	description: ...
//	---
//	# 正文
func parseSkill(raw []byte) (skill, error) {
	text := strings.ReplaceAll(string(raw), "\r\n", "\n")
	if !strings.HasPrefix(text, "---\n") {
		return skill{}, errors.New("缺少 YAML frontmatter")
	}
	rest := strings.TrimPrefix(text, "---\n")
	end := strings.Index(rest, "\n---")
	if end < 0 {
		return skill{}, errors.New("frontmatter 未闭合")
	}
	frontmatter := rest[:end]
	body := strings.TrimPrefix(rest[end+len("\n---"):], "\n")

	var sk skill
	if err := yaml.Unmarshal([]byte(frontmatter), &sk); err != nil {
		return skill{}, fmt.Errorf("frontmatter 不是合法 YAML: %w", err)
	}
	sk.Name = strings.TrimSpace(sk.Name)
	sk.Description = strings.TrimSpace(sk.Description)
	sk.Body = strings.TrimSpace(body)
	if sk.Body == "" {
		return skill{}, errors.New("skill 正文为空")
	}
	return sk, nil
}

// catalog 生成 system prompt 里的 skill 目录。空目录返回空串，调用方据此省略整段，
// 避免给模型一段只有标题没有内容的噪声。
func (r *skillRegistry) catalog() string {
	if len(r.order) == 0 {
		return ""
	}
	var b strings.Builder
	for _, name := range r.order {
		sk := r.byName[name]
		fmt.Fprintf(&b, "- %s：%s\n", sk.Name, sk.Description)
	}
	return b.String()
}

// load 返回指定 skill 的完整正文。未知名称返回错误——模型编造 skill 名称时
// 应看到明确失败，而不是拿到空串继续瞎猜。
func (r *skillRegistry) load(name string) (string, error) {
	sk, ok := r.byName[strings.TrimSpace(name)]
	if !ok {
		return "", fmt.Errorf("未知 skill: %s", name)
	}
	return sk.Body, nil
}

// toolSpec 描述本地 load_skill 工具：只读、入参只有一个受枚举约束的 name。
func (r *skillRegistry) toolSpec() toolSpec {
	schema, _ := json.Marshal(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{
				"type":        "string",
				"enum":        r.order,
				"description": "要加载的 skill 名称",
			},
		},
		"required": []string{"name"},
	})
	return toolSpec{
		Name: skillToolName,
		Description: "加载 mars 内嵌的排查手册正文（SKILL.md）。" +
			"当需要按标准套路排查问题、或不确定 mars 资源模型/工具入参时，先加载对应 skill 再行动。",
		InputSchema: schema,
		ReadOnly:    true,
	}
}
