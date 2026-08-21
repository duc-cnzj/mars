package cronjob

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/biz/schematype"
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/data"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/uploader"
	"github.com/duc-cnzj/mars/v6/internal/util/timer"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// TestTasks_SyncDomainSecret_NonNotFoundErrorNoPanic 回归防护：GetSecret 返回
// 非 NotFound 错误（网络/权限）时不得 nil-deref panic，也不得误走 AddTlsSecret 分支，
// 只记录后跳过该 namespace。同步逻辑全部走端口，不触碰基础设施门面。
func TestTasks_SyncDomainSecret_NonNotFoundErrorNoPanic(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	nsRepo := data.NewMockNamespaceRepo(m)
	nsRepo.EXPECT().ListAll(context.TODO()).Return([]*biz.Namespace{{ID: 1, Name: "ns1"}}, nil)
	k8sRepo := data.NewMockK8sRepo(m)
	k8sRepo.EXPECT().GetSecret(gomock.Any(), "ns1", "secret-name").Return(nil, errors.New("permission denied"))
	// 无 AddTlsSecret EXPECT：若被误调会 panic，反向证明未走创建分支。

	repo := &Tasks{
		logger:   mlog.NewForConfig(nil),
		getCerts: func() (string, string, string) { return "secret-name", "tls-key", "tls-crt" },
		nsRepo:   nsRepo,
		k8sRepo:  k8sRepo,
	}
	assert.NotPanics(t, func() {
		assert.NoError(t, repo.SyncDomainSecret())
	})
}

// TestTasks_SyncDomainSecret_NotFoundRegistersTls 覆盖 TLS 证书缺失分支：
// secret 不存在时经 AddTlsSecret 端口创建，不走门面。
func TestTasks_SyncDomainSecret_NotFoundRegistersTls(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	nsRepo := data.NewMockNamespaceRepo(m)
	nsRepo.EXPECT().ListAll(context.TODO()).Return([]*biz.Namespace{{ID: 1, Name: "ns1"}}, nil)
	k8sRepo := data.NewMockK8sRepo(m)
	k8sRepo.EXPECT().GetSecret(gomock.Any(), "ns1", "secret-name").Return(nil, apierrors.NewNotFound(schema.GroupResource{}, "secret-name"))
	k8sRepo.EXPECT().AddTlsSecret("ns1", "secret-name", "tls-key", "tls-crt").Return(nil, nil)

	repo := &Tasks{
		logger:   mlog.NewForConfig(nil),
		getCerts: func() (string, string, string) { return "secret-name", "tls-key", "tls-crt" },
		nsRepo:   nsRepo,
		k8sRepo:  k8sRepo,
	}
	assert.NoError(t, repo.SyncDomainSecret())
}

// TestTasks_SyncDomainSecret_CertChangedUpdates 覆盖证书内容不一致分支：
// 经 UpdateSecret 端口更新 secret 数据。
func TestTasks_SyncDomainSecret_CertChangedUpdates(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	nsRepo := data.NewMockNamespaceRepo(m)
	nsRepo.EXPECT().ListAll(context.TODO()).Return([]*biz.Namespace{{ID: 1, Name: "ns1"}}, nil)
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Namespace: "ns1", Name: "secret-name"},
		Data:       map[string][]byte{"tls.crt": []byte("old-crt"), "tls.key": []byte("old-key")},
	}
	k8sRepo := data.NewMockK8sRepo(m)
	k8sRepo.EXPECT().GetSecret(gomock.Any(), "ns1", "secret-name").Return(secret, nil)
	k8sRepo.EXPECT().UpdateSecret(gomock.Any(), "ns1", "secret-name", gomock.Any()).Return(secret, nil)

	repo := &Tasks{
		logger:   mlog.NewForConfig(nil),
		getCerts: func() (string, string, string) { return "secret-name", "tls-key", "tls-crt" },
		nsRepo:   nsRepo,
		k8sRepo:  k8sRepo,
	}
	assert.NoError(t, repo.SyncDomainSecret())
}

// TestTasks_SyncDomainSecret_NoCertsSkips 覆盖无证书配置分支：GetCerts 全空时
// 不查 namespace、不触碰 k8s，直接返回。
func TestTasks_SyncDomainSecret_NoCertsSkips(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	repo := &Tasks{
		logger:   mlog.NewForConfig(nil),
		getCerts: func() (string, string, string) { return "", "", "" },
	}
	// 无 nsRepo/k8sRepo EXPECT：若被调用会 panic，反向证明未走同步逻辑。
	assert.NoError(t, repo.SyncDomainSecret())
}

