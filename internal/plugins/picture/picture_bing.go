package picture

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/application"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/rand"
)

var (
	// url 是 Bing 每日壁纸接口模板，n 为请求图片数量。
	url = "https://cn.bing.com/HPImageArchive.aspx?format=js&idx=0&n=%d&mkt=zh-CN"
	// bingname 插件注册名。
	bingname = "picture_bing"
	// bingClient 带超时，避免上游卡死时 http.Get 无限阻塞并长期持有写锁。
	bingClient = http.Client{Timeout: 10 * time.Second}
	// randIntn 随机下标；测试可替换为固定值以确定性触发 key 越界防御分支。
	randIntn = rand.Intn
)

var _ application.Picture = (*bing)(nil)

func init() {
	p := &bing{}
	application.RegisterPlugin(p.Name(), p)
}

// Item Bing 接口返回的单张图片字段。
type Item struct {
	Startdate     string `json:"startdate"`
	Fullstartdate string `json:"fullstartdate"`
	Enddate       string `json:"enddate"`
	URL           string `json:"url"`
	Urlbase       string `json:"urlbase"`
	Copyright     string `json:"copyright"`
	Copyrightlink string `json:"copyrightlink"`
	Title         string `json:"title"`
	Quiz          string `json:"quiz"`
	Wp            bool   `json:"wp"`
	Hsh           string `json:"hsh"`
	Drk           int    `json:"drk"`
	Top           int    `json:"top"`
	Bot           int    `json:"bot"`
	Hs            []any  `json:"hs"`
}

// Res Bing 接口返回的整体响应。
type Res struct {
	Images []Item `json:"images"`
}

// bing 每日从 Bing 壁纸接口拉取图片，按天缓存。
type bing struct {
	sync.RWMutex
	cacheItems []Item
	cacheDay   string
	logger     mlog.Logger
}

// Name 返回插件名 picture_bing。
func (b *bing) Name() string {
	return bingname
}

// Initialize 保存 app.Logger 并输出初始化日志。
func (b *bing) Initialize(app application.PluginApp, args map[string]any) error {
	b.logger = app.Logger()
	b.logger.Info("[Plugin]: " + b.Name() + " plugin Initialize...")
	return nil
}

// Destroy 输出销毁日志。
func (b *bing) Destroy() error {
	b.logger.Info("[Plugin]: " + b.Name() + " plugin Destroy...")
	return nil
}

// Get 返回一张 Bing 壁纸；random 为 true 时随机选一张，否则取当天第一张。当天结果按天缓存。
func (b *bing) Get(ctx context.Context, random bool) (*application.PictureItem, error) {
	key, n := 0, 8
	if random {
		key = randIntn(n)
	}
	var res []Item
	day := time.Now().Format("2006-01-02")

	func() {
		b.RLock()
		defer b.RUnlock()
		if len(b.cacheItems) > 0 && b.cacheDay == day {
			b.logger.Debug("use cache")
			res = b.cacheItems
		}
	}()

	if res == nil {
		b.Lock()
		defer b.Unlock()
		get, err := bingClient.Get(fmt.Sprintf(url, n))
		if err != nil {
			return nil, err
		}
		defer get.Body.Close()
		var response Res
		all, err := io.ReadAll(get.Body)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(all, &response)
		if err != nil {
			return nil, err
		}
		res = response.Images
		b.cacheItems = response.Images
		b.cacheDay = day
	}

	// 防御边界：接口可能返回空图片列表或少于 n 张，直接取下标会越界 panic。
	if len(res) == 0 {
		return nil, errors.New("bing: empty images response")
	}
	if key >= len(res) {
		key = 0
	}
	item := res[key]
	// 防御边界：Copyright 文案里没有 "(©" 时 strings.Index 返回 -1，切片会越界 panic。
	copyright := item.Copyright
	if idx := strings.Index(copyright, "(©"); idx > 0 {
		copyright = copyright[:idx]
	}
	return &application.PictureItem{
		Url:       "https://cn.bing.com/" + strings.TrimLeft(item.URL, "/"),
		Copyright: copyright,
	}, nil
}
