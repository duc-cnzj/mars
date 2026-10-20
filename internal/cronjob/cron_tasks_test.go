package cronjob

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/api/v6/proto/mars"
	"github.com/duc-cnzj/mars/api/v6/proto/types"
	websocket_pb "github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/biz/schematype"
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/data"
	"github.com/duc-cnzj/mars/v6/internal/deploy"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/locker"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/uploader"
	"github.com/duc-cnzj/mars/v6/internal/util/timer"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	appsv1 "k8s.io/api/apps/v1"
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
	nsRepo.EXPECT().ListAll(context.Background()).Return([]*biz.Namespace{{ID: 1, Name: "ns1"}}, nil)
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
	nsRepo.EXPECT().ListAll(context.Background()).Return([]*biz.Namespace{{ID: 1, Name: "ns1"}}, nil)
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
	nsRepo.EXPECT().ListAll(context.Background()).Return([]*biz.Namespace{{ID: 1, Name: "ns1"}}, nil)
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
	nsRepo.EXPECT().ListAll(context.Background()).Return([]*biz.Namespace{{ID: 1, Name: "ns1", ImagePullSecrets: []string{}}}, nil)
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
	nsRepo.EXPECT().ListAll(context.Background()).Return(ns, nil)
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
		data.NewMockHibernateRepo(m),
		data.NewMockHibernateDeployer(m),
		data.NewMockHibernateWindowRepo(m),
		biz.NewMockAgentActivity(m),
		locker.NewMockLocker(m),
		&PluginDeps{GetCerts: func() (string, string, string) { return "", "", "" }},
	)
	assert.NotNil(t, tasks)
	assert.NotNil(t, tasks.getCerts)
	assert.NotNil(t, tasks.logger)
	assert.NotNil(t, tasks.locker)
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
	nsRepo.EXPECT().ListAll(context.Background()).Return(nil, errors.New("list err"))
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
	nsRepo.EXPECT().ListAll(context.Background()).Return(nil, errors.New("list err"))
	repo := &Tasks{logger: mlog.NewForConfig(nil), getCerts: func() (string, string, string) { return "secret-name", "k", "c" }, nsRepo: nsRepo}
	assert.Equal(t, "list err", repo.SyncDomainSecret().Error())
}

// TestTasks_SyncDomainSecret_AddTlsError 覆盖 AddTlsSecret 失败分支：记日志不中断。
func TestTasks_SyncDomainSecret_AddTlsError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	nsRepo := data.NewMockNamespaceRepo(m)
	nsRepo.EXPECT().ListAll(context.Background()).Return([]*biz.Namespace{{ID: 1, Name: "ns1"}}, nil)
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
	nsRepo.EXPECT().ListAll(context.Background()).Return([]*biz.Namespace{{ID: 1, Name: "ns1"}}, nil)
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

// TestTasks_CacheClusterBoard 覆盖集群看板 30s 预热任务：force=true 触发一次刷新，
// 成功后无错误。
func TestTasks_CacheClusterBoard(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	k8sRepo := data.NewMockK8sRepo(m)
	k8sRepo.EXPECT().ClusterBoard(context.Background(), true).Return(&biz.ClusterBoardData{}, nil)
	repo := &Tasks{k8sRepo: k8sRepo}
	assert.NoError(t, repo.CacheClusterBoard())
}

// TestTasks_CacheClusterBoard_Error 覆盖集群看板刷新失败路径：错误原样上抛。
func TestTasks_CacheClusterBoard_Error(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	k8sRepo := data.NewMockK8sRepo(m)
	k8sRepo.EXPECT().ClusterBoard(context.Background(), true).Return(nil, errors.New("cluster board boom"))
	repo := &Tasks{k8sRepo: k8sRepo}
	assert.Equal(t, "cluster board boom", repo.CacheClusterBoard().Error())
}

// TestTasks_CacheResourceSnapshot 覆盖空间资源快照 5m 预热任务：force=true 触发一次
// 刷新，成功后无错误。
func TestTasks_CacheResourceSnapshot(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	k8sRepo := data.NewMockK8sRepo(m)
	k8sRepo.EXPECT().ResourceSnapshot(context.Background(), true).Return(&biz.ResourceSnapshotData{}, nil)
	repo := &Tasks{k8sRepo: k8sRepo}
	assert.NoError(t, repo.CacheResourceSnapshot())
}

// TestTasks_CacheResourceSnapshot_Error 覆盖空间资源快照刷新失败路径：错误原样上抛。
func TestTasks_CacheResourceSnapshot_Error(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	k8sRepo := data.NewMockK8sRepo(m)
	k8sRepo.EXPECT().ResourceSnapshot(context.Background(), true).Return(nil, errors.New("snapshot boom"))
	repo := &Tasks{k8sRepo: k8sRepo}
	assert.Equal(t, "snapshot boom", repo.CacheResourceSnapshot().Error())
}

// TestTasks_CacheClusterInfo 覆盖集群信息统计 30s 预热任务：force 刷新一次，成功后无错误。
func TestTasks_CacheClusterInfo(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	k8sRepo := data.NewMockK8sRepo(m)
	k8sRepo.EXPECT().RefreshClusterInfo().Return(&biz.ClusterInfo{}, nil)
	repo := &Tasks{k8sRepo: k8sRepo}
	assert.NoError(t, repo.CacheClusterInfo())
}

// TestTasks_CacheClusterInfo_Error 覆盖集群信息刷新失败路径：错误原样上抛。
func TestTasks_CacheClusterInfo_Error(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	k8sRepo := data.NewMockK8sRepo(m)
	k8sRepo.EXPECT().RefreshClusterInfo().Return(nil, errors.New("cluster info boom"))
	repo := &Tasks{k8sRepo: k8sRepo}
	assert.Equal(t, "cluster info boom", repo.CacheClusterInfo().Error())
}

// TestTasks_FixDeployStatus_SkipsNilNamespace 项目存活、其所属空间已软删时，ListByDeployStatus
// 的命名空间边加载（走 SELECT，被软删拦截器过滤掉已软删空间）拿不到空间 → Namespace 为 nil；
// 这类项目的 deploy_status 仍是列默认值 0（StatusUnknown），恰好落在失败/未知过滤条件内。
// 缺 nil 守卫即 nil 解引用 panic，让整个定时任务永久失效。
//
// 用例混入一个正常项目：若为 nil 项目误调 helm，gomock 会因未预期调用而失败，双重兜底。
func TestTasks_FixDeployStatus_SkipsNilNamespace(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	projectRepo := data.NewMockProjectRepo(m)
	projectRepo.EXPECT().ListByDeployStatus(gomock.Any(), types.Deploy_StatusFailed, types.Deploy_StatusUnknown).
		Return([]*biz.Project{
			{ID: 1, Name: "orphan", Namespace: nil},
			{ID: 2, Name: "p2", Namespace: &biz.Namespace{Name: "ns2"}},
		}, nil)
	helm := data.NewMockHelmerRepo(m)
	helm.EXPECT().ReleaseStatus("p2", "ns2").Return(types.Deploy_StatusDeployed)
	projectRepo.EXPECT().UpdateDeployStatus(gomock.Any(), 2, types.Deploy_StatusDeployed).Return(&biz.Project{}, nil)
	repo := newTasksBase(m)
	repo.projectRepo = projectRepo
	repo.helm = helm

	assert.NoError(t, repo.FixDeployStatus())
}

// ---------- ReconcileHibernate ----------

// hibernateNowIn 返回落在休眠窗口 [23:00, 06:00) 内的固定时刻，配合固定时钟消除时区/跨日 flake。
func hibernateNowIn() time.Time {
	return time.Date(2026, 10, 10, 23, 30, 0, 0, time.Local)
}

// hibernateNowOut 返回落在休眠窗口外的固定时刻（同一日中午）。
func hibernateNowOut() time.Time {
	return time.Date(2026, 10, 10, 12, 0, 0, 0, time.Local)
}

