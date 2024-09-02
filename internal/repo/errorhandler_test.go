package repo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestToError_WithNilError(t *testing.T) {
	err := ToError(400, nil)
	assert.Nil(t, err)
}

func TestToError_WithStringError(t *testing.T) {
	err := ToError(400, "test error")
	assert.Equal(t, status.Error(codes.InvalidArgument, "test error"), err)
}

func TestToError_WithErrorType(t *testing.T) {
	e := status.Error(codes.Internal, "test error")
	err := ToError(400, e)
	assert.Equal(t, e, err)
}

func TestToError_WithInvalidCode(t *testing.T) {
	err := ToError(999, "test error")
	assert.Equal(t, status.Error(codes.Internal, "test error"), err)
}

func TestToError_WithValidCodes(t *testing.T) {
	testCases := []struct {
		code     uint32
		expected codes.Code
	}{
		{400, codes.InvalidArgument},
		{401, codes.Unauthenticated},
		{403, codes.PermissionDenied},
		{404, codes.NotFound},
		{429, codes.ResourceExhausted},
	}

	for _, tc := range testCases {
		err := ToError(tc.code, "test error")
		assert.Equal(t, status.Error(tc.expected, "test error"), err)
	}
}
