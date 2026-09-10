package data

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/biz/schematype"
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/data/ent"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/uploader"
	"github.com/duc-cnzj/mars/v6/internal/util/timer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// closedCh 返回一个已关闭的空 channel，供 StreamUploadFile 无需真正发数据时使用。
func closedCh() chan []byte {
	ch := make(chan []byte)
	close(ch)
	return ch
}

// fakeFileInfo 返回一个 size 字节的 os.FileInfo，用于 mock File.Stat 的返回值。
func fakeFileInfo(t *testing.T, size int) os.FileInfo {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "fi")
	require.NoError(t, err)
	require.NoError(t, f.Close())
	if size > 0 {
		require.NoError(t, os.WriteFile(f.Name(), bytes.Repeat([]byte("x"), size), 0o644))
	}
	fi, err := os.Stat(f.Name())
	require.NoError(t, err)
	return fi
}

// newMockFileRepo 构造基于 mock uploader 的 fileRepo，d 用真实 sqlite（recorder 操作不触库）。
func newMockFileRepo(t *testing.T, m *gomock.Controller, up uploader.Uploader) *fileRepo {
	t.Helper()
	entdb, err := NewSqliteDB()
	require.NoError(t, err)
	t.Cleanup(func() { entdb.Close() })
	d := NewDataImpl(&NewDataParams{DB: entdb, Cfg: &config.Config{UploadMaxSize: "1M"}})
	return NewFileRepo(mlog.NewForConfig(nil), d, &noCache{}, up, timer.NewReal()).(*fileRepo)
}

// TestFileRepo_Delete_DBError 覆盖 Delete 中记录行删除失败的错误分支
// SQL 序列：First(file)→DeleteOneID。第 1 条写入
// （eAfter=0）即 DELETE，注入成功；物理文件删除不执行。
func TestFileRepo_Delete_DBError(t *testing.T) {
	ctx := context.TODO()
	client, fd := newFailDB(t, -1, 0)
	d := NewDataImpl(&NewDataParams{DB: client, Cfg: &config.Config{}})
	up, err := uploader.NewDiskUploader(t.TempDir(), mlog.NewForConfig(nil))
	require.NoError(t, err)
	repo := NewFileRepo(mlog.NewForConfig(nil), d, &noCache{}, up, timer.NewReal()).(*fileRepo)

	created, err := repo.Create(ctx, &biz.CreateFileInput{Path: "p", Username: "u", Size: 1, UploadType: schematype.Local})
	require.NoError(t, err)
	fd.Arm()
	assert.Error(t, repo.Delete(ctx, created.ID))
}

// TestFileRepo_ListByCreatedAtRange_DBError 覆盖区间查询 DB 错误分支。
func TestFileRepo_ListByCreatedAtRange_DBError(t *testing.T) {
	d := NewDataImpl(&NewDataParams{DB: mustClosedDB(t), Cfg: &config.Config{}})
	up, err := uploader.NewDiskUploader(t.TempDir(), mlog.NewForConfig(nil))
	require.NoError(t, err)
	repo := NewFileRepo(mlog.NewForConfig(nil), d, &noCache{}, up, timer.NewReal()).(*fileRepo)
	now := time.Now()
	_, err = repo.ListByCreatedAtRange(context.TODO(), now.Add(-time.Hour), now)
	assert.Error(t, err)
}