// hibernateCfgRepo 返回启用休眠的仓库：副本字段 app->replicas，引用全局时间窗「夜间」
// （helper 默认把该窗解析为 23:00-06:00）。
func hibernateCfgRepo(id int) *biz.Repo {
	return &biz.Repo{ID: id, MarsConfig: &mars.Config{Hibernate: &mars.Hibernate{
		Enabled: true, ReplicaPaths: []string{"app->replicas"}, WindowNames: []string{"夜间"},
	}}}
}

// hibernateProject 返回副本数为 3 的项目，供休眠对账分支复用。
func hibernateProject() *biz.Project {
	return &biz.Project{ID: 9, Name: "app", RepoID: 2, FinalExtraValues: []*websocket_pb.ExtraValue{
		{Path: "app->replicas", Value: "3"},
		{Path: "app->image", Value: "nginx:1.0"},
	}}
}

// hibernateSleepingProject 返回副本字段已置 0 的项目，代表「确实处于休眠态」——供一致性守卫
// 需判为不命中的分支（真休眠）复用。
func hibernateSleepingProject() *biz.Project {
	return &biz.Project{ID: 9, Name: "app", RepoID: 2, FinalExtraValues: []*websocket_pb.ExtraValue{
		{Path: "app->replicas", Value: "0"},
		{Path: "app->image", Value: "nginx:1.0"},
	}}
}

// newHibernateTasks 构造注入固定时钟与四个端口 mock 的 Tasks，供休眠对账分支矩阵复用。
// 另注册一个时间窗解析的默认期望（把「夜间」解析为 23:00-06:00，AnyTimes）：不触达解析的
// 分支零调用亦无碍；需要「空名/窗口不存在」的用例自行构造 Tasks（见 TestTasks_ResolveHibernateWindow_*）。
func newHibernateTasks(m *gomock.Controller, now time.Time) (*Tasks, *data.MockHibernateRepo, *data.MockHibernateDeployer, *data.MockProjectRepo, *data.MockRepoRepo) {
	hr := data.NewMockHibernateRepo(m)
	hd := data.NewMockHibernateDeployer(m)
	pr := data.NewMockProjectRepo(m)
	rr := data.NewMockRepoRepo(m)
	hw := data.NewMockHibernateWindowRepo(m)
	hw.EXPECT().GetByName(gomock.Any(), gomock.Any()).Return(
		&biz.HibernateWindow{Name: "夜间", Start: "23:00", End: "06:00"}, nil).AnyTimes()
	return &Tasks{
		timer:               fakeTimer{now: now},
		logger:              mlog.NewForConfig(nil),
		hibernateRepo:       hr,
		hibernateDeployer:   hd,
		projectRepo:         pr,
		repoRepo:            rr,
		hibernateWindowRepo: hw,
	}, hr, hd, pr, rr
}

// captureHibernateDeploy 注册 Redeploy 期望并把实参副本字段快照写入 out。
func captureHibernateDeploy(hd *data.MockHibernateDeployer, proj *biz.Project, out *[]*websocket_pb.ExtraValue) {
	hd.EXPECT().Redeploy(gomock.Any(), proj, gomock.Any()).DoAndReturn(
		func(_ context.Context, _ *biz.Project, ev []*websocket_pb.ExtraValue) error {
			*out = ev
			return nil
		})
}

// captureHibernateUpsert 注册 Upsert 期望并把落库记录写入 out。
func captureHibernateUpsert(hr *data.MockHibernateRepo, out **biz.ProjectHibernate) {
	hr.EXPECT().Upsert(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, h *biz.ProjectHibernate) (*biz.ProjectHibernate, error) {
			*out = h
			return h, nil
		})
}

// customHibernateTasks 构造可按名字返回不同时间窗（未列出的名字返回 nil=悬空）的 Tasks，
// 供「多窗并集 / 悬空窗」分支用例复用；其余端口与 newHibernateTasks 一致。
func customHibernateTasks(m *gomock.Controller, now time.Time, wins map[string]*biz.HibernateWindow) (*Tasks, *data.MockHibernateRepo, *data.MockHibernateDeployer, *data.MockProjectRepo, *data.MockRepoRepo) {
	hr := data.NewMockHibernateRepo(m)
	hd := data.NewMockHibernateDeployer(m)
	pr := data.NewMockProjectRepo(m)
	rr := data.NewMockRepoRepo(m)
	hw := data.NewMockHibernateWindowRepo(m)
	hw.EXPECT().GetByName(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, name string) (*biz.HibernateWindow, error) {
			return wins[name], nil
		}).AnyTimes()
	return &Tasks{
		timer:               fakeTimer{now: now},
		logger:              mlog.NewForConfig(nil),
		hibernateRepo:       hr,
		hibernateDeployer:   hd,
		projectRepo:         pr,
		repoRepo:            rr,
		hibernateWindowRepo: hw,
	}, hr, hd, pr, rr
}

// TestTasks_ReconcileHibernate_Sleep 覆盖清醒→休眠：命中窗口时副本字段置 0 重部署、
// 状态转 sleeping 并保留休眠前副本快照。
func TestTasks_ReconcileHibernate_Sleep(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	repo, hr, hd, pr, rr := newHibernateTasks(m, hibernateNowIn())

	hr.EXPECT().ListEnabled(gomock.Any()).Return([]*biz.ProjectHibernate{
		{ProjectID: 9, Enabled: true, State: biz.HibernateStateAwake, WindowNames: []string{"夜间"}},
	}, nil)
	proj := hibernateProject()
	pr.EXPECT().Show(gomock.Any(), 9).Return(proj, nil)
	rr.EXPECT().Get(gomock.Any(), 2).Return(hibernateCfgRepo(2), nil)

	var target []*websocket_pb.ExtraValue
	captureHibernateDeploy(hd, proj, &target)
	var rec *biz.ProjectHibernate
	captureHibernateUpsert(hr, &rec)

	assert.NoError(t, repo.ReconcileHibernate())
	assert.Equal(t, "0", target[0].Value)
	assert.Equal(t, "nginx:1.0", target[1].Value)
	assert.Equal(t, biz.HibernateStateSleeping, rec.State)
	assert.True(t, rec.Enabled)
	assert.Len(t, rec.Snapshot, 1)
	assert.Equal(t, "3", rec.Snapshot[0].Value)
}

// TestTasks_ReconcileHibernate_SleepUnionSecondWindow 覆盖项目多窗子集的并集语义：项目选了
// 两个窗，此刻只有第二个窗命中（第一个已出窗）——命中任一即算在窗内，正常休眠且选择原样回写。
func TestTasks_ReconcileHibernate_SleepUnionSecondWindow(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	repo, hr, hd, pr, rr := customHibernateTasks(m, hibernateNowIn(), map[string]*biz.HibernateWindow{
		"夜间": {Name: "夜间", Start: "23:00", End: "06:00"}, // 午间：错过
		"午休": {Name: "午休", Start: "11:00", End: "14:00"}, // 午间：命中
	})

	hr.EXPECT().ListEnabled(gomock.Any()).Return([]*biz.ProjectHibernate{
		{ProjectID: 9, Enabled: true, State: biz.HibernateStateAwake, WindowNames: []string{"夜间", "午休"}},
	}, nil)
	proj := hibernateProject()
	pr.EXPECT().Show(gomock.Any(), 9).Return(proj, nil)
	rr.EXPECT().Get(gomock.Any(), 2).Return(&biz.Repo{ID: 2, MarsConfig: &mars.Config{Hibernate: &mars.Hibernate{
		Enabled: true, ReplicaPaths: []string{"app->replicas"}, WindowNames: []string{"夜间", "午休"},
	}}}, nil)

	var target []*websocket_pb.ExtraValue
	captureHibernateDeploy(hd, proj, &target)
	var rec *biz.ProjectHibernate
	captureHibernateUpsert(hr, &rec)

	assert.NoError(t, repo.ReconcileHibernate())
	assert.Equal(t, "0", target[0].Value)
	assert.Equal(t, biz.HibernateStateSleeping, rec.State)
	// 项目多窗选择原样回写，不被对账清空。
	assert.Equal(t, []string{"夜间", "午休"}, rec.WindowNames)
}

