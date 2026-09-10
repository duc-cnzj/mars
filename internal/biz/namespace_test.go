package biz

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/pagination"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// 以下 stub 只覆盖 NamespaceBiz 用到的接口方法，其余由嵌入接口兜底。

type fakeNamespaceRepoForNSBiz struct {
	NamespaceRepo
	getMars       func(name string) string
	findByName    func(ctx context.Context, name string) (*Namespace, error)
	create        func(ctx context.Context, input *CreateNamespaceInput) (*Namespace, error)
	delete        func(ctx context.Context, id int) error
	favorite      func(ctx context.Context, input *FavoriteNamespaceInput) error
	favoriteSort  func(ctx context.Context, email string, firstID, secondID int) error
	list          func(ctx context.Context, input *ListNamespaceInput) ([]*Namespace, *pagination.Pagination, error)
	listAll       func(ctx context.Context) ([]*Namespace, error)
	listAdminPage func(ctx context.Context, query *AdminListPageQuery) (*AdminListPageResult, error)
	update        func(ctx context.Context, input *UpdateNamespaceInput) (*Namespace, error)
	show          func(ctx context.Context, id int) (*Namespace, error)
	syncMembers   func(ctx context.Context, namespaceID int, memberEmails []string) (*Namespace, error)
	updatePrivate func(ctx context.Context, namespaceID int, private bool) (*Namespace, error)
	transfer      func(ctx context.Context, id int, email string) (*Namespace, error)
	updateConfig  func(ctx context.Context, input *UpdateConfigInput) (*Namespace, error)
}

func (f *fakeNamespaceRepoForNSBiz) GetMarsNamespace(name string) string { return f.getMars(name) }
func (f *fakeNamespaceRepoForNSBiz) FindByName(ctx context.Context, name string) (*Namespace, error) {
	return f.findByName(ctx, name)
}
func (f *fakeNamespaceRepoForNSBiz) Create(ctx context.Context, input *CreateNamespaceInput) (*Namespace, error) {
	return f.create(ctx, input)
}
func (f *fakeNamespaceRepoForNSBiz) Delete(ctx context.Context, id int) error {
	return f.delete(ctx, id)
}
func (f *fakeNamespaceRepoForNSBiz) Favorite(ctx context.Context, input *FavoriteNamespaceInput) error {
	return f.favorite(ctx, input)
}
func (f *fakeNamespaceRepoForNSBiz) FavoriteSort(ctx context.Context, email string, firstID, secondID int) error {
	return f.favoriteSort(ctx, email, firstID, secondID)
}
func (f *fakeNamespaceRepoForNSBiz) List(ctx context.Context, input *ListNamespaceInput) ([]*Namespace, *pagination.Pagination, error) {
	return f.list(ctx, input)
}
func (f *fakeNamespaceRepoForNSBiz) ListAll(ctx context.Context) ([]*Namespace, error) {
	return f.listAll(ctx)
}
func (f *fakeNamespaceRepoForNSBiz) ListAdminPage(ctx context.Context, query *AdminListPageQuery) (*AdminListPageResult, error) {
	return f.listAdminPage(ctx, query)
}
func (f *fakeNamespaceRepoForNSBiz) Update(ctx context.Context, input *UpdateNamespaceInput) (*Namespace, error) {
	return f.update(ctx, input)
}
func (f *fakeNamespaceRepoForNSBiz) Show(ctx context.Context, id int) (*Namespace, error) {
	return f.show(ctx, id)
}
func (f *fakeNamespaceRepoForNSBiz) SyncMembers(ctx context.Context, namespaceID int, memberEmails []string) (*Namespace, error) {
	return f.syncMembers(ctx, namespaceID, memberEmails)
}
func (f *fakeNamespaceRepoForNSBiz) UpdatePrivate(ctx context.Context, namespaceID int, private bool) (*Namespace, error) {
	return f.updatePrivate(ctx, namespaceID, private)
}
func (f *fakeNamespaceRepoForNSBiz) Transfer(ctx context.Context, id int, email string) (*Namespace, error) {
	return f.transfer(ctx, id, email)
}
func (f *fakeNamespaceRepoForNSBiz) UpdateConfig(ctx context.Context, input *UpdateConfigInput) (*Namespace, error) {
	return f.updateConfig(ctx, input)
}

type fakeK8sRepoForNSBiz struct {
	K8sRepo
	createNamespace    func(ctx context.Context, name string) (*corev1.Namespace, error)
	getNamespace       func(ctx context.Context, name string) (*corev1.Namespace, error)
	createDockerSecret func(ctx context.Context, namespace string) (*corev1.Secret, error)
	deleteNamespace    func(ctx context.Context, name string) error
	deleteSecret       func(ctx context.Context, namespace, secret string) error
}

func (f *fakeK8sRepoForNSBiz) CreateNamespace(ctx context.Context, name string) (*corev1.Namespace, error) {
	return f.createNamespace(ctx, name)
}
func (f *fakeK8sRepoForNSBiz) GetNamespace(ctx context.Context, name string) (*corev1.Namespace, error) {
	return f.getNamespace(ctx, name)
}
func (f *fakeK8sRepoForNSBiz) CreateDockerSecret(ctx context.Context, namespace string) (*corev1.Secret, error) {
	return f.createDockerSecret(ctx, namespace)
}
func (f *fakeK8sRepoForNSBiz) DeleteNamespace(ctx context.Context, name string) error {
	return f.deleteNamespace(ctx, name)
}
func (f *fakeK8sRepoForNSBiz) DeleteSecret(ctx context.Context, namespace, secret string) error {
	return f.deleteSecret(ctx, namespace, secret)
}

type fakeHelmerRepoForNSBiz struct {
	HelmerRepo
	uninstall func(releaseName, namespace string, log LogFn) error
}

func (f *fakeHelmerRepoForNSBiz) Uninstall(releaseName, namespace string, log LogFn) error {
	return f.uninstall(releaseName, namespace, log)
}

type fakeEventRepoForNSBiz struct {
	EventRepo
	dispatch func(created EventKey, createdData any)
}

func (f *fakeEventRepoForNSBiz) Dispatch(created EventKey, createdData any) {
	f.dispatch(created, createdData)
}

// notFoundErr 是 biz 层业务 NotFound（errs.WrapNotFound()），用于 FindByName 预查等
// 走 errs.IsNotFound 判定的路径。
func notFoundErr() error {
	return errs.WrapNotFound(errors.New("not found"), "not found")
}

// k8sNotFoundErr 是 k8s API 层 NotFound（StatusReasonNotFound），Delete 的
// DeleteNamespace/轮询 GetNamespace 用 k8sapierrors.IsNotFound 判定，必须用真 StatusError。
func k8sNotFoundErr() error {
	return &k8sapierrors.StatusError{ErrStatus: metav1.Status{Reason: metav1.StatusReasonNotFound}}
}

