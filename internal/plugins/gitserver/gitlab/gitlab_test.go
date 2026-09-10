package gitlab

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/duc-cnzj/mars/v6/internal/app"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xanzy/go-gitlab"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// gitApp 是 PluginApp 的最小手写 stub。
type gitApp struct {
	logger mlog.Logger
}

func (g gitApp) Logger() mlog.Logger          { return g.logger }
func (g gitApp) ProjectRepo() biz.ProjectRepo { return nil }

// apiHandler 过滤 go-gitlab 首次 Do 时对 /api/v4/ 发送的速率限制探测请求，
// 确保断言只看到真实的业务 API 调用。
func apiHandler(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v4/" {
			w.WriteHeader(http.StatusOK)
			return
		}
		h(w, r)
	}
}

// newServer 通过 Initialize 构造指向假 GitLab API 的 server 实例。
func newServer(t *testing.T, baseURL string) *server {
	t.Helper()
	s := &server{}
	err := s.Initialize(gitApp{logger: mlog.NewForConfig(nil)}, map[string]any{
		"token":   "test-token",
		"baseurl": baseURL,
	})
	require.NoError(t, err)
	return s
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

// jobsJSON 构造 n 个 id 自 startID 起递增、同一 name/status/stage 的 job，用于分页等批量场景。
func jobsJSON(n, startID int, name, status, stage string) []map[string]any {
	out := make([]map[string]any, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, map[string]any{"id": startID + i, "name": name, "status": status, "stage": stage})
	}
	return out
}

// ---------------------------------------------------------------------------
// lifecycle
// ---------------------------------------------------------------------------

func TestGitlabName(t *testing.T) {
	assert.Equal(t, "gitlab", (&server{}).Name())
}

func TestGitlabInitialize_valid(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	t.Cleanup(srv.Close)

	s := &server{}
	err := s.Initialize(gitApp{logger: mlog.NewForConfig(nil)}, map[string]any{
		"token":      "tok",
		"baseurl":    srv.URL,
		"http_proxy": "",
	})
	require.NoError(t, err)
	assert.NotNil(t, s.client)
}

func TestGitlabInitialize_missing_token(t *testing.T) {
	s := &server{}
	err := s.Initialize(gitApp{logger: mlog.NewForConfig(nil)}, map[string]any{"baseurl": "http://x"})
	assert.ErrorContains(t, err, "token required")
}

func TestGitlabInitialize_missing_baseurl(t *testing.T) {
	s := &server{}
	err := s.Initialize(gitApp{logger: mlog.NewForConfig(nil)}, map[string]any{"token": "tok"})
	assert.ErrorContains(t, err, "baseurl required")
}

func TestGitlabInitialize_bad_proxy_type(t *testing.T) {
	s := &server{}
	err := s.Initialize(gitApp{logger: mlog.NewForConfig(nil)}, map[string]any{
		"token":      "tok",
		"baseurl":    "http://x",
		"http_proxy": 123,
	})
	assert.ErrorContains(t, err, "http_proxy must be string")
}

func TestGitlabInitialize_invalid_baseurl(t *testing.T) {
	s := &server{}
	err := s.Initialize(gitApp{logger: mlog.NewForConfig(nil)}, map[string]any{
		"token":   "tok",
		"baseurl": "http://%zz", // 非法 URL，gitlab.NewClient 构造失败
	})
	assert.Error(t, err)
}

func TestGitlabDestroy(t *testing.T) {
	s := &server{logger: mlog.NewForConfig(nil)}
	assert.NoError(t, s.Destroy())
}

// ---------------------------------------------------------------------------
// GetProject
// ---------------------------------------------------------------------------

func TestGetProject_success(t *testing.T) {
	srv := httptest.NewServer(apiHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v4/projects/7", r.URL.Path)
		writeJSON(w, map[string]any{
			"id": 7, "name": "my-app", "default_branch": "main",
			"web_url": "https://gitlab.com/group/my-app", "path": "my-app",
			"avatar_url": "https://avatar", "description": "desc",
		})
	})))
	t.Cleanup(srv.Close)

	g := newServer(t, srv.URL)
	proj, err := g.GetProject("7")
	require.NoError(t, err)
	assert.Equal(t, int64(7), proj.ID)
	assert.Equal(t, "my-app", proj.Name)
	assert.Equal(t, "main", proj.DefaultBranch)
	assert.Equal(t, "desc", proj.Description)
}