// TestTasks_ReconcileHibernate_EmptySelectionNoInherit 覆盖「空选择不再继承仓库全量」：项目未选任何
// 窗（WindowNames 空）时按「未选即未启用」处理——不参与休眠判定。即便此刻恰落在仓库某窗内
// （now 取命中时刻、仓库两窗其一命中），也零动作：无 Redeploy/Upsert（未注册期望，误调即 panic）。
func TestTasks_ReconcileHibernate_EmptySelectionNoInherit(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	repo, hr, _, pr, rr := customHibernateTasks(m, hibernateNowIn(), map[string]*biz.HibernateWindow{
		"夜间": {Name: "夜间", Start: "23:00", End: "06:00"}, // 错过
		"午休": {Name: "午休", Start: "11:00", End: "14:00"}, // 命中
	})

	hr.EXPECT().ListEnabled(gomock.Any()).Return([]*biz.ProjectHibernate{
		{ProjectID: 9, Enabled: true, State: biz.HibernateStateAwake}, // WindowNames 空 = 未选，不继承仓库全量
	}, nil)
	pr.EXPECT().Show(gomock.Any(), 9).Return(hibernateProject(), nil)
	rr.EXPECT().Get(gomock.Any(), 2).Return(&biz.Repo{ID: 2, MarsConfig: &mars.Config{Hibernate: &mars.Hibernate{
		Enabled: true, ReplicaPaths: []string{"app->replicas"}, WindowNames: []string{"夜间", "午休"},
	}}}, nil)

	// 无 Redeploy/Upsert 期望：空选择既不继承窗、也不休眠/唤醒，零副作用即通过。
	assert.NoError(t, repo.ReconcileHibernate())
}

// TestTasks_ReconcileHibernate_WakeDanglingWindow 覆盖「睡眠中唯一被引用的窗被删」：对账按名
// 解析全失败 → 视为无窗口（出窗）→ 确在休眠则正常唤醒，避免悬空引用让项目永久卡在休眠态。
func TestTasks_ReconcileHibernate_WakeDanglingWindow(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	// 所有名字都解析不到（模拟窗被删 / 跨环境未同步）。
	repo, hr, hd, pr, rr := customHibernateTasks(m, hibernateNowIn(), nil)

	hr.EXPECT().ListEnabled(gomock.Any()).Return([]*biz.ProjectHibernate{
		{ProjectID: 9, Enabled: true, State: biz.HibernateStateSleeping,
			Snapshot:    []*websocket_pb.ExtraValue{{Path: "app->replicas", Value: "3"}},
			WindowNames: []string{"已删除的窗"}},
	}, nil)
	proj := hibernateSleepingProject()
	pr.EXPECT().Show(gomock.Any(), 9).Return(proj, nil)
	rr.EXPECT().Get(gomock.Any(), 2).Return(hibernateCfgRepo(2), nil)

	var target []*websocket_pb.ExtraValue
	captureHibernateDeploy(hd, proj, &target)
	var rec *biz.ProjectHibernate
	captureHibernateUpsert(hr, &rec)

	assert.NoError(t, repo.ReconcileHibernate())
	assert.Equal(t, "3", target[0].Value)
	assert.Equal(t, biz.HibernateStateAwake, rec.State)
}

// TestTasks_ReconcileHibernate_SleepPersistsSnapshotBeforeDeploy 是「休眠不幂等丢数」的回归闸门：
// 必须先落「休眠态+快照」再重部署。反序（先部署后落库）会让部署把副本置 0 落库后、若落库失败则
// 下一轮对账以全 0 基座重拍快照，永久丢原始副本数（唤醒只能恢复成 0）。用 InOrder 钉死调用顺序，
// 并在落库回调里断言快照仍是休眠前原值（3），反序时该值会变成 0 而被抓住。
func TestTasks_ReconcileHibernate_SleepPersistsSnapshotBeforeDeploy(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	repo, hr, hd, pr, rr := newHibernateTasks(m, hibernateNowIn())

	hr.EXPECT().ListEnabled(gomock.Any()).Return([]*biz.ProjectHibernate{
		{ProjectID: 9, Enabled: true, State: biz.HibernateStateAwake, WindowNames: []string{"夜间"}},
	}, nil)
	proj := hibernateProject()
	pr.EXPECT().Show(gomock.Any(), 9).Return(proj, nil)
	rr.EXPECT().Get(gomock.Any(), 2).Return(hibernateCfgRepo(2), nil)

	gomock.InOrder(
		hr.EXPECT().Upsert(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, h *biz.ProjectHibernate) (*biz.ProjectHibernate, error) {
				assert.Equal(t, biz.HibernateStateSleeping, h.State)
				assert.Len(t, h.Snapshot, 1)
				assert.Equal(t, "3", h.Snapshot[0].Value)
				return h, nil
			}),
		hd.EXPECT().Redeploy(gomock.Any(), proj, gomock.Any()).Return(nil),
	)

	assert.NoError(t, repo.ReconcileHibernate())
}

// TestTasks_ReconcileHibernate_SleepSkipNotInWindow 覆盖清醒态但未命中窗口：
// 不重部署、不改状态（无 Redeploy/Upsert 期望，误调即 panic）。
func TestTasks_ReconcileHibernate_SleepSkipNotInWindow(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	repo, hr, _, pr, rr := newHibernateTasks(m, hibernateNowOut())

	hr.EXPECT().ListEnabled(gomock.Any()).Return([]*biz.ProjectHibernate{
		{ProjectID: 9, Enabled: true, State: biz.HibernateStateAwake, WindowNames: []string{"夜间"}},
	}, nil)
	pr.EXPECT().Show(gomock.Any(), 9).Return(hibernateProject(), nil)
	rr.EXPECT().Get(gomock.Any(), 2).Return(hibernateCfgRepo(2), nil)

	assert.NoError(t, repo.ReconcileHibernate())
}

// TestTasks_ReconcileHibernate_SleepSkipRepoDisabled 覆盖仓库未配置休眠：直接跳过。
func TestTasks_ReconcileHibernate_SleepSkipRepoDisabled(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	repo, hr, _, pr, rr := newHibernateTasks(m, hibernateNowIn())

	hr.EXPECT().ListEnabled(gomock.Any()).Return([]*biz.ProjectHibernate{
		{ProjectID: 9, Enabled: true, State: biz.HibernateStateAwake, WindowNames: []string{"夜间"}},
	}, nil)
	pr.EXPECT().Show(gomock.Any(), 9).Return(hibernateProject(), nil)
	rr.EXPECT().Get(gomock.Any(), 2).Return(&biz.Repo{ID: 2}, nil)

	assert.NoError(t, repo.ReconcileHibernate())
}

// TestTasks_ReconcileHibernate_SleepSkipNoWindow 覆盖清醒态但引用的全局时间窗不存在
// （被删/未同步环境）：解析失败按无窗口处理，不重部署、不改状态（无 Redeploy/Upsert 期望）。
func TestTasks_ReconcileHibernate_SleepSkipNoWindow(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	hr := data.NewMockHibernateRepo(m)
	pr := data.NewMockProjectRepo(m)
	rr := data.NewMockRepoRepo(m)
	hw := data.NewMockHibernateWindowRepo(m)
	repo := &Tasks{
		timer:               fakeTimer{now: hibernateNowIn()},
		logger:              mlog.NewForConfig(nil),
		hibernateRepo:       hr,
		projectRepo:         pr,
		repoRepo:            rr,
		hibernateWindowRepo: hw,
	}

	hr.EXPECT().ListEnabled(gomock.Any()).Return([]*biz.ProjectHibernate{
		{ProjectID: 9, Enabled: true, State: biz.HibernateStateAwake, WindowNames: []string{"夜间"}},
	}, nil)
	pr.EXPECT().Show(gomock.Any(), 9).Return(hibernateProject(), nil)
	rr.EXPECT().Get(gomock.Any(), 2).Return(hibernateCfgRepo(2), nil)
	hw.EXPECT().GetByName(gomock.Any(), "夜间").Return(nil, nil)

	assert.NoError(t, repo.ReconcileHibernate())
}