// TestFileRepo_StreamUploadFile_Errors 覆盖流式上传的三个错误分支：
// MkDir（242）、NewFile（246）、写入块失败（251）。
func TestFileRepo_StreamUploadFile_Errors(t *testing.T) {
	m := gomock.NewController(t)
	t.Cleanup(m.Finish)
	ctx := context.TODO()

	t.Run("MkDir error", func(t *testing.T) {
		up := uploader.NewMockUploader(m)
		disk := uploader.NewMockUploader(m)
		gomock.InOrder(
			up.EXPECT().Disk(StreamUploadFileDisk).Return(disk),
			disk.EXPECT().AbsolutePath(gomock.Any()).Return("/tmp/x"),
			disk.EXPECT().MkDir(gomock.Any(), true).Return(errors.New("mkdir boom")),
		)
		repo := newMockFileRepo(t, m, up)
		_, err := repo.StreamUploadFile(ctx, &biz.StreamUploadFileRequest{Username: "u", FileName: "a.txt", FileData: closedCh()})
		assert.Error(t, err)
	})

	t.Run("NewFile error", func(t *testing.T) {
		up := uploader.NewMockUploader(m)
		disk := uploader.NewMockUploader(m)
		gomock.InOrder(
			up.EXPECT().Disk(StreamUploadFileDisk).Return(disk),
			disk.EXPECT().AbsolutePath(gomock.Any()).Return("/tmp/x"),
			disk.EXPECT().MkDir(gomock.Any(), true).Return(nil),
			up.EXPECT().NewFile("/tmp/x").Return(nil, errors.New("newfile boom")),
		)
		repo := newMockFileRepo(t, m, up)
		_, err := repo.StreamUploadFile(ctx, &biz.StreamUploadFileRequest{Username: "u", FileName: "a.txt", FileData: closedCh()})
		assert.Error(t, err)
	})

	t.Run("Write error", func(t *testing.T) {
		up := uploader.NewMockUploader(m)
		disk := uploader.NewMockUploader(m)
		file := uploader.NewMockFile(m)
		gomock.InOrder(
			up.EXPECT().Disk(StreamUploadFileDisk).Return(disk),
			disk.EXPECT().AbsolutePath(gomock.Any()).Return("/tmp/x"),
			disk.EXPECT().MkDir(gomock.Any(), true).Return(nil),
			up.EXPECT().NewFile("/tmp/x").Return(file, nil),
			file.EXPECT().Write(gomock.Any()).Return(0, errors.New("write boom")),
			file.EXPECT().Close(),
		)
		repo := newMockFileRepo(t, m, up)
		ch := make(chan []byte, 1)
		ch <- []byte("part1")
		close(ch)
		_, err := repo.StreamUploadFile(ctx, &biz.StreamUploadFileRequest{Username: "u", FileName: "a.txt", FileData: ch})
		assert.Error(t, err)
	})
}

// TestRecorder_Write_NewFileError 覆盖 Write 首次建 tmp 文件失败的分支
func TestRecorder_Write_NewFileError(t *testing.T) {
	m := gomock.NewController(t)
	t.Cleanup(m.Finish)
	up := uploader.NewMockUploader(m)
	local := uploader.NewMockUploader(m)
	tmp := uploader.NewMockUploader(m)
	gomock.InOrder(
		up.EXPECT().LocalUploader().Return(local),
		local.EXPECT().Disk("tmp").Return(tmp),
		tmp.EXPECT().NewFile(gomock.Any()).Return(nil, errors.New("newfile boom")),
	)
	repo := newMockFileRepo(t, m, up)
	r := repo.NewRecorder(&biz.UserInfo{Name: "duc"}, &biz.Container{Namespace: "ns", Pod: "p", Container: "c"})
	_, err := r.Write([]byte("data"))
	assert.Error(t, err)
}

// TestRecorder_Write_BufferWriteError 覆盖 Write 写入缓冲（触发底层落盘）失败的分支
// 写入超 20KB 缓冲强制 flush，底层文件 Write 失败。
func TestRecorder_Write_BufferWriteError(t *testing.T) {
	m := gomock.NewController(t)
	t.Cleanup(m.Finish)
	up := uploader.NewMockUploader(m)
	local := uploader.NewMockUploader(m)
	tmp := uploader.NewMockUploader(m)
	file := uploader.NewMockFile(m)
	gomock.InOrder(
		up.EXPECT().LocalUploader().Return(local),
		local.EXPECT().Disk("tmp").Return(tmp),
		tmp.EXPECT().NewFile(gomock.Any()).Return(file, nil),
		file.EXPECT().WriteString(gomock.Any()).Return(0, errors.New("write boom")),
	)
	repo := newMockFileRepo(t, m, up)
	r := repo.NewRecorder(&biz.UserInfo{Name: "duc"}, &biz.Container{Namespace: "ns", Pod: "p", Container: "c"})
	_, err := r.Write(bytes.Repeat([]byte("a"), 30*1024))
	assert.Error(t, err)
}