func TestGetProject_error(t *testing.T) {
	srv := httptest.NewServer(apiHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		writeJSON(w, map[string]any{"message": "404 Not Found"})
	})))
	t.Cleanup(srv.Close)

	g := newServer(t, srv.URL)
	_, err := g.GetProject("999")
	require.Error(t, err)
	// gitlab 项目不存在是资源缺失而非系统故障，经 classifyGitlabError 映射为 404，
	// data 层 errs.Wrap 保留该状态码，避免客户端把"项目不存在"误判成 500。
	assert.Equal(t, codes.NotFound, status.Code(err))
}

// TestClassifyGitlabError 覆盖 gitlab API 错误按 HTTP 状态码归类的全部分支：
// 404→NotFound、400→InvalidArgument、401→Unauthenticated、403→PermissionDenied，
// 5xx 与非 *gitlab.ErrorResponse 原样透传（data 层 errs.Wrap 落 500）、nil 返回 nil。
func TestClassifyGitlabError(t *testing.T) {
	t.Run("nil input", func(t *testing.T) {
		assert.Nil(t, classifyGitlabError(nil))
	})

	t.Run("non gitlab error passthrough", func(t *testing.T) {
		plain := errors.New("connection refused")
		assert.Same(t, plain, classifyGitlabError(plain))
	})

	testCases := []struct {
		name   string
		code   int
		expect codes.Code
	}{
		{"not found", http.StatusNotFound, codes.NotFound},
		{"bad request", http.StatusBadRequest, codes.InvalidArgument},
		{"unauthorized", http.StatusUnauthorized, codes.Unauthenticated},
		{"forbidden", http.StatusForbidden, codes.PermissionDenied},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// ErrorResponse.Error() 会解引用 Response.Request.URL.Path，
			// 测试需与真实 go-gitlab 一致地补上 Request，否则 status.Convert 构造消息时 panic。
			respErr := &gitlab.ErrorResponse{
				Response: &http.Response{
					StatusCode: tc.code,
					Request:    &http.Request{URL: &url.URL{Path: "/api/v4/projects/999"}},
				},
				Message: "boom",
			}
			got := classifyGitlabError(respErr)
			assert.Equal(t, tc.expect, status.Code(got))
			assert.ErrorIs(t, got, respErr)
		})
	}

	t.Run("server error stays passthrough", func(t *testing.T) {
		respErr := &gitlab.ErrorResponse{Response: &http.Response{StatusCode: http.StatusInternalServerError}, Message: "boom"}
		assert.Same(t, respErr, classifyGitlabError(respErr))
	})

	t.Run("error response without response stays passthrough", func(t *testing.T) {
		respErr := &gitlab.ErrorResponse{Message: "no response"}
		assert.Same(t, respErr, classifyGitlabError(respErr))
	})
}

// ---------------------------------------------------------------------------
// AllProjects / AllBranches pagination
// ---------------------------------------------------------------------------

func TestAllProjects_pagination(t *testing.T) {
	var pageCalls []string
	srv := httptest.NewServer(apiHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		pageCalls = append(pageCalls, page)
		if page == "1" {
			projects := make([]map[string]any, 100) // 满页 → 继续下一页
			for i := range projects {
				projects[i] = map[string]any{"id": i, "name": fmt.Sprintf("p%d", i)}
			}
			writeJSON(w, projects)
			return
		}
		writeJSON(w, []map[string]any{{"id": 100, "name": "last"}}) // 不满页 → 终止
	})))
	t.Cleanup(srv.Close)

	g := newServer(t, srv.URL)
	projects, err := g.AllProjects()
	require.NoError(t, err)
	assert.Len(t, projects, 101)
	assert.Equal(t, []string{"1", "2"}, pageCalls)
}

func TestAllProjects_error(t *testing.T) {
	srv := httptest.NewServer(apiHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		writeJSON(w, map[string]any{"message": "forbidden"})
	})))
	t.Cleanup(srv.Close)

	g := newServer(t, srv.URL)
	_, err := g.AllProjects()
	assert.Error(t, err)
}

