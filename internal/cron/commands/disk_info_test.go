package commands

import (
	"testing"

	"github.com/duc-cnzj/mars/v4/internal/mock"
	"github.com/duc-cnzj/mars/v4/internal/testutil"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_diskInfo(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	up := mock.NewMockUploader(m)
	app.EXPECT().Uploader().Return(up)
	up.EXPECT().DirSize().Times(1).Return(int64(10), nil)
	assert.Nil(t, diskInfo())
}