// TestTasks_SyncImagePullSecrets 覆盖缺 secret 创建 + 回写 imagePullSecrets 分支。
func TestTasks_SyncImagePullSecrets(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	nsRepo := data.NewMockNamespaceRepo(m)
	nsRepo.EXPECT().ListAll(context.TODO()).Return([]*biz.Namespace{{ID: 1, Name: "ns1", ImagePullSecrets: []string{}}}, nil)
	k8sRepo := data.NewMockK8sRepo(m)
	k8sRepo.EXPECT().CreateDockerSecrets(gomock.Any(), "ns1", []string{"reg.io"}).Return(&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "mars-abc"}}, nil)
	nsRepo.EXPECT().UpdateImagePullSecrets(gomock.Any(), 1, []string{"mars-abc"}).Return(nil)

	repo := &Tasks{
		logger:  mlog.NewForConfig(nil),
		cfg:     &config.Config{ImagePullSecrets: config.DockerAuths{{Server: "reg.io", Username: "u", Password: "p", Email: "e"}}},
		nsRepo:  nsRepo,
		k8sRepo: k8sRepo,
	}
	assert.NoError(t, repo.SyncImagePullSecrets())
}

// fakeTimer 固定时钟：让 CleanUploadFiles 的昨日时间窗完全确定，消除午夜跨日 flake。
type fakeTimer struct{ now time.Time }

func (f fakeTimer) Now() time.Time                  { return f.now }
func (f fakeTimer) Since(t time.Time) time.Duration { return f.now.Sub(t) }

// newTasksBase 构造注入真实时钟与日志的 Tasks 空壳，供各用例测试填充端口 mock。
func newTasksBase(m *gomock.Controller) *Tasks {
	return &Tasks{
		timer:  timer.NewReal(),
		logger: mlog.NewForConfig(nil),
	}
}

// newImagePullTasks 构造配置单个 registry 凭据（reg.io）的 Tasks，
// ListAll 返回给定 namespace 列表，供 SyncImagePullSecrets 分支矩阵复用。
func newImagePullTasks(m *gomock.Controller, ns []*biz.Namespace) (*Tasks, *data.MockNamespaceRepo, *data.MockK8sRepo) {
	nsRepo := data.NewMockNamespaceRepo(m)
	nsRepo.EXPECT().ListAll(context.TODO()).Return(ns, nil)
	k8sRepo := data.NewMockK8sRepo(m)
	return &Tasks{
		logger:  mlog.NewForConfig(nil),
		cfg:     &config.Config{ImagePullSecrets: config.DockerAuths{{Server: "reg.io", Username: "u", Password: "p", Email: "e"}}},
		nsRepo:  nsRepo,
		k8sRepo: k8sRepo,
	}, nsRepo, k8sRepo
}

// TestTasks_NewTasks 验证构造器全部依赖注入。
func TestTasks_NewTasks(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	tasks := NewTasks(
		timer.NewReal(),
		mlog.NewForConfig(nil),
		&config.Config{},
		data.NewMockFileRepo(m),
		data.NewMockProjectRepo(m),
		data.NewMockRepoRepo(m),
		data.NewMockNamespaceRepo(m),
		data.NewMockK8sRepo(m),
		data.NewMockEventRepo(m),
		uploader.NewMockUploader(m),
		data.NewMockHelmerRepo(m),
		data.NewMockGitRepo(m),
		&PluginDeps{GetCerts: func() (string, string, string) { return "", "", "" }},
	)
	assert.NotNil(t, tasks)
	assert.NotNil(t, tasks.getCerts)
	assert.NotNil(t, tasks.logger)
}

// TestTasks_CacheAllBranches 覆盖并发拉取分支：重复 GitProjectID 被 UniqBy 去重。
func TestTasks_CacheAllBranches(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	repoRepo := data.NewMockRepoRepo(m)
	repoRepo.EXPECT().All(gomock.Any(), gomock.Any()).Return([]*biz.Repo{
		{GitProjectID: 1}, {GitProjectID: 2}, {GitProjectID: 1},
	}, nil)
	gitRepo := data.NewMockGitRepo(m)
	gitRepo.EXPECT().AllBranches(gomock.Any(), 1, true)
	gitRepo.EXPECT().AllBranches(gomock.Any(), 2, true)
	repo := newTasksBase(m)
	repo.repoRepo = repoRepo
	repo.gitRepo = gitRepo
	assert.NoError(t, repo.CacheAllBranches())
}

