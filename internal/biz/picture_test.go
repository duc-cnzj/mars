package biz

import (
	"context"
	"errors"
	"testing"

	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/stretchr/testify/assert"
)

// fakePictureGetter 记录 Get 调用参数并返回罐头数据/错误。
type fakePictureGetter struct {
	called *bool
	random *bool
	item   *PictureItem
	err    error
}

func (f *fakePictureGetter) Get(ctx context.Context, random bool) (*PictureItem, error) {
	if f.called != nil {
		*f.called = true
	}
	if f.random != nil {
		*f.random = random
	}
	return f.item, f.err
}

// fakePictureProvider 返回固定的 getter。
type fakePictureProvider struct {
	getter PictureGetter
}

func (f *fakePictureProvider) Picture() PictureGetter { return f.getter }

func newPictureBizForTest(getter PictureGetter) PictureBiz {
	return NewPictureBiz(mlog.NewForConfig(nil), &fakePictureProvider{getter: getter})
}

func TestPictureBiz_Get_Valid(t *testing.T) {
	called, random := false, false
	b := newPictureBizForTest(&fakePictureGetter{
		called: &called,
		random: &random,
		item:   &PictureItem{Url: "http://x/1.png", Copyright: "c"},
	})
	got, err := b.Get(context.TODO(), true)
	assert.NoError(t, err)
	assert.True(t, called)
	assert.True(t, random)
	assert.Equal(t, "http://x/1.png", got.Url)
	assert.Equal(t, "c", got.Copyright)
}

func TestPictureBiz_Get_PropagatesError(t *testing.T) {
	b := newPictureBizForTest(&fakePictureGetter{err: errors.New("pic down")})
	got, err := b.Get(context.TODO(), false)
	assert.Nil(t, got)
	assert.EqualError(t, err, "pic down")
}
