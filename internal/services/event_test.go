package services

import (
	"context"
	"errors"
	"testing"

	"github.com/duc-cnzj/mars/api/v6/proto/event"
	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/data"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/pagination"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNewEventSvc(t *testing.T) {
	svc, _ := newEventSvcWithMocks(t)
	assert.NotNil(t, svc)
	assert.NotNil(t, svc.eventBiz)
}

func TestEventSvc_List_Success(t *testing.T) {
	svc, mocks := newEventSvcWithMocks(t)
	eventRepo := mocks.eventRepo

	eventRepo.EXPECT().List(gomock.Any(), &biz.ListEventInput{
		Page:        1,
		PageSize:    12,
		ActionTypes: []types.EventActionType{types.EventActionType_Delete},
		Search:      "x",
		OrderIDDesc: lo.ToPtr(true),
	}).Return([]*biz.Event{}, &pagination.Pagination{}, nil)

	resp, err := svc.List(context.TODO(), &event.ListRequest{
		Page:        lo.ToPtr(int32(1)),
		PageSize:    lo.ToPtr(int32(12)),
		ActionTypes: []types.EventActionType{types.EventActionType_Delete},
		Search:      "x",
	})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestEventSvc_List_Failure(t *testing.T) {
	svc, mocks := newEventSvcWithMocks(t)
	eventRepo := mocks.eventRepo

	req := &event.ListRequest{}

	eventRepo.EXPECT().List(gomock.Any(), &biz.ListEventInput{
		Page:        1,
		PageSize:    15,
		ActionTypes: normalizeActionTypes(req.ActionType, req.ActionTypes),
		Search:      req.Search,
		OrderIDDesc: lo.ToPtr(true),
	}).Return(nil, nil, errors.New("error"))

	_, err := svc.List(context.TODO(), req)
	assert.Error(t, err)
}

func Test_normalizeActionTypes(t *testing.T) {
	cases := []struct {
		name   string
		single types.EventActionType
		multi  []types.EventActionType
		want   []types.EventActionType
	}{
		{name: "empty = all", single: types.EventActionType_Unknown, multi: nil, want: nil},
		{name: "single fallback", single: types.EventActionType_Delete, multi: nil, want: []types.EventActionType{types.EventActionType_Delete}},
		{name: "multi wins", single: types.EventActionType_Create, multi: []types.EventActionType{types.EventActionType_Delete}, want: []types.EventActionType{types.EventActionType_Delete}},
		{name: "multi only", single: types.EventActionType_Unknown, multi: []types.EventActionType{types.EventActionType_Update, types.EventActionType_Shell}, want: []types.EventActionType{types.EventActionType_Update, types.EventActionType_Shell}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.want, normalizeActionTypes(c.single, c.multi))
		})
	}
}

func Test_eventSvc_Show(t *testing.T) {
	svc, mocks := newEventSvcWithMocks(t)
	eventRepo := mocks.eventRepo

	eventRepo.EXPECT().Show(gomock.Any(), 1).Return(nil, errors.New("x"))

	show, err := svc.Show(context.TODO(), &event.ShowRequest{Id: 1})
	assert.Nil(t, show)
	assert.NotNil(t, err)
}

func Test_eventSvc_Show_Success(t *testing.T) {
	svc, mocks := newEventSvcWithMocks(t)
	eventRepo := mocks.eventRepo

	eventRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Event{
		ID:       42,
		Action:   types.EventActionType_Create,
		Username: "duc",
		Message:  "deployed",
	}, nil)

	show, err := svc.Show(context.TODO(), &event.ShowRequest{Id: 1})
	assert.Nil(t, err)
	if assert.NotNil(t, show) && assert.NotNil(t, show.Item) {
		assert.Equal(t, int32(42), show.Item.Id)
		assert.Equal(t, types.EventActionType_Create, show.Item.Action)
		assert.Equal(t, "duc", show.Item.Username)
		assert.Equal(t, "deployed", show.Item.Message)
	}
}

func TestEventSvc_Authorize_AdminUser(t *testing.T) {
	svc, _ := newEventSvcWithMocks(t)

	_, err := svc.Authorize(newAdminUserCtx(), "TestMethod")
	assert.Nil(t, err)
}

func TestEventSvc_Authorize_NonAdminUser(t *testing.T) {
	svc, _ := newEventSvcWithMocks(t)

	_, err := svc.Authorize(newOtherUserCtx(), "TestMethod")
	s, _ := status.FromError(err)
	assert.Equal(t, codes.PermissionDenied, s.Code())
}

type eventSvcMocks struct {
	ctrl      *gomock.Controller
	eventRepo *data.MockEventRepo
}

func newEventSvcWithMocks(t *testing.T) (*eventSvc, *eventSvcMocks) {
	t.Helper()
	ctrl := gomock.NewController(t)
	mocks := &eventSvcMocks{
		ctrl:      ctrl,
		eventRepo: data.NewMockEventRepo(ctrl),
	}
	logger := mlog.NewForConfig(nil)
	s, ok := NewEventSvc(EventSvcDeps{
		Logger:    logger,
		EventBiz:  biz.NewEventBiz(mocks.eventRepo),
		AccessBiz: biz.NewAccessBiz(nil, nil),
	}).(*eventSvc)
	if !ok {
		panic("NewEventSvc returned unexpected type")
	}
	return s, mocks
}