func alreadyExistsErr() error {
	return &k8sapierrors.StatusError{ErrStatus: metav1.Status{Reason: metav1.StatusReasonAlreadyExists}}
}

// nsBizForTest 组装 NamespaceBiz 测试实例，依赖由测试注入的 fake 替身。
func nsBizForTest(ns NamespaceRepo, k8s K8sRepo, helmer HelmerRepo, event EventRepo) NamespaceBiz {
	return NewNamespaceBiz(mlog.NewForConfig(nil), ns, k8s, helmer, event)
}

// withFastDeletePolling 把轮询参数压到 100ms/1ms，避免 Delete 轮询测试支付真实墙钟。
func withFastDeletePolling(t *testing.T) {
	t.Helper()
	oldTimeout, oldInterval := NamespaceDeleteTimeout, NamespacePollInterval
	NamespaceDeleteTimeout = 100 * time.Millisecond
	NamespacePollInterval = time.Millisecond
	t.Cleanup(func() {
		NamespaceDeleteTimeout, NamespacePollInterval = oldTimeout, oldInterval
	})
}

// ---- Create ----

func TestNamespaceBiz_Create_FindByNameNonNotFoundError(t *testing.T) {
	ns := &fakeNamespaceRepoForNSBiz{
		getMars:    func(name string) string { return name },
		findByName: func(ctx context.Context, name string) (*Namespace, error) { return nil, errors.New("db down") },
	}
	n := nsBizForTest(ns, &fakeK8sRepoForNSBiz{}, &fakeHelmerRepoForNSBiz{}, &fakeEventRepoForNSBiz{})

	got, exists, err := n.Create(context.TODO(), "test", "desc", "a@b.c")
	assert.Nil(t, got)
	assert.False(t, exists)
	assert.Error(t, err)
}

func TestNamespaceBiz_Create_ExistsReturnsExisting(t *testing.T) {
	existing := &Namespace{ID: 7, Name: "mars-test"}
	ns := &fakeNamespaceRepoForNSBiz{
		getMars:    func(name string) string { return name },
		findByName: func(ctx context.Context, name string) (*Namespace, error) { return existing, nil },
	}
	n := nsBizForTest(ns, &fakeK8sRepoForNSBiz{}, &fakeHelmerRepoForNSBiz{}, &fakeEventRepoForNSBiz{})

	got, exists, err := n.Create(context.TODO(), "test", "desc", "a@b.c")
	assert.Nil(t, err)
	assert.True(t, exists)
	assert.Equal(t, existing, got)
}

func TestNamespaceBiz_Create_K8sCreateNonAlreadyExistsError(t *testing.T) {
	ns := &fakeNamespaceRepoForNSBiz{
		getMars:    func(name string) string { return name },
		findByName: func(ctx context.Context, name string) (*Namespace, error) { return nil, notFoundErr() },
	}
	k8s := &fakeK8sRepoForNSBiz{
		createNamespace: func(ctx context.Context, name string) (*corev1.Namespace, error) {
			return nil, errors.New("k8s boom")
		},
	}
	n := nsBizForTest(ns, k8s, &fakeHelmerRepoForNSBiz{}, &fakeEventRepoForNSBiz{})

	got, exists, err := n.Create(context.TODO(), "test", "desc", "a@b.c")
	assert.Nil(t, got)
	assert.False(t, exists)
	assert.Error(t, err)
}

func TestNamespaceBiz_Create_AlreadyExistsGetNamespaceError(t *testing.T) {
	ns := &fakeNamespaceRepoForNSBiz{
		getMars:    func(name string) string { return name },
		findByName: func(ctx context.Context, name string) (*Namespace, error) { return nil, notFoundErr() },
	}
	k8s := &fakeK8sRepoForNSBiz{
		createNamespace: func(ctx context.Context, name string) (*corev1.Namespace, error) {
			return nil, alreadyExistsErr()
		},
		getNamespace: func(ctx context.Context, name string) (*corev1.Namespace, error) {
			return nil, errors.New("get boom")
		},
	}
	n := nsBizForTest(ns, k8s, &fakeHelmerRepoForNSBiz{}, &fakeEventRepoForNSBiz{})

	got, exists, err := n.Create(context.TODO(), "test", "desc", "a@b.c")
	assert.Nil(t, got)
	assert.False(t, exists)
	assert.Error(t, err)
}

func TestNamespaceBiz_Create_AlreadyExistsTerminating(t *testing.T) {
	ns := &fakeNamespaceRepoForNSBiz{
		getMars:    func(name string) string { return name },
		findByName: func(ctx context.Context, name string) (*Namespace, error) { return nil, notFoundErr() },
	}
	k8s := &fakeK8sRepoForNSBiz{
		createNamespace: func(ctx context.Context, name string) (*corev1.Namespace, error) {
			return nil, alreadyExistsErr()
		},
		getNamespace: func(ctx context.Context, name string) (*corev1.Namespace, error) {
			return &corev1.Namespace{Status: corev1.NamespaceStatus{Phase: corev1.NamespaceTerminating}}, nil
		},
	}
	n := nsBizForTest(ns, k8s, &fakeHelmerRepoForNSBiz{}, &fakeEventRepoForNSBiz{})

	got, exists, err := n.Create(context.TODO(), "test", "desc", "a@b.c")
	assert.Nil(t, got)
	assert.False(t, exists)
	assert.ErrorIs(t, err, ErrNamespaceTerminating)
}

func TestNamespaceBiz_Create_AdoptExistingHappy(t *testing.T) {
	found := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "mars-adopted"}}
	nsRepo := &fakeNamespaceRepoForNSBiz{
		getMars:    func(name string) string { return name },
		findByName: func(ctx context.Context, name string) (*Namespace, error) { return nil, notFoundErr() },
		create: func(ctx context.Context, input *CreateNamespaceInput) (*Namespace, error) {
			assert.Equal(t, "mars-adopted", input.Name)
			assert.Equal(t, "a@b.c", input.CreatorEmail)
			return &Namespace{ID: 1, Name: input.Name}, nil
		},
		favorite: func(ctx context.Context, input *FavoriteNamespaceInput) error { return nil },
	}
	k8s := &fakeK8sRepoForNSBiz{
		createNamespace: func(ctx context.Context, name string) (*corev1.Namespace, error) {
			return nil, alreadyExistsErr()
		},
		getNamespace: func(ctx context.Context, name string) (*corev1.Namespace, error) { return found, nil },
		createDockerSecret: func(ctx context.Context, namespace string) (*corev1.Secret, error) {
			return &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "secret-1"}}, nil
		},
	}
	var dispatched EventKey
	var createdData NamespaceCreatedData
	event := &fakeEventRepoForNSBiz{dispatch: func(created EventKey, data any) {
		dispatched = created
		createdData, _ = data.(NamespaceCreatedData)
		assert.Equal(t, found, createdData.NsK8sObj)
	}}
	n := nsBizForTest(nsRepo, k8s, &fakeHelmerRepoForNSBiz{}, event)

	got, exists, err := n.Create(context.TODO(), "test", "desc", "a@b.c")
	assert.Nil(t, err)
	assert.False(t, exists)
	assert.Equal(t, 1, got.ID)
	assert.Equal(t, EventNamespaceCreated, dispatched)
}

