package biz

import (
	"context"
	"testing"

	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/release"
)

// fakeHelmerRepoForHelmerBiz 记录各写操作是否被调用，输入校验测试中 repo 不被调用（调用即 panic）。
type fakeHelmerRepoForHelmerBiz struct {
	HelmerRepo
	upgradeCalled, rollbackCalled, uninstallCalled bool
	statusCalled, packageCalled                    bool
}

func (f *fakeHelmerRepoForHelmerBiz) UpgradeOrInstall(ctx context.Context, releaseName, namespace string, ch *chart.Chart, valueOpts *values.Options, fn WrapLogFn, wait bool, timeoutSeconds int64, dryRun bool, desc string) (*release.Release, error) {
	f.upgradeCalled = true
	return &release.Release{Name: releaseName}, nil
}

func (f *fakeHelmerRepoForHelmerBiz) Rollback(releaseName, namespace string, wait bool, log LogFn, dryRun bool) error {
	f.rollbackCalled = true
	return nil
}

func (f *fakeHelmerRepoForHelmerBiz) Uninstall(releaseName, namespace string, log LogFn) error {
	f.uninstallCalled = true
	return nil
}

func (f *fakeHelmerRepoForHelmerBiz) ReleaseStatus(releaseName, namespace string) types.Deploy {
	f.statusCalled = true
	return types.Deploy_StatusDeploying
}

func (f *fakeHelmerRepoForHelmerBiz) PackageChart(path string, destDir string) (string, error) {
	f.packageCalled = true
	return destDir + "/app.tgz", nil
}

func newHelmerBizForTest(repo HelmerRepo) HelmerBiz {
	return NewHelmerBiz(repo)
}

func TestHelmerBiz_UpgradeOrInstall_EmptyReleaseName(t *testing.T) {
	h := newHelmerBizForTest(&fakeHelmerRepoForHelmerBiz{})
	got, err := h.UpgradeOrInstall(context.TODO(), "", "", nil, nil, nil, false, 0, false, "")
	assert.Nil(t, got)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, "releaseName 或 namespace 不能为空", status.Convert(err).Message())
}

func TestHelmerBiz_UpgradeOrInstall_Valid(t *testing.T) {
	f := &fakeHelmerRepoForHelmerBiz{}
	h := newHelmerBizForTest(f)
	got, err := h.UpgradeOrInstall(context.TODO(), "app", "ns", nil, nil, nil, false, 0, false, "")
	assert.NoError(t, err)
	assert.True(t, f.upgradeCalled)
	assert.Equal(t, "app", got.Name)
}

func TestHelmerBiz_Rollback_EmptyReleaseName(t *testing.T) {
	h := newHelmerBizForTest(&fakeHelmerRepoForHelmerBiz{})
	err := h.Rollback("", "", false, nil, false)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, "releaseName 或 namespace 不能为空", status.Convert(err).Message())
}

func TestHelmerBiz_Rollback_Valid(t *testing.T) {
	f := &fakeHelmerRepoForHelmerBiz{}
	h := newHelmerBizForTest(f)
	assert.NoError(t, h.Rollback("app", "ns", false, nil, false))
	assert.True(t, f.rollbackCalled)
}

func TestHelmerBiz_Uninstall_EmptyReleaseName(t *testing.T) {
	h := newHelmerBizForTest(&fakeHelmerRepoForHelmerBiz{})
	err := h.Uninstall("", "", nil)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, "releaseName 或 namespace 不能为空", status.Convert(err).Message())
}

func TestHelmerBiz_Uninstall_Valid(t *testing.T) {
	f := &fakeHelmerRepoForHelmerBiz{}
	h := newHelmerBizForTest(f)
	assert.NoError(t, h.Uninstall("app", "ns", nil))
	assert.True(t, f.uninstallCalled)
}

func TestHelmerBiz_ReleaseStatus(t *testing.T) {
	f := &fakeHelmerRepoForHelmerBiz{}
	h := newHelmerBizForTest(f)
	assert.Equal(t, types.Deploy_StatusDeploying, h.ReleaseStatus("app", "ns"))
	assert.True(t, f.statusCalled)
}

func TestHelmerBiz_PackageChart(t *testing.T) {
	f := &fakeHelmerRepoForHelmerBiz{}
	h := newHelmerBizForTest(f)
	got, err := h.PackageChart("/charts/app", "/dest")
	assert.NoError(t, err)
	assert.Equal(t, "/dest/app.tgz", got)
	assert.True(t, f.packageCalled)
}
