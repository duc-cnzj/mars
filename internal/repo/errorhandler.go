package repo

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ToError(code uint32, err any) error {
	if err == nil {
		return nil
	}

	var errMessage string
	var co codes.Code
	switch e := err.(type) {
	case error:
		if s, ok := status.FromError(e); ok {
			return s.Err()
		}
		errMessage = e.Error()
	case string:
		errMessage = e
	}

	switch code {
	case 400:
		co = codes.InvalidArgument
	case 401:
		co = codes.Unauthenticated
	case 403:
		co = codes.PermissionDenied
	case 404:
		co = codes.NotFound
	case 429:
		co = codes.ResourceExhausted
	default:
		co = codes.Internal
	}

	return status.Error(co, errMessage)
}