// TestRecorder_Close_FlushError 覆盖 Close 刷新缓冲失败的分支。
func TestRecorder_Close_FlushError(t *testing.T) {
	m := gomock.NewController(t)
	t.Cleanup(m.Finish)
	up := uploader.NewMockUploader(m)
	local := uploader.NewMockUploader(m)
	tmp := uploader.NewMockUploader(m)
	file := uploader.NewMockFile(m)
	gomock.InOrder(
		up.EXPECT().LocalUploader().Return(local),
		local.EXPECT().Disk("tmp").Return(tmp),
		tmp.EXPECT().NewFile(gomock.Any()).Return(file, nil),
		file.EXPECT().Write(gomock.Any()).Return(0, errors.New("flush boom")),
	)
	repo := newMockFileRepo(t, m, up)
	r := repo.NewRecorder(&biz.UserInfo{Name: "duc"}, &biz.Container{Namespace: "ns", Pod: "p", Container: "c"})
	_, err := r.Write([]byte("small"))
	assert.NoError(t, err)
	assert.Error(t, r.Close())
}

// TestRecorder_Close_ShellNewFileError 覆盖 shell 目标文件创建失败的分支
func TestRecorder_Close_ShellNewFileError(t *testing.T) {
	m := gomock.NewController(t)
	t.Cleanup(m.Finish)
	up := uploader.NewMockUploader(m)
	local := uploader.NewMockUploader(m)
	tmp := uploader.NewMockUploader(m)
	shell := uploader.NewMockUploader(m)
	file := uploader.NewMockFile(m)
	gomock.InOrder(
		up.EXPECT().LocalUploader().Return(local),
		local.EXPECT().Disk("tmp").Return(tmp),
		tmp.EXPECT().NewFile(gomock.Any()).Return(file, nil),
		file.EXPECT().Write(gomock.Any()).DoAndReturn(func(p []byte) (int, error) { return len(p), nil }),
		up.EXPECT().Disk("shell").Return(shell),
		shell.EXPECT().NewFile(gomock.Any()).Return(nil, errors.New("shell newfile boom")),
	)
	repo := newMockFileRepo(t, m, up)
	r := repo.NewRecorder(&biz.UserInfo{Name: "duc"}, &biz.Container{Namespace: "ns", Pod: "p", Container: "c"})
	_, err := r.Write([]byte("data"))
	require.NoError(t, err)
	assert.Error(t, r.Close())
}