func TestNamespaceBiz_Create_DockerSecretDegrade(t *testing.T) {
	nsRepo := &fakeNamespaceRepoForNSBiz{
		getMars:    func(name string) string { return name },
		findByName: func(ctx context.Context, name string) (*Namespace, error) { return nil, notFoundErr() },
		create: func(ctx context.Context, input *CreateNamespaceInput) (*Namespace, error) {
			// 降级后 ImagePullSecrets 为空，不影响创建。
			assert.Empty(t, input.ImagePullSecrets)
			return &Namespace{ID: 1, Name: input.Name}, nil
		},
		favorite: func(ctx context.Context, input *FavoriteNamespaceInput) error { return nil },
	}
	k8s := &fakeK8sRepoForNSBiz{
		createNamespace: func(ctx context.Context, name string) (*corev1.Namespace, error) {
			return &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "mars-test"}}, nil
		},
		createDockerSecret: func(ctx context.Context, namespace string) (*corev1.Secret, error) {
			return nil, errors.New("secret boom")
		},
	}
	event := &fakeEventRepoForNSBiz{dispatch: func(created EventKey, createdData any) {}}
	n := nsBizForTest(nsRepo, k8s, &fakeHelmerRepoForNSBiz{}, event)

	got, exists, err := n.Create(context.TODO(), "test", "desc", "a@b.c")
	assert.Nil(t, err)
	assert.False(t, exists)
	assert.Equal(t, 1, got.ID)
}

func TestNamespaceBiz_Create_DbCreateErrorRollsBackCreatedNamespace(t *testing.T) {
	nsRepo := &fakeNamespaceRepoForNSBiz{
		getMars:    func(name string) string { return name },
		findByName: func(ctx context.Context, name string) (*Namespace, error) { return nil, notFoundErr() },
		create: func(ctx context.Context, input *CreateNamespaceInput) (*Namespace, error) {
			return nil, errors.New("db down")
		},
	}
	var deleted string
	k8s := &fakeK8sRepoForNSBiz{
		createNamespace: func(ctx context.Context, name string) (*corev1.Namespace, error) {
			return &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "mars-test"}}, nil
		},
		createDockerSecret: func(ctx context.Context, namespace string) (*corev1.Secret, error) {
			return &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "s"}}, nil
		},
		deleteNamespace: func(ctx context.Context, name string) error { deleted = name; return nil },
	}
	n := nsBizForTest(nsRepo, k8s, &fakeHelmerRepoForNSBiz{}, &fakeEventRepoForNSBiz{})

	got, exists, err := n.Create(context.TODO(), "test", "desc", "a@b.c")
	assert.Nil(t, got)
	assert.False(t, exists)
	assert.Error(t, err)
	// 本请求新建（createdByUs=true）失败必须回滚 k8s namespace，避免孤儿资源。
	assert.Equal(t, "mars-test", deleted)
}

func TestNamespaceBiz_Create_DbCreateErrorAdoptedNoRollback(t *testing.T) {
	nsRepo := &fakeNamespaceRepoForNSBiz{
		getMars:    func(name string) string { return name },
		findByName: func(ctx context.Context, name string) (*Namespace, error) { return nil, notFoundErr() },
		create: func(ctx context.Context, input *CreateNamespaceInput) (*Namespace, error) {
			return nil, errors.New("db down")
		},
	}
	var deleted bool
	k8s := &fakeK8sRepoForNSBiz{
		createNamespace: func(ctx context.Context, name string) (*corev1.Namespace, error) {
			return nil, alreadyExistsErr()
		},
		getNamespace: func(ctx context.Context, name string) (*corev1.Namespace, error) {
			return &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "mars-existing"}}, nil
		},
		// 收养路径也必经 CreateDockerSecret，给一个成功返回值即可（DB 失败在 create 处返回）。
		createDockerSecret: func(ctx context.Context, namespace string) (*corev1.Secret, error) {
			return &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "s"}}, nil
		},
		deleteNamespace: func(ctx context.Context, name string) error { deleted = true; return nil },
	}
	n := nsBizForTest(nsRepo, k8s, &fakeHelmerRepoForNSBiz{}, &fakeEventRepoForNSBiz{})

	got, exists, err := n.Create(context.TODO(), "test", "desc", "a@b.c")
	assert.Nil(t, got)
	assert.False(t, exists)
	assert.Error(t, err)
	// 收养已存在的 k8s namespace（createdByUs=false）不得回滚删除。
	assert.False(t, deleted)
}

func TestNamespaceBiz_Create_RollbackErrorLogged(t *testing.T) {
	nsRepo := &fakeNamespaceRepoForNSBiz{
		getMars:    func(name string) string { return name },
		findByName: func(ctx context.Context, name string) (*Namespace, error) { return nil, notFoundErr() },
		create: func(ctx context.Context, input *CreateNamespaceInput) (*Namespace, error) {
			return nil, errors.New("db down")
		},
	}
	k8s := &fakeK8sRepoForNSBiz{
		createNamespace: func(ctx context.Context, name string) (*corev1.Namespace, error) {
			return &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "mars-test"}}, nil
		},
		createDockerSecret: func(ctx context.Context, namespace string) (*corev1.Secret, error) {
			return &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "s"}}, nil
		},
		deleteNamespace: func(ctx context.Context, name string) error { return errors.New("rollback boom") },
	}
	n := nsBizForTest(nsRepo, k8s, &fakeHelmerRepoForNSBiz{}, &fakeEventRepoForNSBiz{})

	_, _, err := n.Create(context.TODO(), "test", "desc", "a@b.c")
	assert.Error(t, err)
}

func TestNamespaceBiz_Create_FavoriteDegrade(t *testing.T) {
	nsRepo := &fakeNamespaceRepoForNSBiz{
		getMars:    func(name string) string { return name },
		findByName: func(ctx context.Context, name string) (*Namespace, error) { return nil, notFoundErr() },
		create: func(ctx context.Context, input *CreateNamespaceInput) (*Namespace, error) {
			return &Namespace{ID: 1, Name: input.Name}, nil
		},
		favorite: func(ctx context.Context, input *FavoriteNamespaceInput) error {
			return errors.New("favorite boom")
		},
	}
	k8s := &fakeK8sRepoForNSBiz{
		createNamespace: func(ctx context.Context, name string) (*corev1.Namespace, error) {
			return &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "mars-test"}}, nil
		},
		createDockerSecret: func(ctx context.Context, namespace string) (*corev1.Secret, error) {
			return &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "s"}}, nil
		},
	}
	var dispatched bool
	event := &fakeEventRepoForNSBiz{dispatch: func(created EventKey, createdData any) { dispatched = true }}
	n := nsBizForTest(nsRepo, k8s, &fakeHelmerRepoForNSBiz{}, event)

	got, exists, err := n.Create(context.TODO(), "test", "desc", "a@b.c")
	assert.Nil(t, err)
	assert.False(t, exists)
	assert.Equal(t, 1, got.ID)
	// 自动关注失败仅降级，不阻断创建结果与事件派发。
	assert.True(t, dispatched)
}

