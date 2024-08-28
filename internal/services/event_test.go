package services

import (
	"context"
	"errors"
	"testing"

	"github.com/duc-cnzj/mars/api/v5/event"
	"github.com/duc-cnzj/mars/api/v5/types"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/duc-cnzj/mars/v5/internal/repo"
	"github.com/duc-cnzj/mars/v5/internal/util/pagination"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNewEventSvc(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	svc := NewEventSvc(mlog.NewLogger(nil), repo.NewMockEventRepo(m))
	assert.NotNil(t, svc)
	assert.NotNil(t, svc.(*eventSvc).eventRepo)
}

func TestEventSvc_List_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	eventRepo := repo.NewMockEventRepo(m)
	svc := NewEventSvc(mlog.NewLogger(nil), eventRepo)

	eventRepo.EXPECT().List(gomock.Any(), &repo.ListEventInput{
		Page:        1,
		PageSize:    12,
		ActionType:  types.EventActionType_Delete,
		Search:      "x",
		OrderIDDesc: lo.ToPtr(true),
	}).Return([]*repo.Event{}, &pagination.Pagination{}, nil)

	resp, err := svc.List(context.Background(), &event.ListRequest{
		Page:       lo.ToPtr(int32(1)),
		PageSize:   lo.ToPtr(int32(12)),
		ActionType: types.EventActionType_Delete,
		Search:     "x",
	})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestEventSvc_List_Failure(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	eventRepo := repo.NewMockEventRepo(m)
	svc := NewEventSvc(mlog.NewLogger(nil), eventRepo)

	eventRepo.EXPECT().List(gomock.Any(), gomock.Any()).Return(nil, nil, errors.New("error"))

	_, err := svc.List(context.Background(), &event.ListRequest{})
	assert.Error(t, err)
}

func Test_eventSvc_Show(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	eventRepo := repo.NewMockEventRepo(m)
	svc := NewEventSvc(mlog.NewLogger(nil), eventRepo)

	eventRepo.EXPECT().Show(gomock.Any(), 1).Return(nil, errors.New("x"))

	show, err := svc.Show(context.TODO(), &event.ShowRequest{Id: 1})
	assert.Nil(t, show)
	assert.NotNil(t, err)
}

func Test_eventSvc_Show_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	eventRepo := repo.NewMockEventRepo(m)
	svc := NewEventSvc(mlog.NewLogger(nil), eventRepo)

	eventRepo.EXPECT().Show(gomock.Any(), 1).Return(&repo.Event{}, nil)

	show, err := svc.Show(context.TODO(), &event.ShowRequest{Id: 1})
	assert.NotNil(t, show)
	assert.Nil(t, err)
}

func TestEventSvc_Authorize_AdminUser(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	eventRepo := repo.NewMockEventRepo(m)
	svc := NewEventSvc(mlog.NewLogger(nil), eventRepo).(*eventSvc)

	_, err := svc.Authorize(newAdminUserCtx(), "TestMethod")
	assert.Nil(t, err)
}

func TestEventSvc_Authorize_NonAdminUser(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	eventRepo := repo.NewMockEventRepo(m)
	svc := NewEventSvc(mlog.NewLogger(nil), eventRepo).(*eventSvc)

	_, err := svc.Authorize(newOtherUserCtx(), "TestMethod")
	s, _ := status.FromError(err)
	assert.Equal(t, codes.PermissionDenied, s.Code())
}