// TestTasks_CacheAllBranches_ConcurrentLimit 覆盖仓库数超 10 时 worker 收敛到 8。
func TestTasks_CacheAllBranches_ConcurrentLimit(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	var repos []*biz.Repo
	for i := 0; i < 11; i++ {
		repos = append(repos, &biz.Repo{GitProjectID: int32(i)})
	}
	repoRepo := data.NewMockRepoRepo(m)
	repoRepo.EXPECT().All(gomock.Any(), gomock.Any()).Return(repos, nil)
	gitRepo := data.NewMockGitRepo(m)
	for i := 0; i < 11; i++ {
		gitRepo.EXPECT().AllBranches(gomock.Any(), i, true).Times(1)
	}
	repo := newTasksBase(m)
	repo.repoRepo = repoRepo
	repo.gitRepo = gitRepo
	assert.NoError(t, repo.CacheAllBranches())
}

func TestTasks_CacheAllBranches_RepoError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	repoRepo := data.NewMockRepoRepo(m)
	repoRepo.EXPECT().All(gomock.Any(), gomock.Any()).Return(nil, errors.New("repo err"))
	repo := newTasksBase(m)
	repo.repoRepo = repoRepo
	assert.Equal(t, "repo err", repo.CacheAllBranches().Error())
}

func TestTasks_CacheAllProjects(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	gitRepo := data.NewMockGitRepo(m)
	gitRepo.EXPECT().AllProjects(gomock.Any(), true).Return(nil, nil)
	repo := newTasksBase(m)
	repo.gitRepo = gitRepo
	assert.NoError(t, repo.CacheAllProjects())
}

func TestTasks_CacheAllProjects_Error(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	gitRepo := data.NewMockGitRepo(m)
	gitRepo.EXPECT().AllProjects(gomock.Any(), true).Return(nil, errors.New("git err"))
	repo := newTasksBase(m)
	repo.gitRepo = gitRepo
	assert.Equal(t, "git err", repo.CacheAllProjects().Error())
}

func TestTasks_SyncImagePullSecrets_ListAllError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	nsRepo := data.NewMockNamespaceRepo(m)
	nsRepo.EXPECT().ListAll(context.TODO()).Return(nil, errors.New("list err"))
	repo := &Tasks{logger: mlog.NewForConfig(nil), cfg: &config.Config{}, nsRepo: nsRepo}
	assert.Equal(t, "list err", repo.SyncImagePullSecrets().Error())
}

// TestTasks_SyncImagePullSecrets_GetSecretNotFoundDeletes 覆盖 GetSecret NotFound
// 分支：删除旧 secret、清空列表后补建缺失 registry。
func TestTasks_SyncImagePullSecrets_GetSecretNotFoundDeletes(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	repo, nsRepo, k8sRepo := newImagePullTasks(m, []*biz.Namespace{{ID: 1, Name: "ns1", ImagePullSecrets: []string{"old"}}})
	k8sRepo.EXPECT().GetSecret(gomock.Any(), "ns1", "old").Return(nil, apierrors.NewNotFound(schema.GroupResource{}, "old"))
	k8sRepo.EXPECT().DeleteSecret(gomock.Any(), "ns1", "old").Return(nil)
	nsRepo.EXPECT().UpdateImagePullSecrets(gomock.Any(), 1, []string(nil)).Return(nil)
	k8sRepo.EXPECT().CreateDockerSecrets(gomock.Any(), "ns1", []string{"reg.io"}).Return(&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "mars-abc"}}, nil)
	nsRepo.EXPECT().UpdateImagePullSecrets(gomock.Any(), 1, []string{"mars-abc"}).Return(nil)
	assert.NoError(t, repo.SyncImagePullSecrets())
}

