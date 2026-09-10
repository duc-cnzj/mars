package biz

import (
	"context"
	"errors"
	"testing"

	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"helm.sh/helm/v3/pkg/storage/driver"
)

// 以下 stub 只覆盖 DeployBiz.DeleteProject 用到的接口方法，其余由嵌入接口兜底。

type fakeHelmerRepo struct {
	HelmerRepo
	uninstall func(releaseName, namespace string, log LogFn) error
}

func (f *fakeHelmerRepo) Uninstall(releaseName, namespace string, log LogFn) error {
	return f.uninstall(releaseName, namespace, log)
}

type fakeProjectRepoForDeploy struct {
	ProjectRepo
	delete func(ctx context.Context, id int) error
}

func (f *fakeProjectRepoForDeploy) Delete(ctx context.Context, id int) error {
	return f.delete(ctx, id)
}

type fakeEventRepoForDeploy struct {
	EventRepo
	dispatch func(created EventKey, createdData any)
}

func (f *fakeEventRepoForDeploy) Dispatch(created EventKey, createdData any) {
	f.dispatch(created, createdData)
}

var deployTestProj = &Project{ID: 2, Name: "app", NamespaceID: 3, Namespace: &Namespace{Name: "ns"}}

func deployBizForTest(proj ProjectRepo, helmer HelmerRepo, event EventRepo) DeployBiz {
	return NewDeployBiz(mlog.NewForConfig(nil), proj, helmer, event)
}

func TestDeployBiz_DeleteProject_HappyPath(t *testing.T) {
	var deletedID int
	var dispatched EventKey
	helmer := &fakeHelmerRepo{uninstall: func(releaseName, namespace string, log LogFn) error {
		assert.Equal(t, "app", releaseName)
		assert.Equal(t, "ns", namespace)
		return nil
	}}
	proj := &fakeProjectRepoForDeploy{delete: func(ctx context.Context, id int) error {
		deletedID = id
		return nil
	}}
	event := &fakeEventRepoForDeploy{dispatch: func(created EventKey, createdData any) { dispatched = created }}
	d := deployBizForTest(proj, helmer, event)

	err := d.DeleteProject(context.TODO(), 1, deployTestProj, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, deletedID)
	assert.Equal(t, EventProjectDeleted, dispatched)
}

func TestDeployBiz_DeleteProject_UninstallNotFoundStillDeletes(t *testing.T) {
	// release 已不存在（手动清理/孤儿）不算失败，继续删 DB + 派发事件，
	// 避免把已经没 release 的项目锁死无法删除。
	var deleted bool
	helmer := &fakeHelmerRepo{uninstall: func(releaseName, namespace string, log LogFn) error {
		return driver.ErrReleaseNotFound
	}}
	proj := &fakeProjectRepoForDeploy{delete: func(ctx context.Context, id int) error { deleted = true; return nil }}
	event := &fakeEventRepoForDeploy{dispatch: func(created EventKey, createdData any) {}}
	d := deployBizForTest(proj, helmer, event)

	err := d.DeleteProject(context.TODO(), 1, deployTestProj, nil)
	assert.Nil(t, err)
	assert.True(t, deleted)
}

func TestDeployBiz_DeleteProject_UninstallErrorAborts(t *testing.T) {
	// 回归防护：卸载 release 失败（非 not-found）必须中止删除、保留 DB 记录，
	// 否则会留下无记录、无法重试的孤儿 release。
	var deleted, dispatched bool
	helmer := &fakeHelmerRepo{uninstall: func(releaseName, namespace string, log LogFn) error {
		return errors.New("uninstall boom")
	}}
	proj := &fakeProjectRepoForDeploy{delete: func(ctx context.Context, id int) error { deleted = true; return nil }}
	event := &fakeEventRepoForDeploy{dispatch: func(created EventKey, createdData any) { dispatched = true }}
	d := deployBizForTest(proj, helmer, event)

	err := d.DeleteProject(context.TODO(), 1, deployTestProj, nil)
	assert.Error(t, err)
	assert.False(t, deleted)
	assert.False(t, dispatched)
}

func TestDeployBiz_DeleteProject_DeleteErrorNoDispatch(t *testing.T) {
	// DB 删除失败时不派发项目删除事件（事件代表删除已生效）。
	var dispatched bool
	helmer := &fakeHelmerRepo{uninstall: func(releaseName, namespace string, log LogFn) error { return nil }}
	proj := &fakeProjectRepoForDeploy{delete: func(ctx context.Context, id int) error { return errors.New("db down") }}
	event := &fakeEventRepoForDeploy{dispatch: func(created EventKey, createdData any) { dispatched = true }}
	d := deployBizForTest(proj, helmer, event)

	err := d.DeleteProject(context.TODO(), 1, deployTestProj, nil)
	assert.Error(t, err)
	assert.False(t, dispatched)
}

func TestDeployBiz_DeleteProject_NilProject(t *testing.T) {
	d := deployBizForTest(nil, nil, nil)
	err := d.DeleteProject(context.TODO(), 1, nil, nil)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, "project 不能为空", status.Convert(err).Message())
}
