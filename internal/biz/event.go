package biz

import (
	"context"
	"time"

	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/util/pagination"
)

// EventBiz 封装事件审计与事件派发。
type EventBiz interface {
	// List 分页列出事件。
	List(ctx context.Context, input *ListEventInput) ([]*Event, *pagination.Pagination, error)
	// Show 按 id 查询事件。
	Show(ctx context.Context, id int) (*Event, error)
	// AuditLogWithChange 记录带变更前后内容的审计日志。
	AuditLogWithChange(action types.EventActionType, username string, operatorEmail string, msg string, oldS, newS YamlPrettier)
	// AuditLog 记录不带变更内容的审计日志。
	AuditLog(action types.EventActionType, username string, operatorEmail string, msg string)
	// AuditLogWithRequest 记录携带请求体摘要的审计日志。
	AuditLogWithRequest(action types.EventActionType, username string, operatorEmail string, msg string, req any)
	// FileAuditLog 记录与文件关联的审计日志。
	FileAuditLog(action types.EventActionType, username string, operatorEmail string, msg string, fileId int)
	// FileAuditLogWithDuration 记录与文件关联并携带执行时长的审计日志。
	FileAuditLogWithDuration(action types.EventActionType, username string, operatorEmail string, msg string, fileId int, duration time.Duration)
}

type eventBiz struct {
	eventRepo EventRepo
}

// NewEventBiz 构造 event biz。
func NewEventBiz(eventRepo EventRepo) EventBiz {
	return &eventBiz{eventRepo: eventRepo}
}

// List 分页列出事件（透传 repo）。
func (e *eventBiz) List(ctx context.Context, input *ListEventInput) ([]*Event, *pagination.Pagination, error) {
	return e.eventRepo.List(ctx, input)
}

// Show 按 id 查询事件（透传 repo）。
func (e *eventBiz) Show(ctx context.Context, id int) (*Event, error) {
	return e.eventRepo.Show(ctx, id)
}

// AuditLog 记录不带变更内容的审计日志（透传 repo）。
func (e *eventBiz) AuditLog(action types.EventActionType, username string, operatorEmail string, msg string) {
	e.eventRepo.AuditLog(action, username, operatorEmail, msg)
}

// AuditLogWithChange 记录带变更前后内容的审计日志（透传 repo）。
func (e *eventBiz) AuditLogWithChange(action types.EventActionType, username string, operatorEmail string, msg string, oldS, newS YamlPrettier) {
	e.eventRepo.AuditLogWithChange(action, username, operatorEmail, msg, oldS, newS)
}

// AuditLogWithRequest 记录携带请求体摘要的审计日志（透传 repo）。
func (e *eventBiz) AuditLogWithRequest(action types.EventActionType, username string, operatorEmail string, msg string, req any) {
	e.eventRepo.AuditLogWithRequest(action, username, operatorEmail, msg, req)
}

// FileAuditLog 记录与文件关联的审计日志（透传 repo）。
func (e *eventBiz) FileAuditLog(action types.EventActionType, username string, operatorEmail string, msg string, fileId int) {
	e.eventRepo.FileAuditLog(action, username, operatorEmail, msg, fileId)
}

// FileAuditLogWithDuration 记录与文件关联并携带执行时长的审计日志（透传 repo）。
func (e *eventBiz) FileAuditLogWithDuration(action types.EventActionType, username string, operatorEmail string, msg string, fileId int, duration time.Duration) {
	e.eventRepo.FileAuditLogWithDuration(action, username, operatorEmail, msg, fileId, duration)
}

// EventRepo 是事件与审计日志仓库端口。
type EventRepo interface {
	// List 分页列出事件。
	List(ctx context.Context, input *ListEventInput) (events []*Event, pag *pagination.Pagination, err error)
	// Show 按 id 查询事件。
	Show(ctx context.Context, id int) (*Event, error)
	// Dispatch 派发领域事件。
	Dispatch(created EventKey, createdData any)
	// AuditLogWithChange 记录带变更前后内容的审计日志。
	AuditLogWithChange(action types.EventActionType, username string, operatorEmail string, msg string, oldS, newS YamlPrettier)
	// AuditLog 记录不带变更内容的审计日志。
	AuditLog(action types.EventActionType, username string, operatorEmail string, msg string)
	// AuditLogWithRequest 记录携带请求体摘要的审计日志。
	AuditLogWithRequest(action types.EventActionType, username string, operatorEmail string, msg string, req any)
	// FileAuditLog 记录与文件关联的审计日志。
	FileAuditLog(action types.EventActionType, username string, operatorEmail string, msg string, fileId int)
	// FileAuditLogWithDuration 记录与文件关联并携带执行时长的审计日志。
	FileAuditLogWithDuration(action types.EventActionType, username string, operatorEmail string, msg string, fileId int, duration time.Duration)
	// HandleAuditLog 消费审计事件并写入。
	HandleAuditLog(data any, e EventKey) error
}