// TestTasks_SyncImagePullSecrets_GetSecretOtherErrorSkips 覆盖非 NotFound 错误：
// 不删除只记日志，随后仍补建缺失 registry。
func TestTasks_SyncImagePullSecrets_GetSecretOtherErrorSkips(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	repo, nsRepo, k8sRepo := newImagePullTasks(m, []*biz.Namespace{{ID: 1, Name: "ns1", ImagePullSecrets: []string{"old"}}})
	k8sRepo.EXPECT().GetSecret(gomock.Any(), "ns1", "old").Return(nil, errors.New("boom"))
	k8sRepo.EXPECT().CreateDockerSecrets(gomock.Any(), "ns1", []string{"reg.io"}).Return(&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "mars-abc"}}, nil)
	// 非 NotFound 不删除：原列表保留，回写时拼接新 secret 名。
	nsRepo.EXPECT().UpdateImagePullSecrets(gomock.Any(), 1, []string{"old", "mars-abc"}).Return(nil)
	assert.NoError(t, repo.SyncImagePullSecrets())
}

// TestTasks_SyncImagePullSecrets_DockerSecretNoDiff 覆盖 docker 配置与凭据完全一致：
// 无 UpdateSecret、无 CreateDockerSecrets（被误调即 panic，反向证明未走分支）。
func TestTasks_SyncImagePullSecrets_DockerSecretNoDiff(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	auth := base64.StdEncoding.EncodeToString([]byte("u:p"))
	secret := &corev1.Secret{
		Type: corev1.SecretTypeDockerConfigJson,
		Data: map[string][]byte{corev1.DockerConfigJsonKey: []byte(`{"auths":{"reg.io":{"username":"u","password":"p","email":"e","auth":"` + auth + `"}}}`)},
	}
	repo, _, k8sRepo := newImagePullTasks(m, []*biz.Namespace{{ID: 1, Name: "ns1", ImagePullSecrets: []string{"json"}}})
	k8sRepo.EXPECT().GetSecret(gomock.Any(), "ns1", "json").Return(secret, nil)
	assert.NoError(t, repo.SyncImagePullSecrets())
}

// TestTasks_SyncImagePullSecrets_DockerSecretDiffSyncs 覆盖凭据不一致：自动同步更新 secret。
func TestTasks_SyncImagePullSecrets_DockerSecretDiffSyncs(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "json", Namespace: "ns1"},
		Type:       corev1.SecretTypeDockerConfigJson,
		Data:       map[string][]byte{corev1.DockerConfigJsonKey: []byte(`{"auths":{"reg.io":{"username":"u","password":"OLD","email":"e","auth":"b2xk"}}}`)},
	}
	repo, _, k8sRepo := newImagePullTasks(m, []*biz.Namespace{{ID: 1, Name: "ns1", ImagePullSecrets: []string{"json"}}})
	k8sRepo.EXPECT().GetSecret(gomock.Any(), "ns1", "json").Return(secret, nil)
	k8sRepo.EXPECT().UpdateSecret(gomock.Any(), "ns1", "json", gomock.Any()).Return(secret, nil)
	assert.NoError(t, repo.SyncImagePullSecrets())
}

func TestTasks_SyncImagePullSecrets_DockerSecretUpdateError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "json", Namespace: "ns1"},
		Type:       corev1.SecretTypeDockerConfigJson,
		Data:       map[string][]byte{corev1.DockerConfigJsonKey: []byte(`{"auths":{"reg.io":{"username":"u","password":"OLD","email":"e","auth":"b2xk"}}}`)},
	}
	repo, _, k8sRepo := newImagePullTasks(m, []*biz.Namespace{{ID: 1, Name: "ns1", ImagePullSecrets: []string{"json"}}})
	k8sRepo.EXPECT().GetSecret(gomock.Any(), "ns1", "json").Return(secret, nil)
	k8sRepo.EXPECT().UpdateSecret(gomock.Any(), "ns1", "json", gomock.Any()).Return(nil, errors.New("upd err"))
	assert.NoError(t, repo.SyncImagePullSecrets())
}

// TestTasks_SyncImagePullSecrets_DockerSecretNoMatchDeletes 覆盖 docker 配置里没有
// 任何受管 registry：删除该 secret 后补建缺失。
func TestTasks_SyncImagePullSecrets_DockerSecretNoMatchDeletes(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "json", Namespace: "ns1"},
		Type:       corev1.SecretTypeDockerConfigJson,
		Data:       map[string][]byte{corev1.DockerConfigJsonKey: []byte(`{"auths":{"other.io":{"username":"x"}}}`)},
	}
	repo, nsRepo, k8sRepo := newImagePullTasks(m, []*biz.Namespace{{ID: 1, Name: "ns1", ImagePullSecrets: []string{"json"}}})
	k8sRepo.EXPECT().GetSecret(gomock.Any(), "ns1", "json").Return(secret, nil)
	k8sRepo.EXPECT().DeleteSecret(gomock.Any(), "ns1", "json").Return(nil)
	nsRepo.EXPECT().UpdateImagePullSecrets(gomock.Any(), 1, []string(nil)).Return(nil)
	k8sRepo.EXPECT().CreateDockerSecrets(gomock.Any(), "ns1", []string{"reg.io"}).Return(&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "mars-abc"}}, nil)
	nsRepo.EXPECT().UpdateImagePullSecrets(gomock.Any(), 1, []string{"mars-abc"}).Return(nil)
	assert.NoError(t, repo.SyncImagePullSecrets())
}

