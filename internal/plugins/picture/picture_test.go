package picture

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/cenkalti/backoff/v4"
	"github.com/duc-cnzj/mars/v6/internal/app"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/data"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// pictureApp 是 PluginApp 的最小手写 stub。
type pictureApp struct {
	cache  data.Cache
	logger mlog.Logger
}

func (p pictureApp) Logger() mlog.Logger          { return p.logger }
func (p pictureApp) Cache() data.Cache            { return p.cache }
func (p pictureApp) ProjectRepo() biz.ProjectRepo { return nil }

// restore 记录并恢复包级测试缝。
func restore[B any](t *testing.T, target *B, old B) {
	t.Helper()
	t.Cleanup(func() { *target = old })
}

// ---------------------------------------------------------------------------
// picture_cartoon
// ---------------------------------------------------------------------------

// newCartoonServer 返回一个返回 302 + Location 的假图源服务器，并覆写 urls 指向它。
func newCartoonServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", "https://img.example.com/cartoon.png")
		w.WriteHeader(http.StatusFound)
	}))
	t.Cleanup(srv.Close)

	oldURLs := urls
	urls = []string{srv.URL}
	restore(t, &urls, oldURLs)
	return srv
}

// stopBackOff 让 backoff.Retry 单次失败立即返回，避免错误路径重试 15 分钟。
func stopBackOff(t *testing.T) {
	t.Helper()
	old := newBackOff
	newBackOff = func() backoff.BackOff { return &backoff.StopBackOff{} }
	restore(t, &newBackOff, old)
}

func TestCartoonName(t *testing.T) {
	assert.Equal(t, "picture_cartoon", (&cartoon{}).Name())
}

func TestCartoonInitialize_and_Destroy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	cache := data.NewMockCache(ctrl)

	c := &cartoon{}
	require.NoError(t, c.Initialize(pictureApp{cache: cache, logger: mlog.NewForConfig(nil)}, nil))
	assert.Same(t, cache, c.cache)
	assert.NoError(t, c.Destroy())
}

func TestCartoonGet_success_with_location(t *testing.T) {
	newCartoonServer(t)
	// 注意：这里刻意不覆写 newBackOff，让默认的 ExponentialBackOff 体执行一次以覆盖默认分支。

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	cache := data.NewMockCache(ctrl)
	cache.EXPECT().Remember(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ data.CacheKey, _ int, fn func() ([]byte, error), _ bool) ([]byte, error) {
			return fn()
		})

	c := &cartoon{}
	require.NoError(t, c.Initialize(pictureApp{cache: cache, logger: mlog.NewForConfig(nil)}, nil))

	item, err := c.Get(context.TODO(), false)
	require.NoError(t, err)
	assert.Equal(t, "https://img.example.com/cartoon.png", item.Url)
	assert.Empty(t, item.Copyright)
}

func TestCartoonGet_cache_error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	cache := data.NewMockCache(ctrl)
	cache.EXPECT().Remember(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, assert.AnError)

	c := &cartoon{}
	require.NoError(t, c.Initialize(pictureApp{cache: cache, logger: mlog.NewForConfig(nil)}, nil))

	_, err := c.Get(context.TODO(), false)
	assert.ErrorIs(t, err, assert.AnError)
}

func TestCartoonGet_http_error(t *testing.T) {
	stopBackOff(t)

	oldURLs := urls
	urls = []string{"http://127.0.0.1:1/"} // 关闭端口，连接必失败
	restore(t, &urls, oldURLs)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	cache := data.NewMockCache(ctrl)
	cache.EXPECT().Remember(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ data.CacheKey, _ int, fn func() ([]byte, error), _ bool) ([]byte, error) {
			return fn()
		})

	c := &cartoon{}
	require.NoError(t, c.Initialize(pictureApp{cache: cache, logger: mlog.NewForConfig(nil)}, nil))

	_, err := c.Get(context.TODO(), false)
	assert.Error(t, err)
}