// ---- Delete ----

func nsForDelete() *Namespace {
	return &Namespace{
		ID:               1,
		Name:             "ns-1",
		ImagePullSecrets: []string{"sec-1", "sec-2"},
		Projects:         []*Project{{Name: "app-1"}, {Name: "app-2"}},
	}
}

func TestNamespaceBiz_Delete_HappyPath(t *testing.T) {
	withFastDeletePolling(t)
	// 卸载是并发 goroutine，共享切片追加必须加锁（race 检测器会抓裸 append）。
	var mu sync.Mutex
	var uninstalled []string
	helmer := &fakeHelmerRepoForNSBiz{uninstall: func(releaseName, namespace string, log LogFn) error {
		assert.Equal(t, "ns-1", namespace)
		mu.Lock()
		uninstalled = append(uninstalled, releaseName)
		mu.Unlock()
		return nil
	}}
	var deletedSecrets []string
	k8s := &fakeK8sRepoForNSBiz{
		deleteSecret: func(ctx context.Context, namespace, secret string) error {
			assert.Equal(t, "ns-1", namespace)
			deletedSecrets = append(deletedSecrets, secret)
			return nil
		},
		deleteNamespace: func(ctx context.Context, name string) error {
			assert.Equal(t, "ns-1", name)
			return nil
		},
		getNamespace: func(ctx context.Context, name string) (*corev1.Namespace, error) {
			return nil, k8sNotFoundErr()
		},
	}
	var deletedID int
	nsRepo := &fakeNamespaceRepoForNSBiz{delete: func(ctx context.Context, id int) error {
		deletedID = id
		return nil
	}}
	var dispatched EventKey
	var deletedData NamespaceDeletedData
	event := &fakeEventRepoForNSBiz{dispatch: func(created EventKey, createdData any) {
		dispatched = created
		d, _ := createdData.(NamespaceDeletedData)
		deletedData = d
	}}
	n := nsBizForTest(nsRepo, k8s, helmer, event)

	names, err := n.Delete(context.TODO(), nsForDelete())
	assert.Nil(t, err)
	assert.Equal(t, []string{"app-1", "app-2"}, names)
	// 并发卸载 goroutine 完成顺序不定，只断言集合相等。
	assert.ElementsMatch(t, []string{"app-1", "app-2"}, uninstalled)
	assert.Equal(t, []string{"sec-1", "sec-2"}, deletedSecrets)
	assert.Equal(t, 1, deletedID)
	assert.Equal(t, EventNamespaceDeleted, dispatched)
	assert.Equal(t, 1, deletedData.ID)
}

func TestNamespaceBiz_Delete_UninstallErrorContinue(t *testing.T) {
	withFastDeletePolling(t)
	helmer := &fakeHelmerRepoForNSBiz{uninstall: func(releaseName, namespace string, log LogFn) error {
		return errors.New("uninstall boom")
	}}
	k8s := &fakeK8sRepoForNSBiz{
		deleteSecret:    func(ctx context.Context, namespace, secret string) error { return nil },
		deleteNamespace: func(ctx context.Context, name string) error { return nil },
		getNamespace:    func(ctx context.Context, name string) (*corev1.Namespace, error) { return nil, k8sNotFoundErr() },
	}
	nsRepo := &fakeNamespaceRepoForNSBiz{delete: func(ctx context.Context, id int) error { return nil }}
	event := &fakeEventRepoForNSBiz{dispatch: func(created EventKey, createdData any) {}}
	n := nsBizForTest(nsRepo, k8s, helmer, event)

	// 单个 release 卸载失败仅记日志，不阻断整条删除链路。
	names, err := n.Delete(context.TODO(), nsForDelete())
	assert.Nil(t, err)
	assert.Equal(t, []string{"app-1", "app-2"}, names)
}

func TestNamespaceBiz_Delete_DeleteSecretErrorContinue(t *testing.T) {
	withFastDeletePolling(t)
	helmer := &fakeHelmerRepoForNSBiz{uninstall: func(releaseName, namespace string, log LogFn) error { return nil }}
	k8s := &fakeK8sRepoForNSBiz{
		deleteSecret:    func(ctx context.Context, namespace, secret string) error { return errors.New("secret boom") },
		deleteNamespace: func(ctx context.Context, name string) error { return nil },
		getNamespace:    func(ctx context.Context, name string) (*corev1.Namespace, error) { return nil, k8sNotFoundErr() },
	}
	nsRepo := &fakeNamespaceRepoForNSBiz{delete: func(ctx context.Context, id int) error { return nil }}
	event := &fakeEventRepoForNSBiz{dispatch: func(created EventKey, createdData any) {}}
	n := nsBizForTest(nsRepo, k8s, helmer, event)

	names, err := n.Delete(context.TODO(), nsForDelete())
	assert.Nil(t, err)
	assert.Equal(t, []string{"app-1", "app-2"}, names)
}

func TestNamespaceBiz_Delete_DeleteNamespaceNonNotFoundAborts(t *testing.T) {
	helmer := &fakeHelmerRepoForNSBiz{uninstall: func(releaseName, namespace string, log LogFn) error { return nil }}
	var dbDeleted bool
	k8s := &fakeK8sRepoForNSBiz{
		deleteSecret:    func(ctx context.Context, namespace, secret string) error { return nil },
		deleteNamespace: func(ctx context.Context, name string) error { return errors.New("ns boom") },
	}
	nsRepo := &fakeNamespaceRepoForNSBiz{delete: func(ctx context.Context, id int) error {
		dbDeleted = true
		return nil
	}}
	var dispatched bool
	event := &fakeEventRepoForNSBiz{dispatch: func(created EventKey, createdData any) { dispatched = true }}
	n := nsBizForTest(nsRepo, k8s, helmer, event)

	// k8s namespace 未真正删除（非 NotFound）时必须中止，不得删 DB 记录/派发事件。
	names, err := n.Delete(context.TODO(), nsForDelete())
	assert.Nil(t, names)
	assert.Error(t, err)
	assert.False(t, dbDeleted)
	assert.False(t, dispatched)
}