// TestTasks_SyncImagePullSecrets_DockerSecretDecodeError 覆盖 docker 配置解析失败：跳过该 secret。
func TestTasks_SyncImagePullSecrets_DockerSecretDecodeError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	secret := &corev1.Secret{
		Type: corev1.SecretTypeDockerConfigJson,
		Data: map[string][]byte{corev1.DockerConfigJsonKey: []byte(`{invalid`)},
	}
	repo, nsRepo, k8sRepo := newImagePullTasks(m, []*biz.Namespace{{ID: 1, Name: "ns1", ImagePullSecrets: []string{"json"}}})
	k8sRepo.EXPECT().GetSecret(gomock.Any(), "ns1", "json").Return(secret, nil)
	k8sRepo.EXPECT().CreateDockerSecrets(gomock.Any(), "ns1", []string{"reg.io"}).Return(&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "mars-abc"}}, nil)
	nsRepo.EXPECT().UpdateImagePullSecrets(gomock.Any(), 1, []string{"json", "mars-abc"}).Return(nil)
	assert.NoError(t, repo.SyncImagePullSecrets())
}

// TestTasks_SyncImagePullSecrets_NonDockerTypeSkips 覆盖非 docker 类型 secret：跳过不做处理。
func TestTasks_SyncImagePullSecrets_NonDockerTypeSkips(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	secret := &corev1.Secret{Type: corev1.SecretTypeOpaque}
	repo, nsRepo, k8sRepo := newImagePullTasks(m, []*biz.Namespace{{ID: 1, Name: "ns1", ImagePullSecrets: []string{"opaque"}}})
	k8sRepo.EXPECT().GetSecret(gomock.Any(), "ns1", "opaque").Return(secret, nil)
	k8sRepo.EXPECT().CreateDockerSecrets(gomock.Any(), "ns1", []string{"reg.io"}).Return(&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "mars-abc"}}, nil)
	nsRepo.EXPECT().UpdateImagePullSecrets(gomock.Any(), 1, []string{"opaque", "mars-abc"}).Return(nil)
	assert.NoError(t, repo.SyncImagePullSecrets())
}

// TestTasks_SyncImagePullSecrets_CreateDockerSecretsError 覆盖创建失败分支：静默跳过。
func TestTasks_SyncImagePullSecrets_CreateDockerSecretsError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	repo, _, k8sRepo := newImagePullTasks(m, []*biz.Namespace{{ID: 1, Name: "ns1", ImagePullSecrets: []string{}}})
	k8sRepo.EXPECT().CreateDockerSecrets(gomock.Any(), "ns1", []string{"reg.io"}).Return(nil, errors.New("create err"))
	assert.NoError(t, repo.SyncImagePullSecrets())
}

// TestTasks_SyncImagePullSecrets_UpdateNsError 覆盖回写 namespace 失败分支：记日志不中断。
func TestTasks_SyncImagePullSecrets_UpdateNsError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	repo, nsRepo, k8sRepo := newImagePullTasks(m, []*biz.Namespace{{ID: 1, Name: "ns1", ImagePullSecrets: []string{}}})
	k8sRepo.EXPECT().CreateDockerSecrets(gomock.Any(), "ns1", []string{"reg.io"}).Return(&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "mars-abc"}}, nil)
	nsRepo.EXPECT().UpdateImagePullSecrets(gomock.Any(), 1, []string{"mars-abc"}).Return(errors.New("ns err"))
	assert.NoError(t, repo.SyncImagePullSecrets())
}

