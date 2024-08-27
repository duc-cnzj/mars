package services

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/duc-cnzj/mars/api/v4/file"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/repo"
	"github.com/duc-cnzj/mars/v4/internal/util/pagination"
	"github.com/dustin/go-humanize"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNewFileSvc(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	svc := NewFileSvc(repo.NewMockEventRepo(m), repo.NewMockFileRepo(m), mlog.NewLogger(nil))
	assert.NotNil(t, svc)
	assert.NotNil(t, svc.(*fileSvc).logger)
	assert.NotNil(t, svc.(*fileSvc).eventRepo)
	assert.NotNil(t, svc.(*fileSvc).fileRepo)
}

func TestFileSvc_Authorize_AdminUser(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	eventRepo := repo.NewMockEventRepo(m)
	fileRepo := repo.NewMockFileRepo(m)
	svc := NewFileSvc(eventRepo, fileRepo, mlog.NewLogger(nil)).(*fileSvc)

	_, err := svc.Authorize(newAdminUserCtx(), "TestMethod")
	assert.NoError(t, err)
}

func TestFileSvc_Authorize_NonAdminUser(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	eventRepo := repo.NewMockEventRepo(m)
	fileRepo := repo.NewMockFileRepo(m)
	svc := NewFileSvc(eventRepo, fileRepo, mlog.NewLogger(nil)).(*fileSvc)

	_, err := svc.Authorize(newOtherUserCtx(), "TestMethod")
	assert.Error(t, err)
	assert.Equal(t, codes.PermissionDenied, status.Code(err))
}

func TestFileSvc_Authorize_MaxUploadSize(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	eventRepo := repo.NewMockEventRepo(m)
	fileRepo := repo.NewMockFileRepo(m)
	svc := NewFileSvc(eventRepo, fileRepo, mlog.NewLogger(nil)).(*fileSvc)

	_, err := svc.Authorize(newOtherUserCtx(), "MaxUploadSize")
	assert.Nil(t, err)
	_, err = svc.Authorize(context.TODO(), "MaxUploadSize")
	assert.Nil(t, err)
}

func Test_fileSvc_Delete(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	fileRepo := repo.NewMockFileRepo(m)
	eventRepo := repo.NewMockEventRepo(m)
	svc := NewFileSvc(eventRepo, fileRepo, mlog.NewLogger(nil))

	fileRepo.EXPECT().GetByID(gomock.Any(), int(1)).Return(nil, errors.New("xx"))

	response, err := svc.Delete(context.TODO(), &file.DeleteRequest{Id: 1})
	assert.Nil(t, response)
	assert.Error(t, err)
}

func Test_fileSvc_Delete2(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	fileRepo := repo.NewMockFileRepo(m)
	eventRepo := repo.NewMockEventRepo(m)
	svc := NewFileSvc(eventRepo, fileRepo, mlog.NewLogger(nil))

	fileRepo.EXPECT().GetByID(gomock.Any(), int(1)).Return(&repo.File{}, nil)
	fileRepo.EXPECT().Delete(gomock.Any(), int(1)).Return(errors.New("xx"))
	response, err := svc.Delete(context.TODO(), &file.DeleteRequest{Id: 1})
	assert.Nil(t, response)
	assert.Error(t, err)
}

