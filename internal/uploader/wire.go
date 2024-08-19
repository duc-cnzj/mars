package uploader

//go:generate mockgen -destination ./mock_uploader.go -package uploader github.com/duc-cnzj/mars/v4/internal/uploader Uploader,File
import "github.com/google/wire"

var WireUploader = wire.NewSet(NewUploader)