func TestNamespaceBiz_Delete_DeleteNamespaceNotFoundIsClean(t *testing.T) {
	withFastDeletePolling(t)
	helmer := &fakeHelmerRepoForNSBiz{uninstall: func(releaseName, namespace string, log LogFn) error { return nil }}
	k8s := &fakeK8sRepoForNSBiz{
		deleteSecret:    func(ctx context.Context, namespace, secret string) error { return nil },
		deleteNamespace: func(ctx context.Context, name string) error { return k8sNotFoundErr() },
		getNamespace:    func(ctx context.Context, name string) (*corev1.Namespace, error) { return nil, k8sNotFoundErr() },
	}
	nsRepo := &fakeNamespaceRepoForNSBiz{delete: func(ctx context.Context, id int) error { return nil }}
	var dispatched bool
	event := &fakeEventRepoForNSBiz{dispatch: func(created EventKey, createdData any) { dispatched = true }}
	n := nsBizForTest(nsRepo, k8s, helmer, event)

	names, err := n.Delete(context.TODO(), nsForDelete())
	assert.Nil(t, err)
	assert.Equal(t, []string{"app-1", "app-2"}, names)
	assert.True(t, dispatched)
}

func TestNamespaceBiz_Delete_DbDeleteErrorAborts(t *testing.T) {
	helmer := &fakeHelmerRepoForNSBiz{uninstall: func(releaseName, namespace string, log LogFn) error { return nil }}
	k8s := &fakeK8sRepoForNSBiz{
		deleteSecret:    func(ctx context.Context, namespace, secret string) error { return nil },
		deleteNamespace: func(ctx context.Context, name string) error { return nil },
	}
	nsRepo := &fakeNamespaceRepoForNSBiz{delete: func(ctx context.Context, id int) error { return errors.New("db down") }}
	var dispatched bool
	event := &fakeEventRepoForNSBiz{dispatch: func(created EventKey, createdData any) { dispatched = true }}
	n := nsBizForTest(nsRepo, k8s, helmer, event)

	names, err := n.Delete(context.TODO(), nsForDelete())
	assert.Nil(t, names)
	assert.Error(t, err)
	assert.False(t, dispatched)
}

func TestNamespaceBiz_Delete_PollTransientThenNotFound(t *testing.T) {
	withFastDeletePolling(t)
	helmer := &fakeHelmerRepoForNSBiz{uninstall: func(releaseName, namespace string, log LogFn) error { return nil }}
	// 先返回瞬态错误（API 抖动），再 NotFound：瞬态错误必须继续轮询而不是提前 break。
	calls := 0
	k8s := &fakeK8sRepoForNSBiz{
		deleteSecret:    func(ctx context.Context, namespace, secret string) error { return nil },
		deleteNamespace: func(ctx context.Context, name string) error { return nil },
		getNamespace: func(ctx context.Context, name string) (*corev1.Namespace, error) {
			calls++
			if calls == 1 {
				return nil, errors.New("transient api error")
			}
			return nil, k8sNotFoundErr()
		},
	}
	nsRepo := &fakeNamespaceRepoForNSBiz{delete: func(ctx context.Context, id int) error { return nil }}
	var dispatched bool
	event := &fakeEventRepoForNSBiz{dispatch: func(created EventKey, createdData any) { dispatched = true }}
	n := nsBizForTest(nsRepo, k8s, helmer, event)

	names, err := n.Delete(context.TODO(), nsForDelete())
	assert.Nil(t, err)
	assert.Equal(t, []string{"app-1", "app-2"}, names)
	assert.True(t, dispatched)
	assert.GreaterOrEqual(t, calls, 2)
}

func TestNamespaceBiz_Delete_PollTimeoutFallback(t *testing.T) {
	withFastDeletePolling(t)
	helmer := &fakeHelmerRepoForNSBiz{uninstall: func(releaseName, namespace string, log LogFn) error { return nil }}
	// GetNamespace 一直返回存在（nil error）：轮询直至超时兜底 break，仍派发删除事件。
	k8s := &fakeK8sRepoForNSBiz{
		deleteSecret:    func(ctx context.Context, namespace, secret string) error { return nil },
		deleteNamespace: func(ctx context.Context, name string) error { return nil },
		getNamespace:    func(ctx context.Context, name string) (*corev1.Namespace, error) { return &corev1.Namespace{}, nil },
	}
	nsRepo := &fakeNamespaceRepoForNSBiz{delete: func(ctx context.Context, id int) error { return nil }}
	var dispatched bool
	event := &fakeEventRepoForNSBiz{dispatch: func(created EventKey, createdData any) { dispatched = true }}
	n := nsBizForTest(nsRepo, k8s, helmer, event)

	names, err := n.Delete(context.TODO(), nsForDelete())
	assert.Nil(t, err)
	assert.Equal(t, []string{"app-1", "app-2"}, names)
	assert.True(t, dispatched)
}

// ---- CRUD 门面（输入校验 + 透传）----

// nsFacadeForTest 组装仅测 CRUD 门面的 NamespaceBiz：门面方法只触达 nsRepo，
// k8s/helm/event 传空替身（嵌入接口零实现，被调用即 panic，确保门面不透传越界）。
func nsFacadeForTest(ns NamespaceRepo) NamespaceBiz {
	return nsBizForTest(ns, &fakeK8sRepoForNSBiz{}, &fakeHelmerRepoForNSBiz{}, &fakeEventRepoForNSBiz{})
}

func TestNamespaceBiz_Update_InvalidID(t *testing.T) {
	n := nsFacadeForTest(&fakeNamespaceRepoForNSBiz{})
	got, err := n.Update(context.TODO(), &UpdateNamespaceInput{ID: 0})
	assert.Nil(t, got)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, "namespace 不能为空或 id 不能小于等于 0", status.Convert(err).Message())
}

func TestNamespaceBiz_Update_Valid(t *testing.T) {
	var gotInput *UpdateNamespaceInput
	ns := &fakeNamespaceRepoForNSBiz{update: func(ctx context.Context, input *UpdateNamespaceInput) (*Namespace, error) {
		gotInput = input
		return &Namespace{ID: input.ID}, nil
	}}
	n := nsFacadeForTest(ns)
	got, err := n.Update(context.TODO(), &UpdateNamespaceInput{ID: 1})
	assert.NoError(t, err)
	assert.Equal(t, 1, got.ID)
	assert.Equal(t, 1, gotInput.ID)
}

func TestNamespaceBiz_Favorite_InvalidNamespaceID(t *testing.T) {
	n := nsFacadeForTest(&fakeNamespaceRepoForNSBiz{})
	err := n.Favorite(context.TODO(), &FavoriteNamespaceInput{NamespaceID: 0, UserEmail: "a@b.c", Favorite: true})
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, "namespace 不能为空或 id 不能小于等于 0", status.Convert(err).Message())
}

