package uploader

//go:generate go tool mockgen -destination ./mock_uploader.go -package uploader github.com/duc-cnzj/mars/v6/internal/uploader Uploader,File,FileInfo
import "github.com/google/wire"

var WireUploader = wire.NewSet(NewUploader)
