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

// TestFileSvc_Authorize_ShowRecords 覆盖 ShowRecords 进入 allowlist：非 admin 与
// 无用户 ctx 均可过 Authorize 门禁（方法体内 RequireFileAccess 再做所有者/admin 判定）。
func TestFileSvc_Authorize_ShowRecords(t *testing.T) {
	svc, _ := newFileSvcWithMocks(t)

	_, err := svc.Authorize(newOtherUserCtx(), "/file.File/ShowRecords")
	assert.Nil(t, err)
	_, err = svc.Authorize(context.TODO(), "/file.File/ShowRecords")
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
		biz.MustGetUser(newAdminUserCtx()).Email,
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

// TestFileSvc_ShowRecords_Success 覆盖文件所有者（非 admin）回放自己的会话：应放行并整体回传记录。
func TestFileSvc_ShowRecords_Success(t *testing.T) {
	svc, mocks := newFileSvcWithMocks(t)
	fileRepo := mocks.fileRepo

	fileRepo.EXPECT().GetByID(gomock.Any(), 1).Return(&biz.File{ID: 1, Username: "user1"}, nil)
	fileRepo.EXPECT().ShowRecords(gomock.Any(), 1).Return(io.NopCloser(strings.NewReader("record1\nrecord2\n")), nil)

	resp, err := svc.ShowRecords(newOtherUserCtx(), &file.ShowRecordsRequest{Id: 1})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, []string{"record1\nrecord2\n"}, resp.Items)
}

// TestFileSvc_ShowRecords_Denied 覆盖非所有者非 admin 的越权回放：返回
// PermissionDenied，且不触达 ShowRecords 读取（未设置该调用期望，误调即 gomock 失败）。
func TestFileSvc_ShowRecords_Denied(t *testing.T) {
	svc, mocks := newFileSvcWithMocks(t)
	fileRepo := mocks.fileRepo

	fileRepo.EXPECT().GetByID(gomock.Any(), 1).Return(&biz.File{ID: 1, Username: "someone_else"}, nil)

	resp, err := svc.ShowRecords(newOtherUserCtx(), &file.ShowRecordsRequest{Id: 1})
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Equal(t, codes.PermissionDenied, status.Code(err))
}

// TestFileSvc_ShowRecords_Admin 覆盖 admin 查看非本人文件：admin 任意放行，正常读取。
func TestFileSvc_ShowRecords_Admin(t *testing.T) {
	svc, mocks := newFileSvcWithMocks(t)
	fileRepo := mocks.fileRepo

	fileRepo.EXPECT().GetByID(gomock.Any(), 1).Return(&biz.File{ID: 1, Username: "someone_else"}, nil)
	fileRepo.EXPECT().ShowRecords(gomock.Any(), 1).Return(io.NopCloser(strings.NewReader("record")), nil)

	resp, err := svc.ShowRecords(newAdminUserCtx(), &file.ShowRecordsRequest{Id: 1})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, []string{"record"}, resp.Items)
}

// TestFileSvc_ShowRecords_GetByIDError 覆盖加载文件元数据失败：直接返回错误，不读取记录。
func TestFileSvc_ShowRecords_GetByIDError(t *testing.T) {
	svc, mocks := newFileSvcWithMocks(t)
	fileRepo := mocks.fileRepo

	fileRepo.EXPECT().GetByID(gomock.Any(), 1).Return(nil, errors.New("not found"))

	resp, err := svc.ShowRecords(newAdminUserCtx(), &file.ShowRecordsRequest{Id: 1})
	assert.Nil(t, resp)
	assert.Error(t, err)
}

// TestFileSvc_ShowRecords_Failure 覆盖文件元数据通过但读取记录失败：返回错误。
func TestFileSvc_ShowRecords_Failure(t *testing.T) {
	svc, mocks := newFileSvcWithMocks(t)
	fileRepo := mocks.fileRepo

	fileRepo.EXPECT().GetByID(gomock.Any(), 1).Return(&biz.File{ID: 1, Username: "user1"}, nil)
	fileRepo.EXPECT().ShowRecords(gomock.Any(), gomock.Any()).Return(nil, errors.New("error"))

	_, err := svc.ShowRecords(newOtherUserCtx(), &file.ShowRecordsRequest{Id: 1})
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

	fileRepo.EXPECT().GetByID(gomock.Any(), 1).Return(&biz.File{ID: 1, Username: "user1"}, nil)
	fileRepo.EXPECT().ShowRecords(gomock.Any(), 1).Return(errorReadCloser{}, nil)

	resp, err := svc.ShowRecords(newOtherUserCtx(), &file.ShowRecordsRequest{Id: 1})
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