func TestCartoonGet_status_greater_than_400(t *testing.T) {
	stopBackOff(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	oldURLs := urls
	urls = []string{srv.URL}
	restore(t, &urls, oldURLs)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	cache := data.NewMockCache(ctrl)
	cache.EXPECT().Remember(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ data.CacheKey, _ int, fn func() ([]byte, error), _ bool) ([]byte, error) {
			return fn()
		})

	c := &cartoon{}
	require.NoError(t, c.Initialize(pictureApp{cache: cache, logger: mlog.NewForConfig(nil)}, nil))

	_, err := c.Get(context.TODO(), false)
	assert.ErrorContains(t, err, "status code > 400")
}

// ---------------------------------------------------------------------------
// picture_bing
// ---------------------------------------------------------------------------

func bingJSON(images ...string) string {
	return `{"images": [` + strings.Join(images, ",") + `]}`
}

func bingImage(url, copyright string) string {
	return `{"url": "` + url + `", "copyright": "` + copyright + `"}`
}

// newBingServer 返回一个固定返回给定 JSON 的假 Bing 服务器，并覆写 url 指向它。
func newBingServer(t *testing.T, body string, hits *atomic.Int64) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)

	oldURL := url
	// url 是含 %d 占位符的格式串（fmt.Sprintf(url, n)），覆写时必须保留占位符。
	url = srv.URL + "?n=%d"
	restore(t, &url, oldURL)
	return srv
}

func TestBingName(t *testing.T) {
	assert.Equal(t, "picture_bing", (&bing{}).Name())
}

func TestBingInitialize_and_Destroy(t *testing.T) {
	b := &bing{}
	require.NoError(t, b.Initialize(pictureApp{logger: mlog.NewForConfig(nil)}, nil))
	assert.NoError(t, b.Destroy())
}

func TestBingGet_fetch_and_cache_hit(t *testing.T) {
	var hits atomic.Int64
	newBingServer(t, bingJSON(bingImage("/th?id=1", "Photo (© Bing)")), &hits)

	b := &bing{}
	require.NoError(t, b.Initialize(pictureApp{logger: mlog.NewForConfig(nil)}, nil))

	item, err := b.Get(context.TODO(), false)
	require.NoError(t, err)
	assert.Equal(t, "https://cn.bing.com/th?id=1", item.Url)
	assert.Equal(t, "Photo ", item.Copyright)
	assert.Equal(t, int64(1), hits.Load())

	// 同一天第二次获取走缓存，不再请求上游。
	item2, err := b.Get(context.TODO(), false)
	require.NoError(t, err)
	assert.Equal(t, item.Url, item2.Url)
	assert.Equal(t, int64(1), hits.Load())
}

func TestBingGet_random_picks_an_image(t *testing.T) {
	var hits atomic.Int64
	newBingServer(t, bingJSON(bingImage("/th?id=1", ""), bingImage("/th?id=2", "")), &hits)

	b := &bing{}
	require.NoError(t, b.Initialize(pictureApp{logger: mlog.NewForConfig(nil)}, nil))

	item, err := b.Get(context.TODO(), true)
	require.NoError(t, err)
	assert.Contains(t, item.Url, "https://cn.bing.com/th?id=")
}

func TestBingGet_cache_expired_refetches(t *testing.T) {
	var hits atomic.Int64
	newBingServer(t, bingJSON(bingImage("/th?id=new", "")), &hits)

	b := &bing{}
	require.NoError(t, b.Initialize(pictureApp{logger: mlog.NewForConfig(nil)}, nil))

	// 手动制造跨天缓存：命中失败后应重新拉取并覆盖缓存。
	b.Lock()
	b.cacheItems = []Item{{URL: "/th?id=stale"}}
	b.cacheDay = "2000-01-01"
	b.Unlock()

	item, err := b.Get(context.TODO(), false)
	require.NoError(t, err)
	assert.Equal(t, "https://cn.bing.com/th?id=new", item.Url)
	assert.Equal(t, int64(1), hits.Load())

	b.RLock()
	assert.Equal(t, "/th?id=new", b.cacheItems[0].URL)
	b.RUnlock()
}