func TestAllBranches_error(t *testing.T) {
	srv := httptest.NewServer(apiHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		writeJSON(w, map[string]any{"message": "forbidden"})
	})))
	t.Cleanup(srv.Close)

	g := newServer(t, srv.URL)
	_, err := g.AllBranches("1")
	assert.Error(t, err)
}

func TestAllBranches_pagination(t *testing.T) {
	var pageCalls []string
	srv := httptest.NewServer(apiHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		pageCalls = append(pageCalls, page)
		if page == "1" {
			branches := make([]map[string]any, 100)
			for i := range branches {
				branches[i] = map[string]any{"name": fmt.Sprintf("b%d", i), "default": false}
			}
			writeJSON(w, branches)
			return
		}
		writeJSON(w, []map[string]any{{"name": "dev", "default": false}})
	})))
	t.Cleanup(srv.Close)

	g := newServer(t, srv.URL)
	branches, err := g.AllBranches("1")
	require.NoError(t, err)
	assert.Len(t, branches, 101)
	assert.Equal(t, []string{"1", "2"}, pageCalls)
}

// ---------------------------------------------------------------------------
// GetCommit / ListCommits
// ---------------------------------------------------------------------------

func TestGetCommit_error(t *testing.T) {
	srv := httptest.NewServer(apiHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		writeJSON(w, map[string]any{"message": "404"})
	})))
	t.Cleanup(srv.Close)

	g := newServer(t, srv.URL)
	_, err := g.GetCommit("1", "abc")
	assert.Error(t, err)
}

func TestGetCommit_success(t *testing.T) {
	srv := httptest.NewServer(apiHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v4/projects/1/repository/commits/abc123", r.URL.Path)
		writeJSON(w, map[string]any{
			"id": "abc123", "short_id": "abc", "title": "fix", "message": "fix bug",
			"author_name": "a", "author_email": "a@b.c",
			"committer_name": "c", "committer_email": "c@b.c",
			"web_url": "https://gitlab.com/1/-/commit/abc123",
		})
	})))
	t.Cleanup(srv.Close)

	g := newServer(t, srv.URL)
	c, err := g.GetCommit("1", "abc123")
	require.NoError(t, err)
	assert.Equal(t, "abc123", c.ID)
	assert.Equal(t, "fix", c.Title)
	assert.Equal(t, "a@b.c", c.AuthorEmail)
}

func TestListCommits_success(t *testing.T) {
	srv := httptest.NewServer(apiHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v4/projects/1/repository/commits", r.URL.Path)
		writeJSON(w, []map[string]any{
			{"id": "c1", "short_id": "c1", "title": "one"},
			{"id": "c2", "short_id": "c2", "title": "two"},
		})
	})))
	t.Cleanup(srv.Close)

	g := newServer(t, srv.URL)
	commits, err := g.ListCommits("1", "main")
	require.NoError(t, err)
	assert.Len(t, commits, 2)
	assert.Equal(t, "one", commits[0].Title)
}

func TestListCommits_error_returns_empty(t *testing.T) {
	srv := httptest.NewServer(apiHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})))
	t.Cleanup(srv.Close)

	g := newServer(t, srv.URL)
	commits, err := g.ListCommits("1", "main")
	assert.Error(t, err)
	assert.Empty(t, commits)
}

// ---------------------------------------------------------------------------
// GetCommitPipeline
// ---------------------------------------------------------------------------

func TestGetCommitPipeline_success(t *testing.T) {
	srv := httptest.NewServer(apiHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v4/projects/1/pipelines":
			writeJSON(w, []map[string]any{
				{"id": 11, "project_id": 1, "status": "success", "source": "push", "ref": "main", "sha": "s"},
			})
		case "/api/v4/projects/1/pipelines/11/jobs":
			writeJSON(w, []map[string]any{
				{"id": 1, "name": "build", "status": "success", "stage": "compile"},
			})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	})))
	t.Cleanup(srv.Close)

	g := newServer(t, srv.URL)
	p, err := g.GetCommitPipeline("1", "main", "s")
	require.NoError(t, err)
	assert.Equal(t, int64(11), p.ID)
	assert.Equal(t, biz.StatusSuccess, p.Status)
	require.Len(t, p.Jobs, 1)
	assert.Equal(t, "build", p.Jobs[0].Name)
	assert.Equal(t, biz.StatusSuccess, p.Jobs[0].Status)
	assert.Equal(t, "compile", p.Jobs[0].StageName)
}