// TestTasks_ResolveHibernateWindows 覆盖全局时间窗列表解析的各分支：空列表/纯空白名不查库、
// 逐名解析（悬空与报错逐个跳过、不牵连其余），返回全部可解析窗；全部不可解析才 ok=false。
func TestTasks_ResolveHibernateWindows(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	hw := data.NewMockHibernateWindowRepo(m)
	tk := &Tasks{logger: mlog.NewForConfig(nil), hibernateWindowRepo: hw}

	// 空列表 / 纯空白名：不解析、直接 ok=false（无 GetByName 期望，误调即 panic）。
	_, ok := tk.resolveHibernateWindows(context.Background(), nil)
	assert.False(t, ok)
	_, ok = tk.resolveHibernateWindows(context.Background(), []string{"", "  "})
	assert.False(t, ok)

	// 唯一名不存在（被删/未同步环境）：GetByName 返回 nil → 全部不可解析 → ok=false。
	hw.EXPECT().GetByName(gomock.Any(), "不存在").Return(nil, nil)
	_, ok = tk.resolveHibernateWindows(context.Background(), []string{"不存在"})
	assert.False(t, ok)

	// 唯一名查询报错：不阻断对账，全部不可解析 → ok=false。
	hw.EXPECT().GetByName(gomock.Any(), "出错").Return(nil, errors.New("boom"))
	_, ok = tk.resolveHibernateWindows(context.Background(), []string{"出错"})
	assert.False(t, ok)

	// 命中：返回该窗。
	hw.EXPECT().GetByName(gomock.Any(), "夜间").Return(
		&biz.HibernateWindow{Name: "夜间", Start: "23:00", End: "06:00"}, nil)
	wins, ok := tk.resolveHibernateWindows(context.Background(), []string{"夜间"})
	assert.True(t, ok)
	assert.Len(t, wins, 1)
	assert.Equal(t, "夜间", wins[0].Name)

	// 多窗：悬空 + 空白 + 命中混合——只有可解析的进结果，整体 ok=true（有任一可解析即可）。
	hw.EXPECT().GetByName(gomock.Any(), "悬空").Return(nil, nil)
	hw.EXPECT().GetByName(gomock.Any(), "午休").Return(
		&biz.HibernateWindow{Name: "午休", Start: "12:00", End: "13:00"}, nil)
	wins, ok = tk.resolveHibernateWindows(context.Background(), []string{"悬空", "  ", "午休"})
	assert.True(t, ok)
	assert.Len(t, wins, 1)
	assert.Equal(t, "午休", wins[0].Name)

	// 多窗全部悬空：ok=false（视为无窗口）。
	hw.EXPECT().GetByName(gomock.Any(), "悬空A").Return(nil, nil)
	hw.EXPECT().GetByName(gomock.Any(), "悬空B").Return(nil, nil)
	wins, ok = tk.resolveHibernateWindows(context.Background(), []string{"悬空A", "悬空B"})
	assert.False(t, ok)
	assert.Empty(t, wins)
}

// TestTasks_ReconcileHibernate_KeepSleeping 覆盖已休眠（副本确实全 0）且仍在窗口内：保持
// sleeping，不重部署、不改状态、不清空快照（无 Redeploy/Upsert 期望，误调即 panic）。
func TestTasks_ReconcileHibernate_KeepSleeping(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	repo, hr, _, pr, rr := newHibernateTasks(m, hibernateNowIn())

	hr.EXPECT().ListEnabled(gomock.Any()).Return([]*biz.ProjectHibernate{
		{ProjectID: 9, Enabled: true, State: biz.HibernateStateSleeping,
			Snapshot:    []*websocket_pb.ExtraValue{{Path: "app->replicas", Value: "3"}},
			WindowNames: []string{"夜间"}},
	}, nil)
	pr.EXPECT().Show(gomock.Any(), 9).Return(hibernateSleepingProject(), nil)
	rr.EXPECT().Get(gomock.Any(), 2).Return(hibernateCfgRepo(2), nil)

	assert.NoError(t, repo.ReconcileHibernate())
}

// TestTasks_ReconcileHibernate_Wake 覆盖休眠→唤醒（已出窗口）：按快照回填副本字段重部署、
// 状态转 awake 并清空快照。
func TestTasks_ReconcileHibernate_Wake(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	repo, hr, hd, pr, rr := newHibernateTasks(m, hibernateNowOut())

	hr.EXPECT().ListEnabled(gomock.Any()).Return([]*biz.ProjectHibernate{
		{ProjectID: 9, Enabled: true, State: biz.HibernateStateSleeping,
			Snapshot:    []*websocket_pb.ExtraValue{{Path: "app->replicas", Value: "3"}},
			WindowNames: []string{"夜间"}},
	}, nil)
	proj := hibernateSleepingProject()
	pr.EXPECT().Show(gomock.Any(), 9).Return(proj, nil)
	rr.EXPECT().Get(gomock.Any(), 2).Return(hibernateCfgRepo(2), nil)

	var target []*websocket_pb.ExtraValue
	captureHibernateDeploy(hd, proj, &target)
	var rec *biz.ProjectHibernate
	captureHibernateUpsert(hr, &rec)

	assert.NoError(t, repo.ReconcileHibernate())
	assert.Equal(t, "3", target[0].Value)
	assert.Equal(t, biz.HibernateStateAwake, rec.State)
	assert.True(t, rec.Enabled)
	assert.Empty(t, rec.Snapshot)
}

// TestTasks_ReconcileHibernate_WakeRepoDisabled 覆盖仓库不再启用休眠：即便仍在窗口内也唤醒。
func TestTasks_ReconcileHibernate_WakeRepoDisabled(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	repo, hr, hd, pr, rr := newHibernateTasks(m, hibernateNowIn())

	hr.EXPECT().ListEnabled(gomock.Any()).Return([]*biz.ProjectHibernate{
		{ProjectID: 9, Enabled: true, State: biz.HibernateStateSleeping,
			Snapshot:    []*websocket_pb.ExtraValue{{Path: "app->replicas", Value: "3"}},
			WindowNames: []string{"夜间"}},
	}, nil)
	proj := hibernateSleepingProject()
	pr.EXPECT().Show(gomock.Any(), 9).Return(proj, nil)
	rr.EXPECT().Get(gomock.Any(), 2).Return(&biz.Repo{ID: 2}, nil)

	var target []*websocket_pb.ExtraValue
	captureHibernateDeploy(hd, proj, &target)
	var rec *biz.ProjectHibernate
	captureHibernateUpsert(hr, &rec)

	assert.NoError(t, repo.ReconcileHibernate())
	assert.Equal(t, "3", target[0].Value)
	assert.Equal(t, biz.HibernateStateAwake, rec.State)
}

// TestTasks_ReconcileHibernate_WakeRepoCapabilityDisabled 覆盖仓库休眠配置仍在、但仓库级能力已被
// 关掉（Enabled=false）且项目仍标记休眠：视同无窗口——确在休眠即唤醒，不因配置残留而卡在休眠。
func TestTasks_ReconcileHibernate_WakeRepoCapabilityDisabled(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	repo, hr, hd, pr, rr := newHibernateTasks(m, hibernateNowIn())

	hr.EXPECT().ListEnabled(gomock.Any()).Return([]*biz.ProjectHibernate{
		{ProjectID: 9, Enabled: true, State: biz.HibernateStateSleeping,
			Snapshot:    []*websocket_pb.ExtraValue{{Path: "app->replicas", Value: "3"}},
			WindowNames: []string{"夜间"}},
	}, nil)
	proj := hibernateSleepingProject()
	pr.EXPECT().Show(gomock.Any(), 9).Return(proj, nil)
	rr.EXPECT().Get(gomock.Any(), 2).Return(&biz.Repo{ID: 2, MarsConfig: &mars.Config{
		Hibernate: &mars.Hibernate{Enabled: false, ReplicaPaths: []string{"app->replicas"}, WindowNames: []string{"夜间"}},
	}}, nil)

	var target []*websocket_pb.ExtraValue
	captureHibernateDeploy(hd, proj, &target)
	var rec *biz.ProjectHibernate
	captureHibernateUpsert(hr, &rec)

	assert.NoError(t, repo.ReconcileHibernate())
	assert.Equal(t, "3", target[0].Value)
	assert.Equal(t, biz.HibernateStateAwake, rec.State)
}

