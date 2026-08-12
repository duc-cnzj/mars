package biz

import (
	"context"
	"testing"

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