func TestNamespaceBiz_Favorite_Valid(t *testing.T) {
	var gotInput *FavoriteNamespaceInput
	ns := &fakeNamespaceRepoForNSBiz{favorite: func(ctx context.Context, input *FavoriteNamespaceInput) error {
		gotInput = input
		return nil
	}}
	n := nsFacadeForTest(ns)
	err := n.Favorite(context.TODO(), &FavoriteNamespaceInput{NamespaceID: 1, UserEmail: "a@b.c", Favorite: true})
	assert.NoError(t, err)
	assert.Equal(t, 1, gotInput.NamespaceID)
}

func TestNamespaceBiz_FavoriteSort_EmptyList(t *testing.T) {
	n := nsFacadeForTest(&fakeNamespaceRepoForNSBiz{})
	// nil 输入：两个空间 id 必须报 400
	err := n.FavoriteSort(context.TODO(), nil)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, "两个空间 id 必须大于 0", status.Convert(err).Message())
	// id 缺省同样拒绝
	err = n.FavoriteSort(context.TODO(), &FavoriteSortNamespaceInput{UserEmail: "a@b.c"})
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	err = n.FavoriteSort(context.TODO(), &FavoriteSortNamespaceInput{UserEmail: "a@b.c", FirstID: 1})
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestNamespaceBiz_FavoriteSort_Valid(t *testing.T) {
	var gotEmail string
	var gotFirst, gotSecond int
	ns := &fakeNamespaceRepoForNSBiz{favoriteSort: func(ctx context.Context, email string, firstID, secondID int) error {
		gotEmail = email
		gotFirst = firstID
		gotSecond = secondID
		return nil
	}}
	n := nsFacadeForTest(ns)
	err := n.FavoriteSort(context.TODO(), &FavoriteSortNamespaceInput{
		UserEmail: "a@b.c",
		FirstID:   3,
		SecondID:  1,
	})
	assert.NoError(t, err)
	assert.Equal(t, "a@b.c", gotEmail)
	assert.Equal(t, 3, gotFirst)
	assert.Equal(t, 1, gotSecond)
}

func TestNamespaceBiz_SyncMembers_InvalidNamespaceID(t *testing.T) {
	n := nsFacadeForTest(&fakeNamespaceRepoForNSBiz{})
	got, err := n.SyncMembers(context.TODO(), 0, nil)
	assert.Nil(t, got)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, "namespace id 不能小于等于 0", status.Convert(err).Message())
}

func TestNamespaceBiz_SyncMembers_Valid(t *testing.T) {
	var gotID int
	var gotEmails []string
	ns := &fakeNamespaceRepoForNSBiz{syncMembers: func(ctx context.Context, namespaceID int, memberEmails []string) (*Namespace, error) {
		gotID = namespaceID
		gotEmails = memberEmails
		return &Namespace{ID: namespaceID}, nil
	}}
	n := nsFacadeForTest(ns)
	got, err := n.SyncMembers(context.TODO(), 1, []string{"a@b.c"})
	assert.NoError(t, err)
	assert.Equal(t, 1, got.ID)
	assert.Equal(t, 1, gotID)
	assert.Equal(t, []string{"a@b.c"}, gotEmails)
}

func TestNamespaceBiz_UpdatePrivate_InvalidNamespaceID(t *testing.T) {
	n := nsFacadeForTest(&fakeNamespaceRepoForNSBiz{})
	got, err := n.UpdatePrivate(context.TODO(), 0, true)
	assert.Nil(t, got)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, "namespace id 不能小于等于 0", status.Convert(err).Message())
}

func TestNamespaceBiz_UpdatePrivate_Valid(t *testing.T) {
	var gotID int
	var gotPrivate bool
	ns := &fakeNamespaceRepoForNSBiz{updatePrivate: func(ctx context.Context, namespaceID int, private bool) (*Namespace, error) {
		gotID = namespaceID
		gotPrivate = private
		return &Namespace{ID: namespaceID}, nil
	}}
	n := nsFacadeForTest(ns)
	got, err := n.UpdatePrivate(context.TODO(), 1, true)
	assert.NoError(t, err)
	assert.Equal(t, 1, got.ID)
	assert.Equal(t, 1, gotID)
	assert.True(t, gotPrivate)
}

func TestNamespaceBiz_Transfer_InvalidID(t *testing.T) {
	n := nsFacadeForTest(&fakeNamespaceRepoForNSBiz{})
	got, err := n.Transfer(context.TODO(), 0, "a@b.c")
	assert.Nil(t, got)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, "namespace id 不能小于等于 0", status.Convert(err).Message())
}

func TestNamespaceBiz_Transfer_EmptyEmail(t *testing.T) {
	n := nsFacadeForTest(&fakeNamespaceRepoForNSBiz{})
	got, err := n.Transfer(context.TODO(), 1, "")
	assert.Nil(t, got)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, "transfer email 不能为空", status.Convert(err).Message())
}

func TestNamespaceBiz_Transfer_Valid(t *testing.T) {
	var gotID int
	var gotEmail string
	ns := &fakeNamespaceRepoForNSBiz{transfer: func(ctx context.Context, id int, email string) (*Namespace, error) {
		gotID = id
		gotEmail = email
		return &Namespace{ID: id}, nil
	}}
	n := nsFacadeForTest(ns)
	got, err := n.Transfer(context.TODO(), 1, "a@b.c")
	assert.NoError(t, err)
	assert.Equal(t, 1, got.ID)
	assert.Equal(t, 1, gotID)
	assert.Equal(t, "a@b.c", gotEmail)
}

func TestNamespaceBiz_UpdateConfig_InvalidID(t *testing.T) {
	n := nsFacadeForTest(&fakeNamespaceRepoForNSBiz{})
	got, err := n.UpdateConfig(context.TODO(), &UpdateConfigInput{ID: 0})
	assert.Nil(t, got)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, "namespace id 不能小于等于 0", status.Convert(err).Message())
}

func TestNamespaceBiz_UpdateConfig_Valid(t *testing.T) {
	var gotInput *UpdateConfigInput
	ns := &fakeNamespaceRepoForNSBiz{updateConfig: func(ctx context.Context, input *UpdateConfigInput) (*Namespace, error) {
		gotInput = input
		return &Namespace{ID: input.ID}, nil
	}}
	n := nsFacadeForTest(ns)
	got, err := n.UpdateConfig(context.TODO(), &UpdateConfigInput{ID: 1})
	assert.NoError(t, err)
	assert.Equal(t, 1, got.ID)
	assert.Equal(t, 1, gotInput.ID)
}

func TestNamespaceBiz_UpdateConfig_RepoError(t *testing.T) {
	ns := &fakeNamespaceRepoForNSBiz{updateConfig: func(ctx context.Context, input *UpdateConfigInput) (*Namespace, error) {
		return nil, errors.New("db down")
	}}
	n := nsFacadeForTest(ns)
	got, err := n.UpdateConfig(context.TODO(), &UpdateConfigInput{ID: 1})
	assert.Nil(t, got)
	assert.Equal(t, "db down", err.Error())
}