// TestTasks_ReconcileHibernate_ReleaseMismatchOutWindow 覆盖窗口外「标记休眠中但副本已被手动
// 扩容」（非全 0）：放弃按快照回填（否则会把人的意图覆盖回旧值），仅释放接管权——状态转 awake、
// 清空快照，保留用户改后的值，开关态原样保留。
func TestTasks_ReconcileHibernate_ReleaseMismatchOutWindow(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	repo, hr, _, pr, rr := newHibernateTasks(m, hibernateNowOut())

	hr.EXPECT().ListEnabled(gomock.Any()).Return([]*biz.ProjectHibernate{
		{ProjectID: 9, Enabled: true, State: biz.HibernateStateSleeping,
			Snapshot:    []*websocket_pb.ExtraValue{{Path: "app->replicas", Value: "3"}},
			WindowNames: []string{"夜间"}},
	}, nil)
	proj := &biz.Project{ID: 9, Name: "app", RepoID: 2, FinalExtraValues: []*websocket_pb.ExtraValue{
		{Path: "app->replicas", Value: "5"}, // 用户手动扩容后的值
		{Path: "app->image", Value: "nginx:1.0"},
	}}
	pr.EXPECT().Show(gomock.Any(), 9).Return(proj, nil)
	rr.EXPECT().Get(gomock.Any(), 2).Return(hibernateCfgRepo(2), nil)

	// 无 Redeploy 期望：释放接管权不重部署（误调即 panic）。
	var rec *biz.ProjectHibernate
	captureHibernateUpsert(hr, &rec)

	assert.NoError(t, repo.ReconcileHibernate())
	assert.Equal(t, biz.HibernateStateAwake, rec.State)
	assert.True(t, rec.Enabled)
	assert.Empty(t, rec.Snapshot)
}

// TestTasks_ReconcileHibernate_FixMismatchSameTick 覆盖窗口内「标记休眠中但副本已被手动扩容」
// （非全 0）的错位：同一轮即修正并重新休眠——快照当前值、置 0 重部署、状态仍 sleeping，
// 无需等下一轮对账（这是单趟对账相对旧「两趟」的关键改进）。
func TestTasks_ReconcileHibernate_FixMismatchSameTick(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	repo, hr, hd, pr, rr := newHibernateTasks(m, hibernateNowIn())

	hr.EXPECT().ListEnabled(gomock.Any()).Return([]*biz.ProjectHibernate{
		{ProjectID: 9, Enabled: true, State: biz.HibernateStateSleeping,
			Snapshot:    []*websocket_pb.ExtraValue{{Path: "app->replicas", Value: "3"}},
			WindowNames: []string{"夜间"}},
	}, nil)
	proj := &biz.Project{ID: 9, Name: "app", RepoID: 2, FinalExtraValues: []*websocket_pb.ExtraValue{
		{Path: "app->replicas", Value: "5"},
		{Path: "app->image", Value: "nginx:1.0"},
	}}
	pr.EXPECT().Show(gomock.Any(), 9).Return(proj, nil)
	rr.EXPECT().Get(gomock.Any(), 2).Return(hibernateCfgRepo(2), nil)

	var target []*websocket_pb.ExtraValue
	captureHibernateDeploy(hd, proj, &target)
	var rec *biz.ProjectHibernate
	captureHibernateUpsert(hr, &rec)

	assert.NoError(t, repo.ReconcileHibernate())
	assert.Equal(t, "0", target[0].Value) // 同轮置 0
	assert.Equal(t, "nginx:1.0", target[1].Value)
	assert.Equal(t, biz.HibernateStateSleeping, rec.State)
	// 快照刷新为当前值 5（不是旧快照 3）。
	assert.Len(t, rec.Snapshot, 1)
	assert.Equal(t, "5", rec.Snapshot[0].Value)
}

// TestTasks_ReconcileHibernate_ReleaseUpsertError 覆盖释放接管权落库失败（窗口外错位）：仅告警，
// 本轮无错，不重部署。
func TestTasks_ReconcileHibernate_ReleaseUpsertError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	repo, hr, _, pr, rr := newHibernateTasks(m, hibernateNowOut())

	hr.EXPECT().ListEnabled(gomock.Any()).Return([]*biz.ProjectHibernate{
		{ProjectID: 9, Enabled: true, State: biz.HibernateStateSleeping,
			Snapshot:    []*websocket_pb.ExtraValue{{Path: "app->replicas", Value: "3"}},
			WindowNames: []string{"夜间"}},
	}, nil)
	pr.EXPECT().Show(gomock.Any(), 9).Return(&biz.Project{ID: 9, Name: "app", RepoID: 2,
		FinalExtraValues: []*websocket_pb.ExtraValue{
			{Path: "app->replicas", Value: "5"},
		}}, nil)
	rr.EXPECT().Get(gomock.Any(), 2).Return(hibernateCfgRepo(2), nil)
	hr.EXPECT().Upsert(gomock.Any(), gomock.Any()).Return(nil, errors.New("upsert err"))

	assert.NoError(t, repo.ReconcileHibernate())
}

// TestTasks_ReconcileHibernate_ListEnabledError 覆盖启用列表查询失败：原样上抛。
func TestTasks_ReconcileHibernate_ListEnabledError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	repo, hr, _, _, _ := newHibernateTasks(m, hibernateNowIn())

	hr.EXPECT().ListEnabled(gomock.Any()).Return(nil, errors.New("enabled err"))

	assert.Equal(t, "enabled err", repo.ReconcileHibernate().Error())
}

// TestTasks_ReconcileHibernate_OrphanSelfHeal 覆盖项目已删（NotFound）：清理孤儿休眠记录，
// 不再重部署。
func TestTasks_ReconcileHibernate_OrphanSelfHeal(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	repo, hr, _, pr, _ := newHibernateTasks(m, hibernateNowIn())

	hr.EXPECT().ListEnabled(gomock.Any()).Return([]*biz.ProjectHibernate{
		{ProjectID: 9, Enabled: true, State: biz.HibernateStateAwake, WindowNames: []string{"夜间"}},
	}, nil)
	pr.EXPECT().Show(gomock.Any(), 9).Return(nil, errs.NotFound("project gone"))
	hr.EXPECT().Delete(gomock.Any(), 9).Return(nil)

	assert.NoError(t, repo.ReconcileHibernate())
}

// TestTasks_ReconcileHibernate_OrphanDeleteError 覆盖孤儿记录清理失败：仅告警，本轮无错。
func TestTasks_ReconcileHibernate_OrphanDeleteError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	repo, hr, _, pr, _ := newHibernateTasks(m, hibernateNowIn())

	hr.EXPECT().ListEnabled(gomock.Any()).Return([]*biz.ProjectHibernate{
		{ProjectID: 9, Enabled: true, State: biz.HibernateStateAwake, WindowNames: []string{"夜间"}},
	}, nil)
	pr.EXPECT().Show(gomock.Any(), 9).Return(nil, errs.NotFound("project gone"))
	hr.EXPECT().Delete(gomock.Any(), 9).Return(errors.New("del err"))

	assert.NoError(t, repo.ReconcileHibernate())
}

