package picture

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/duc-cnzj/mars/v5/internal/application"
	"github.com/duc-cnzj/mars/v5/internal/cache"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/duc-cnzj/mars/v5/internal/util/rand"
)

var (
	nameCartoon          = "picture_cartoon"
	urls        []string = []string{
		"https://api.btstu.cn/sjbz/?lx=dongman",
		"https://www.dmoe.cc/random.php",
	}
)

var _ application.Picture = (*cartoon)(nil)

func init() {
	p := &cartoon{}
	application.RegisterPlugin(p.Name(), p)
}

type cartoon struct {
	cache  cache.Cache
	logger mlog.Logger
}

var client = http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

func (c *cartoon) Get(ctx context.Context, random bool) (*application.PictureItem, error) {
	day := time.Now().Format("2006-01-02")
	seconds := 0
	if !random {
		seconds = 24 * 60 * 60
	}
	bg, _ := c.cache.Remember(cache.NewKey("picture-%s-%d", day, seconds), seconds, func() ([]byte, error) {
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
		}, backoff.NewExponentialBackOff()); err != nil {
			return nil, err
		}

		return []byte(response.Header.Get("Location")), nil
	}, false)

	return &application.PictureItem{
		Url:       string(bg),
		Copyright: "",
	}, nil
}

func (c *cartoon) Name() string {
	return nameCartoon
}

func (c *cartoon) Initialize(app application.App, args map[string]any) error {
	c.logger = app.Logger()
	c.logger.Info("[Plugin]: " + c.Name() + " plugin Initialize...")
	return nil
}

func (c *cartoon) Destroy() error {
	c.logger.Info("[Plugin]: " + c.Name() + " plugin Destroy...")
	return nil
}