// TestRecorder_Close_IOCopyAndStatError 覆盖 io.Copy 失败与
// upFile.Stat 失败两个分支：io.Copy 读源失败触发日志，
// 随后 Stat 失败进入清理分支。
func TestRecorder_Close_IOCopyAndStatError(t *testing.T) {
	m := gomock.NewController(t)
	t.Cleanup(m.Finish)
	up := uploader.NewMockUploader(m)
	local := uploader.NewMockUploader(m)
	tmp := uploader.NewMockUploader(m)
	shell := uploader.NewMockUploader(m)
	src := uploader.NewMockFile(m)
	dst := uploader.NewMockFile(m)
	gomock.InOrder(
		up.EXPECT().LocalUploader().Return(local),
		local.EXPECT().Disk("tmp").Return(tmp),
		tmp.EXPECT().NewFile(gomock.Any()).Return(src, nil),
		src.EXPECT().Write(gomock.Any()).DoAndReturn(func(p []byte) (int, error) { return len(p), nil }),
		up.EXPECT().Disk("shell").Return(shell),
		shell.EXPECT().NewFile(gomock.Any()).Return(dst, nil),
		// 内层 func：Seek → WriteString → io.Copy(Read)，defer 里的 Close/Name/Delete 最后执行。
		src.EXPECT().Seek(gomock.Any(), gomock.Any()).Return(int64(0), nil),
		dst.EXPECT().WriteString(gomock.Any()).Return(1, nil),
		src.EXPECT().Read(gomock.Any()).Return(0, errors.New("copy boom")),
		src.EXPECT().Close(),
		src.EXPECT().Name().Return("/tmp/rec.cast.tmp"),
		local.EXPECT().Delete("/tmp/rec.cast.tmp").Return(nil),
		dst.EXPECT().Stat().Return(nil, errors.New("stat boom")),
		dst.EXPECT().Close(),
		dst.EXPECT().Name().Return("/tmp/rec.cast"),
		up.EXPECT().Delete("/tmp/rec.cast").Return(nil),
	)
	repo := newMockFileRepo(t, m, up)
	r := repo.NewRecorder(&biz.UserInfo{Name: "duc"}, &biz.Container{Namespace: "ns", Pod: "p", Container: "c"})
	_, err := r.Write([]byte("data"))
	require.NoError(t, err)
	assert.Error(t, r.Close())
}

// TestRecorder_Close_EmptyFile 覆盖 Stat 为空文件时清理产物、跳过建记录的分支
// Stat 返回 size=0 → emptyFile 恒 true → up.Delete 被调。
func TestRecorder_Close_EmptyFile(t *testing.T) {
	m := gomock.NewController(t)
	t.Cleanup(m.Finish)
	up := uploader.NewMockUploader(m)
	local := uploader.NewMockUploader(m)
	tmp := uploader.NewMockUploader(m)
	shell := uploader.NewMockUploader(m)
	src := uploader.NewMockFile(m)
	dst := uploader.NewMockFile(m)
	gomock.InOrder(
		up.EXPECT().LocalUploader().Return(local),
		local.EXPECT().Disk("tmp").Return(tmp),
		tmp.EXPECT().NewFile(gomock.Any()).Return(src, nil),
		src.EXPECT().Write(gomock.Any()).DoAndReturn(func(p []byte) (int, error) { return len(p), nil }),
		up.EXPECT().Disk("shell").Return(shell),
		shell.EXPECT().NewFile(gomock.Any()).Return(dst, nil),
		src.EXPECT().Seek(gomock.Any(), gomock.Any()).Return(int64(0), nil),
		dst.EXPECT().WriteString(gomock.Any()).Return(1, nil),
		src.EXPECT().Read(gomock.Any()).Return(0, io.EOF),
		src.EXPECT().Close(),
		src.EXPECT().Name().Return("/tmp/rec.cast.tmp"),
		local.EXPECT().Delete("/tmp/rec.cast.tmp").Return(nil),
		dst.EXPECT().Stat().Return(fakeFileInfo(t, 0), nil),
		dst.EXPECT().Close().Return(nil),
		dst.EXPECT().Name().Return("/tmp/rec.cast"),
		up.EXPECT().Delete("/tmp/rec.cast").Return(nil),
	)
	repo := newMockFileRepo(t, m, up)
	r := repo.NewRecorder(&biz.UserInfo{Name: "duc"}, &biz.Container{Namespace: "ns", Pod: "p", Container: "c"})
	_, err := r.Write([]byte("data"))
	require.NoError(t, err)
	assert.NoError(t, r.Close())
}

