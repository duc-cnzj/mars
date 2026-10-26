package mcp

import (
	"context"

	repopb "github.com/duc-cnzj/mars/api/v6/proto/repo"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// listReposIn 是 list_repos 的入参：全部可选，省略即不筛选、取第一页。
// enabled / ai_agent 用指针是为了区分「不传」与「显式 false」——后者是有效筛选条件。
type listReposIn struct {
	paginationIn
	Name    string `json:"name,omitempty" jsonschema:"按仓库名模糊过滤（可选，子串匹配）"`
	Enabled *bool  `json:"enabled,omitempty" jsonschema:"按启用状态过滤（可选）：true 只看已启用，false 只看已停用，不传则不筛选"`
	AiAgent *bool  `json:"ai_agent,omitempty" jsonschema:"按 AI Agent 仓库标记过滤（可选）：true 只看 AI Agent 仓库，false 只看非 AI Agent 仓库"`
}

// listRepos 列出仓库（管理员专属）：复用 RepoSvc.List 做分页与名称/状态过滤，
// 出参即仓库列表模型（items 不含 mars_config——数据层列表查询本就只取概要列）。
//
// 门禁必须先于调用：RepoSvc.List 在 gRPC 侧对登录用户放行（Authorize allowlist），
// 但仓库后台属管理面能力，MCP 只对管理员开放。
func (t *toolset) listRepos(ctx context.Context, _ *sdk.CallToolRequest, in listReposIn) (*sdk.CallToolResult, any, error) {
	if err := t.requireAdmin(ctx, "list_repos"); err != nil {
		return nil, nil, err
	}
	resp, err := t.repoSvc.List(ctx, &repopb.ListRequest{
		Page:     &in.Page,
		PageSize: &in.PageSize,
		Name:     in.Name,
		Enabled:  in.Enabled,
		AiAgent:  in.AiAgent,
	})
	if err != nil {
		return nil, nil, err
	}
	return nil, resp, nil
}