// TestTasks_ReconcileHibernate_ShowOtherErrorSkips 覆盖项目查询非 NotFound 错误：
// 跳过本轮（无 Delete 期望，反向证明未走孤儿清理）。
func TestTasks_ReconcileHibernate_ShowOtherErrorSkips(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	repo, hr, _, pr, _ := newHibernateTasks(m, hibernateNowIn())

	hr.EXPECT().ListEnabled(gomock.Any()).Return([]*biz.ProjectHibernate{
		{ProjectID: 9, Enabled: true, State: biz.HibernateStateAwake, WindowNames: []string{"夜间"}},
	}, nil)
	pr.EXPECT().Show(gomock.Any(), 9).Return(nil, errors.New("db down"))

	assert.NoError(t, repo.ReconcileHibernate())
}

// TestTasks_ReconcileHibernate_RepoGetErrorSkips 覆盖仓库查询失败：跳过本轮。
func TestTasks_ReconcileHibernate_RepoGetErrorSkips(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	repo, hr, _, pr, rr := newHibernateTasks(m, hibernateNowIn())

	hr.EXPECT().ListEnabled(gomock.Any()).Return([]*biz.ProjectHibernate{
		{ProjectID: 9, Enabled: true, State: biz.HibernateStateAwake, WindowNames: []string{"夜间"}},
	}, nil)
	pr.EXPECT().Show(gomock.Any(), 9).Return(hibernateProject(), nil)
	rr.EXPECT().Get(gomock.Any(), 2).Return(nil, errors.New("repo err"))

	assert.NoError(t, repo.ReconcileHibernate())
}

// TestTasks_ReconcileHibernate_SleepRedeployError 覆盖休眠「先落库成功、后重部署失败」：
// 状态与快照已落库（转 sleeping），仅记重部署告警、本轮无错。
func TestTasks_ReconcileHibernate_SleepRedeployError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	repo, hr, hd, pr, rr := newHibernateTasks(m, hibernateNowIn())

	hr.EXPECT().ListEnabled(gomock.Any()).Return([]*biz.ProjectHibernate{
		{ProjectID: 9, Enabled: true, State: biz.HibernateStateAwake, WindowNames: []string{"夜间"}},
	}, nil)
	proj := hibernateProject()
	pr.EXPECT().Show(gomock.Any(), 9).Return(proj, nil)
	rr.EXPECT().Get(gomock.Any(), 2).Return(hibernateCfgRepo(2), nil)
	// 先落「休眠态+快照」成功，再重部署失败（顺序见 SleepPersistsSnapshotBeforeDeploy）。
	hr.EXPECT().Upsert(gomock.Any(), gomock.Any()).Return(&biz.ProjectHibernate{}, nil)
	hd.EXPECT().Redeploy(gomock.Any(), proj, gomock.Any()).Return(errors.New("deploy err"))

	assert.NoError(t, repo.ReconcileHibernate())
}

// TestTasks_ReconcileHibernate_SleepUpsertError 覆盖休眠状态落库失败：落库失败即中止本轮、
// 不重部署（无 Redeploy 期望，误调即 panic），避免「部署已置 0 但快照未留」的丢数窗口。
func TestTasks_ReconcileHibernate_SleepUpsertError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	repo, hr, _, pr, rr := newHibernateTasks(m, hibernateNowIn())

	hr.EXPECT().ListEnabled(gomock.Any()).Return([]*biz.ProjectHibernate{
		{ProjectID: 9, Enabled: true, State: biz.HibernateStateAwake, WindowNames: []string{"夜间"}},
	}, nil)
	proj := hibernateProject()
	pr.EXPECT().Show(gomock.Any(), 9).Return(proj, nil)
	rr.EXPECT().Get(gomock.Any(), 2).Return(hibernateCfgRepo(2), nil)
	hr.EXPECT().Upsert(gomock.Any(), gomock.Any()).Return(nil, errors.New("upsert err"))

	assert.NoError(t, repo.ReconcileHibernate())
}

// TestTasks_ReconcileHibernate_WakeRedeployError 覆盖唤醒重部署失败：记日志不改状态。
func TestTasks_ReconcileHibernate_WakeRedeployError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	repo, hr, hd, pr, rr := newHibernateTasks(m, hibernateNowOut())

	hr.EXPECT().ListEnabled(gomock.Any()).Return([]*biz.ProjectHibernate{
		{ProjectID: 9, Enabled: true, State: biz.HibernateStateSleeping,
			Snapshot:    []*websocket_pb.ExtraValue{{Path: "app->replicas", Value: "3"}},
			WindowNames: []string{"夜间"}},
	}, nil)
	proj := hibernateSleepingProject()
	pr.EXPECT().Show(gomock.Any(), 9).Return(proj, nil)
	rr.EXPECT().Get(gomock.Any(), 2).Return(hibernateCfgRepo(2), nil)
	hd.EXPECT().Redeploy(gomock.Any(), proj, gomock.Any()).Return(errors.New("deploy err"))

	assert.NoError(t, repo.ReconcileHibernate())
}

// TestTasks_ReconcileHibernate_WakeUpsertError 覆盖唤醒状态落库失败：仅告警，本轮无错。
func TestTasks_ReconcileHibernate_WakeUpsertError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	repo, hr, hd, pr, rr := newHibernateTasks(m, hibernateNowOut())

	hr.EXPECT().ListEnabled(gomock.Any()).Return([]*biz.ProjectHibernate{
		{ProjectID: 9, Enabled: true, State: biz.HibernateStateSleeping,
			Snapshot:    []*websocket_pb.ExtraValue{{Path: "app->replicas", Value: "3"}},
			WindowNames: []string{"夜间"}},
	}, nil)
	proj := hibernateSleepingProject()
	pr.EXPECT().Show(gomock.Any(), 9).Return(proj, nil)
	rr.EXPECT().Get(gomock.Any(), 2).Return(hibernateCfgRepo(2), nil)
	hd.EXPECT().Redeploy(gomock.Any(), proj, gomock.Any()).Return(nil)
	hr.EXPECT().Upsert(gomock.Any(), gomock.Any()).Return(nil, errors.New("upsert err"))

	assert.NoError(t, repo.ReconcileHibernate())
}

// TestTasks_ReconcileHibernate_WakeOrphanSelfHeal 覆盖休眠态记录对应的项目已删：
// 唤醒轮同样触发孤儿清理并跳过（无 Redeploy 期望，反向证明未重部署）。
func TestTasks_ReconcileHibernate_WakeOrphanSelfHeal(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	repo, hr, _, pr, _ := newHibernateTasks(m, hibernateNowOut())

	hr.EXPECT().ListEnabled(gomock.Any()).Return([]*biz.ProjectHibernate{
		{ProjectID: 9, Enabled: true, State: biz.HibernateStateSleeping, WindowNames: []string{"夜间"}},
	}, nil)
	pr.EXPECT().Show(gomock.Any(), 9).Return(nil, errs.NotFound("project gone"))
	hr.EXPECT().Delete(gomock.Any(), 9).Return(nil)

	assert.NoError(t, repo.ReconcileHibernate())
}

// TestTasks_ReconcileHibernate_MultipleProjectsAllProcessed 覆盖有界并发消费：一趟内记录数
// 超过并发上限时全部仍被处理（各自唤醒重部署），证明 worker 池封顶后不遗漏、不串号。
func TestTasks_ReconcileHibernate_MultipleProjectsAllProcessed(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	repo, hr, hd, pr, rr := newHibernateTasks(m, hibernateNowOut())

	snap := []*websocket_pb.ExtraValue{{Path: "app->replicas", Value: "3"}}
	ids := []int{9, 10, 11, 12, 13, 14} // 6 条 > hibernateReconcileWorkers(5)，触发封顶分支
	records := make([]*biz.ProjectHibernate, 0, len(ids))
	for _, id := range ids {
		records = append(records, &biz.ProjectHibernate{ProjectID: id, Enabled: true, State: biz.HibernateStateSleeping, Snapshot: snap})
	}
	hr.EXPECT().ListEnabled(gomock.Any()).Return(records, nil)
	for _, id := range ids {
		pr.EXPECT().Show(gomock.Any(), id).Return(&biz.Project{ID: id, Name: "app", RepoID: 2, FinalExtraValues: []*websocket_pb.ExtraValue{
			{Path: "app->replicas", Value: "0"},
		}}, nil)
	}
	// 六个项目共用仓库 2，各自唤醒都要取一次仓库配置；六次唤醒、六次落库。
	rr.EXPECT().Get(gomock.Any(), 2).Return(hibernateCfgRepo(2), nil).Times(len(ids))
	hd.EXPECT().Redeploy(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(len(ids))
	hr.EXPECT().Upsert(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, h *biz.ProjectHibernate) (*biz.ProjectHibernate, error) { return h, nil }).Times(len(ids))

	assert.NoError(t, repo.ReconcileHibernate())
}

