package gitlab

import (
	"errors"
	"sort"

	"github.com/duc-cnzj/mars/v6/internal/app"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/proxy"
	"github.com/xanzy/go-gitlab"
)

var _ app.GitServer = (*server)(nil)

// gitlabName 插件注册名。
const gitlabName = "gitlab"

func init() {
	dr := &server{}
	app.RegisterPlugin(dr.Name(), dr)
}

// toGitProject 将 go-gitlab 的 Project 转成业务层 GitProject；nil 输入返回 nil。
func toGitProject(p *gitlab.Project) *biz.GitProject {
	if p == nil {
		return nil
	}
	return &biz.GitProject{
		ID:            int64(p.ID),
		Name:          p.Name,
		DefaultBranch: p.DefaultBranch,
		WebURL:        p.WebURL,
		Path:          p.Path,
		AvatarURL:     p.AvatarURL,
		Description:   p.Description,
	}
}

// toBranch 将 go-gitlab 的 Branch 转成业务层 Branch；nil 输入返回 nil。
func toBranch(b *gitlab.Branch) *biz.Branch {
	if b == nil {
		return nil
	}
	return &biz.Branch{
		Name:      b.Name,
		IsDefault: b.Default,
		WebURL:    b.WebURL,
	}
}

// toCommit 将 go-gitlab 的 Commit 转成业务层 Commit；nil 输入返回 nil。
func toCommit(c *gitlab.Commit) *biz.Commit {
	if c == nil {
		return nil
	}
	return &biz.Commit{
		ID:             c.ID,
		ShortID:        c.ShortID,
		Title:          c.Title,
		CommittedDate:  c.CommittedDate,
		AuthorName:     c.AuthorName,
		AuthorEmail:    c.AuthorEmail,
		CommitterName:  c.CommitterName,
		CommitterEmail: c.CommitterEmail,
		CreatedAt:      c.CreatedAt,
		Message:        c.Message,
		WebURL:         c.WebURL,
	}
}

// pipelineStatus 将 GitLab pipeline 状态字符串映射为业务层 Status。
// 可能值：created, waiting_for_resource, preparing, pending, running, success,
// failed, canceled, skipped, manual, scheduled；其中 manual 单独映射为
// StatusManual（区别于 success：存在手动触发的 job，等待人工确认）。
func pipelineStatus(s string) biz.Status {
	switch s {
	case "failed":
		return biz.StatusFailed
	case "running":
		return biz.StatusRunning
	case "success":
		return biz.StatusSuccess
	case "manual":
		return biz.StatusManual
	default:
		return biz.StatusUnknown
	}
}

// toPipeline 将 go-gitlab 的 PipelineInfo 转成业务层 Pipeline；nil 输入返回 nil。
func toPipeline(p *gitlab.PipelineInfo) *biz.Pipeline {
	if p == nil {
		return nil
	}
	return &biz.Pipeline{
		ID:        int64(p.ID),
		ProjectID: int64(p.ProjectID),
		Status:    pipelineStatus(p.Status),
		Ref:       p.Ref,
		SHA:       p.SHA,
		WebURL:    p.WebURL,
		UpdatedAt: p.UpdatedAt,
		CreatedAt: p.CreatedAt,
	}
}

// server 是 gitlab 插件实现：持有 go-gitlab 客户端与日志器，实现 app.GitServer。
type server struct {
	client *gitlab.Client
	logger mlog.Logger
}

// Name 返回插件名 gitlab。
func (g *server) Name() string {
	return gitlabName
}

// Initialize 从 args 读取 token/baseurl/http_proxy，校验必填项后创建 go-gitlab 客户端。
func (g *server) Initialize(pluginApp app.PluginApp, args map[string]any) error {
	token, ok := args["token"].(string)
	if !ok || token == "" {
		return errors.New("gitlab: token required")
	}
	baseurl, ok := args["baseurl"].(string)
	if !ok || baseurl == "" {
		return errors.New("gitlab: baseurl required")
	}
	var proxyStr string
	if found, ok := args["http_proxy"]; ok {
		proxyStr, ok = found.(string)
		if !ok {
			return errors.New("gitlab: http_proxy must be string")
		}
	}
	client, err := gitlab.NewClient(
		token,
		gitlab.WithBaseURL(baseurl),
		gitlab.WithHTTPClient(proxy.NewHTTPProxyClient(proxyStr)),
	)
	if err != nil {
		return err
	}
	g.client = client
	g.logger = pluginApp.Logger()
	g.logger.Info("[Plugin]: " + g.Name() + " plugin Initialize...")
	return nil
}

// Destroy 输出销毁日志。
func (g *server) Destroy() error {
	g.logger.Info("[Plugin]: " + g.Name() + " plugin Destroy...")
	return nil
}

// GetProject 按项目 id 返回项目信息。
func (g *server) GetProject(pid string) (*biz.GitProject, error) {
	p, _, err := g.client.Projects.GetProject(pid, &gitlab.GetProjectOptions{})
	return toGitProject(p), err
}