func Test_fileSvc_Delete3(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	fileRepo := repo.NewMockFileRepo(m)
	eventRepo := repo.NewMockEventRepo(m)
	svc := NewFileSvc(eventRepo, fileRepo, mlog.NewLogger(nil))

	eventRepo.EXPECT().FileAuditLog(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
	fileRepo.EXPECT().GetByID(gomock.Any(), int(1)).Return(&repo.File{}, nil)
	fileRepo.EXPECT().Delete(gomock.Any(), int(1)).Return(nil)
	response, err := svc.Delete(newAdminUserCtx(), &file.DeleteRequest{Id: 1})
	assert.NotNil(t, response)
	assert.Nil(t, err)
}

func Test_fileSvc_MaxUploadSize(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	fileRepo := repo.NewMockFileRepo(m)
	eventRepo := repo.NewMockEventRepo(m)
	svc := NewFileSvc(eventRepo, fileRepo, mlog.NewLogger(nil))

	fileRepo.EXPECT().MaxUploadSize().Return(uint64(10000))
	size, err := svc.MaxUploadSize(newAdminUserCtx(), &file.MaxUploadSizeRequest{})
	assert.Nil(t, err)
	assert.NotNil(t, size)
}

func TestFileSvc_DiskInfo_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	eventRepo := repo.NewMockEventRepo(m)
	fileRepo := repo.NewMockFileRepo(m)
	svc := NewFileSvc(eventRepo, fileRepo, mlog.NewLogger(nil))

	fileRepo.EXPECT().DiskInfo(false).Return(int64(10000), nil)

	resp, err := svc.DiskInfo(context.Background(), &file.DiskInfoRequest{})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int64(10000), resp.Usage)
	assert.Equal(t, humanize.Bytes(uint64(10000)), resp.HumanizeUsage)
}

func TestFileSvc_DiskInfo_Failure(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	eventRepo := repo.NewMockEventRepo(m)
	fileRepo := repo.NewMockFileRepo(m)
	svc := NewFileSvc(eventRepo, fileRepo, mlog.NewLogger(nil))

	fileRepo.EXPECT().DiskInfo(false).Return(int64(0), errors.New("error"))

	_, err := svc.DiskInfo(context.Background(), &file.DiskInfoRequest{})
	assert.Error(t, err)
}

func TestFileSvc_List_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	eventRepo := repo.NewMockEventRepo(m)
	fileRepo := repo.NewMockFileRepo(m)
	svc := NewFileSvc(eventRepo, fileRepo, mlog.NewLogger(nil))

	fileRepo.EXPECT().List(gomock.Any(), &repo.ListFileInput{
		Page:           1,
		PageSize:       11,
		OrderIDDesc:    lo.ToPtr(true),
		WithSoftDelete: true,
	}).Return([]*repo.File{}, &pagination.Pagination{}, nil)

	resp, err := svc.List(context.Background(), &file.ListRequest{
		Page:           lo.ToPtr(int32(1)),
		PageSize:       lo.ToPtr(int32(11)),
		WithoutDeleted: true,
	})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestFileSvc_List_Failure(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	eventRepo := repo.NewMockEventRepo(m)
	fileRepo := repo.NewMockFileRepo(m)
	svc := NewFileSvc(eventRepo, fileRepo, mlog.NewLogger(nil))

	fileRepo.EXPECT().List(gomock.Any(), gomock.Any()).Return(nil, nil, errors.New("error"))

	_, err := svc.List(context.Background(), &file.ListRequest{})
	assert.Error(t, err)
}

func TestFileSvc_ShowRecords_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	eventRepo := repo.NewMockEventRepo(m)
	fileRepo := repo.NewMockFileRepo(m)
	svc := NewFileSvc(eventRepo, fileRepo, mlog.NewLogger(nil))

	fileRepo.EXPECT().ShowRecords(gomock.Any(), gomock.Any()).Return(io.NopCloser(strings.NewReader("record1\nrecord2\n")), nil)

	resp, err := svc.ShowRecords(context.Background(), &file.ShowRecordsRequest{Id: 1})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, []string{"record1\nrecord2\n"}, resp.Items)
}

func TestFileSvc_ShowRecords_Failure(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	eventRepo := repo.NewMockEventRepo(m)
	fileRepo := repo.NewMockFileRepo(m)
	svc := NewFileSvc(eventRepo, fileRepo, mlog.NewLogger(nil))

	fileRepo.EXPECT().ShowRecords(gomock.Any(), gomock.Any()).Return(nil, errors.New("error"))

	_, err := svc.ShowRecords(context.Background(), &file.ShowRecordsRequest{Id: 1})
	assert.Error(t, err)
}
