package data

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/data/ent"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/file"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/schema/mixin"
	"github.com/duc-cnzj/mars/v6/internal/data/filters"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/uploader"
	"github.com/duc-cnzj/mars/v6/internal/util/pagination"
	"github.com/duc-cnzj/mars/v6/internal/util/rand"
	"github.com/duc-cnzj/mars/v6/internal/util/slice"
	"github.com/duc-cnzj/mars/v6/internal/util/timer"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// toFile 把 ent.File 转换为 biz.File（nil 安全）。
func toFile(file *ent.File) *biz.File {
	if file == nil {
		return nil
	}
	return &biz.File{
		ID:            file.ID,
		CreatedAt:     file.CreatedAt,
		UpdatedAt:     file.UpdatedAt,
		DeletedAt:     file.DeletedAt,
		UploadType:    file.UploadType,
		Path:          file.Path,
		Size:          file.Size,
		Username:      file.Username,
		Namespace:     file.Namespace,
		Pod:           file.Pod,
		Container:     file.Container,
		ContainerPath: file.ContainerPath,
	}
}

var _ biz.FileRepo = (*fileRepo)(nil)

// fileRepo 是 biz.FileRepo 的实现：封装文件记录 CRUD 与磁盘用量缓存，
// 并装配 recorder（终端会话录制器）与流式上传。
type fileRepo struct {
	logger   mlog.Logger
	uploader uploader.Uploader
	timer    timer.Timer

	cache         Cache
	maxUploadSize uint64
	d             dataStore
}

// NewFileRepo 构造文件 repo 实现，注入日志、dataStore、缓存、上传器与计时器，
// 并读取配置的最大上传字节数。
func NewFileRepo(
	logger mlog.Logger,
	d dataStore,
	c Cache,
	up uploader.Uploader,
	t timer.Timer,
) biz.FileRepo {
	return &fileRepo{
		cache:         c,
		logger:        logger.WithModule("repo/file"),
		uploader:      up,
		timer:         t,
		d:             d,
		maxUploadSize: d.Config().MaxUploadSize(),
	}
}

// List 分页查询文件记录，支持按 ID 倒序与软删除可见开关。
func (repo *fileRepo) List(ctx context.Context, input *biz.ListFileInput) ([]*biz.File, *pagination.Pagination, error) {
	var db = repo.d.DB()
	queryCtx := ctx
	if input.WithSoftDelete {
		queryCtx = mixin.SkipSoftDelete(queryCtx)
	}
	query := db.File.Query().
		Where(filters.IfOrderByIDDesc(input.OrderIDDesc))
	files := query.Clone().
		Offset(pagination.GetPageOffset(input.Page, input.PageSize)).
		Limit(int(input.PageSize)).
		AllX(queryCtx)
	count := query.Clone().
		CountX(queryCtx)

	return slice.Map(files, toFile), pagination.NewPagination(input.Page, input.PageSize, count), nil
}

// Create 新建一条文件记录并落库。
func (repo *fileRepo) Create(todo context.Context, input *biz.CreateFileInput) (*biz.File, error) {
	var db = repo.d.DB()
	save, err := db.File.Create().
		SetPath(input.Path).
		SetUsername(input.Username).
		SetNamespace(input.Namespace).
		SetPod(input.Pod).
		SetContainer(input.Container).
		SetSize(input.Size).
		SetUploadType(input.UploadType).
		Save(todo)
	return toFile(save), errs.Wrap(err, "create file")
}

// GetByID 按 ID 查询单条文件记录。
func (repo *fileRepo) GetByID(ctx context.Context, id int) (*biz.File, error) {
	var db = repo.d.DB()
	first, err := db.File.Query().Where(file.ID(id)).First(ctx)
	return toFile(first), errs.Wrap(err, "get file")
}

// Update 更新文件记录的执行上下文（容器/命名空间/Pod/容器路径）。
func (repo *fileRepo) Update(ctx context.Context, i *biz.UpdateFileRequest) (*biz.File, error) {
	var db = repo.d.DB()
	first, err := db.File.UpdateOneID(i.ID).
		SetContainerPath(i.ContainerPath).
		SetNamespace(i.Namespace).
		SetPod(i.Pod).
		SetContainer(i.Container).
		Save(ctx)
	return toFile(first), errs.Wrap(err, "update file")
}

