package services

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/duc-cnzj/mars/api/v6/proto/types"

	"github.com/duc-cnzj/mars/api/v6/proto/file"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/data"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/pagination"
	"github.com/dustin/go-humanize"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNewFileSvc(t *testing.T) {
	svc, _ := newFileSvcWithMocks(t)
	assert.NotNil(t, svc)
	assert.NotNil(t, svc.logger)
	assert.NotNil(t, svc.eventBiz)
	assert.NotNil(t, svc.fileBiz)
}

func TestFileSvc_Authorize_AdminUser(t *testing.T) {
	svc, _ := newFileSvcWithMocks(t)

	_, err := svc.Authorize(newAdminUserCtx(), "TestMethod")
	assert.NoError(t, err)
}

func TestFileSvc_Authorize_NonAdminUser(t *testing.T) {
	svc, _ := newFileSvcWithMocks(t)

	_, err := svc.Authorize(newOtherUserCtx(), "TestMethod")
	assert.Error(t, err)
	assert.Equal(t, codes.PermissionDenied, status.Code(err))
}

func TestFileSvc_Authorize_MaxUploadSize(t *testing.T) {
	svc, _ := newFileSvcWithMocks(t)

	_, err := svc.Authorize(newOtherUserCtx(), "/file.File/MaxUploadSize")
	assert.Nil(t, err)
	_, err = svc.Authorize(context.TODO(), "/file.File/MaxUploadSize")
	assert.Nil(t, err)
}

func Test_fileSvc_Delete(t *testing.T) {
	svc, mocks := newFileSvcWithMocks(t)
	fileRepo := mocks.fileRepo

	fileRepo.EXPECT().GetByID(gomock.Any(), 1).Return(nil, errors.New("xx"))

	response, err := svc.Delete(context.TODO(), &file.DeleteRequest{Id: 1})
	assert.Nil(t, response)
	assert.Error(t, err)
}

func Test_fileSvc_Delete2(t *testing.T) {
	svc, mocks := newFileSvcWithMocks(t)
	fileRepo := mocks.fileRepo

	fileRepo.EXPECT().GetByID(gomock.Any(), 1).Return(&biz.File{}, nil)
	fileRepo.EXPECT().Delete(gomock.Any(), 1).Return(errors.New("xx"))
	response, err := svc.Delete(context.TODO(), &file.DeleteRequest{Id: 1})
	assert.Nil(t, response)
	assert.Error(t, err)
}

func Test_fileSvc_Delete3(t *testing.T) {
	svc, mocks := newFileSvcWithMocks(t)
	fileRepo := mocks.fileRepo
	eventRepo := mocks.eventRepo

	eventRepo.EXPECT().FileAuditLog(
		types.EventActionType_Delete,
		biz.MustGetUser(newAdminUserCtx()).Name,
		gomock.Any(),
		999,
	)
	fileRepo.EXPECT().GetByID(gomock.Any(), 1).Return(&biz.File{ID: 999}, nil)
	fileRepo.EXPECT().Delete(gomock.Any(), 1).Return(nil)
	response, err := svc.Delete(newAdminUserCtx(), &file.DeleteRequest{Id: 1})
	assert.NotNil(t, response)
	assert.Nil(t, err)
}

func Test_fileSvc_MaxUploadSize(t *testing.T) {
	svc, mocks := newFileSvcWithMocks(t)
	fileRepo := mocks.fileRepo

	fileRepo.EXPECT().MaxUploadSize().Return(uint64(10000))
	size, err := svc.MaxUploadSize(newAdminUserCtx(), &file.MaxUploadSizeRequest{})
	assert.Nil(t, err)
	if assert.NotNil(t, size) {
		assert.Equal(t, uint32(10000), size.Bytes)
		assert.Equal(t, humanize.Bytes(10000), size.HumanizeSize)
	}
}

func TestFileSvc_DiskInfo_Success(t *testing.T) {
	svc, mocks := newFileSvcWithMocks(t)
	fileRepo := mocks.fileRepo

	fileRepo.EXPECT().DiskInfo(false).Return(int64(10000), nil)

	resp, err := svc.DiskInfo(context.TODO(), &file.DiskInfoRequest{})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int64(10000), resp.Usage)
	assert.Equal(t, humanize.Bytes(uint64(10000)), resp.HumanizeUsage)
}

