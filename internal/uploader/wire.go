package uploader

//go:generate mockgen -destination ./mock_uploader.go -package uploader github.com/duc-cnzj/mars/v5/internal/uploader Uploader,File,FileInfo
import "github.com/google/wire"

var WireUploader = wire.NewSet(NewUploader)