func TestTasks_deleteSecret_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	k8sRepo := data.NewMockK8sRepo(m)
	nsRepo := data.NewMockNamespaceRepo(m)
	k8sRepo.EXPECT().DeleteSecret(gomock.Any(), "ns1", "old").Return(nil)
	nsRepo.EXPECT().UpdateImagePullSecrets(gomock.Any(), 1, []string{"keep"}).Return(nil)
	repo := &Tasks{logger: mlog.NewForConfig(nil), k8sRepo: k8sRepo, nsRepo: nsRepo}
	ns := repo.deleteSecret(&biz.Namespace{ID: 1, Name: "ns1", ImagePullSecrets: []string{"keep", "old"}}, "old")
	assert.Equal(t, []string{"keep"}, ns.ImagePullSecrets)
}

func TestTasks_deleteSecret_DeleteSecretError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	k8sRepo := data.NewMockK8sRepo(m)
	nsRepo := data.NewMockNamespaceRepo(m)
	k8sRepo.EXPECT().DeleteSecret(gomock.Any(), "ns1", "old").Return(errors.New("del err"))
	nsRepo.EXPECT().UpdateImagePullSecrets(gomock.Any(), 1, []string(nil)).Return(nil)
	repo := &Tasks{logger: mlog.NewForConfig(nil), k8sRepo: k8sRepo, nsRepo: nsRepo}
	ns := repo.deleteSecret(&biz.Namespace{ID: 1, Name: "ns1", ImagePullSecrets: []string{"old"}}, "old")
	assert.Empty(t, ns.ImagePullSecrets)
}

func TestTasks_deleteSecret_UpdateNsError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	k8sRepo := data.NewMockK8sRepo(m)
	nsRepo := data.NewMockNamespaceRepo(m)
	k8sRepo.EXPECT().DeleteSecret(gomock.Any(), "ns1", "old").Return(nil)
	nsRepo.EXPECT().UpdateImagePullSecrets(gomock.Any(), 1, []string(nil)).Return(errors.New("ns err"))
	repo := &Tasks{logger: mlog.NewForConfig(nil), k8sRepo: k8sRepo, nsRepo: nsRepo}
	ns := repo.deleteSecret(&biz.Namespace{ID: 1, Name: "ns1", ImagePullSecrets: []string{"old"}}, "old")
	assert.Empty(t, ns.ImagePullSecrets)
}

func TestTasks_SyncDomainSecret_ListAllError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	nsRepo := data.NewMockNamespaceRepo(m)
	nsRepo.EXPECT().ListAll(context.TODO()).Return(nil, errors.New("list err"))
	repo := &Tasks{logger: mlog.NewForConfig(nil), getCerts: func() (string, string, string) { return "secret-name", "k", "c" }, nsRepo: nsRepo}
	assert.Equal(t, "list err", repo.SyncDomainSecret().Error())
}

// TestTasks_SyncDomainSecret_AddTlsError 覆盖 AddTlsSecret 失败分支：记日志不中断。
func TestTasks_SyncDomainSecret_AddTlsError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	nsRepo := data.NewMockNamespaceRepo(m)
	nsRepo.EXPECT().ListAll(context.TODO()).Return([]*biz.Namespace{{ID: 1, Name: "ns1"}}, nil)
	k8sRepo := data.NewMockK8sRepo(m)
	k8sRepo.EXPECT().GetSecret(gomock.Any(), "ns1", "secret-name").Return(nil, apierrors.NewNotFound(schema.GroupResource{}, "secret-name"))
	k8sRepo.EXPECT().AddTlsSecret("ns1", "secret-name", "k", "c").Return(nil, errors.New("add err"))
	repo := &Tasks{logger: mlog.NewForConfig(nil), getCerts: func() (string, string, string) { return "secret-name", "k", "c" }, nsRepo: nsRepo, k8sRepo: k8sRepo}
	assert.NoError(t, repo.SyncDomainSecret())
}

// TestTasks_SyncDomainSecret_UpdateSecretError 覆盖证书更新失败分支：记日志不中断。
func TestTasks_SyncDomainSecret_UpdateSecretError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	nsRepo := data.NewMockNamespaceRepo(m)
	nsRepo.EXPECT().ListAll(context.TODO()).Return([]*biz.Namespace{{ID: 1, Name: "ns1"}}, nil)
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Namespace: "ns1", Name: "secret-name"},
		Data:       map[string][]byte{"tls.crt": []byte("old-crt"), "tls.key": []byte("old-key")},
	}
	k8sRepo := data.NewMockK8sRepo(m)
	k8sRepo.EXPECT().GetSecret(gomock.Any(), "ns1", "secret-name").Return(secret, nil)
	k8sRepo.EXPECT().UpdateSecret(gomock.Any(), "ns1", "secret-name", gomock.Any()).Return(nil, errors.New("upd err"))
	repo := &Tasks{logger: mlog.NewForConfig(nil), getCerts: func() (string, string, string) { return "secret-name", "k", "c" }, nsRepo: nsRepo, k8sRepo: k8sRepo}
	assert.NoError(t, repo.SyncDomainSecret())
}