func TestBingGet_empty_images_error(t *testing.T) {
	newBingServer(t, bingJSON(), new(atomic.Int64))

	b := &bing{}
	require.NoError(t, b.Initialize(pictureApp{logger: mlog.NewForConfig(nil)}, nil))

	_, err := b.Get(context.TODO(), false)
	assert.ErrorContains(t, err, "empty images response")
}

func TestBingGet_invalid_json_error(t *testing.T) {
	newBingServer(t, "this is not json", new(atomic.Int64))

	b := &bing{}
	require.NoError(t, b.Initialize(pictureApp{logger: mlog.NewForConfig(nil)}, nil))

	_, err := b.Get(context.TODO(), false)
	assert.Error(t, err)
}

func TestBingGet_http_error(t *testing.T) {
	// 指向关闭端口，bingClient.Get 必失败。
	oldURL := url
	url = "http://127.0.0.1:1/?n=%d"
	restore(t, &url, oldURL)

	b := &bing{}
	require.NoError(t, b.Initialize(pictureApp{logger: mlog.NewForConfig(nil)}, nil))

	_, err := b.Get(context.TODO(), false)
	assert.Error(t, err)
}

// errReadCloser 读取即失败，用于触发 io.ReadAll 错误分支。
type errReadCloser struct{}

func (errReadCloser) Read([]byte) (int, error) { return 0, errors.New("read boom") }
func (errReadCloser) Close() error             { return nil }

// failBodyRT 返回 HTTP 200 但 body 读取失败的响应。
type failBodyRT struct{}

func (failBodyRT) RoundTrip(*http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       errReadCloser{},
		Header:     make(http.Header),
	}, nil
}

func TestBingGet_read_all_error(t *testing.T) {
	// Get 成功但 io.ReadAll 失败：应透传错误，而非吞掉后用空 body 继续。
	oldClient := bingClient
	bingClient = http.Client{Transport: failBodyRT{}}
	restore(t, &bingClient, oldClient)

	b := &bing{}
	require.NoError(t, b.Initialize(pictureApp{logger: mlog.NewForConfig(nil)}, nil))

	_, err := b.Get(context.TODO(), false)
	assert.ErrorContains(t, err, "read boom")
}

func TestBingGet_key_out_of_range_resets_to_zero(t *testing.T) {
	var hits atomic.Int64
	newBingServer(t, bingJSON(bingImage("/th?id=only", "")), &hits)

	// 固定随机下标为 99，触发 key >= len(res) 防御分支。
	oldRand := randIntn
	randIntn = func(int) int { return 99 }
	restore(t, &randIntn, oldRand)

	b := &bing{}
	require.NoError(t, b.Initialize(pictureApp{logger: mlog.NewForConfig(nil)}, nil))

	item, err := b.Get(context.TODO(), true)
	require.NoError(t, err)
	assert.Equal(t, "https://cn.bing.com/th?id=only", item.Url)
}

func TestBingGet_copyright_without_paren_stays(t *testing.T) {
	newBingServer(t, bingJSON(bingImage("/th?id=1", "Plain copyright text")), new(atomic.Int64))

	b := &bing{}
	require.NoError(t, b.Initialize(pictureApp{logger: mlog.NewForConfig(nil)}, nil))

	item, err := b.Get(context.TODO(), false)
	require.NoError(t, err)
	assert.Equal(t, "Plain copyright text", item.Copyright)
}

// TestRegister_interface ensures the plugin satisfies app.Picture.
func TestRegister_interface(t *testing.T) {
	var _ app.Picture = (*cartoon)(nil)
	var _ app.Picture = (*bing)(nil)
}