// listProjects 是分页内部实现，供 AllProjects 迭代拉取全部项目。
func (g *server) listProjects(page, pageSize int) ([]*biz.GitProject, error) {
	res, _, err := g.client.Projects.ListProjects(&gitlab.ListProjectsOptions{
		MinAccessLevel: gitlab.AccessLevel(gitlab.DeveloperPermissions),
		ListOptions:    gitlab.ListOptions{PerPage: pageSize, Page: page},
	})
	if err != nil {
		return nil, err
	}
	projects := make([]*biz.GitProject, 0, len(res))
	for _, re := range res {
		projects = append(projects, toGitProject(re))
	}
	return projects, nil
}

// AllProjects 分页拉取当前 token 有权限的全部项目。
func (g *server) AllProjects() ([]*biz.GitProject, error) {
	var ps []*biz.GitProject
	page := 1
	pageSize := 100
	for page != -1 {
		projects, err := g.listProjects(page, pageSize)
		if err != nil {
			return nil, err
		}
		if len(projects) < pageSize {
			page = -1
		} else {
			page++
		}
		ps = append(ps, projects...)
	}

	return ps, nil
}

// listBranches 是分页内部实现，供 AllBranches 迭代拉取全部分支。
func (g *server) listBranches(pid string, page, pageSize int) ([]*biz.Branch, error) {
	gitlabBranches, _, e := g.client.Branches.ListBranches(pid, &gitlab.ListBranchesOptions{ListOptions: gitlab.ListOptions{PerPage: pageSize, Page: page}})
	if e != nil {
		return nil, e
	}
	branches := make([]*biz.Branch, 0, len(gitlabBranches))
	for _, gitlabBranch := range gitlabBranches {
		branches = append(branches, toBranch(gitlabBranch))
	}
	return branches, nil
}

// AllBranches 分页拉取项目全部分支，直至不足一页或取完。
func (g *server) AllBranches(pid string) ([]*biz.Branch, error) {
	var branches []*biz.Branch
	page := 1
	pageSize := 100
	for page != -1 {
		gitlabBranches, err := g.listBranches(pid, page, pageSize)
		if err != nil {
			return nil, err
		}
		if len(gitlabBranches) < pageSize {
			page = -1
		} else {
			page++
		}
		branches = append(branches, gitlabBranches...)
	}

	return branches, nil
}

// GetCommit 返回指定 sha 的提交信息。
func (g *server) GetCommit(pid string, sha string) (*biz.Commit, error) {
	c, _, err := g.client.Commits.GetCommit(pid, sha)
	if err != nil {
		return nil, err
	}
	return toCommit(c), nil
}

// ListCommits 返回指定分支最近的提交列表。
func (g *server) ListCommits(pid string, branch string) ([]*biz.Commit, error) {
	commits, _, err := g.client.Commits.ListCommits(pid, &gitlab.ListCommitsOptions{RefName: gitlab.String(branch), ListOptions: gitlab.ListOptions{PerPage: 100}})

	res := make([]*biz.Commit, 0, len(commits))
	for _, c := range commits {
		res = append(res, toCommit(c))
	}

	return res, err
}

// GetCommitPipeline 返回指定分支/提交对应的 push/web pipeline；没有则报错。
// 除流水线整体状态外，还会拉取 pipeline 的 job 列表返回每个 job 的名称与状态。
func (g *server) GetCommitPipeline(pid string, branch string, sha string) (*biz.Pipeline, error) {
	var p *gitlab.PipelineInfo
	pipelines, _, err := g.client.Pipelines.ListProjectPipelines(pid, &gitlab.ListProjectPipelinesOptions{
		ListOptions: gitlab.ListOptions{
			Page:    1,
			PerPage: 100,
		},
		Ref: gitlab.String(branch),
		SHA: gitlab.String(sha),
	})
	if err != nil {
		return nil, err
	}
	// 只拿 push/web 的 pipeline
	for _, info := range pipelines {
		if info.Source == "push" || info.Source == "web" {
			p = info
			break
		}
	}

	if p == nil {
		// 确定语义：指定分支/提交没有 push/web pipeline，属"资源不存在"而非系统故障。
		return nil, errs.NotFound("pipeline not found")
	}

	pipeline := toPipeline(p)
	pipeline.Jobs, err = g.pipelineJobs(pid, int(p.ID))
	if err != nil {
		return nil, err
	}
	return pipeline, nil
}

