package services

import (
	"context"
	"fmt"
	"io"

	"github.com/duc-cnzj/mars/api/v6/proto/file"
	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/transformer"
	"github.com/duc-cnzj/mars/v6/internal/util/pagination"
	"github.com/duc-cnzj/mars/v6/internal/util/slice"
	"github.com/dustin/go-humanize"
	"github.com/samber/lo"
)

var _ file.FileServer = (*fileSvc)(nil)

// fileSvc 是 file.FileServer 的 gRPC 实现：管理上传文件记录（列表/磁盘信息/删除），
// 经 access 校验访问权限，由 NewFileSvc 构造。
type fileSvc struct {
	file.UnimplementedFileServer
	eventBiz  biz.EventBiz
	fileBiz   biz.FileBiz
	logger    mlog.Logger
	accessBiz biz.AccessBiz
}

// FileSvcDeps 收口 NewFileSvc 的构造依赖，由 wire 按字段注入。
type FileSvcDeps struct {
	EventBiz  biz.EventBiz
	FileBiz   biz.FileBiz
	Logger    mlog.Logger
	AccessBiz biz.AccessBiz
}

// NewFileSvc 收口文件管理服务的构造依赖，由 wire 按字段注入。
func NewFileSvc(deps FileSvcDeps) file.FileServer {
	return &fileSvc{eventBiz: deps.EventBiz, fileBiz: deps.FileBiz, logger: deps.Logger.WithModule("services/file"), accessBiz: deps.AccessBiz}
}

// List 分页列出文件元数据（默认不含软删除记录），按 id 倒序返回。
func (f *fileSvc) List(ctx context.Context, request *file.ListRequest) (*file.ListResponse, error) {
	page, size := pagination.InitByDefault(request.Page, request.PageSize)
	files, pag, err := f.fileBiz.List(ctx, &biz.ListFileInput{
		Page:        page,
		PageSize:    size,
		OrderIDDesc: lo.ToPtr(true),
		// 默认只展示未删除文件（软删过滤常开）。无"回收站"视图；如需，再新增语义明确的字段。
		WithSoftDelete: false,
	})
	if err != nil {
		return nil, logError(ctx, f.logger, err)
	}

	return &file.ListResponse{
		Page:     pag.Page,
		PageSize: pag.PageSize,
		Items:    slice.Map(files, transformer.FromFile),
		Count:    pag.Count,
	}, nil
}

// DiskInfo 返回文件存储盘当前占用大小，并以人类可读格式输出。
func (f *fileSvc) DiskInfo(ctx context.Context, request *file.DiskInfoRequest) (*file.DiskInfoResponse, error) {
	size, err := f.fileBiz.DiskInfo(false)
	if err != nil {
		return nil, logError(ctx, f.logger, err)
	}
	return &file.DiskInfoResponse{
		Usage:         size,
		HumanizeUsage: humanize.Bytes(uint64(size)),
	}, nil
}

// ShowRecords 返回指定文件的上传/执行记录文本（读取存储对象后整体回传）。
func (f *fileSvc) ShowRecords(ctx context.Context, request *file.ShowRecordsRequest) (*file.ShowRecordsResponse, error) {
	records, err := f.fileBiz.ShowRecords(ctx, int(request.Id))
	if err != nil {
		return nil, logError(ctx, f.logger, err)
	}
	defer records.Close()
	all, readErr := io.ReadAll(records)
	if readErr != nil {
		f.logger.ErrorCtx(ctx, readErr)
		return nil, readErr
	}

	return &file.ShowRecordsResponse{Items: []string{string(all)}}, nil
}

// Delete 删除指定文件（软删除）：先取原记录用于审计，再删并落删除审计日志。
func (f *fileSvc) Delete(ctx context.Context, request *file.DeleteRequest) (*file.DeleteResponse, error) {
	record, err := f.fileBiz.GetByID(ctx, int(request.Id))
	if err != nil {
		return nil, logError(ctx, f.logger, err)
	}
	if err := f.fileBiz.Delete(ctx, int(request.Id)); err != nil {
		return nil, logError(ctx, f.logger, err)
	}
	f.eventBiz.FileAuditLog(
		types.EventActionType_Delete,
		biz.MustGetUser(ctx).Name,
		fmt.Sprintf("删除文件: '%s', 该文件由 %s 上传, 大小是 %s", record.Path, record.Username, humanize.Bytes(record.Size)),
		record.ID,
	)

	return &file.DeleteResponse{}, nil
}

// MaxUploadSize 返回上传大小上限，同时给出字节数与人类可读格式。
func (f *fileSvc) MaxUploadSize(ctx context.Context, request *file.MaxUploadSizeRequest) (*file.MaxUploadSizeResponse, error) {
	size := f.fileBiz.MaxUploadSize()
	return &file.MaxUploadSizeResponse{
		HumanizeSize: humanize.Bytes(size),
		Bytes:        uint32(size),
	}, nil
}

// Authorize 是文件服务的 admin 门禁：MaxUploadSize 放行给任意登录用户，
// 其余文件管理方法（列表/删除/详情）仅 admin 可调用。
func (f *fileSvc) Authorize(ctx context.Context, fullMethodName string) (context.Context, error) {
	return f.accessBiz.RequireAdmin(ctx, fullMethodName, file.File_MaxUploadSize_FullMethodName)
}
