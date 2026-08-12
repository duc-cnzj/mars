package biz

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/util/pagination"
)

// FileBiz 收口文件记录业务：上传/拷贝记录的持久化、读取与磁盘用量。
type FileBiz interface {
	// MaxUploadSize 返回上传大小上限（字节）。
	MaxUploadSize() uint64
	// Delete 删除文件记录。
	Delete(ctx context.Context, id int) error
	// ShowRecords 读取文件记录的内容流。
	ShowRecords(ctx context.Context, id int) (io.ReadCloser, error)
	// DiskInfo 返回磁盘用量信息，force 控制是否强制刷新缓存。
	DiskInfo(force bool) (int64, error)
	// List 分页列出文件记录。
	List(ctx context.Context, input *ListFileInput) ([]*File, *pagination.Pagination, error)
	// GetByID 按 id 查询文件记录。
	GetByID(ctx context.Context, id int) (*File, error)
	// Create 创建文件记录。
	Create(ctx context.Context, input *CreateFileInput) (*File, error)
	// NewRecorder 为指定用户/容器构造终端会话录制器。
	NewRecorder(user *UserInfo, container *Container) Recorder
	// Update 更新文件记录（拷贝目标路径等）。
	Update(ctx context.Context, i *UpdateFileRequest) (*File, error)
	// StreamUploadFile 流式上传文件并落库。
	StreamUploadFile(ctx context.Context, input *StreamUploadFileRequest) (*File, error)
}

type fileBiz struct {
	fileRepo FileRepo
}

// NewFileBiz 构造 file biz。
func NewFileBiz(fileRepo FileRepo) FileBiz {
	return &fileBiz{fileRepo: fileRepo}
}

// MaxUploadSize 返回上传大小上限（透传 repo）。
func (f *fileBiz) MaxUploadSize() uint64 { return f.fileRepo.MaxUploadSize() }

// Delete 校验 id 后删除文件。
func (f *fileBiz) Delete(ctx context.Context, id int) error {
	if id <= 0 {
		return errs.WrapInvalidArgument(errors.New("file id 不能小于等于 0"), "delete file")
	}
	return f.fileRepo.Delete(ctx, id)
}

// ShowRecords 返回文件的读取流（透传 repo）。
func (f *fileBiz) ShowRecords(ctx context.Context, id int) (io.ReadCloser, error) {
	return f.fileRepo.ShowRecords(ctx, id)
}

// DiskInfo 返回存储磁盘占用（透传 repo）。
func (f *fileBiz) DiskInfo(force bool) (int64, error) { return f.fileRepo.DiskInfo(force) }

// List 分页列出文件（透传 repo）。
func (f *fileBiz) List(ctx context.Context, input *ListFileInput) ([]*File, *pagination.Pagination, error) {
	return f.fileRepo.List(ctx, input)
}

// GetByID 按 id 查询文件（透传 repo）。
func (f *fileBiz) GetByID(ctx context.Context, id int) (*File, error) {
	return f.fileRepo.GetByID(ctx, id)
}

// Create 校验输入后创建文件记录。
func (f *fileBiz) Create(ctx context.Context, input *CreateFileInput) (*File, error) {
	if input == nil || input.Path == "" {
		return nil, errs.WrapInvalidArgument(errors.New("file 不能为空或 path 不能为空"), "create file")
	}
	return f.fileRepo.Create(ctx, input)
}

// NewRecorder 基于用户与容器创建文件记录器（透传 repo）。
func (f *fileBiz) NewRecorder(user *UserInfo, container *Container) Recorder {
	return f.fileRepo.NewRecorder(user, container)
}

// Update 校验输入后更新文件。
func (f *fileBiz) Update(ctx context.Context, i *UpdateFileRequest) (*File, error) {
	if i == nil || i.ID <= 0 {
		return nil, errs.WrapInvalidArgument(errors.New("file 不能为空或 id 不能小于等于 0"), "update file")
	}
	return f.fileRepo.Update(ctx, i)
}

// StreamUploadFile 校验输入后流式上传文件。
func (f *fileBiz) StreamUploadFile(ctx context.Context, input *StreamUploadFileRequest) (*File, error) {
	if input == nil || input.FileName == "" {
		return nil, errs.WrapInvalidArgument(errors.New("file 不能为空或文件名不能为空"), "stream upload file")
	}
	return f.fileRepo.StreamUploadFile(ctx, input)
}

// FileRepo 是文件记录仓库端口，覆盖上传/拷贝文件记录的持久化与检索。
type FileRepo interface {
	// MaxUploadSize 返回上传大小上限（字节）。
	MaxUploadSize() uint64
	// Delete 删除文件记录。
	Delete(ctx context.Context, id int) error
	// ShowRecords 读取文件记录的内容流。
	ShowRecords(ctx context.Context, id int) (io.ReadCloser, error)
	// DiskInfo 返回磁盘用量信息，force 控制是否强制刷新缓存。
	DiskInfo(force bool) (int64, error)
	// List 分页列出文件记录。
	List(ctx context.Context, input *ListFileInput) ([]*File, *pagination.Pagination, error)
	// GetByID 按 id 查询文件记录。
	GetByID(ctx context.Context, id int) (*File, error)
	// Create 创建文件记录。
	Create(todo context.Context, input *CreateFileInput) (*File, error)
	// NewRecorder 为指定用户/容器构造终端会话录制器。
	NewRecorder(user *UserInfo, container *Container) Recorder
	// Update 更新文件记录（拷贝目标路径等）。
	Update(ctx context.Context, i *UpdateFileRequest) (*File, error)
	// StreamUploadFile 流式上传文件并落库。
	StreamUploadFile(ctx context.Context, input *StreamUploadFileRequest) (*File, error)
	// DeleteRecord 仅删除文件记录行，不触碰物理文件。用于 cron 清理对账场景：
	// 物理文件已不在存储（Exists 为 false），只剩孤儿记录待移除。
	DeleteRecord(ctx context.Context, id int) error
	// ListByCreatedAtRange 按创建时间范围取文件（cron 昨日清理对账用）。
	ListByCreatedAtRange(ctx context.Context, start, end time.Time) ([]*File, error)
}

// Recorder 是终端会话录制器端口，负责持久化 shell 会话的输入输出。
type Recorder interface {
	// Resize 处理终端尺寸变化。
	Resize(width, height uint16)
	// Write 写入终端输出字节。
	Write(p []byte) (n int, err error)
	// Close 关闭录制。
	Close() error
	// SetShell 设置 shell 类型。
	SetShell(string)
	// GetShell 获取 shell 类型。
	GetShell() string
	// File 获取关联的文件记录。
	File() *File
	// Duration 获取会话已录制时长。
	Duration() time.Duration
	// User 获取会话所属用户。
	User() *UserInfo
	// Container 获取会话所属容器三元组。
	Container() *Container
}
