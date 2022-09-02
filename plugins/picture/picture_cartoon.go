package picture

import (
	"context"
	"errors"
	"math/rand"
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v4"

	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/cache"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/plugins"
)

var (
	nameCartoon          = "picture_cartoon"
	urls        []string = []string{
		"https://api.btstu.cn/sjbz/?lx=dongman",
		"https://www.dmoe.cc/random.php",
	}
)

var _ plugins.PictureInterface = (*Cartoon)(nil)

func init() {
	p := &Cartoon{}
	plugins.RegisterPlugin(p.Name(), p)
}

type Cartoon struct{}

var client = http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

func (c *Cartoon) Get(ctx context.Context, random bool) (*contracts.Picture, error) {
	day := time.Now().Format("2006-01-02")
	seconds := 0
	if !random {
		seconds = 24 * 60 * 60
	}
	bg, _ := app.Cache().Remember(cache.NewKey("picture-%s-%d", day, seconds), seconds, func() ([]byte, error) {
		var (
			response *http.Response
			err      error
		)
		if err := backoff.Retry(func() error {
			weburl := urls[rand.Intn(len(urls))]
			mlog.Debugf("[Picture]: request %s", weburl)
			response, err = client.Get(weburl)
			if err != nil {
				return err
			}
			defer response.Body.Close()
			if response.StatusCode > 400 {
				mlog.Debug(errors.New(weburl + ": status code > 400"))
				return errors.New(weburl + ": status code > 400")
			}
			return nil
		}, backoff.NewExponentialBackOff()); err != nil {
			return nil, err
		}

		return []byte(response.Header.Get("Location")), nil
	})

	return &contracts.Picture{
		Url:       string(bg),
		Copyright: "",
	}, nil
}

func (c *Cartoon) Name() string {
	return nameCartoon
}

func (c *Cartoon) Initialize(args map[string]any) error {
	mlog.Info("[Plugin]: " + c.Name() + " plugin Initialize...")
	return nil
}

func (c *Cartoon) Destroy() error {
	mlog.Info("[Plugin]: " + c.Name() + " plugin Destroy...")
	return nil
}
