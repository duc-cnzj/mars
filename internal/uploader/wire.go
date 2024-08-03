package uploader

import "github.com/google/wire"

var WireUploader = wire.NewSet(NewUploader)