// TestRecorder_Close_CreateError 覆盖建文件记录失败的分支。
// fileRepo 用 mock 返回错误；注意 Close 的 defer 会用 upFile.Close 的返回值
// 覆盖 err，因此断言 Error 依赖 upFile.Close 也返回错误。
func TestRecorder_Close_CreateError(t *testing.T) {
	m := gomock.NewController(t)
	t.Cleanup(m.Finish)
	up := uploader.NewMockUploader(m)
	local := uploader.NewMockUploader(m)
	tmp := uploader.NewMockUploader(m)
	shell := uploader.NewMockUploader(m)
	src := uploader.NewMockFile(m)
	dst := uploader.NewMockFile(m)
	fileRepo := NewMockFileRepo(m)
	gomock.InOrder(
		local.EXPECT().Disk("tmp").Return(tmp),
		tmp.EXPECT().NewFile(gomock.Any()).Return(src, nil),
		src.EXPECT().Write(gomock.Any()).DoAndReturn(func(p []byte) (int, error) { return len(p), nil }),
		up.EXPECT().Disk("shell").Return(shell),
		shell.EXPECT().NewFile(gomock.Any()).Return(dst, nil),
		src.EXPECT().Seek(gomock.Any(), gomock.Any()).Return(int64(0), nil),
		dst.EXPECT().WriteString(gomock.Any()).Return(1, nil),
		src.EXPECT().Read(gomock.Any()).Return(0, io.EOF),
		src.EXPECT().Close(),
		src.EXPECT().Name().Return("/tmp/rec.cast.tmp"),
		local.EXPECT().Delete("/tmp/rec.cast.tmp").Return(nil),
		dst.EXPECT().Stat().Return(fakeFileInfo(t, 8), nil),
		up.EXPECT().Type().Return(schematype.Local),
		dst.EXPECT().Name().Return("/tmp/rec.cast"),
		fileRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, errors.New("create boom")),
		dst.EXPECT().Close().Return(errors.New("close boom")),
		dst.EXPECT().Name().Return("/tmp/rec.cast"),
		up.EXPECT().Delete("/tmp/rec.cast").Return(nil),
	)
	rec := &recorder{
		logger:        mlog.NewForConfig(nil),
		timer:         timer.NewReal(),
		container:     &biz.Container{Namespace: "ns", Pod: "p", Container: "c"},
		user:          &biz.UserInfo{Name: "duc"},
		localUploader: local,
		uploader:      up,
		fileRepo:      fileRepo,
	}
	_, err := rec.Write([]byte("data"))
	require.NoError(t, err)
	assert.Error(t, rec.Close())
}

// newFileRepo 构造基于 sqlite + 真实磁盘 uploader 的 fileRepo。
func newFileRepo(t *testing.T) (*fileRepo, *ent.Client, uploader.Uploader) {
	entdb, err := NewSqliteDB()
	require.NoError(t, err)
	t.Cleanup(func() { entdb.Close() })
	d := NewDataImpl(&NewDataParams{
		Cfg:    &config.Config{UploadMaxSize: "10M"},
		DB:     entdb,
		Logger: mlog.NewForConfig(nil),
	})
	up, err := uploader.NewDiskUploader(t.TempDir(), mlog.NewForConfig(nil))
	require.NoError(t, err)
	repo := NewFileRepo(mlog.NewForConfig(nil), d, &noCache{}, up, timer.NewReal()).(*fileRepo)
	return repo, entdb, up
}

// TestFileRepo_List 覆盖分页/软删除包含两种查询路径。
func TestFileRepo_List(t *testing.T) {
	repo, entdb, _ := newFileRepo(t)
	ctx := context.TODO()
	entdb.File.Create().SetPath("a").SetUsername("u").SetSize(1).SetUploadType(schematype.Local).SaveX(ctx)
	entdb.File.Create().SetPath("b").SetUsername("u").SetSize(2).SetUploadType(schematype.Local).SaveX(ctx)

	t.Run("default excludes soft deleted", func(t *testing.T) {
		items, pag, err := repo.List(ctx, &biz.ListFileInput{Page: 1, PageSize: 10})
		assert.NoError(t, err)
		assert.Len(t, items, 2)
		assert.Equal(t, int32(2), pag.Count)
	})

	t.Run("with soft delete includes deleted", func(t *testing.T) {
		first := entdb.File.Query().FirstX(ctx)
		entdb.File.DeleteOneID(first.ID).Exec(ctx)
		items, _, err := repo.List(ctx, &biz.ListFileInput{Page: 1, PageSize: 10, WithSoftDelete: true})
		assert.NoError(t, err)
		assert.Len(t, items, 2)
	})
}