func TestGetCommitPipeline_skips_other_sources(t *testing.T) {
	srv := httptest.NewServer(apiHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v4/projects/1/pipelines":
			writeJSON(w, []map[string]any{
				{"id": 1, "source": "schedule", "status": "failed"}, // 跳过
				{"id": 2, "source": "web", "status": "running"},     // 命中
			})
		case "/api/v4/projects/1/pipelines/2/jobs":
			writeJSON(w, []map[string]any{
				{"id": 1, "name": "build", "status": "running", "stage": "compile"},
			})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	})))
	t.Cleanup(srv.Close)

	g := newServer(t, srv.URL)
	p, err := g.GetCommitPipeline("1", "main", "s")
	require.NoError(t, err)
	assert.Equal(t, int64(2), p.ID)
	assert.Equal(t, biz.StatusRunning, p.Status)
	require.Len(t, p.Jobs, 1)
	assert.Equal(t, "compile", p.Jobs[0].StageName)
}

func TestGetCommitPipeline_error(t *testing.T) {
	srv := httptest.NewServer(apiHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		writeJSON(w, map[string]any{"message": "forbidden"})
	})))
	t.Cleanup(srv.Close)

	g := newServer(t, srv.URL)
	_, err := g.GetCommitPipeline("1", "main", "s")
	assert.Error(t, err)
}

func TestGetCommitPipeline_not_found(t *testing.T) {
	srv := httptest.NewServer(apiHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, []map[string]any{{"id": 1, "source": "schedule", "status": "failed"}})
	})))
	t.Cleanup(srv.Close)

	g := newServer(t, srv.URL)
	_, err := g.GetCommitPipeline("1", "main", "s")
	assert.ErrorContains(t, err, "pipeline not found")
	// 无 pipeline 属"资源不存在"而非系统故障，须映射 404 而非 500。
	assert.Equal(t, codes.NotFound, status.Code(err))
}

// TestGetCommitPipeline_jobs_mixed 验证每个 job 返回自己的名称与状态，空名 job 被忽略。
func TestGetCommitPipeline_jobs_mixed(t *testing.T) {
	srv := httptest.NewServer(apiHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v4/projects/1/pipelines":
			writeJSON(w, []map[string]any{
				{"id": 11, "project_id": 1, "status": "running", "source": "web", "ref": "main", "sha": "s"},
			})
		case "/api/v4/projects/1/pipelines/11/jobs":
			writeJSON(w, []map[string]any{
				{"id": 1, "name": "compile", "status": "success", "stage": "build"},
				{"id": 2, "name": "lint", "status": "failed", "stage": "build"},
				{"id": 3, "name": "unit-test", "status": "running", "stage": "test"},
				{"id": 4, "name": "deploy", "status": "skipped", "stage": "deploy"},
				{"id": 5, "name": "", "status": "success", "stage": "deploy"},
			})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	})))
	t.Cleanup(srv.Close)

	g := newServer(t, srv.URL)
	p, err := g.GetCommitPipeline("1", "main", "s")
	require.NoError(t, err)
	require.Len(t, p.Jobs, 4)
	assert.Equal(t, "compile", p.Jobs[0].Name)
	assert.Equal(t, biz.StatusSuccess, p.Jobs[0].Status)
	assert.Equal(t, "build", p.Jobs[0].StageName)
	assert.Equal(t, "lint", p.Jobs[1].Name)
	assert.Equal(t, biz.StatusFailed, p.Jobs[1].Status)
	assert.Equal(t, "unit-test", p.Jobs[2].Name)
	assert.Equal(t, biz.StatusRunning, p.Jobs[2].Status)
	assert.Equal(t, "test", p.Jobs[2].StageName)
	assert.Equal(t, "deploy", p.Jobs[3].Name)
	assert.Equal(t, biz.StatusFailed, p.Jobs[3].Status) // skipped → failed（被跳过当失败处理）
	assert.Equal(t, "deploy", p.Jobs[3].StageName)
}