// ---- 纯透传查询 ----

func TestNamespaceBiz_List_Passthrough(t *testing.T) {
	var gotInput *ListNamespaceInput
	ns := &fakeNamespaceRepoForNSBiz{list: func(ctx context.Context, input *ListNamespaceInput) ([]*Namespace, *pagination.Pagination, error) {
		gotInput = input
		return []*Namespace{{ID: 1}}, nil, nil
	}}
	n := nsFacadeForTest(ns)
	got, pag, err := n.List(context.TODO(), &ListNamespaceInput{})
	assert.NoError(t, err)
	assert.Len(t, got, 1)
	assert.Nil(t, pag)
	assert.NotNil(t, gotInput)
}

// TestNamespaceBiz_ListAllNames 成功路径：ListAll 全量空间名抽成 k8s 名称切片。
func TestNamespaceBiz_ListAllNames(t *testing.T) {
	ns := &fakeNamespaceRepoForNSBiz{listAll: func(ctx context.Context) ([]*Namespace, error) {
		return []*Namespace{{Name: "mars-dev"}, {Name: "mars-prod"}}, nil
	}}
	n := nsFacadeForTest(ns)
	got, err := n.ListAllNames(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, []string{"mars-dev", "mars-prod"}, got)
}

// TestNamespaceBiz_ListAllNames_RepoError 失败路径：ListAll 报错时原样上抛。
func TestNamespaceBiz_ListAllNames_RepoError(t *testing.T) {
	ns := &fakeNamespaceRepoForNSBiz{listAll: func(ctx context.Context) ([]*Namespace, error) {
		return nil, errors.New("db down")
	}}
	n := nsFacadeForTest(ns)
	got, err := n.ListAllNames(context.TODO())
	assert.Nil(t, got)
	assert.Error(t, err)
}

// TestNamespaceBiz_AdminList 管理列表成功路径：ListAdminPage（搜索/私有/分页/Now 透传）返回
// 已分页+已分类+已统计结果，biz 按已加载的边计算最近活跃时间与行级活跃度分类，统计/计数原样透传。
func TestNamespaceBiz_AdminList(t *testing.T) {
	now := time.Now()
	var gotQuery *AdminListPageQuery
	ns := &fakeNamespaceRepoForNSBiz{listAdminPage: func(ctx context.Context, query *AdminListPageQuery) (*AdminListPageResult, error) {
		gotQuery = query
		return &AdminListPageResult{
			Namespaces: []*Namespace{
				// 无项目：从未活跃 → 僵尸
				{ID: 1, Name: "mars-a"},
				// 项目 10 天前更新：活跃
				{ID: 2, Name: "mars-b", Projects: []*Project{{UpdatedAt: now.Add(-10 * 24 * time.Hour)}}},
				// 项目 120 天前更新：僵尸
				{ID: 3, Name: "mars-c", Projects: []*Project{{UpdatedAt: now.Add(-120 * 24 * time.Hour)}}},
			},
			Count: 3,
			Stats: AdminLivenessStats{Total: 3, Active: 1, Zombie: 2},
		}, nil
	}}
	n := nsBizForTest(ns, &fakeK8sRepoForNSBiz{}, &fakeHelmerRepoForNSBiz{}, &fakeEventRepoForNSBiz{})

	items, stats, pag, err := n.AdminList(context.TODO(), &AdminListInput{Page: 1, PageSize: 15, Search: "mars", PrivateOnly: true})
	assert.NoError(t, err)
	// 查询参数透传：Search/PrivateOnly/分页 + 分类基准 Now（非零）
	if assert.NotNil(t, gotQuery) {
		assert.Equal(t, "mars", gotQuery.Search)
		assert.True(t, gotQuery.PrivateOnly)
		assert.Equal(t, int32(1), gotQuery.Page)
		assert.Equal(t, int32(15), gotQuery.PageSize)
		assert.False(t, gotQuery.Now.IsZero())
	}
	assert.Equal(t, int32(3), pag.Count)
	assert.Equal(t, AdminLivenessStats{Total: 3, Active: 1, Zombie: 2}, *stats)
	if assert.Len(t, items, 3) {
		assert.Equal(t, 1, items[0].Namespace.ID)
		// 无项目：最近活跃为零值，分类僵尸
		assert.True(t, items[0].LastActiveAt.IsZero())
		assert.Equal(t, LivenessZombie, items[0].LivenessKind)
		assert.Equal(t, 2, items[1].Namespace.ID)
		// 带项目：最近活跃 = 最大 UpdatedAt，分类活跃
		assert.Equal(t, now.Add(-10*24*time.Hour).Truncate(time.Second), items[1].LastActiveAt.Truncate(time.Second))
		assert.Equal(t, LivenessActive, items[1].LivenessKind)
		assert.Equal(t, 3, items[2].Namespace.ID)
		assert.Equal(t, LivenessZombie, items[2].LivenessKind)
	}
}

// TestNamespaceBiz_AdminList_FilterByKind 分类过滤：Liveness 参数透传，条目/计数/统计由 repo 决定原样透传。
func TestNamespaceBiz_AdminList_FilterByKind(t *testing.T) {
	now := time.Now()
	var gotQuery *AdminListPageQuery
	ns := &fakeNamespaceRepoForNSBiz{listAdminPage: func(ctx context.Context, query *AdminListPageQuery) (*AdminListPageResult, error) {
		gotQuery = query
		return &AdminListPageResult{
			Namespaces: []*Namespace{
				{ID: 2, Name: "b", Projects: []*Project{{UpdatedAt: now.Add(-60 * 24 * time.Hour)}}}, // 低活跃
			},
			Count: 1,
			Stats: AdminLivenessStats{Total: 3, Active: 1, Dormant: 1, Zombie: 1},
		}, nil
	}}
	n := nsBizForTest(ns, &fakeK8sRepoForNSBiz{}, &fakeHelmerRepoForNSBiz{}, &fakeEventRepoForNSBiz{})

	items, stats, pag, err := n.AdminList(context.TODO(), &AdminListInput{Page: 1, PageSize: 10, Liveness: "dormant"})
	assert.NoError(t, err)
	// 分类参数透传给 repo（分类过滤由 repo 下沉 SQL）
	if assert.NotNil(t, gotQuery) {
		assert.Equal(t, "dormant", gotQuery.Liveness)
	}
	// 统计不随分类过滤裁剪（repo 返回 search 命中全量口径）
	assert.Equal(t, AdminLivenessStats{Total: 3, Active: 1, Dormant: 1, Zombie: 1}, *stats)
	assert.Equal(t, int32(1), pag.Count)
	if assert.Len(t, items, 1) {
		assert.Equal(t, 2, items[0].Namespace.ID)
		assert.Equal(t, LivenessDormant, items[0].LivenessKind)
	}
}