func TestTasks_FixDeployStatus(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	projectRepo := data.NewMockProjectRepo(m)
	projectRepo.EXPECT().ListByDeployStatus(gomock.Any(), types.Deploy_StatusFailed, types.Deploy_StatusUnknown).
		Return([]*biz.Project{{ID: 1, Name: "p1", Namespace: &biz.Namespace{Name: "ns1"}}}, nil)
	helm := data.NewMockHelmerRepo(m)
	helm.EXPECT().ReleaseStatus("p1", "ns1").Return(types.Deploy_StatusDeployed)
	projectRepo.EXPECT().UpdateDeployStatus(gomock.Any(), 1, types.Deploy_StatusDeployed).Return(&biz.Project{}, nil)
	repo := newTasksBase(m)
	repo.projectRepo = projectRepo
	repo.helm = helm
	assert.NoError(t, repo.FixDeployStatus())
}

// TestTasks_FixDeployStatus_SkipsUnchanged 覆盖 helm 实测状态仍为失败：不更新（误调即 panic）。
func TestTasks_FixDeployStatus_SkipsUnchanged(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	projectRepo := data.NewMockProjectRepo(m)
	projectRepo.EXPECT().ListByDeployStatus(gomock.Any(), types.Deploy_StatusFailed, types.Deploy_StatusUnknown).
		Return([]*biz.Project{{ID: 1, Name: "p1", Namespace: &biz.Namespace{Name: "ns1"}}}, nil)
	helm := data.NewMockHelmerRepo(m)
	helm.EXPECT().ReleaseStatus("p1", "ns1").Return(types.Deploy_StatusFailed)
	repo := newTasksBase(m)
	repo.projectRepo = projectRepo
	repo.helm = helm
	assert.NoError(t, repo.FixDeployStatus())
}

func TestTasks_FixDeployStatus_ListError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	projectRepo := data.NewMockProjectRepo(m)
	projectRepo.EXPECT().ListByDeployStatus(gomock.Any(), types.Deploy_StatusFailed, types.Deploy_StatusUnknown).Return(nil, errors.New("list err"))
	repo := newTasksBase(m)
	repo.projectRepo = projectRepo
	assert.Equal(t, "list err", repo.FixDeployStatus().Error())
}

func TestTasks_FixDeployStatus_UpdateError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	projectRepo := data.NewMockProjectRepo(m)
	projectRepo.EXPECT().ListByDeployStatus(gomock.Any(), types.Deploy_StatusFailed, types.Deploy_StatusUnknown).
		Return([]*biz.Project{{ID: 1, Name: "p1", Namespace: &biz.Namespace{Name: "ns1"}}}, nil)
	helm := data.NewMockHelmerRepo(m)
	helm.EXPECT().ReleaseStatus("p1", "ns1").Return(types.Deploy_StatusDeployed)
	projectRepo.EXPECT().UpdateDeployStatus(gomock.Any(), 1, types.Deploy_StatusDeployed).Return(nil, errors.New("upd err"))
	repo := newTasksBase(m)
	repo.projectRepo = projectRepo
	repo.helm = helm
	assert.Equal(t, "upd err", repo.FixDeployStatus().Error())
}

func TestTasks_DiskInfo(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	fileRepo := data.NewMockFileRepo(m)
	fileRepo.EXPECT().DiskInfo(true).Return(int64(1024), nil)
	repo := newTasksBase(m)
	repo.fileRepo = fileRepo
	size, err := repo.DiskInfo()
	assert.NoError(t, err)
	assert.Equal(t, int64(1024), size)
}

func TestTasks_DiskInfo_Error(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	fileRepo := data.NewMockFileRepo(m)
	fileRepo.EXPECT().DiskInfo(true).Return(int64(0), errors.New("disk err"))
	repo := newTasksBase(m)
	repo.fileRepo = fileRepo
	_, err := repo.DiskInfo()
	assert.Equal(t, "disk err", err.Error())
}