// TestGetCommitPipeline_jobs_error 验证拉 job 失败时整个请求返回错误（fail-fast）。
func TestGetCommitPipeline_jobs_error(t *testing.T) {
	srv := httptest.NewServer(apiHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v4/projects/1/pipelines":
			writeJSON(w, []map[string]any{
				{"id": 11, "project_id": 1, "status": "success", "source": "push", "ref": "main", "sha": "s"},
			})
		default:
			w.WriteHeader(http.StatusInternalServerError)
			writeJSON(w, map[string]any{"message": "jobs boom"})
		}
	})))
	t.Cleanup(srv.Close)

	g := newServer(t, srv.URL)
	_, err := g.GetCommitPipeline("1", "main", "s")
	assert.Error(t, err)
}

// TestGetCommitPipeline_jobs_pagination 验证 job 超过一页时分页拉全。
func TestGetCommitPipeline_jobs_pagination(t *testing.T) {
	srv := httptest.NewServer(apiHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v4/projects/1/pipelines":
			writeJSON(w, []map[string]any{
				{"id": 11, "project_id": 1, "status": "success", "source": "push", "ref": "main", "sha": "s"},
			})
		case "/api/v4/projects/1/pipelines/11/jobs":
			if r.URL.Query().Get("page") == "2" {
				writeJSON(w, jobsJSON(2, 101, "test-2", "success", "test"))
			} else {
				writeJSON(w, jobsJSON(100, 1, "build-1", "success", "build"))
			}
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	})))
	t.Cleanup(srv.Close)

	g := newServer(t, srv.URL)
	p, err := g.GetCommitPipeline("1", "main", "s")
	require.NoError(t, err)
	require.Len(t, p.Jobs, 102)
	assert.Equal(t, "build-1", p.Jobs[0].Name)
	assert.Equal(t, "build", p.Jobs[0].StageName)
	assert.Equal(t, "test-2", p.Jobs[101].Name)
	assert.Equal(t, "test", p.Jobs[101].StageName)
}

// TestGetCommitPipeline_jobs_sorted_by_id 验证 job 按 id 升序（stage 执行顺序）返回：
// 低版本 GitLab 的 pipeline 详情无 stages 数组、job 接口也不保证顺序，
// 但 job id 按 stage 声明顺序分配，故按 id 手动排序即可还原执行顺序。
func TestGetCommitPipeline_jobs_sorted_by_id(t *testing.T) {
	srv := httptest.NewServer(apiHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v4/projects/1/pipelines":
			writeJSON(w, []map[string]any{
				{"id": 11, "project_id": 1, "status": "success", "source": "push", "ref": "main", "sha": "s"},
			})
		case "/api/v4/projects/1/pipelines/11/jobs":
			// 故意按 id 倒序返回，模拟低版本 GitLab 的 job 原始顺序。
			writeJSON(w, []map[string]any{
				{"id": 5, "name": "deploy", "status": "success", "stage": "deploy"},
				{"id": 1, "name": "build", "status": "success", "stage": "build"},
				{"id": 3, "name": "lint", "status": "success", "stage": "test"},
				{"id": 2, "name": "unit", "status": "success", "stage": "test"},
				{"id": 4, "name": "manual", "status": "skipped", "stage": "deploy"},
			})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	})))
	t.Cleanup(srv.Close)

	g := newServer(t, srv.URL)
	p, err := g.GetCommitPipeline("1", "main", "s")
	require.NoError(t, err)
	require.Len(t, p.Jobs, 5)
	assert.Equal(t, "build", p.Jobs[0].Name)
	assert.Equal(t, "unit", p.Jobs[1].Name)
	assert.Equal(t, "lint", p.Jobs[2].Name)
	assert.Equal(t, "manual", p.Jobs[3].Name, "手动/未运行 job 按所属 stage 位置排列")
	assert.Equal(t, "deploy", p.Jobs[4].Name)
}

// ---------------------------------------------------------------------------
// PipelineJobOptions
// ---------------------------------------------------------------------------

