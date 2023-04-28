package uploader

import (
	"fmt"
	"strconv"

	"github.com/duc-cnzj/mars/v4/internal/cache"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
)

type cacheUploader struct {
	contracts.Uploader
	cache contracts.CacheInterface
}

func NewCacheUploader(uploader contracts.Uploader, cache contracts.CacheInterface) contracts.Uploader {
	return &cacheUploader{Uploader: uploader, cache: cache}
}

func toByteNum(i int64) []byte {
	return []byte(fmt.Sprintf("%d", i))
}

func byteNum(remember []byte) int64 {
	atoi, _ := strconv.Atoi(string(remember))
	return int64(atoi)
}

func (ca *cacheUploader) UnWrap() contracts.Uploader {
	return ca.Uploader
}

func (ca *cacheUploader) DirSize() (int64, error) {
	remember, err := ca.cache.Remember(cache.NewKey("dir-size"), 60, func() ([]byte, error) {
		size, err := ca.Uploader.DirSize()
		return toByteNum(size), err
	})
	return byteNum(remember), err
}
