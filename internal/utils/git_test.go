package utils

import (
	"errors"
	"strings"
	"testing"

	"github.com/duc-cnzj/mars/v4/internal/config"
	"github.com/duc-cnzj/mars/v4/internal/mock"
	"github.com/duc-cnzj/mars/v4/internal/testutil"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func mockGitServer(m *gomock.Controller, app *mock.MockApplicationInterface) *mock.MockGitServer {
	gitS := mock.NewMockGitServer(m)
	app.EXPECT().Config().Return(&config.Config{
		GitServerPlugin: config.Plugin{
			Name: "test_git_server",
		},
	}).AnyTimes()
	app.EXPECT().GetPluginByName("test_git_server").Return(gitS).AnyTimes()
	app.EXPECT().RegisterAfterShutdownFunc(gomock.All()).AnyTimes()
	gitS.EXPECT().Initialize(gomock.All()).AnyTimes()
	return gitS
}

func TestDownloadFiles(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	up := mock.NewMockUploader(m)
	app.EXPECT().LocalUploader().Return(up).AnyTimes()
	up.EXPECT().MkDir(gomock.Any(), false).Times(1)
	up.EXPECT().AbsolutePath(gomock.Any()).Return("/tmp")

	gits := mockGitServer(m, app)
	gits.EXPECT().GetFileContentWithSha("1", "xxx", "/app/a.txt").Times(1).Return("aaa", nil)
	gits.EXPECT().GetFileContentWithSha("1", "xxx", "/app/b.txt").Times(1).Return("bbb", nil)

	up.EXPECT().Put(gomock.Any(), strings.NewReader("aaa")).Times(1).Return(nil, nil)
	up.EXPECT().Put(gomock.Any(), strings.NewReader("bbb")).Times(1).Return(nil, nil)

	dir, f, err := DownloadFiles("1", "xxx", []string{"/app/a.txt", "/app/b.txt"})
	assert.Equal(t, "/tmp", dir)
	assert.Nil(t, err)
	up.EXPECT().DeleteDir("/tmp")
	f()

	up.EXPECT().MkDir(gomock.Any(), gomock.Any()).Return(errors.New("xx"))
	_, _, err = DownloadFiles("1", "xxx", []string{"/app/a.txt", "/app/b.txt"})
	assert.Equal(t, "xx", err.Error())
}