// TestPipelineJobOptions_success_with_branch 验证传 branch 时带 ref 过滤，返回去重后的 stage/job。
func TestPipelineJobOptions_success_with_branch(t *testing.T) {
	srv := httptest.NewServer(apiHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v4/projects/1/pipelines":
			assert.Equal(t, "main", r.URL.Query().Get("ref"), "传 branch 时应带 ref 过滤")
			writeJSON(w, []map[string]any{
				{"id": 11, "project_id": 1, "status": "success", "source": "push", "ref": "main"},
			})
		case "/api/v4/projects/1/pipelines/11/jobs":
			writeJSON(w, []map[string]any{
				{"id": 1, "name": "compile", "status": "success", "stage": "build"},
				{"id": 2, "name": "deploy", "status": "success", "stage": "deploy"},
			})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	})))
	t.Cleanup(srv.Close)

	g := newServer(t, srv.URL)
	stages, jobs, err := g.PipelineJobOptions("1", "main")
	require.NoError(t, err)
	assert.Equal(t, []string{"build", "deploy"}, stages)
	assert.Equal(t, []string{"compile", "deploy"}, jobs)
}

// TestPipelineJobOptions_success_without_branch 验证 branch 为空时不带 ref 过滤（取项目最近 pipeline）。
func TestPipelineJobOptions_success_without_branch(t *testing.T) {
	srv := httptest.NewServer(apiHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v4/projects/1/pipelines":
			assert.Empty(t, r.URL.Query().Get("ref"), "branch 为空时不应带 ref 过滤")
			writeJSON(w, []map[string]any{
				{"id": 11, "project_id": 1, "status": "success", "source": "push"},
			})
		case "/api/v4/projects/1/pipelines/11/jobs":
			writeJSON(w, []map[string]any{
				{"id": 1, "name": "build", "status": "success", "stage": "compile"},
			})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	})))
	t.Cleanup(srv.Close)

	g := newServer(t, srv.URL)
	stages, jobs, err := g.PipelineJobOptions("1", "")
	require.NoError(t, err)
	assert.Equal(t, []string{"compile"}, stages)
	assert.Equal(t, []string{"build"}, jobs)
}

// TestPipelineJobOptions_dedup 验证重复的 stage/job 名按出现顺序去重。
func TestPipelineJobOptions_dedup(t *testing.T) {
	srv := httptest.NewServer(apiHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v4/projects/1/pipelines":
			writeJSON(w, []map[string]any{
				{"id": 11, "project_id": 1, "status": "success", "source": "web"},
			})
		case "/api/v4/projects/1/pipelines/11/jobs":
			writeJSON(w, []map[string]any{
				{"id": 1, "name": "build", "status": "success", "stage": "build"},
				{"id": 2, "name": "build", "status": "success", "stage": "build"},
				{"id": 3, "name": "lint", "status": "success", "stage": "build"},
				{"id": 4, "name": "", "status": "success", "stage": "test"},
				{"id": 5, "name": "deploy", "status": "success", "stage": ""},
			})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	})))
	t.Cleanup(srv.Close)

	g := newServer(t, srv.URL)
	stages, jobs, err := g.PipelineJobOptions("1", "main")
	require.NoError(t, err)
	// job4 空 name 不影响 stage；job5 空 stage 不影响 name。
	assert.Equal(t, []string{"build", "test"}, stages, "重复 stage 去重，空 stage 忽略")
	assert.Equal(t, []string{"build", "lint", "deploy"}, jobs, "重复 job 去重，空 job 忽略")
}

// TestPipelineJobOptions_skips_other_sources 验证 schedule 等非 push/web pipeline 被跳过。
func TestPipelineJobOptions_skips_other_sources(t *testing.T) {
	srv := httptest.NewServer(apiHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v4/projects/1/pipelines":
			writeJSON(w, []map[string]any{
				{"id": 1, "source": "schedule", "status": "failed"},
				{"id": 2, "source": "web", "status": "success"},
			})
		case "/api/v4/projects/1/pipelines/2/jobs":
			writeJSON(w, []map[string]any{
				{"id": 1, "name": "build", "status": "success", "stage": "compile"},
			})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	})))
	t.Cleanup(srv.Close)

	g := newServer(t, srv.URL)
	stages, jobs, err := g.PipelineJobOptions("1", "main")
	require.NoError(t, err)
	assert.Equal(t, []string{"compile"}, stages)
	assert.Equal(t, []string{"build"}, jobs)
}