// TestTasks_ReconcileHibernate_ProjectPanicIsolated 覆盖单项目 panic 隔离：一个项目在加载期
// panic（模拟代码缺陷）只作废该项目，同趟其它项目照常处理，整趟不 panic 上抛（串行实现下
// panic 会顺循环上抛、令本轮其余项目全部跳过）。项目 10 的唤醒期望即反向证据。
func TestTasks_ReconcileHibernate_ProjectPanicIsolated(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	repo, hr, hd, pr, rr := newHibernateTasks(m, hibernateNowOut())

	snap := []*websocket_pb.ExtraValue{{Path: "app->replicas", Value: "3"}}
	hr.EXPECT().ListEnabled(gomock.Any()).Return([]*biz.ProjectHibernate{
		{ProjectID: 9, Enabled: true, State: biz.HibernateStateSleeping, Snapshot: snap},
		{ProjectID: 10, Enabled: true, State: biz.HibernateStateSleeping, Snapshot: snap},
	}, nil)
	pr.EXPECT().Show(gomock.Any(), 9).DoAndReturn(func(context.Context, int) (*biz.Project, error) {
		panic("boom")
	})
	pr.EXPECT().Show(gomock.Any(), 10).Return(&biz.Project{ID: 10, Name: "app", RepoID: 2, FinalExtraValues: []*websocket_pb.ExtraValue{
		{Path: "app->replicas", Value: "0"},
	}}, nil)
	rr.EXPECT().Get(gomock.Any(), 2).Return(hibernateCfgRepo(2), nil)
	hd.EXPECT().Redeploy(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	hr.EXPECT().Upsert(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, h *biz.ProjectHibernate) (*biz.ProjectHibernate, error) { return h, nil })

	assert.NoError(t, repo.ReconcileHibernate())
}

// TestTasks_ReconcileHibernate_NoProjects 覆盖无启用记录：直接返回、不起 worker。
// 无任何 Show/Redeploy/Upsert 期望，误调即 panic（反向证明空列表未触发下游端口）。
func TestTasks_ReconcileHibernate_NoProjects(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	repo, hr, _, _, _ := newHibernateTasks(m, hibernateNowOut())

	hr.EXPECT().ListEnabled(gomock.Any()).Return([]*biz.ProjectHibernate{}, nil)

	assert.NoError(t, repo.ReconcileHibernate())
}

// TestTasks_ReconcileHibernate_PanicIsolatedInDebugMode 是「worker 崩进程」的回归闸门：
// worker 是独立 goroutine，若用 logger.HandlePanic（debug 下把 panic 重新抛出）则单项目
// panic 会逃出 goroutine、直接崩掉整个进程。用 debug=true 的 logger 触发同款 panic，断言整趟
// 仍返回 nil——一旦回退到 HandlePanic，本测试会因进程崩溃而失败。
func TestTasks_ReconcileHibernate_PanicIsolatedInDebugMode(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	repo, hr, _, pr, _ := newHibernateTasks(m, hibernateNowOut())
	repo.logger = mlog.NewForConfig(&config.Config{LogChannel: "zap", Debug: true})

	hr.EXPECT().ListEnabled(gomock.Any()).Return([]*biz.ProjectHibernate{
		{ProjectID: 9, Enabled: true, State: biz.HibernateStateAwake, WindowNames: []string{"夜间"}},
	}, nil)
	pr.EXPECT().Show(gomock.Any(), 9).DoAndReturn(func(context.Context, int) (*biz.Project, error) {
		panic("boom-debug")
	})

	assert.NoError(t, repo.ReconcileHibernate())
}

// ---- agent 空闲关闭 ----

// newAgentIdleTasks 构造只挂 agent 空闲关闭所需端口的 Tasks，便于逐分支断言。
func newAgentIdleTasks(m *gomock.Controller, timeout time.Duration) (*Tasks, *data.MockNamespaceRepo, *data.MockProjectRepo, *data.MockK8sRepo, *biz.MockAgentActivity, *locker.MockLocker, *data.MockEventRepo) {
	nsRepo := data.NewMockNamespaceRepo(m)
	projRepo := data.NewMockProjectRepo(m)
	k8sRepo := data.NewMockK8sRepo(m)
	activity := biz.NewMockAgentActivity(m)
	lk := locker.NewMockLocker(m)
	eventRepo := data.NewMockEventRepo(m)
	return &Tasks{
		logger:        mlog.NewForConfig(nil),
		cfg:           &config.Config{AI: config.AIConfig{IdleTimeout: timeout}},
		nsRepo:        nsRepo,
		projectRepo:   projRepo,
		k8sRepo:       k8sRepo,
		agentActivity: activity,
		locker:        lk,
		event:         eventRepo,
	}, nsRepo, projRepo, k8sRepo, activity, lk, eventRepo
}

// expectAgentIdleLock 预置缩容临界区的项目级锁获取成功，键与部署全局锁一致。
func expectAgentIdleLock(lk *locker.MockLocker, namespaceID int, name string) {
	lk.EXPECT().
		RenewalAcquire(deploy.GetSlugName(namespaceID, name), deploy.LockTTLSeconds, deploy.LockRenewalSeconds).
		Return(func() {}, true)
}

// TestTasks_ReconcileAgentIdle_Disabled 空闲阈值 <= 0（显式关闭）时不触碰任何依赖。
func TestTasks_ReconcileAgentIdle_Disabled(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	repo, _, _, _, _, _, _ := newAgentIdleTasks(m, -time.Minute)
	assert.NoError(t, repo.ReconcileAgentIdle())
}

// TestTasks_ReconcileAgentIdle_NamespaceNotFound 空间不存在（AI 尚未启用）：静默返回。
func TestTasks_ReconcileAgentIdle_NamespaceNotFound(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	repo, nsRepo, _, _, _, _, _ := newAgentIdleTasks(m, time.Minute)
	nsRepo.EXPECT().FindByName(gomock.Any(), biz.AgentNamespace).
		Return(nil, errs.WrapNotFound(errors.New("no ns"), "find ns"))
	assert.NoError(t, repo.ReconcileAgentIdle())
}

// TestTasks_ReconcileAgentIdle_QueryErrors 空间/项目列表查询失败向上返回。
func TestTasks_ReconcileAgentIdle_QueryErrors(t *testing.T) {
	t.Run("空间查询失败", func(t *testing.T) {
		m := gomock.NewController(t)
		defer m.Finish()
		repo, nsRepo, _, _, _, _, _ := newAgentIdleTasks(m, time.Minute)
		nsRepo.EXPECT().FindByName(gomock.Any(), biz.AgentNamespace).Return(nil, errors.New("db down"))
		assert.EqualError(t, repo.ReconcileAgentIdle(), "db down")
	})
	t.Run("项目列表失败", func(t *testing.T) {
		m := gomock.NewController(t)
		defer m.Finish()
		repo, nsRepo, projRepo, _, _, _, _ := newAgentIdleTasks(m, time.Minute)
		nsRepo.EXPECT().FindByName(gomock.Any(), biz.AgentNamespace).Return(&biz.Namespace{ID: 3}, nil)
		projRepo.EXPECT().ListAllProjectBriefs(gomock.Any()).Return(nil, errors.New("db down"))
		assert.EqualError(t, repo.ReconcileAgentIdle(), "db down")
	})
}

