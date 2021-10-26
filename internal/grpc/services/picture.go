package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/pkg/picture"
)

type Picture struct {
	sync.Mutex
	cacheItems []Item
	cacheDay   string
	picture.UnimplementedPictureServer
}

type Item struct {
	Startdate     string        `json:"startdate"`
	Fullstartdate string        `json:"fullstartdate"`
	Enddate       string        `json:"enddate"`
	URL           string        `json:"url"`
	Urlbase       string        `json:"urlbase"`
	Copyright     string        `json:"copyright"`
	Copyrightlink string        `json:"copyrightlink"`
	Title         string        `json:"title"`
	Quiz          string        `json:"quiz"`
	Wp            bool          `json:"wp"`
	Hsh           string        `json:"hsh"`
	Drk           int           `json:"drk"`
	Top           int           `json:"top"`
	Bot           int           `json:"bot"`
	Hs            []interface{} `json:"hs"`
}

type Res struct {
	Images []Item `json:"images"`
}

var url = "https://cn.bing.com/HPImageArchive.aspx?format=js&idx=0&n=%d&mkt=zh-CN"

func init() {
	rand.Seed(time.Now().UnixNano())
}

func (p *Picture) Background(ctx context.Context, req *picture.BackgroundRequest) (*picture.BackgroundResponse, error) {
	key, n := 0, 8
	if req.Random {
		key = rand.Intn(n - 1)
	}
	var res []Item

	p.Lock()
	defer p.Unlock()

	day := time.Now().Format("2006-01-02")
	if len(p.cacheItems) > 0 && p.cacheDay == day {
		mlog.Debug("use cache")
		res = p.cacheItems
	} else {
		get, err := http.Get(fmt.Sprintf(url, n))
		if err != nil {
			return nil, err
		}
		defer get.Body.Close()
		var response Res
		all, _ := io.ReadAll(get.Body)
		err = json.Unmarshal(all, &response)
		if err != nil {
			return nil, err
		}
		res = response.Images
		p.cacheItems = response.Images
		p.cacheDay = day
	}

	return &picture.BackgroundResponse{
		Url:       path.Clean("https://cn.bing.com/" + res[key].URL),
		Copyright: res[key].Copyright[:strings.Index(res[key].Copyright, "(©")],
	}, nil
}

func (p *Picture) AuthFuncOverride(ctx context.Context, fullMethodName string) (context.Context, error) {
	mlog.Debug("client is calling method:", fullMethodName)
	return ctx, nil
}