func TestPipelineJobOptions_not_found(t *testing.T) {
	srv := httptest.NewServer(apiHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, []map[string]any{{"id": 1, "source": "schedule", "status": "failed"}})
	})))
	t.Cleanup(srv.Close)

	g := newServer(t, srv.URL)
	_, _, err := g.PipelineJobOptions("1", "main")
	assert.ErrorContains(t, err, "pipeline not found")
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestPipelineJobOptions_pipelines_error(t *testing.T) {
	srv := httptest.NewServer(apiHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		writeJSON(w, map[string]any{"message": "forbidden"})
	})))
	t.Cleanup(srv.Close)

	g := newServer(t, srv.URL)
	_, _, err := g.PipelineJobOptions("1", "main")
	assert.Error(t, err)
}

func TestPipelineJobOptions_jobs_error(t *testing.T) {
	srv := httptest.NewServer(apiHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v4/projects/1/pipelines":
			writeJSON(w, []map[string]any{
				{"id": 11, "project_id": 1, "status": "success", "source": "push"},
			})
		case "/api/v4/projects/1/pipelines/11/jobs":
			w.WriteHeader(http.StatusForbidden)
			writeJSON(w, map[string]any{"message": "forbidden"})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	})))
	t.Cleanup(srv.Close)

	g := newServer(t, srv.URL)
	_, _, err := g.PipelineJobOptions("1", "main")
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// file content
// ---------------------------------------------------------------------------

func TestGetFileContentWithSha_and_Branch(t *testing.T) {
	srv := httptest.NewServer(apiHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/repository/files/README.md/raw")
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("# readme"))
	})))
	t.Cleanup(srv.Close)

	g := newServer(t, srv.URL)
	bySha, err := g.GetFileContentWithSha("1", "sha1", "README.md")
	require.NoError(t, err)
	assert.Equal(t, "# readme", bySha)

	byBranch, err := g.GetFileContentWithBranch("1", "main", "README.md")
	require.NoError(t, err)
	assert.Equal(t, "# readme", byBranch)
}

func TestGetRawFile_empty_ref(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("no-ref"))
	}))
	t.Cleanup(srv.Close)

	g := newServer(t, srv.URL)
	// shaOrBranch 为空 → 不设置 Ref，直接调用包内函数覆盖该分支。
	content, err := getRawFile(g.client, "1", "", "README.md")
	require.NoError(t, err)
	assert.Equal(t, "no-ref", content)
}

// ---------------------------------------------------------------------------
// directory files
// ---------------------------------------------------------------------------

func TestGetDirectoryFiles_filters_blobs_and_paginates(t *testing.T) {
	var pageCalls []string
	srv := httptest.NewServer(apiHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		pageCalls = append(pageCalls, page)
		if page == "1" {
			tree := make([]map[string]any, 100)
			for i := range tree {
				if i == 0 {
					tree[i] = map[string]any{"id": "d", "type": "tree", "path": "sub/"} // 非 blob，过滤
					continue
				}
				tree[i] = map[string]any{"id": fmt.Sprintf("f%d", i), "type": "blob", "path": fmt.Sprintf("file%d.go", i)}
			}
			writeJSON(w, tree)
			return
		}
		writeJSON(w, []map[string]any{
			{"id": "last", "type": "blob", "path": "last.go"},
			{"id": "dir", "type": "tree", "path": "dir/"},
		})
	})))
	t.Cleanup(srv.Close)

	g := newServer(t, srv.URL)
	files, err := g.GetDirectoryFilesWithBranch("1", "main", "src", true)
	require.NoError(t, err)
	// 第一页 99 个 blob + 第二页 1 个 blob = 100
	assert.Len(t, files, 100)
	assert.Equal(t, []string{"1", "2"}, pageCalls)
}

func TestGetDirectoryFiles_error(t *testing.T) {
	srv := httptest.NewServer(apiHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		writeJSON(w, map[string]any{"message": "forbidden"})
	})))
	t.Cleanup(srv.Close)

	g := newServer(t, srv.URL)
	_, err := g.GetDirectoryFilesWithBranch("1", "main", "src", false)
	assert.Error(t, err)
}

