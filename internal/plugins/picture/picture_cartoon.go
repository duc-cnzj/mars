package picture

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/duc-cnzj/mars/v6/internal/application"
	"github.com/duc-cnzj/mars/v6/internal/data"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/rand"
)

var (
	// nameCartoon 插件注册名。
	nameCartoon = "picture_cartoon"
	// urls 候选图片源，随机挑选一个请求。
	urls = []string{
		"https://www.dmoe.cc/random.php",
	}
)

var _ application.Picture = (*cartoon)(nil)

// cartoonDeps 是 cartoon 插件的依赖视图：只用 Logger 与 Cache。
// Cache 是单插件独有能力，不在 PluginApp 公共接口里，经 Resolve 断言取用。
type cartoonDeps interface {
	Logger() mlog.Logger
	Cache() data.Cache
}

func init() {
	p := &cartoon{}
	application.RegisterPlugin(p.Name(), p)
}

// cartoon 从随机图源请求一张二次元图片，图片地址按天缓存。
type cartoon struct {
	cache  data.Cache
	logger mlog.Logger
}

// client 不跟随重定向，通过响应头 Location 拿到最终图片地址；带超时避免图源挂起时无限阻塞。
var client = http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
	Timeout: 10 * time.Second,
}

// newBackOff 默认指数退避；测试可替换为短超时版本，避免失败路径重试 15 分钟。
var newBackOff = func() backoff.BackOff { return backoff.NewExponentialBackOff() }

// Get 随机请求一个图源并返回图片地址；非 random 时结果缓存 24 小时，random 时不缓存。
func (c *cartoon) Get(ctx context.Context, random bool) (*application.PictureItem, error) {
	day := time.Now().Format("2006-01-02")
	seconds := 0
	if !random {
		seconds = 24 * 60 * 60
	}
	bg, err := c.cache.Remember(data.NewKey("picture-%s-%d", day, seconds), seconds, func() ([]byte, error) {
		var (
			response *http.Response
			err      error
		)
		if err := backoff.Retry(func() error {
			weburl := urls[rand.Intn(len(urls))]
			c.logger.Debugf("[PictureItem]: request %s", weburl)
			response, err = client.Get(weburl)
			if err != nil {
				return err
			}
			defer response.Body.Close()
			if response.StatusCode > 400 {
				c.logger.Debug(errors.New(weburl + ": status code > 400"))
				return errors.New(weburl + ": status code > 400")
			}
			return nil
		}, newBackOff()); err != nil {
			return nil, err
		}

		return []byte(response.Header.Get("Location")), nil
	}, false)
	if err != nil {
		return nil, err
	}

	return &application.PictureItem{
		Url:       string(bg),
		Copyright: "",
	}, nil
}

// Name 返回插件名 picture_cartoon。
func (c *cartoon) Name() string {
	return nameCartoon
}

// Initialize 从宽入口 Resolve 出窄依赖视图，注入 cache 与 logger 并输出初始化日志。
func (c *cartoon) Initialize(app application.PluginApp, args map[string]any) error {
	d := application.Resolve[cartoonDeps](app)
	c.cache = d.Cache()
	c.logger = d.Logger()
	c.logger.Info("[Plugin]: " + c.Name() + " plugin Initialize...")
	return nil
}

// Destroy 输出销毁日志。
func (c *cartoon) Destroy() error {
	c.logger.Info("[Plugin]: " + c.Name() + " plugin Destroy...")
	return nil
}
