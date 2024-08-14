package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_NewGrpcRegistry_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	registry := NewGrpcRegistry(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	assert.NotNil(t, registry)
}
