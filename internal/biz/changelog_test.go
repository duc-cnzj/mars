package biz

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// fakeClRepoForChangelogBiz 记录各方法是否被调用，输入校验测试中 repo 不被调用（调用即 panic）。
type fakeClRepoForChangelogBiz struct {
	ChangelogRepo
	createCalled           bool
	listCalled, lastCalled bool
}

func (f *fakeClRepoForChangelogBiz) Create(ctx context.Context, input *CreateChangeLogInput) (*Changelog, error) {
	f.createCalled = true
	return &Changelog{ID: 1, ProjectID: input.ProjectID}, nil
}

func (f *fakeClRepoForChangelogBiz) FindLastChangelogsByProjectID(ctx context.Context, input *FindLastChangelogsByProjectIDChangeLogInput) ([]*Changelog, error) {
	f.listCalled = true
	return []*Changelog{{ID: 1, ProjectID: input.ProjectID}}, nil
}

func (f *fakeClRepoForChangelogBiz) FindLastChangeByProjectID(ctx context.Context, projectID int) (*Changelog, error) {
	f.lastCalled = true
	return &Changelog{ID: 1, ProjectID: projectID}, nil
}

func newClBizForTest(repo ChangelogRepo) ChangelogBiz {
	return NewChangelogBiz(repo)
}

func TestChangelogBiz_Create_NilInput(t *testing.T) {
	c := newClBizForTest(&fakeClRepoForChangelogBiz{})
	got, err := c.Create(context.TODO(), nil)
	assert.Nil(t, got)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, "changelog 不能为空或 project id 不能小于等于 0", status.Convert(err).Message())
}

func TestChangelogBiz_Create_InvalidProjectID(t *testing.T) {
	c := newClBizForTest(&fakeClRepoForChangelogBiz{})
	got, err := c.Create(context.TODO(), &CreateChangeLogInput{ProjectID: 0})
	assert.Nil(t, got)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, "changelog 不能为空或 project id 不能小于等于 0", status.Convert(err).Message())
}

func TestChangelogBiz_Create_Valid(t *testing.T) {
	f := &fakeClRepoForChangelogBiz{}
	c := newClBizForTest(f)
	got, err := c.Create(context.TODO(), &CreateChangeLogInput{ProjectID: 1})
	assert.NoError(t, err)
	assert.True(t, f.createCalled)
	assert.Equal(t, 1, got.ProjectID)
}

func TestChangelogBiz_FindLastChangelogsByProjectID(t *testing.T) {
	f := &fakeClRepoForChangelogBiz{}
	c := newClBizForTest(f)
	got, err := c.FindLastChangelogsByProjectID(context.TODO(), &FindLastChangelogsByProjectIDChangeLogInput{ProjectID: 1})
	assert.NoError(t, err)
	assert.True(t, f.listCalled)
	assert.Len(t, got, 1)
}

func TestChangelogBiz_FindLastChangeByProjectID(t *testing.T) {
	f := &fakeClRepoForChangelogBiz{}
	c := newClBizForTest(f)
	got, err := c.FindLastChangeByProjectID(context.TODO(), 5)
	assert.NoError(t, err)
	assert.True(t, f.lastCalled)
	assert.Equal(t, 5, got.ProjectID)
}

// fakeClRepoDaily 只实现 SelectCreatedAtBetween，记录收到的窗口参数并回放预设时间戳。
type fakeClRepoDaily struct {
	ChangelogRepo
	since, until time.Time
	created      []time.Time
	err          error
}

func (f *fakeClRepoDaily) SelectCreatedAtBetween(ctx context.Context, since, until time.Time) ([]time.Time, error) {
	f.since, f.until = since, until
	return f.created, f.err
}

// todayLocalStartForTest 返回服务端本地时区下的今日 00:00，与被测实现同口径。
func todayLocalStartForTest() time.Time {
	now := time.Now().In(time.Local)
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
}

// 分桶 + 零填充 + 升序：同一天多条累加，无部署的天补 0，长度恒等于 days，末位为今天；
// 同时校验传给 repo 的窗口是 [今天-(days-1) 00:00, 明天 00:00)。
func TestChangelogBiz_DeployDailyCounts_ZeroFillAndOrder(t *testing.T) {
	today := todayLocalStartForTest()
	f := &fakeClRepoDaily{created: []time.Time{
		today.Add(2 * time.Hour),
		today.Add(5 * time.Hour),
		today.AddDate(0, 0, -2).Add(10 * time.Hour),
	}}
	c := newClBizForTest(f)

	got, err := c.DeployDailyCounts(context.TODO(), 3)
	assert.NoError(t, err)
	assert.Equal(t, []*DeployDailyCount{
		{Date: dayKey(today.AddDate(0, 0, -2)), Count: 1},
		{Date: dayKey(today.AddDate(0, 0, -1)), Count: 0},
		{Date: dayKey(today), Count: 2},
	}, got)
	assert.Equal(t, today.AddDate(0, 0, -2), f.since)
	assert.Equal(t, today.AddDate(0, 0, 1), f.until)
}

// days=1 是窗口下界：只回今天一个桶，since 即今日 00:00。
func TestChangelogBiz_DeployDailyCounts_SingleDay(t *testing.T) {
	today := todayLocalStartForTest()
	f := &fakeClRepoDaily{created: []time.Time{today.Add(time.Minute)}}
	c := newClBizForTest(f)

	got, err := c.DeployDailyCounts(context.TODO(), 1)
	assert.NoError(t, err)
	assert.Equal(t, []*DeployDailyCount{{Date: dayKey(today), Count: 1}}, got)
	assert.Equal(t, today, f.since)
	assert.Equal(t, today.AddDate(0, 0, 1), f.until)
}

// repo 失败原样上抛，不吞错、不返回半成品切片。
func TestChangelogBiz_DeployDailyCounts_RepoError(t *testing.T) {
	f := &fakeClRepoDaily{err: errors.New("db down")}
	c := newClBizForTest(f)

	got, err := c.DeployDailyCounts(context.TODO(), 7)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "db down")
}

// 窗口内无任何记录：仍返回满长度全 0，而不是空切片——前端按固定长度画趋势图。
func TestChangelogBiz_DeployDailyCounts_EmptyWindow(t *testing.T) {
	f := &fakeClRepoDaily{}
	c := newClBizForTest(f)

	got, err := c.DeployDailyCounts(context.TODO(), 5)
	assert.NoError(t, err)
	assert.Len(t, got, 5)
	for _, d := range got {
		assert.Equal(t, 0, d.Count)
	}
}

// dayKey 折叠为本地时区的 YYYY-MM-DD，个位月/日须补零。
func TestDayKey(t *testing.T) {
	assert.Equal(t, "2026-03-05", dayKey(time.Date(2026, 3, 5, 23, 59, 59, 0, time.Local)))
	assert.Equal(t, "2026-01-09", dayKey(time.Date(2026, 1, 9, 0, 0, 0, 0, time.Local)))
}