// MaxUploadSize 返回配置的最大上传字节数。
func (repo *fileRepo) MaxUploadSize() uint64 {
	return repo.maxUploadSize
}

// Delete 删除文件记录并连带删除物理文件。
func (repo *fileRepo) Delete(ctx context.Context, id int) error {
	var db = repo.d.DB()
	f, err := db.File.Query().Where(file.ID(id)).First(ctx)
	if err != nil {
		return errs.Wrap(err, "delete file")
	}
	if err = db.File.DeleteOneID(id).Exec(ctx); err != nil {
		return errs.Wrap(err, "delete file")
	}
	return errs.Wrap(repo.uploader.Delete(f.Path), "delete file upload")
}

// DeleteRecord 仅删除文件记录行，不触碰物理文件。cron CleanUploadFiles 对账时
// 已确认物理文件不在存储（Exists 为 false），只需移除孤儿记录，避免 fileRepo.Delete
// 对缺失物理文件的 os.Remove 报错。
func (repo *fileRepo) DeleteRecord(ctx context.Context, id int) error {
	var db = repo.d.DB()
	return errs.Wrap(db.File.DeleteOneID(id).Exec(ctx), "delete file record")
}

// ListByCreatedAtRange 按创建时间区间查询文件，cron CleanUploadFiles 取昨日文件对账。
func (repo *fileRepo) ListByCreatedAtRange(ctx context.Context, start, end time.Time) ([]*biz.File, error) {
	var db = repo.d.DB()
	all, err := db.File.Query().Where(file.CreatedAtGTE(start), file.CreatedAtLTE(end)).All(ctx)
	if err != nil {
		return nil, errs.Wrap(err, "list file by created range")
	}
	return slice.Map(all, toFile), nil
}

// ShowRecords 按 ID 读取文件记录的物理文件内容流（按上传类型选择本地或远端上传器）。
func (repo *fileRepo) ShowRecords(ctx context.Context, id int) (io.ReadCloser, error) {
	var db = repo.d.DB()
	f, err := db.File.Query().Where(file.ID(id)).First(ctx)
	if err != nil {
		return nil, errs.Wrap(err, "show file records")
	}
	var up uploader.Uploader
	switch f.UploadType {
	case repo.uploader.LocalUploader().Type():
		up = repo.uploader.LocalUploader()
	case repo.uploader.Type():
		up = repo.uploader
	}
	read, err := up.Read(f.Path)
	return read, errs.Wrap(err, "read file records")
}

// int64ToByte 把磁盘占用字节数编码为可缓存的字节切片（字符串十进制）。
func int64ToByte(i int64) []byte {
	return []byte(fmt.Sprintf("%d", i))
}

// byteToInt64 把 int64ToByte 的产物解码回字节数；空/非法值按 0 处理。
func byteToInt64(remember []byte) int64 {
	atoi, _ := strconv.Atoi(string(remember))
	return int64(atoi)
}

// dirSizeCacheSeconds 是磁盘占用缓存的有效秒数（15 分钟）。
const dirSizeCacheSeconds = 900

// DiskInfo 读取磁盘占用字节数，经 cache 缓存并支持 force 强刷。
func (repo *fileRepo) DiskInfo(force bool) (int64, error) {
	remember, err := repo.cache.Remember(NewKey("dir-size"), dirSizeCacheSeconds, func() ([]byte, error) {
		size, err := repo.uploader.DirSize()
		return int64ToByte(size), err
	}, force)

	return byteToInt64(remember), errs.Wrap(err, "disk info")
}

// NewRecorder 构造终端会话录制器，绑定文件 repo、本地/远端上传器与计时器。
func (repo *fileRepo) NewRecorder(user *biz.UserInfo, container *biz.Container) biz.Recorder {
	return &recorder{
		fileRepo:      repo,
		logger:        repo.logger,
		timer:         repo.timer,
		container:     container,
		user:          user,
		localUploader: repo.uploader.LocalUploader(),
		uploader:      repo.uploader,
	}
}

// StreamUploadFileDisk 是流式上传落盘的目标磁盘名。
const StreamUploadFileDisk = "grpc_upload"