// PipelineJobOptions 返回项目最近一条 push/web pipeline 的 stage/job 去重选项，
// 供配置通过规则下拉选择；branch 为空时不按分支过滤（取项目最近 pipeline）。
// stage 按出现顺序、job 按执行顺序（job id 升序）去重。
func (g *server) PipelineJobOptions(pid string, branch string) (stages []string, jobs []string, err error) {
	opts := &gitlab.ListProjectPipelinesOptions{
		ListOptions: gitlab.ListOptions{Page: 1, PerPage: 100},
	}
	if branch != "" {
		opts.Ref = gitlab.String(branch)
	}
	pipelines, _, err := g.client.Pipelines.ListProjectPipelines(pid, opts)
	if err != nil {
		return nil, nil, err
	}
	var p *gitlab.PipelineInfo
	for _, info := range pipelines {
		if info.Source == "push" || info.Source == "web" {
			p = info
			break
		}
	}
	if p == nil {
		return nil, nil, errs.NotFound("pipeline not found")
	}
	jobList, err := g.listPipelineJobs(pid, int(p.ID))
	if err != nil {
		return nil, nil, err
	}
	stageSet := make(map[string]struct{})
	jobSet := make(map[string]struct{})
	for _, j := range jobList {
		if j.Stage != "" {
			if _, ok := stageSet[j.Stage]; !ok {
				stageSet[j.Stage] = struct{}{}
				stages = append(stages, j.Stage)
			}
		}
		if j.Name != "" {
			if _, ok := jobSet[j.Name]; !ok {
				jobSet[j.Name] = struct{}{}
				jobs = append(jobs, j.Name)
			}
		}
	}
	return stages, jobs, nil
}

// pipelineJobs 拉取 pipeline 的全部 job，按 job id 升序（即 stage 执行顺序）返回
// 名称/状态/所属 stage 列表；空名 job 被忽略。
func (g *server) pipelineJobs(pid string, pipelineID int) ([]biz.PipelineJob, error) {
	jobs, err := g.listPipelineJobs(pid, pipelineID)
	if err != nil {
		return nil, err
	}
	// 部分 GitLab 实例/低版本 SDK 的 pipeline 详情不返回 stages 数组，job 接口也不保证顺序；
	// job id 在 pipeline 创建时按 stage 声明顺序递增分配，故按 id 升序即可还原 stage 执行顺序，
	// 且无需依赖 started_at（手动/未运行的 job 也排在所属 stage 的位置）。
	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].ID < jobs[j].ID
	})
	out := make([]biz.PipelineJob, 0, len(jobs))
	for _, j := range jobs {
		if j.Name == "" {
			continue
		}
		out = append(out, biz.PipelineJob{Name: j.Name, Status: pipelineStatus(j.Status), StageName: j.Stage})
	}
	return out, nil
}

// listPipelineJobs 分页拉取 pipeline 的全部 job（每页 100，直至不足一页）。
func (g *server) listPipelineJobs(pid string, pipelineID int) ([]*gitlab.Job, error) {
	var jobs []*gitlab.Job
	page := 1
	for page != -1 {
		batch, _, err := g.client.Jobs.ListPipelineJobs(pid, pipelineID, &gitlab.ListJobsOptions{
			ListOptions: gitlab.ListOptions{Page: page, PerPage: 100},
		})
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, batch...)
		if len(batch) < 100 {
			page = -1
		} else {
			page++
		}
	}
	return jobs, nil
}

// getRawFile 按 ref 或分支名拉取仓库文件原始内容。
func getRawFile(client *gitlab.Client, pid string, shaOrBranch string, filename string) (string, error) {
	opt := gitlab.GetRawFileOptions{}
	if shaOrBranch != "" {
		opt.Ref = gitlab.String(shaOrBranch)
	}
	raw, _, err := client.RepositoryFiles.GetRawFile(pid, filename, &opt)
	return string(raw), err
}

// GetFileContentWithSha 按 commit sha 返回文件内容。
func (g *server) GetFileContentWithSha(pid string, sha string, filename string) (string, error) {
	return getRawFile(g.client, pid, sha, filename)
}

// GetFileContentWithBranch 按分支名返回文件内容。
func (g *server) GetFileContentWithBranch(pid string, branch string, filename string) (string, error) {
	return getRawFile(g.client, pid, branch, filename)
}

// getDirectoryFiles 分页遍历仓库目录树，收集所有 blob 类型文件路径。
func getDirectoryFiles(g *gitlab.Client, pid string, commit string, path string, recursive bool) ([]string, error) {
	var files []string

	opt := &gitlab.ListTreeOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 100,
			Page:    1,
		},
		Path:      gitlab.String(path),
		Recursive: gitlab.Bool(recursive),
	}
	if commit != "" {
		opt.Ref = gitlab.String(commit)
	}

	for opt.Page != -1 {
		tree, _, err := g.Repositories.ListTree(pid, opt)
		if err != nil {
			return nil, err
		}
		if len(tree) != opt.PerPage {
			opt.Page = -1
		} else {
			opt.Page++
		}
		for _, node := range tree {
			if node.Type == "blob" {
				files = append(files, node.Path)
			}
		}
	}

	return files, nil
}

// GetDirectoryFilesWithBranch 按分支名返回目录下文件路径列表。
func (g *server) GetDirectoryFilesWithBranch(pid string, branch string, path string, recursive bool) ([]string, error) {
	return getDirectoryFiles(g.client, pid, branch, path, recursive)
}

// GetDirectoryFilesWithSha 按 commit sha 返回目录下文件路径列表。
func (g *server) GetDirectoryFilesWithSha(pid string, sha string, path string, recursive bool) ([]string, error) {
	return getDirectoryFiles(g.client, pid, sha, path, recursive)
}