func TestGetDirectoryFilesWithSha(t *testing.T) {
	srv := httptest.NewServer(apiHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "1", r.URL.Query().Get("page"))
		writeJSON(w, []map[string]any{{"id": "f", "type": "blob", "path": "a.go"}})
	})))
	t.Cleanup(srv.Close)

	g := newServer(t, srv.URL)
	files, err := g.GetDirectoryFilesWithSha("1", "sha", "dir", false)
	require.NoError(t, err)
	assert.Equal(t, []string{"a.go"}, files)
}

// ---------------------------------------------------------------------------
// pure mapping functions
// ---------------------------------------------------------------------------

func TestToGitProject_nil(t *testing.T) {
	assert.Nil(t, toGitProject(nil))
}

func TestToBranch_nil(t *testing.T) {
	assert.Nil(t, toBranch(nil))
}

func TestToCommit_nil(t *testing.T) {
	assert.Nil(t, toCommit(nil))
}

func TestToPipeline_nil(t *testing.T) {
	assert.Nil(t, toPipeline(nil))
}

func TestToGitProject_maps_fields(t *testing.T) {
	p := toGitProject(&gitlab.Project{ID: 5, Name: "n", DefaultBranch: "main", WebURL: "w", Path: "p", AvatarURL: "a", Description: "d"})
	require.NotNil(t, p)
	assert.Equal(t, int64(5), p.ID)
	assert.Equal(t, "n", p.Name)
	assert.Equal(t, "main", p.DefaultBranch)
	assert.Equal(t, "w", p.WebURL)
	assert.Equal(t, "p", p.Path)
	assert.Equal(t, "a", p.AvatarURL)
	assert.Equal(t, "d", p.Description)
}

func TestToBranch_maps_fields(t *testing.T) {
	b := toBranch(&gitlab.Branch{Name: "dev", Default: true, WebURL: "w"})
	require.NotNil(t, b)
	assert.Equal(t, "dev", b.Name)
	assert.True(t, b.IsDefault)
	assert.Equal(t, "w", b.WebURL)
}

func TestToCommit_maps_fields(t *testing.T) {
	c := toCommit(&gitlab.Commit{ID: "id", ShortID: "sid", Title: "t", Message: "m", WebURL: "w", AuthorName: "an", AuthorEmail: "ae", CommitterName: "cn", CommitterEmail: "ce"})
	require.NotNil(t, c)
	assert.Equal(t, "id", c.ID)
	assert.Equal(t, "sid", c.ShortID)
	assert.Equal(t, "t", c.Title)
	assert.Equal(t, "m", c.Message)
	assert.Equal(t, "w", c.WebURL)
	assert.Equal(t, "an", c.AuthorName)
	assert.Equal(t, "ae", c.AuthorEmail)
	assert.Equal(t, "cn", c.CommitterName)
	assert.Equal(t, "ce", c.CommitterEmail)
}

func TestPipelineStatus_mapping(t *testing.T) {
	assert.Equal(t, biz.StatusFailed, pipelineStatus("failed"))
	assert.Equal(t, biz.StatusFailed, pipelineStatus("skipped"))
	assert.Equal(t, biz.StatusRunning, pipelineStatus("running"))
	assert.Equal(t, biz.StatusSuccess, pipelineStatus("success"))
	assert.Equal(t, biz.StatusManual, pipelineStatus("manual"))
	assert.Equal(t, biz.StatusUnknown, pipelineStatus("canceled"))
	assert.Equal(t, biz.StatusUnknown, pipelineStatus(""))
}

func TestToPipeline_maps_fields(t *testing.T) {
	p := toPipeline(&gitlab.PipelineInfo{ID: 9, ProjectID: 1, Status: "success", Ref: "main", SHA: "s", WebURL: "w"})
	require.NotNil(t, p)
	assert.Equal(t, int64(9), p.ID)
	assert.Equal(t, int64(1), p.ProjectID)
	assert.Equal(t, biz.StatusSuccess, p.Status)
	assert.Equal(t, "main", p.Ref)
	assert.Equal(t, "s", p.SHA)
	assert.Equal(t, "w", p.WebURL)
}

// TestRegister_interface ensures the plugin satisfies app.GitServer.
func TestRegister_interface(t *testing.T) {
	var _ app.GitServer = (*server)(nil)
}