// TestTasks_ReconcileAgentIdle_ScalesIdle 只处理 mars-agent 空间下 agent- 前缀的项目；
// 其余项目（空间不符/前缀不符/无空间）直接跳过。
func TestTasks_ReconcileAgentIdle_ScalesIdle(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	repo, nsRepo, projRepo, k8sRepo, activity, lk, eventRepo := newAgentIdleTasks(m, 30*time.Minute)
	ns := &biz.Namespace{ID: 3, Name: "prefix-mars-agent"}
	nsRepo.EXPECT().FindByName(gomock.Any(), biz.AgentNamespace).Return(ns, nil)
	projRepo.EXPECT().ListAllProjectBriefs(gomock.Any()).Return([]*biz.Project{
		{ID: 1, Name: "agent-other", Namespace: &biz.Namespace{Name: "another-ns"}},
		{ID: 2, Name: "normal-project", Namespace: ns},
		{ID: 3, Name: "agent-no-ns", Namespace: nil},
		{ID: 4, Name: "agent-idle", Namespace: ns},
	}, nil)
	// 缩容前会再复查一次活跃度，故两次都返回未活跃。
	activity.EXPECT().IsActive(gomock.Any(), "agent-idle").Return(false, nil).Times(2)
	name := biz.AgentDeploymentName("agent-idle")
	k8sRepo.EXPECT().GetDeployment(ns.Name, name).
		Return(&appsv1.Deployment{Spec: appsv1.DeploymentSpec{Replicas: lo.ToPtr(int32(1))}}, nil)
	k8sRepo.EXPECT().ScaleDeployment(gomock.Any(), ns.Name, name, int32(0)).Return(nil)
	expectAgentIdleLock(lk, ns.ID, "agent-idle")
	// 缩容成功必须留系统审计（与唤醒侧审计对称），且只记标识不记令牌等敏感内容。
	eventRepo.EXPECT().AuditLogWithRequest(
		types.EventActionType_Update, "system", "",
		"AI Agent 沙箱空闲缩容到 0",
		map[string]any{"project_id": 4, "project_name": "agent-idle"},
	)

	assert.NoError(t, repo.ReconcileAgentIdle())
}

// TestTasks_ReconcileAgentIdle_Skips 活跃中、活跃度查询失败、Deployment 已缩容/不存在/
// 查询失败、缩容失败都只跳过，不影响对账返回。
func TestTasks_ReconcileAgentIdle_Skips(t *testing.T) {
	cases := map[string]struct {
		setup func(*data.MockK8sRepo, *biz.MockAgentActivity, *locker.MockLocker)
	}{
		"活跃中": {
			setup: func(_ *data.MockK8sRepo, activity *biz.MockAgentActivity, _ *locker.MockLocker) {
				activity.EXPECT().IsActive(gomock.Any(), "agent-idle").Return(true, nil)
			},
		},
		"活跃度查询失败": {
			setup: func(_ *data.MockK8sRepo, activity *biz.MockAgentActivity, _ *locker.MockLocker) {
				activity.EXPECT().IsActive(gomock.Any(), "agent-idle").Return(false, errors.New("cache down"))
			},
		},
		"Deployment 不存在": {
			setup: func(k8sRepo *data.MockK8sRepo, activity *biz.MockAgentActivity, _ *locker.MockLocker) {
				activity.EXPECT().IsActive(gomock.Any(), "agent-idle").Return(false, nil)
				k8sRepo.EXPECT().GetDeployment(gomock.Any(), gomock.Any()).
					Return(nil, errs.WrapNotFound(errors.New("not found"), "get deployment"))
			},
		},
		"Deployment 查询失败": {
			setup: func(k8sRepo *data.MockK8sRepo, activity *biz.MockAgentActivity, _ *locker.MockLocker) {
				activity.EXPECT().IsActive(gomock.Any(), "agent-idle").Return(false, nil)
				k8sRepo.EXPECT().GetDeployment(gomock.Any(), gomock.Any()).Return(nil, errors.New("k8s down"))
			},
		},
		"已缩容": {
			setup: func(k8sRepo *data.MockK8sRepo, activity *biz.MockAgentActivity, _ *locker.MockLocker) {
				activity.EXPECT().IsActive(gomock.Any(), "agent-idle").Return(false, nil)
				k8sRepo.EXPECT().GetDeployment(gomock.Any(), gomock.Any()).
					Return(&appsv1.Deployment{Spec: appsv1.DeploymentSpec{Replicas: lo.ToPtr(int32(0))}}, nil)
			},
		},
		"抢不到锁": {
			// 同一项目正在唤醒/部署：直接跳过，不再复查活跃度也不缩容。
			setup: func(k8sRepo *data.MockK8sRepo, activity *biz.MockAgentActivity, lk *locker.MockLocker) {
				activity.EXPECT().IsActive(gomock.Any(), "agent-idle").Return(false, nil)
				k8sRepo.EXPECT().GetDeployment(gomock.Any(), gomock.Any()).Return(&appsv1.Deployment{}, nil)
				lk.EXPECT().
					RenewalAcquire(deploy.GetSlugName(3, "agent-idle"), deploy.LockTTLSeconds, deploy.LockRenewalSeconds).
					Return(nil, false)
			},
		},
		"缩容失败": {
			setup: func(k8sRepo *data.MockK8sRepo, activity *biz.MockAgentActivity, lk *locker.MockLocker) {
				activity.EXPECT().IsActive(gomock.Any(), "agent-idle").Return(false, nil).Times(2)
				k8sRepo.EXPECT().GetDeployment(gomock.Any(), gomock.Any()).Return(&appsv1.Deployment{}, nil)
				k8sRepo.EXPECT().ScaleDeployment(gomock.Any(), gomock.Any(), gomock.Any(), int32(0)).
					Return(errors.New("scale failed"))
				expectAgentIdleLock(lk, 3, "agent-idle")
			},
		},
		"缩容前被唤醒": {
			setup: func(k8sRepo *data.MockK8sRepo, activity *biz.MockAgentActivity, lk *locker.MockLocker) {
				gomock.InOrder(
					activity.EXPECT().IsActive(gomock.Any(), "agent-idle").Return(false, nil),
					activity.EXPECT().IsActive(gomock.Any(), "agent-idle").Return(true, nil),
				)
				k8sRepo.EXPECT().GetDeployment(gomock.Any(), gomock.Any()).Return(&appsv1.Deployment{}, nil)
				expectAgentIdleLock(lk, 3, "agent-idle")
			},
		},
		"缩容前复查失败": {
			setup: func(k8sRepo *data.MockK8sRepo, activity *biz.MockAgentActivity, lk *locker.MockLocker) {
				gomock.InOrder(
					activity.EXPECT().IsActive(gomock.Any(), "agent-idle").Return(false, nil),
					activity.EXPECT().IsActive(gomock.Any(), "agent-idle").Return(false, errors.New("cache down")),
				)
				k8sRepo.EXPECT().GetDeployment(gomock.Any(), gomock.Any()).Return(&appsv1.Deployment{}, nil)
				expectAgentIdleLock(lk, 3, "agent-idle")
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			m := gomock.NewController(t)
			defer m.Finish()
			repo, nsRepo, projRepo, k8sRepo, activity, lk, _ := newAgentIdleTasks(m, time.Minute)
			ns := &biz.Namespace{ID: 3, Name: "prefix-mars-agent"}
			nsRepo.EXPECT().FindByName(gomock.Any(), biz.AgentNamespace).Return(ns, nil)
			projRepo.EXPECT().ListAllProjectBriefs(gomock.Any()).
				Return([]*biz.Project{{ID: 4, Name: "agent-idle", Namespace: ns}}, nil)
			tc.setup(k8sRepo, activity, lk)
			assert.NoError(t, repo.ReconcileAgentIdle())
		})
	}
}
