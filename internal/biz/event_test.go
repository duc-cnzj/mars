package biz

import (
	"context"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/util/pagination"
	"github.com/stretchr/testify/assert"
)

// fakeEventRepoForEventBiz 记录各审计方法是否被调用，返回值固定为罐头数据。
type fakeEventRepoForEventBiz struct {
	EventRepo
	listCalled, showCalled              bool
	auditCalled, auditChangeCalled      bool
	auditRequestCalled                  bool
	fileAuditCalled, fileDurationCalled bool
}

func (f *fakeEventRepoForEventBiz) List(ctx context.Context, input *ListEventInput) ([]*Event, *pagination.Pagination, error) {
	f.listCalled = true
	return nil, nil, nil
}

func (f *fakeEventRepoForEventBiz) Show(ctx context.Context, id int) (*Event, error) {
	f.showCalled = true
	return &Event{ID: id}, nil
}

func (f *fakeEventRepoForEventBiz) AuditLog(action types.EventActionType, username, operatorEmail, msg string) {
	f.auditCalled = true
}

func (f *fakeEventRepoForEventBiz) AuditLogWithChange(action types.EventActionType, username, operatorEmail, msg string, oldS, newS YamlPrettier) {
	f.auditChangeCalled = true
}

func (f *fakeEventRepoForEventBiz) AuditLogWithRequest(action types.EventActionType, username, operatorEmail, msg string, req any) {
	f.auditRequestCalled = true
}

func (f *fakeEventRepoForEventBiz) FileAuditLog(action types.EventActionType, username, operatorEmail, msg string, fileId int) {
	f.fileAuditCalled = true
}

func (f *fakeEventRepoForEventBiz) FileAuditLogWithDuration(action types.EventActionType, username, operatorEmail, msg string, fileId int, duration time.Duration) {
	f.fileDurationCalled = true
}

func newEventBizForTest(repo EventRepo) EventBiz {
	return NewEventBiz(repo)
}

func TestEventBiz_List(t *testing.T) {
	f := &fakeEventRepoForEventBiz{}
	e := newEventBizForTest(f)
	_, _, err := e.List(context.TODO(), &ListEventInput{})
	assert.NoError(t, err)
	assert.True(t, f.listCalled)
}

func TestEventBiz_Show(t *testing.T) {
	f := &fakeEventRepoForEventBiz{}
	e := newEventBizForTest(f)
	got, err := e.Show(context.TODO(), 1)
	assert.NoError(t, err)
	assert.True(t, f.showCalled)
	assert.Equal(t, 1, got.ID)
}

func TestEventBiz_AuditLog(t *testing.T) {
	f := &fakeEventRepoForEventBiz{}
	e := newEventBizForTest(f)
	e.AuditLog(types.EventActionType_Delete, "user", "user@example.com", "msg")
	assert.True(t, f.auditCalled)
}

func TestEventBiz_AuditLogWithChange(t *testing.T) {
	f := &fakeEventRepoForEventBiz{}
	e := newEventBizForTest(f)
	e.AuditLogWithChange(types.EventActionType_Update, "user", "user@example.com", "msg", nil, nil)
	assert.True(t, f.auditChangeCalled)
}

func TestEventBiz_AuditLogWithRequest(t *testing.T) {
	f := &fakeEventRepoForEventBiz{}
	e := newEventBizForTest(f)
	e.AuditLogWithRequest(types.EventActionType_Create, "user", "user@example.com", "msg", map[string]string{"a": "b"})
	assert.True(t, f.auditRequestCalled)
}

func TestEventBiz_FileAuditLog(t *testing.T) {
	f := &fakeEventRepoForEventBiz{}
	e := newEventBizForTest(f)
	e.FileAuditLog(types.EventActionType_Create, "user", "user@example.com", "msg", 1)
	assert.True(t, f.fileAuditCalled)
}

func TestEventBiz_FileAuditLogWithDuration(t *testing.T) {
	f := &fakeEventRepoForEventBiz{}
	e := newEventBizForTest(f)
	e.FileAuditLogWithDuration(types.EventActionType_Create, "user", "user@example.com", "msg", 1, time.Second)
	assert.True(t, f.fileDurationCalled)
}
