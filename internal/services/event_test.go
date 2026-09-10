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

	// admin 显式传 all=true 才展开全量（不注入 OperatorEmail）
	resp, err := svc.List(newAdminUserCtx(), &event.ListRequest{
		Page:        lo.ToPtr(int32(1)),
		PageSize:    lo.ToPtr(int32(12)),
		ActionTypes: []types.EventActionType{types.EventActionType_Delete},
		Search:      "x",
		All:         true,
	})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

// admin 未传 all（默认收敛本人，镜像 access_token 语义）：按操作人邮箱过滤为本人事件。
// 这是「下拉入口 /events 对 admin 也只看到自己」的契约基础——全量视图只属显式 all=true 的后台入口。
func TestEventSvc_List_Admin_WithoutAll_FiltersByOwnEmail(t *testing.T) {
	svc, mocks := newEventSvcWithMocks(t)
	eventRepo := mocks.eventRepo
	admin := biz.MustGetUser(newAdminUserCtx())

	eventRepo.EXPECT().List(gomock.Any(), &biz.ListEventInput{
		Page:          1,
		PageSize:      15,
		OrderIDDesc:   lo.ToPtr(true),
		OperatorEmail: &admin.Email,
	}).Return([]*biz.Event{}, &pagination.Pagination{}, nil)

	resp, err := svc.List(newAdminUserCtx(), &event.ListRequest{})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

// 普通用户 List 必须按操作人邮箱（operator_email）过滤为本人事件，
// 归属条件由 ctx 身份推导注入，不接受请求参数（防传他人邮箱枚举全量）。
func TestEventSvc_List_NonAdmin_FiltersByOwnEmail(t *testing.T) {
	svc, mocks := newEventSvcWithMocks(t)
	eventRepo := mocks.eventRepo
	user := biz.MustGetUser(newOtherUserCtx())

	eventRepo.EXPECT().List(gomock.Any(), &biz.ListEventInput{
		Page:          1,
		PageSize:      15,
		OrderIDDesc:   lo.ToPtr(true),
		OperatorEmail: &user.Email,
	}).Return([]*biz.Event{}, &pagination.Pagination{}, nil)

	resp, err := svc.List(newOtherUserCtx(), &event.ListRequest{})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

// 无邮箱的普通用户（非 admin）必须返回空列表：若按空串过滤，IfStrEQ 的
// "空串不过滤"语义会退化成全量可见，违反"普通用户只能看自己的事件"约束。
func TestEventSvc_List_NonAdmin_EmptyEmail_ReturnsEmpty(t *testing.T) {
	svc, mocks := newEventSvcWithMocks(t)
	// 不设置 eventRepo.List 期望：空邮箱路径必须提前返回，任何 DB 查询都会 FAIL。
	_ = mocks

	ctx := biz.SetUser(context.TODO(), &biz.UserInfo{Name: "u"})
	resp, err := svc.List(ctx, &event.ListRequest{})
	assert.NoError(t, err)
	if assert.NotNil(t, resp) {
		// 空邮箱返回空 slice（非 nil），与正常空结果路径的响应形状一致（避免 null vs []）。
		assert.NotNil(t, resp.Items)
		assert.Empty(t, resp.Items)
	}
}

func TestEventSvc_List_Failure(t *testing.T) {
	svc, mocks := newEventSvcWithMocks(t)
	eventRepo := mocks.eventRepo

	// admin + all=true（全量视图路径）：repo 报错 → List 传播错误
	req := &event.ListRequest{All: true}

	eventRepo.EXPECT().List(gomock.Any(), &biz.ListEventInput{
		Page:        1,
		PageSize:    15,
		ActionTypes: normalizeActionTypes(req.ActionType, req.ActionTypes),
		Search:      req.Search,
		OrderIDDesc: lo.ToPtr(true),
	}).Return(nil, nil, errors.New("error"))

	_, err := svc.List(newAdminUserCtx(), req)
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

	show, err := svc.Show(newAdminUserCtx(), &event.ShowRequest{Id: 1})
	assert.Nil(t, show)
	assert.NotNil(t, err)
}

func Test_eventSvc_Show_Success(t *testing.T) {
	svc, mocks := newEventSvcWithMocks(t)
	eventRepo := mocks.eventRepo

	eventRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Event{
		ID:            42,
		Action:        types.EventActionType_Create,
		Username:      "duc",
		OperatorEmail: "duc@example.com",
		Message:       "deployed",
	}, nil)

	show, err := svc.Show(newAdminUserCtx(), &event.ShowRequest{Id: 1})
	assert.Nil(t, err)
	if assert.NotNil(t, show) && assert.NotNil(t, show.Item) {
		assert.Equal(t, int32(42), show.Item.Id)
		assert.Equal(t, types.EventActionType_Create, show.Item.Action)
		assert.Equal(t, "duc", show.Item.Username)
		assert.Equal(t, "deployed", show.Item.Message)
	}
}

// 普通用户 Show 只能查看操作人邮箱为自己的事件。
func Test_eventSvc_Show_NonAdmin_OwnEvent(t *testing.T) {
	svc, mocks := newEventSvcWithMocks(t)
	eventRepo := mocks.eventRepo

	eventRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Event{
		ID:            42,
		OperatorEmail: "user@mars.com",
	}, nil)

	show, err := svc.Show(newOtherUserCtx(), &event.ShowRequest{Id: 1})
	assert.NoError(t, err)
	assert.Equal(t, int32(42), show.Item.Id)
}

// 普通用户 Show 他人事件返回 404（视同不存在），防审计日志 id 枚举侧信道。
func Test_eventSvc_Show_NonAdmin_OtherEvent_NotFound(t *testing.T) {
	svc, mocks := newEventSvcWithMocks(t)
	eventRepo := mocks.eventRepo

	eventRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Event{
		ID:            42,
		OperatorEmail: "other@x.com",
	}, nil)

	show, err := svc.Show(newOtherUserCtx(), &event.ShowRequest{Id: 1})
	assert.Nil(t, show)
	assert.Error(t, err)
	s, _ := status.FromError(err)
	assert.Equal(t, codes.NotFound, s.Code())
}

// 无邮箱的普通用户 Show 必须一律 404，与 List 的"空邮箱返回空列表"对齐。
// 否则 operator_email 为空的事件（迁移前历史行 / cron 系统事件）会被空邮箱用户
// 通过 "" == "" 等值比较绕过归属校验，泄露他人审计日志。
func Test_eventSvc_Show_NonAdmin_EmptyEmail_AlwaysNotFound(t *testing.T) {
	svc, mocks := newEventSvcWithMocks(t)
	eventRepo := mocks.eventRepo

	eventRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Event{
		ID:            42,
		OperatorEmail: "", // 空邮箱事件：迁移前历史行 / cron 系统事件
	}, nil)

	show, err := svc.Show(biz.SetUser(context.TODO(), &biz.UserInfo{Name: "u"}), &event.ShowRequest{Id: 1})
	assert.Nil(t, show)
	assert.Error(t, err)
	s, _ := status.FromError(err)
	assert.Equal(t, codes.NotFound, s.Code())
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
		Logger:   logger,
		EventBiz: biz.NewEventBiz(mocks.eventRepo),
	}).(*eventSvc)
	if !ok {
		panic("NewEventSvc returned unexpected type")
	}
	return s, mocks
}