// StreamUploadFile 流式接收文件分片落盘为一条文件记录（grpc 上传路径）。
func (repo *fileRepo) StreamUploadFile(ctx context.Context, input *biz.StreamUploadFileRequest) (*biz.File, error) {
	tracer := otel.Tracer("")
	ctx, span := tracer.Start(ctx, "fileRepo/StreamUploadFile")
	defer span.End()
	span.SetAttributes(
		attribute.Key("username").String(input.Username),
		attribute.Key("namespace").String(input.Namespace),
		attribute.Key("pod").String(input.Pod),
		attribute.Key("container").String(input.Container),
		attribute.Key("file_name").String(input.FileName),
	)
	disk := repo.uploader.Disk(StreamUploadFileDisk)
	now := repo.timer.Now()
	p := fmt.Sprintf("%s/%s/%s/%s",
		input.Username,
		now.Format("2006-01-02"),
		fmt.Sprintf("%s-%s", now.Format("15-04-05"), rand.String(20)),
		filepath.Base(input.FileName))
	fpath := disk.AbsolutePath(p)
	err := disk.MkDir(filepath.Dir(fpath), true)
	if err != nil {
		return nil, errs.Wrap(err, "stream upload mkdir")
	}
	newFile, err := repo.uploader.NewFile(fpath)
	if err != nil {
		return nil, errs.Wrap(err, "stream upload new file")
	}
	defer newFile.Close()
	for data := range input.FileData {
		if _, err := newFile.Write(data); err != nil {
			return nil, errs.Wrap(err, "stream upload write")
		}
	}
	stat, _ := newFile.Stat()
	return repo.Create(ctx, &biz.CreateFileInput{
		Path:       newFile.Name(),
		Username:   input.Username,
		Size:       uint64(stat.Size()),
		UploadType: disk.Type(),
		Namespace:  input.Namespace,
		Pod:        input.Pod,
		Container:  input.Container,
	})
}

// recorder 是 biz.Recorder 的实现：把终端输出缓冲写入 .cast 临时文件，
// 关闭时转存为正式录制文件并生成文件记录。
type recorder struct {
	sync.RWMutex

	file     *biz.File
	fileRepo biz.FileRepo

	logger    mlog.Logger
	timer     timer.Timer
	container *biz.Container
	f         uploader.File
	startTime time.Time

	user *biz.UserInfo
	once sync.Once

	buffer *bufio.Writer

	shellMu sync.RWMutex
	shell   string

	rcMu          sync.RWMutex
	width, height uint16

	localUploader uploader.Uploader
	uploader      uploader.Uploader
}

// Container 返回录制关联的容器信息。
func (r *recorder) Container() *biz.Container {
	return r.container
}

// User 返回录制关联的用户信息。
func (r *recorder) User() *biz.UserInfo {
	return r.user
}

// Duration 返回录制已持续的时间。startTime 在 Write 的 once.Do 里持 r.Lock 写入，
// 这里必须同样持锁读，否则 closeAll 与 k8s 输出流并发时（Write 设置 startTime）会触发数据竞争。
func (r *recorder) Duration() time.Duration {
	r.RLock()
	defer r.RUnlock()
	return r.timer.Since(r.startTime)
}

// File 返回录制结束后生成的文件记录。
func (r *recorder) File() *biz.File {
	r.Lock()
	defer r.Unlock()
	return r.file
}

// max 返回两数中较大者（终端行列尺寸取历史最大值）。
func max[T int | uint16 | uint64](a, b T) T {
	if a < b {
		return b
	}
	return a
}

// GetShell 返回录制的 shell 路径。
func (r *recorder) GetShell() string {
	r.shellMu.RLock()
	defer r.shellMu.RUnlock()
	return r.shell
}

// SetShell 设置录制的 shell 路径。
func (r *recorder) SetShell(sh string) {
	r.shellMu.Lock()
	defer r.shellMu.Unlock()
	r.shell = sh
}

// Resize 更新终端行列尺寸（透传给 HeadLineColRow 记录历史最大值）。
func (r *recorder) Resize(width, height uint16) {
	r.HeadLineColRow(width, height)
}

