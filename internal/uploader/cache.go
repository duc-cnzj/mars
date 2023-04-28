package uploader

import (
	"fmt"
	"strconv"
	"time"

	"github.com/duc-cnzj/mars/v4/internal/cache"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
)

type cacheUploader struct {
	contracts.Uploader
	cacheFn func() contracts.CacheInterface
}

func NewCacheUploader(uploader contracts.Uploader, cache func() contracts.CacheInterface) contracts.Uploader {
	return &cacheUploader{Uploader: uploader, cacheFn: cache}
}

func int64ToByte(i int64) []byte {
	return []byte(fmt.Sprintf("%d", i))
}

func byteToInt64(remember []byte) int64 {
	atoi, _ := strconv.Atoi(string(remember))
	return int64(atoi)
}

func (ca *cacheUploader) UnWrap() contracts.Uploader {
	return ca.Uploader
}

var DirSizeCacheSeconds = int((15 * time.Minute).Seconds())

func (ca *cacheUploader) DirSize() (int64, error) {
	remember, err := ca.cacheFn().Remember(cache.NewKey("dir-size"), DirSizeCacheSeconds, func() ([]byte, error) {
		mlog.Debug("dir-size cache missing")
		size, err := ca.Uploader.DirSize()
		return int64ToByte(size), err
	})
	return byteToInt64(remember), err
}
