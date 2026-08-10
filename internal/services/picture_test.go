package services

import (
	"context"
	"testing"

	"github.com/duc-cnzj/mars/api/v6/proto/picture"
	"github.com/duc-cnzj/mars/v6/internal/application"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewPictureSvc(t *testing.T) {
	svc, _ := newPictureSvcWithMocks(t)
	assert.NotNil(t, svc)
	assert.NotNil(t, svc.picBiz)
}

func Test_pictureSvc_Background(t *testing.T) {
	svc, mocks := newPictureSvcWithMocks(t)
	picBiz := mocks.picBiz
	picBiz.EXPECT().Get(gomock.Any(), true).Return(&application.PictureItem{Url: "http://pic", Copyright: "© 2026"}, nil)
	resp, err := svc.Background(context.TODO(), &picture.BackgroundRequest{Random: true})
	assert.Nil(t, err)
	if assert.NotNil(t, resp) {
		assert.Equal(t, "http://pic", resp.Url)
		assert.Equal(t, "© 2026", resp.Copyright)
	}

	picBiz.EXPECT().Get(gomock.Any(), true).Return(nil, assert.AnError)
	_, err = svc.Background(context.TODO(), &picture.BackgroundRequest{Random: true})
	assert.NotNil(t, err)
}

type pictureSvcMocks struct {
	ctrl   *gomock.Controller
	picBiz *biz.MockPictureBiz
}

func newPictureSvcWithMocks(t *testing.T) (*pictureSvc, *pictureSvcMocks) {
	t.Helper()
	ctrl := gomock.NewController(t)
	mocks := &pictureSvcMocks{
		ctrl:   ctrl,
		picBiz: biz.NewMockPictureBiz(ctrl),
	}
	s, ok := NewPictureSvc(PictureSvcDeps{
		PicBiz: mocks.picBiz,
		Logger: mlog.NewForConfig(nil),
	}).(*pictureSvc)
	if !ok {
		panic("NewPictureSvc returned unexpected type")
	}
	return s, mocks
}
