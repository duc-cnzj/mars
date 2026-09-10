package biz

import (
	"context"
	"io"
	"testing"

	"github.com/duc-cnzj/mars/v6/internal/util/pagination"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// fakeFileRepoForFileBiz 记录各写操作是否被调用，输入校验测试中 repo 不被调用（调用即 panic）。
type fakeFileRepoForFileBiz struct {
	FileRepo
	createCalled, deleteCalled, updateCalled, streamCalled bool
	maxUploadCalled, showRecordsCalled, diskInfoCalled     bool
	listCalled, getByIDCalled, newRecorderCalled           bool
}

func (f *fakeFileRepoForFileBiz) MaxUploadSize() uint64 {
	f.maxUploadCalled = true
	return 1024
}

func (f *fakeFileRepoForFileBiz) ShowRecords(ctx context.Context, id int) (io.ReadCloser, error) {
	f.showRecordsCalled = true
	return nil, nil
}

func (f *fakeFileRepoForFileBiz) DiskInfo(force bool) (int64, error) {
	f.diskInfoCalled = true
	return 42, nil
}

func (f *fakeFileRepoForFileBiz) List(ctx context.Context, input *ListFileInput) ([]*File, *pagination.Pagination, error) {
	f.listCalled = true
	return []*File{{ID: 1}}, nil, nil
}

func (f *fakeFileRepoForFileBiz) GetByID(ctx context.Context, id int) (*File, error) {
	f.getByIDCalled = true
	return &File{ID: id}, nil
}

func (f *fakeFileRepoForFileBiz) NewRecorder(user *UserInfo, container *Container) Recorder {
	f.newRecorderCalled = true
	return nil
}

func (f *fakeFileRepoForFileBiz) Create(ctx context.Context, input *CreateFileInput) (*File, error) {
	f.createCalled = true
	return &File{ID: 1, Path: input.Path}, nil
}

func (f *fakeFileRepoForFileBiz) Delete(ctx context.Context, id int) error {
	f.deleteCalled = true
	return nil
}

func (f *fakeFileRepoForFileBiz) Update(ctx context.Context, i *UpdateFileRequest) (*File, error) {
	f.updateCalled = true
	return &File{ID: i.ID}, nil
}

func (f *fakeFileRepoForFileBiz) StreamUploadFile(ctx context.Context, input *StreamUploadFileRequest) (*File, error) {
	f.streamCalled = true
	return &File{ID: 1, Path: input.FileName}, nil
}

func newFileBizForTest(repo FileRepo) FileBiz {
	return NewFileBiz(repo)
}

func TestFileBiz_Create_EmptyPath(t *testing.T) {
	b := newFileBizForTest(&fakeFileRepoForFileBiz{})
	got, err := b.Create(context.TODO(), &CreateFileInput{Path: ""})
	assert.Nil(t, got)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, "file 不能为空或 path 不能为空", status.Convert(err).Message())
}

func TestFileBiz_Create_Valid(t *testing.T) {
	f := &fakeFileRepoForFileBiz{}
	b := newFileBizForTest(f)
	got, err := b.Create(context.TODO(), &CreateFileInput{Path: "/tmp/a"})
	assert.NoError(t, err)
	assert.True(t, f.createCalled)
	assert.Equal(t, "/tmp/a", got.Path)
}

func TestFileBiz_Update_InvalidID(t *testing.T) {
	b := newFileBizForTest(&fakeFileRepoForFileBiz{})
	got, err := b.Update(context.TODO(), &UpdateFileRequest{ID: 0})
	assert.Nil(t, got)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, "file 不能为空或 id 不能小于等于 0", status.Convert(err).Message())
}

func TestFileBiz_Update_Valid(t *testing.T) {
	f := &fakeFileRepoForFileBiz{}
	b := newFileBizForTest(f)
	got, err := b.Update(context.TODO(), &UpdateFileRequest{ID: 1})
	assert.NoError(t, err)
	assert.True(t, f.updateCalled)
	assert.Equal(t, 1, got.ID)
}

func TestFileBiz_Delete_InvalidID(t *testing.T) {
	b := newFileBizForTest(&fakeFileRepoForFileBiz{})
	err := b.Delete(context.TODO(), 0)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, "file id 不能小于等于 0", status.Convert(err).Message())
}

func TestFileBiz_Delete_Valid(t *testing.T) {
	f := &fakeFileRepoForFileBiz{}
	b := newFileBizForTest(f)
	assert.NoError(t, b.Delete(context.TODO(), 1))
	assert.True(t, f.deleteCalled)
}

func TestFileBiz_StreamUploadFile_EmptyFileName(t *testing.T) {
	b := newFileBizForTest(&fakeFileRepoForFileBiz{})
	got, err := b.StreamUploadFile(context.TODO(), &StreamUploadFileRequest{FileName: ""})
	assert.Nil(t, got)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, "file 不能为空或文件名不能为空", status.Convert(err).Message())
}

func TestFileBiz_StreamUploadFile_Valid(t *testing.T) {
	f := &fakeFileRepoForFileBiz{}
	b := newFileBizForTest(f)
	got, err := b.StreamUploadFile(context.TODO(), &StreamUploadFileRequest{FileName: "a.txt"})
	assert.NoError(t, err)
	assert.True(t, f.streamCalled)
	assert.Equal(t, "a.txt", got.Path)
}

// ---- 纯透传查询 ----

func TestFileBiz_MaxUploadSize(t *testing.T) {
	f := &fakeFileRepoForFileBiz{}
	b := newFileBizForTest(f)
	assert.Equal(t, uint64(1024), b.MaxUploadSize())
	assert.True(t, f.maxUploadCalled)
}

func TestFileBiz_ShowRecords(t *testing.T) {
	f := &fakeFileRepoForFileBiz{}
	b := newFileBizForTest(f)
	rc, err := b.ShowRecords(context.TODO(), 1)
	assert.NoError(t, err)
	assert.Nil(t, rc)
	assert.True(t, f.showRecordsCalled)
}

func TestFileBiz_DiskInfo(t *testing.T) {
	f := &fakeFileRepoForFileBiz{}
	b := newFileBizForTest(f)
	n, err := b.DiskInfo(false)
	assert.NoError(t, err)
	assert.Equal(t, int64(42), n)
	assert.True(t, f.diskInfoCalled)
}

func TestFileBiz_List(t *testing.T) {
	f := &fakeFileRepoForFileBiz{}
	b := newFileBizForTest(f)
	files, pag, err := b.List(context.TODO(), &ListFileInput{})
	assert.NoError(t, err)
	assert.Len(t, files, 1)
	assert.Nil(t, pag)
	assert.True(t, f.listCalled)
}

func TestFileBiz_GetByID(t *testing.T) {
	f := &fakeFileRepoForFileBiz{}
	b := newFileBizForTest(f)
	got, err := b.GetByID(context.TODO(), 7)
	assert.NoError(t, err)
	assert.Equal(t, 7, got.ID)
	assert.True(t, f.getByIDCalled)
}

func TestFileBiz_NewRecorder(t *testing.T) {
	f := &fakeFileRepoForFileBiz{}
	b := newFileBizForTest(f)
	r := b.NewRecorder(&UserInfo{Name: "u"}, &Container{Namespace: "ns", Pod: "p", Container: "c"})
	assert.Nil(t, r)
	assert.True(t, f.newRecorderCalled)
}