// TestTasks_CleanUploadFiles 覆盖对账主路径：孤儿记录删除、游离文件清理、
// 目录里已记录文件跳过、非受管类型跳过、删除失败仅警告。
func TestTasks_CleanUploadFiles(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	startOfDay := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, time.Local)

	fileRepo := data.NewMockFileRepo(m)
	fileRepo.EXPECT().ListByCreatedAtRange(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*biz.File{
		{ID: 1, Path: "s3-orphan", UploadType: schematype.S3},
		{ID: 2, Path: "local-keep", UploadType: schematype.Local},
		{ID: 3, Path: "local-orphan2", UploadType: schematype.Local},
		{ID: 4, Path: "weird", UploadType: schematype.UploadType("weird")},
	}, nil)
	fileRepo.EXPECT().DeleteRecord(gomock.Any(), 1).Return(nil)
	fileRepo.EXPECT().DeleteRecord(gomock.Any(), 3).Return(errors.New("rec err"))

	upldr := uploader.NewMockUploader(m)
	local := uploader.NewMockUploader(m)
	upldr.EXPECT().LocalUploader().Return(local)
	upldr.EXPECT().Type().Return(schematype.S3).AnyTimes()
	local.EXPECT().Type().Return(schematype.Local).AnyTimes()
	upldr.EXPECT().Exists("s3-orphan").Return(false)
	local.EXPECT().Exists("local-keep").Return(true)
	local.EXPECT().Exists("local-orphan2").Return(false)

	fiIn := uploader.NewMockFileInfo(m)
	fiIn.EXPECT().LastModified().Return(startOfDay.Add(time.Minute)).AnyTimes()
	fiIn.EXPECT().Path().Return("local-orphan1").AnyTimes()
	fiIn.EXPECT().Size().Return(uint64(123))
	fiSkip := uploader.NewMockFileInfo(m)
	fiSkip.EXPECT().LastModified().Return(startOfDay.Add(2 * time.Minute)).AnyTimes()
	fiSkip.EXPECT().Path().Return("local-keep")
	fiOut := uploader.NewMockFileInfo(m)
	// 时间窗判断 Before/After 各调一次 LastModified，故 AnyTimes。
	fiOut.EXPECT().LastModified().Return(startOfDay.Add(-time.Minute)).AnyTimes()
	local.EXPECT().AllDirectoryFiles("").Return([]uploader.FileInfo{fiIn, fiSkip, fiOut}, nil)
	local.EXPECT().Delete("local-orphan1").Return(errors.New("del err"))

	// S3 上传器：AllDirectoryFiles 报错被忽略（directoryFiles, _ :=）。
	upldr.EXPECT().AllDirectoryFiles("").Return(nil, errors.New("boom"))

	local.EXPECT().RemoveEmptyDir().Return(nil)

	event := data.NewMockEventRepo(m)
	event.EXPECT().AuditLogWithChange(types.EventActionType_Delete, "system", "", "删除未被记录的文件", gomock.Any(), nil)

	repo := newTasksBase(m)
	repo.timer = fakeTimer{now: now}
	repo.fileRepo = fileRepo
	repo.up = upldr
	repo.event = event
	assert.NoError(t, repo.CleanUploadFiles())
}

func TestTasks_CleanUploadFiles_ListError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	fileRepo := data.NewMockFileRepo(m)
	fileRepo.EXPECT().ListByCreatedAtRange(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("list err"))
	up := uploader.NewMockUploader(m)
	up.EXPECT().LocalUploader().Return(nil)
	repo := newTasksBase(m)
	repo.fileRepo = fileRepo
	repo.up = up
	assert.Equal(t, "list err", repo.CleanUploadFiles().Error())
}

// Test_listFiles_PrettyYaml 覆盖列表的 YAML 序列化与空列表边界。
func Test_listFiles_PrettyYaml(t *testing.T) {
	files := listFiles{
		{Path: "/a.txt", HumanizeSize: "1.0 kB"},
		{Path: "/b.txt", HumanizeSize: "2.0 kB"},
	}
	out := files.PrettyYaml()
	assert.Contains(t, out, "name: /a.txt")
	assert.Contains(t, out, "size: 1.0 kB")
	assert.Contains(t, out, "name: /b.txt")
	assert.Contains(t, out, "size: 2.0 kB")
	assert.Equal(t, "[]\n", listFiles{}.PrettyYaml())
}
