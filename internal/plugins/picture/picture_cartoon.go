package picture

import (
	"math/rand"
	"time"

	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/plugins"
)

var (
	nameCartoon          = "picture_cartoon"
	urls        []string = []string{
		//"https://api.ixiaowai.cn/api/api.php", // sina 防盗链 403 了
		"https://cloud.qqshabi.cn/api/images/api.php",
	}
)

func init() {
	rand.Seed(time.Now().UnixNano())
	p := &Cartoon{}
	plugins.RegisterPlugin(p.Name(), p)
}

type Cartoon struct{}

func (c *Cartoon) Get(random bool) (*plugins.Picture, error) {
	return &plugins.Picture{
		Url:       urls[rand.Intn(len(urls))],
		Copyright: "",
	}, nil
}

func (c *Cartoon) Name() string {
	return nameCartoon
}

func (c *Cartoon) Initialize(args map[string]interface{}) error {
	mlog.Info("[Plugin]: " + c.Name() + " plugin Initialize...")
	return nil
}

func (c *Cartoon) Destroy() error {
	mlog.Info("[Plugin]: " + c.Name() + " plugin Destroy...")
	return nil
}
