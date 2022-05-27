package picture

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/plugins"
)

var (
	nameCartoon          = "picture_cartoon"
	urls        []string = []string{
		"https://api.btstu.cn/sjbz/?lx=dongman",
		"https://acg.toubiec.cn/random.php",
		"https://www.dmoe.cc/random.php",
		"https://api.ixiaowai.cn/api/api.php",
	}
)

var _ plugins.PictureInterface = (*Cartoon)(nil)

func init() {
	rand.Seed(time.Now().UnixNano())
	p := &Cartoon{}
	plugins.RegisterPlugin(p.Name(), p)
}

type Cartoon struct{}

var client = http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return errors.New("not redirect")
	},
}

func (c *Cartoon) Get(ctx context.Context, random bool) (*contracts.Picture, error) {
	day := time.Now().Format("2006-01-02")
	seconds := 0
	if !random {
		seconds = 24 * 60 * 60
	}
	bg, _ := app.Cache().Remember(fmt.Sprintf("picture-%s-%d", day, seconds), seconds, func() ([]byte, error) {
		weburl := urls[rand.Intn(len(urls))]
		response, _ := client.Get(weburl)
		defer response.Body.Close()
		mlog.Debugf("[Picture]: request %s", weburl)
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