// TestFileRepo_Update 覆盖更新容器路径等字段与不存在记录报错。
func TestFileRepo_Update(t *testing.T) {
	repo, _, _ := newFileRepo(t)
	ctx := context.TODO()
	created, err := repo.Create(ctx, &biz.CreateFileInput{Path: "p", Username: "u", Size: 1, UploadType: schematype.Local})
	require.NoError(t, err)

	updated, err := repo.Update(ctx, &biz.UpdateFileRequest{
		ID: created.ID, ContainerPath: "/tmp/x", Namespace: "ns", Pod: "pod", Container: "c",
	})
	assert.NoError(t, err)
	assert.Equal(t, "/tmp/x", updated.ContainerPath)
	assert.Equal(t, "ns", updated.Namespace)

	_, err = repo.Update(ctx, &biz.UpdateFileRequest{ID: 999999})
	assert.Error(t, err)
}

// TestFileRepo_MaxUploadSize 覆盖构造器从 config 解析的上传上限。
func TestFileRepo_MaxUploadSize(t *testing.T) {
	repo, _, _ := newFileRepo(t)
	assert.Equal(t, uint64(10_000_000), repo.MaxUploadSize())
}

// TestFileRepo_Delete 覆盖删记录 + 删物理文件 + 不存在记录报错。
func TestFileRepo_Delete(t *testing.T) {
	repo, _, up := newFileRepo(t)
	ctx := context.TODO()

	nf, err := up.NewFile("del.txt")
	require.NoError(t, err)
	_, _ = nf.Write([]byte("data"))
	require.NoError(t, nf.Close())
	fpath := nf.Name()

	created, err := repo.Create(ctx, &biz.CreateFileInput{Path: fpath, Username: "u", Size: 4, UploadType: schematype.Local})
	require.NoError(t, err)

	assert.NoError(t, repo.Delete(ctx, created.ID))
	_, err = repo.GetByID(ctx, created.ID)
	assert.Error(t, err)

	err = repo.Delete(ctx, 999999)
	assert.Error(t, err)
}

// TestFileRepo_ShowRecords 覆盖本地读取分支、S3 mock 分支与 DB 错误分支。
func TestFileRepo_ShowRecords(t *testing.T) {
	ctx := context.TODO()

	t.Run("local branch reads content", func(t *testing.T) {
		repo, _, up := newFileRepo(t)
		nf, err := up.NewFile("show.txt")
		require.NoError(t, err)
		_, _ = nf.Write([]byte("hello"))
		require.NoError(t, nf.Close())
		created, err := repo.Create(ctx, &biz.CreateFileInput{Path: nf.Name(), Username: "u", Size: 5, UploadType: schematype.Local})
		require.NoError(t, err)

		rc, err := repo.ShowRecords(ctx, created.ID)
		assert.NoError(t, err)
		data, _ := io.ReadAll(rc)
		assert.Equal(t, "hello", string(data))
		rc.Close()
	})

	t.Run("s3 branch via mock", func(t *testing.T) {
		m := gomock.NewController(t)
		t.Cleanup(m.Finish)
		entdb, _ := NewSqliteDB()
		t.Cleanup(func() { entdb.Close() })
		d := NewDataImpl(&NewDataParams{DB: entdb, Cfg: &config.Config{UploadMaxSize: "1M"}})

		s3mock := uploader.NewMockUploader(m)
		localMock := uploader.NewMockUploader(m)
		gomock.InOrder(
			s3mock.EXPECT().LocalUploader().Return(localMock),
			localMock.EXPECT().Type().Return(schematype.Local),
			s3mock.EXPECT().Type().Return(schematype.S3),
			s3mock.EXPECT().Read("s3path").Return(io.NopCloser(strings.NewReader("s3data")), nil),
		)
		repo := NewFileRepo(mlog.NewForConfig(nil), d, &noCache{}, s3mock, timer.NewReal()).(*fileRepo)

		created := entdb.File.Create().SetPath("s3path").SetUsername("u").SetSize(1).SetUploadType(schematype.S3).SaveX(ctx)
		rc, err := repo.ShowRecords(ctx, created.ID)
		assert.NoError(t, err)
		data, _ := io.ReadAll(rc)
		assert.Equal(t, "s3data", string(data))
	})

	t.Run("missing file errors", func(t *testing.T) {
		repo, _, _ := newFileRepo(t)
		_, err := repo.ShowRecords(ctx, 999999)
		assert.Error(t, err)
	})
}