func TestFileSvc_DiskInfo_Failure(t *testing.T) {
	svc, mocks := newFileSvcWithMocks(t)
	fileRepo := mocks.fileRepo

	fileRepo.EXPECT().DiskInfo(false).Return(int64(0), errors.New("error"))

	_, err := svc.DiskInfo(context.TODO(), &file.DiskInfoRequest{})
	assert.Error(t, err)
}

func TestFileSvc_List_Success(t *testing.T) {
	svc, mocks := newFileSvcWithMocks(t)
	fileRepo := mocks.fileRepo

	fileRepo.EXPECT().List(gomock.Any(), &biz.ListFileInput{
		Page:           1,
		PageSize:       11,
		OrderIDDesc:    lo.ToPtr(true),
		WithSoftDelete: false,
	}).Return([]*biz.File{}, &pagination.Pagination{}, nil)

	resp, err := svc.List(context.TODO(), &file.ListRequest{
		Page:     lo.ToPtr(int32(1)),
		PageSize: lo.ToPtr(int32(11)),
	})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestFileSvc_List_Failure(t *testing.T) {
	svc, mocks := newFileSvcWithMocks(t)
	fileRepo := mocks.fileRepo

	fileRepo.EXPECT().List(gomock.Any(), gomock.Any()).Return(nil, nil, errors.New("error"))

	_, err := svc.List(context.TODO(), &file.ListRequest{})
	assert.Error(t, err)
}

func TestFileSvc_ShowRecords_Success(t *testing.T) {
	svc, mocks := newFileSvcWithMocks(t)
	fileRepo := mocks.fileRepo

	fileRepo.EXPECT().ShowRecords(gomock.Any(), 1).Return(io.NopCloser(strings.NewReader("record1\nrecord2\n")), nil)

	resp, err := svc.ShowRecords(context.TODO(), &file.ShowRecordsRequest{Id: 1})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, []string{"record1\nrecord2\n"}, resp.Items)
}

func TestFileSvc_ShowRecords_Failure(t *testing.T) {
	svc, mocks := newFileSvcWithMocks(t)
	fileRepo := mocks.fileRepo

	fileRepo.EXPECT().ShowRecords(gomock.Any(), gomock.Any()).Return(nil, errors.New("error"))

	_, err := svc.ShowRecords(context.TODO(), &file.ShowRecordsRequest{Id: 1})
	assert.Error(t, err)
}

// errorReadCloser 实现 io.ReadCloser，但 Read 立即返回错误，用于覆盖
// ShowRecords 中 io.ReadAll 的失败分支。
type errorReadCloser struct{}

func (errorReadCloser) Read(_ []byte) (int, error) {
	return 0, errors.New("read boom")
}

func (errorReadCloser) Close() error { return nil }

func TestFileSvc_ShowRecords_ReadError(t *testing.T) {
	svc, mocks := newFileSvcWithMocks(t)
	fileRepo := mocks.fileRepo

	fileRepo.EXPECT().ShowRecords(gomock.Any(), 1).Return(errorReadCloser{}, nil)

	resp, err := svc.ShowRecords(context.TODO(), &file.ShowRecordsRequest{Id: 1})
	assert.Error(t, err)
	assert.Nil(t, resp)
}

type fileSvcMocks struct {
	ctrl      *gomock.Controller
	eventRepo *data.MockEventRepo
	fileRepo  *data.MockFileRepo
}

func newFileSvcWithMocks(t *testing.T) (*fileSvc, *fileSvcMocks) {
	t.Helper()
	ctrl := gomock.NewController(t)
	mocks := &fileSvcMocks{
		ctrl:      ctrl,
		eventRepo: data.NewMockEventRepo(ctrl),
		fileRepo:  data.NewMockFileRepo(ctrl),
	}
	logger := mlog.NewForConfig(nil)
	s, ok := NewFileSvc(FileSvcDeps{
		EventBiz:  biz.NewEventBiz(mocks.eventRepo),
		FileBiz:   biz.NewFileBiz(mocks.fileRepo),
		Logger:    logger,
		AccessBiz: biz.NewAccessBiz(nil, nil),
	}).(*fileSvc)
	if !ok {
		panic("NewFileSvc returned unexpected type")
	}
	return s, mocks
}