// TestNamespaceBiz_AdminList_Pagination 分页：Page/PageSize 透传，条目/计数由 repo 决定原样透传。
func TestNamespaceBiz_AdminList_Pagination(t *testing.T) {
	now := time.Now()
	var gotQuery *AdminListPageQuery
	var list []*Namespace
	for i := 1; i <= 4; i++ {
		list = append(list, &Namespace{ID: i, Name: "ns", Projects: []*Project{{UpdatedAt: now.Add(-10 * 24 * time.Hour)}}})
	}
	ns := &fakeNamespaceRepoForNSBiz{listAdminPage: func(ctx context.Context, query *AdminListPageQuery) (*AdminListPageResult, error) {
		gotQuery = query
		return &AdminListPageResult{Namespaces: list[2:4], Count: 4}, nil
	}}
	n := nsBizForTest(ns, &fakeK8sRepoForNSBiz{}, &fakeHelmerRepoForNSBiz{}, &fakeEventRepoForNSBiz{})

	items, _, pag, err := n.AdminList(context.TODO(), &AdminListInput{Page: 2, PageSize: 2})
	assert.NoError(t, err)
	if assert.NotNil(t, gotQuery) {
		assert.Equal(t, int32(2), gotQuery.Page)
		assert.Equal(t, int32(2), gotQuery.PageSize)
	}
	assert.Equal(t, int32(4), pag.Count)
	if assert.Len(t, items, 2) {
		assert.Equal(t, 3, items[0].Namespace.ID)
		assert.Equal(t, 4, items[1].Namespace.ID)
	}
}

// TestNamespaceBiz_AdminList_PaginationOutOfRange 越界页：条目为空但计数保留全量（由 repo 返回）。
func TestNamespaceBiz_AdminList_PaginationOutOfRange(t *testing.T) {
	ns := &fakeNamespaceRepoForNSBiz{listAdminPage: func(ctx context.Context, query *AdminListPageQuery) (*AdminListPageResult, error) {
		return &AdminListPageResult{Namespaces: nil, Count: 2}, nil
	}}
	n := nsBizForTest(ns, &fakeK8sRepoForNSBiz{}, &fakeHelmerRepoForNSBiz{}, &fakeEventRepoForNSBiz{})

	items, _, pag, err := n.AdminList(context.TODO(), &AdminListInput{Page: 99, PageSize: 10})
	assert.NoError(t, err)
	assert.Empty(t, items)
	assert.Equal(t, int32(2), pag.Count)
}

// Test_FormatResourceUsage 用量格式化：CPU 固定 "%d m"，内存走 humanize.Bytes（十进制基数），
// 零值输出 "0 m"/"0 B"。
func Test_FormatResourceUsage(t *testing.T) {
	cpu, mem := FormatResourceUsage(100, 500000000)
	assert.Equal(t, "100 m", cpu)
	assert.Equal(t, "500 MB", mem)
	cpu, mem = FormatResourceUsage(0, 0)
	assert.Equal(t, "0 m", cpu)
	assert.Equal(t, "0 B", mem)
}

// Test_lastActiveAt 最近活跃时间取命名空间下所有项目 UpdatedAt 的最大值：
// 无项目返回零值；有项目取最大；nil 项目指针跳过不 panic。
func Test_lastActiveAt(t *testing.T) {
	base := time.Date(2026, 8, 20, 10, 0, 0, 0, time.UTC)
	cases := []struct {
		name     string
		projects []*Project
		want     time.Time
	}{
		{name: "无项目返回零值", projects: nil, want: time.Time{}},
		{name: "空列表返回零值", projects: []*Project{}, want: time.Time{}},
		{name: "nil 项目跳过返回零值", projects: []*Project{nil}, want: time.Time{}},
		{
			name:     "单项目取自身 UpdatedAt",
			projects: []*Project{{UpdatedAt: base}},
			want:     base,
		},
		{
			name: "多项目取最大 UpdatedAt",
			projects: []*Project{
				{UpdatedAt: base.Add(-2 * time.Hour)},
				{UpdatedAt: base.Add(3 * time.Hour)},
				{UpdatedAt: base.Add(-1 * time.Hour)},
			},
			want: base.Add(3 * time.Hour),
		},
		{
			name: "nil 项目混入不影响最大值",
			projects: []*Project{
				nil,
				{UpdatedAt: base.Add(time.Hour)},
				nil,
				{UpdatedAt: base.Add(-time.Hour)},
			},
			want: base.Add(time.Hour),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, lastActiveAt(&Namespace{Projects: tc.projects}))
		})
	}
}

// TestNamespaceBiz_AdminList_RepoError 管理列表失败路径：ListAdminPage 查询错误原样上抛。
func TestNamespaceBiz_AdminList_RepoError(t *testing.T) {
	ns := &fakeNamespaceRepoForNSBiz{listAdminPage: func(ctx context.Context, query *AdminListPageQuery) (*AdminListPageResult, error) {
		return nil, errors.New("boom")
	}}
	n := nsBizForTest(ns, &fakeK8sRepoForNSBiz{}, &fakeHelmerRepoForNSBiz{}, &fakeEventRepoForNSBiz{})

	items, stats, pag, err := n.AdminList(context.TODO(), &AdminListInput{})
	assert.Nil(t, items)
	assert.Nil(t, stats)
	assert.Nil(t, pag)
	assert.Error(t, err)
}

func TestNamespaceBiz_Show_Passthrough(t *testing.T) {
	var gotID int
	ns := &fakeNamespaceRepoForNSBiz{show: func(ctx context.Context, id int) (*Namespace, error) {
		gotID = id
		return &Namespace{ID: id}, nil
	}}
	n := nsFacadeForTest(ns)
	got, err := n.Show(context.TODO(), 3)
	assert.NoError(t, err)
	assert.Equal(t, 3, got.ID)
	assert.Equal(t, 3, gotID)
}

func TestNamespaceBiz_GetMarsNamespace_Passthrough(t *testing.T) {
	ns := &fakeNamespaceRepoForNSBiz{getMars: func(name string) string { return "mars-" + name }}
	n := nsFacadeForTest(ns)
	assert.Equal(t, "mars-ns", n.GetMarsNamespace("ns"))
}

func TestNamespaceBiz_FindByName_Passthrough(t *testing.T) {
	var gotName string
	ns := &fakeNamespaceRepoForNSBiz{findByName: func(ctx context.Context, name string) (*Namespace, error) {
		gotName = name
		return &Namespace{Name: name}, nil
	}}
	n := nsFacadeForTest(ns)
	got, err := n.FindByName(context.TODO(), "ns")
	assert.NoError(t, err)
	assert.Equal(t, "ns", got.Name)
	assert.Equal(t, "ns", gotName)
}