// HeadLineColRow 记录终端行列尺寸的历史最大值。
func (r *recorder) HeadLineColRow(width, height uint16) {
	r.rcMu.Lock()
	defer r.rcMu.Unlock()
	r.width = max(r.width, width)
	r.height = max(r.height, height)
}

// Write 把终端输出按 asciinema 格式缓冲写入临时录制文件，首次写入时初始化文件。
func (r *recorder) Write(data []byte) (n int, err error) {
	r.Lock()
	defer r.Unlock()
	r.once.Do(func() {
		var file uploader.File
		file, err = r.localUploader.Disk("tmp").NewFile(fmt.Sprintf("%s/%s/%s",
			r.user.Name,
			r.timer.Now().Format("2006-01-02"),
			fmt.Sprintf("recorder-%s-%s-%s-%s.cast.tmp", r.container.Namespace, r.container.Pod, r.container.Container, rand.String(20))))
		if err != nil {
			return
		}
		r.f = file
		r.buffer = bufio.NewWriterSize(r.f, 1024*20)
		r.startTime = r.timer.Now()
		r.HeadLineColRow(106, 25)
	})
	if err != nil {
		return 0, errs.Wrap(err, "recorder create file")
	}
	marshal, _ := json.Marshal(string(data))
	_, err = io.WriteString(r.buffer, fmt.Sprintf(writeLine, float64(r.timer.Since(r.startTime).Microseconds())/1000000, string(marshal)))
	if err != nil {
		return 0, errs.Wrap(err, "recorder write buffer")
	}
	return len(data), nil
}

// startLine/writeLine 是 asciinema 录制文件的头部与行格式模板。
const (
	startLine = "{\"version\": 2, \"width\": %d, \"height\": %d, \"timestamp\": %d, \"env\": {\"SHELL\": \"%s\", \"TERM\": \"xterm-256color\"}}\n"
	writeLine = "[%.6f, \"o\", %s]\n"
)

// Close 冲刷缓冲、转存正式录制文件，并生成文件记录（空文件不落库）。
func (r *recorder) Close() error {
	r.Lock()
	defer r.Unlock()
	defer r.logger.Info("recorder close")
	var (
		err error

		localUploader = r.localUploader
		up            = r.uploader
	)
	if r.buffer == nil || r.startTime.IsZero() {
		return nil
	}
	if err := r.buffer.Flush(); err != nil {
		return errs.Wrap(err, "recorder flush")
	}

	upFile, err := up.Disk("shell").NewFile(fmt.Sprintf("%s/%s/%s",
		r.user.Name,
		r.timer.Now().Format("2006-01-02"),
		fmt.Sprintf("recorder-%s-%s-%s-%s.cast", r.container.Namespace, r.container.Pod, r.container.Container, rand.String(20))))
	if err != nil {
		return errs.Wrap(err, "recorder create shell file")
	}
	func() {
		defer func() {
			r.f.Close()
			localUploader.Delete(r.f.Name())
		}()
		r.f.Seek(0, 0)
		func() {
			r.rcMu.RLock()
			defer r.rcMu.RUnlock()
			// shell 由 shellMu 保护，不能借 rcMu 的锁读，否则与 SetShell 并发会数据竞争。
			r.shellMu.RLock()
			shell := r.shell
			r.shellMu.RUnlock()
			io.WriteString(upFile, fmt.Sprintf(startLine, r.width, r.height, r.startTime.Unix(), shell))
		}()
		if _, err := io.Copy(upFile, r.f); err != nil {
			r.logger.Error(err)
		}
	}()

	stat, e := upFile.Stat()
	if e != nil {
		upFile.Close()
		up.Delete(upFile.Name())
		return errs.Wrap(e, "recorder stat")
	}
	var emptyFile = true
	defer func() {
		err = upFile.Close()
		if emptyFile {
			up.Delete(upFile.Name())
		}
	}()
	if stat.Size() > 0 {
		r.file, err = r.fileRepo.Create(context.TODO(), &biz.CreateFileInput{
			UploadType: up.Type(),
			Path:       upFile.Name(),
			Size:       uint64(stat.Size()),
			Username:   r.user.Name,
			Namespace:  r.container.Namespace,
			Pod:        r.container.Pod,
			Container:  r.container.Container,
		})
		if err != nil {
			return err
		}

		emptyFile = false
	}
	return err
}