// TestFileRepo_DiskInfo 覆盖缓存回调、字节编解码、强刷路径。
func TestFileRepo_DiskInfo(t *testing.T) {
	repo, _, _ := newFileRepo(t)
	size, err := repo.DiskInfo(false)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), size)

	size, err = repo.DiskInfo(true)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), size)

	assert.Equal(t, []byte("42"), int64ToByte(42))
	assert.Equal(t, int64(42), byteToInt64([]byte("42")))
	assert.Equal(t, int64(0), byteToInt64([]byte("")))
}

// TestFileRepo_StreamUploadFile 覆盖流式上传：分块写入 → 建记录返回。
func TestFileRepo_StreamUploadFile(t *testing.T) {
	repo, _, _ := newFileRepo(t)
	ch := make(chan []byte, 2)
	ch <- []byte("part1")
	ch <- []byte("part2")
	close(ch)

	f, err := repo.StreamUploadFile(context.TODO(), &biz.StreamUploadFileRequest{
		Username: "duc", Namespace: "ns", Pod: "pod", Container: "c", FileName: "a.txt", FileData: ch,
	})
	assert.NoError(t, err)
	assert.NotNil(t, f)
	data, err := os.ReadFile(f.Path)
	require.NoError(t, err)
	assert.Equal(t, "part1part2", string(data))
}

// TestRecorder_Accessors 覆盖 recorder 简单访问器：Container/User/GetShell/SetShell。
func TestRecorder_Accessors(t *testing.T) {
	repo, _, _ := newFileRepo(t)
	r := repo.NewRecorder(&biz.UserInfo{Name: "duc"}, &biz.Container{Namespace: "ns", Pod: "p", Container: "c"})
	rec := r.(*recorder)

	assert.Equal(t, "ns", rec.Container().Namespace)
	assert.Equal(t, "duc", rec.User().Name)
	assert.Equal(t, "", rec.GetShell())
	rec.SetShell("bash")
	assert.Equal(t, "bash", rec.GetShell())
}

// TestToFile_Nil 覆盖 nil 转换。
func TestToFile_Nil(t *testing.T) {
	assert.Nil(t, toFile(nil))
}

// TestRecorder_Close_WithoutWrite 覆盖未写入直接 Close 的早退分支（buffer 为空）。
func TestRecorder_Close_WithoutWrite(t *testing.T) {
	repo, _, _ := newFileRepo(t)
	r := repo.NewRecorder(&biz.UserInfo{Name: "duc"}, &biz.Container{Namespace: "ns", Pod: "p", Container: "c"})
	assert.NoError(t, r.Close())
}

