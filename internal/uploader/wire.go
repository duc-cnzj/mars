package uploader

//go:generate go tool mockgen -destination ./mock_uploader.go -package uploader github.com/duc-cnzj/mars/v6/internal/uploader Uploader,File,FileInfo
import "github.com/google/wire"

// WireUploader 提供 uploader.Uploader 的装配集。
var WireUploader = wire.NewSet(NewUploader)
