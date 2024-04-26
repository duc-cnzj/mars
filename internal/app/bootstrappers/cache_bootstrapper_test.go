package bootstrappers

import (
	"testing"

	"github.com/duc-cnzj/mars/v4/internal/adapter"

	"github.com/duc-cnzj/mars/v4/internal/cache"
	"github.com/duc-cnzj/mars/v4/internal/config"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/mock"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type cacheMatcher struct {
	wantsType any
	t         *testing.T
}

func (c *cacheMatcher) Matches(x any) bool {
	assert.IsType(c.t, c.wantsType, x.(contracts.CacheInterface).Store())
	return true
}

func (c *cacheMatcher) String() string {
	return ""
}

type callbackMatcher struct {
	cb func(any) bool
	t  *testing.T
}

func (c *callbackMatcher) Matches(x any) bool {
	return c.cb(x)
}

func (c *callbackMatcher) String() string {
	return ""
}

func TestCacheBootstrapper_Bootstrap(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	app := mock.NewMockApplicationInterface(controller)
	app.EXPECT().Config().Return(&config.Config{
		DBDriver:    "sqlite",
		CacheDriver: "db",
	}).Times(1)
	app.EXPECT().Singleflight().Times(2)
	app.EXPECT().SetCache(&cacheMatcher{
		wantsType: adapter.NewGoCacheAdapter(nil),
		t:         t,
	})
	app.EXPECT().SetCacheLock(&callbackMatcher{
		cb: func(a any) bool {
			l, ok := a.(contracts.Locker)
			if !ok {
				return false
			}
			return l.Type() == "memory"
		},
		t: t,
	}).Times(1)
	assert.Nil(t, (&CacheBootstrapper{}).Bootstrap(app))
	app.EXPECT().Config().Return(&config.Config{
		DBDriver:    "mysql",
		CacheDriver: "db",
	}).Times(1)
	app.EXPECT().SetCache(&cacheMatcher{
		wantsType: cache.NewDBStore(nil),
		t:         t,
	})
	app.EXPECT().SetCacheLock(&callbackMatcher{
		cb: func(a any) bool {
			l, ok := a.(contracts.Locker)
			if !ok {
				return false
			}
			return l.Type() == "database"
		},
		t: t,
	}).Times(1)
	assert.Nil(t, (&CacheBootstrapper{}).Bootstrap(app))
	app.EXPECT().Config().Return(&config.Config{
		CacheDriver: "xxxx",
	}).Times(1)
	assert.Error(t, (&CacheBootstrapper{}).Bootstrap(app))
}

func TestCacheBootstrapper_Tags(t *testing.T) {
	assert.Equal(t, []string{}, (&CacheBootstrapper{}).Tags())
}