// TestRecorder_ConcurrentAccessNoRace 复现 Exec 流程里 closeAll 与 k8s 输出流
// 并发访问 recorder 的竞态：Write 持 r.Lock 写 startTime，Duration 曾无锁读 startTime。
// 修复后应在 `go test -race` 下无告警；也一并覆盖 Resize（rcMu）与 File（r.Lock）的并发访问。
func TestRecorder_ConcurrentAccessNoRace(t *testing.T) {
	db, err := NewSqliteDB()
	if err != nil {
		t.Fatalf("NewSqliteDB: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	d := NewDataImpl(&NewDataParams{
		Cfg:    &config.Config{},
		DB:     db,
		Logger: mlog.NewForConfig(nil),
	})

	up, err := uploader.NewDiskUploader(t.TempDir(), mlog.NewForConfig(nil))
	if err != nil {
		t.Fatalf("NewDiskUploader: %v", err)
	}

	repo := NewFileRepo(mlog.NewForConfig(nil), d, &noCache{}, up, timer.NewReal())
	r := repo.NewRecorder(&biz.UserInfo{Name: "duc"}, &biz.Container{Namespace: "n", Pod: "p", Container: "c"})

	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(3)
	// 模拟 k8s 输出流：持续 Write（内部 once.Do 会写 startTime）
	go func() {
		defer wg.Done()
		<-start
		for i := 0; i < 1000; i++ {
			if _, err := r.Write([]byte("hello world\n")); err != nil {
				t.Errorf("Write: %v", err)
				return
			}
		}
	}()
	// 模拟 closeAll：并发读 Duration / File
	go func() {
		defer wg.Done()
		<-start
		for i := 0; i < 1000; i++ {
			_ = r.Duration()
			_ = r.File()
		}
	}()
	// 模拟 resize 队列：并发 Resize
	go func() {
		defer wg.Done()
		<-start
		for i := 0; i < 1000; i++ {
			r.Resize(uint16(i%200)+1, uint16(i%50)+1)
		}
	}()
	close(start)
	wg.Wait()
	_ = r.Close()
}

// TestFileRepo_DeleteRecord_OnlyDeletesRow 覆盖仅删记录不删物理的端口语义：
// 物理文件缺失时 Delete 会因 os.Remove 报错，DeleteRecord 不触碰物理仍成功。
func TestFileRepo_DeleteRecord_OnlyDeletesRow(t *testing.T) {
	db, _ := NewSqliteDB()
	t.Cleanup(func() { db.Close() })
	d := NewDataImpl(&NewDataParams{DB: db, Cfg: &config.Config{}})
	up, err := uploader.NewDiskUploader(t.TempDir(), mlog.NewForConfig(nil))
	assert.NoError(t, err)
	repo := NewFileRepo(mlog.NewForConfig(nil), d, &noCache{}, up, timer.NewReal())

	f, err := repo.Create(context.TODO(), &biz.CreateFileInput{Path: "a.txt", Username: "u", Size: 1, UploadType: up.Type()})
	assert.NoError(t, err)
	assert.NoError(t, repo.DeleteRecord(context.TODO(), f.ID))

	_, err = repo.GetByID(context.TODO(), f.ID)
	assert.Error(t, err)
}

// TestFileRepo_ListByCreatedAtRange 覆盖按创建时间区间取文件的端口，cron 昨日清理用。
func TestFileRepo_ListByCreatedAtRange(t *testing.T) {
	db, _ := NewSqliteDB()
	t.Cleanup(func() { db.Close() })
	d := NewDataImpl(&NewDataParams{DB: db, Cfg: &config.Config{}})
	up, err := uploader.NewDiskUploader(t.TempDir(), mlog.NewForConfig(nil))
	assert.NoError(t, err)
	repo := NewFileRepo(mlog.NewForConfig(nil), d, &noCache{}, up, timer.NewReal())

	_, err = repo.Create(context.TODO(), &biz.CreateFileInput{Path: "p1", Username: "u", Size: 1, UploadType: up.Type()})
	assert.NoError(t, err)
	now := time.Now()

	inside, err := repo.ListByCreatedAtRange(context.TODO(), now.Add(-time.Hour), now.Add(time.Hour))
	assert.NoError(t, err)
	assert.Len(t, inside, 1)
	assert.Equal(t, "p1", inside[0].Path)

	outside, err := repo.ListByCreatedAtRange(context.TODO(), now.Add(-48*time.Hour), now.Add(-47*time.Hour))
	assert.NoError(t, err)
	assert.Len(t, outside, 0)
}
