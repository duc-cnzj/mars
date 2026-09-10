package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewVersionSvc(t *testing.T) {
	t.Parallel()
	svc := NewVersionSvc()
	assert.NotNil(t, svc)
}

func Test_versionSvc_Version(t *testing.T) {
	t.Parallel()
	svc := NewVersionSvc()
	res, err := svc.Version(context.TODO(), nil)
	assert.NotNil(t, res)
	assert.Nil(t, err)
}
