package services

import (
	"context"
	"errors"
	"testing"

	"github.com/duc-cnzj/mars-client/v4/picture"
	"github.com/duc-cnzj/mars/internal/app/instance"
	"github.com/duc-cnzj/mars/internal/config"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/duc-cnzj/mars/internal/plugins"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestPictureSvc_AuthFuncOverride(t *testing.T) {
	_, err := new(PictureSvc).AuthFuncOverride(context.TODO(), "")
	assert.Nil(t, err)
}

type testPicturePlugin struct {
}

func (t *testPicturePlugin) Name() string {
	return "test_picture_plugins"
}

func (t *testPicturePlugin) Initialize(args map[string]any) error {
	return nil
}

func (t *testPicturePlugin) Destroy() error {
	return nil
}

type errCtx struct{}

func (t *testPicturePlugin) Get(ctx context.Context, random bool) (*plugins.Picture, error) {
	v := ctx.Value(&errCtx{})
	if v != nil {
		return nil, errors.New("err ctx")
	}
	if random {
		return &plugins.Picture{
			Url:       "https://test.com/random.png",
			Copyright: "@duc-random",
		}, nil
	}
	return &plugins.Picture{
		Url:       "https://test.com/image.png",
		Copyright: "@duc",
	}, nil
}

func TestPictureSvc_Background(t *testing.T) {
	p := new(PictureSvc)
	ctrl := gomock.NewController(t)
	app := mock.NewMockApplicationInterface(ctrl)
	defer ctrl.Finish()
	instance.SetInstance(app)
	app.EXPECT().Config().Return(&config.Config{PicturePlugin: config.Plugin{
		Name: "test_picture_plugins",
		Args: nil,
	}}).AnyTimes()
	pl := &testPicturePlugin{}
	app.EXPECT().RegisterAfterShutdownFunc(gomock.Any()).AnyTimes()
	app.EXPECT().GetPluginByName("test_picture_plugins").Return(pl).AnyTimes()
	background, err := p.Background(context.TODO(), &picture.BackgroundRequest{Random: false})
	assert.Nil(t, err)
	assert.Equal(t, "@duc", background.Copyright)
	assert.Equal(t, "https://test.com/image.png", background.Url)
	background, _ = p.Background(context.TODO(), &picture.BackgroundRequest{Random: true})
	assert.Equal(t, "@duc-random", background.Copyright)
	assert.Equal(t, "https://test.com/random.png", background.Url)
	_, err = p.Background(context.WithValue(context.TODO(), &errCtx{}, "err"), &picture.BackgroundRequest{Random: true})
	assert.NotNil(t, err)
}
