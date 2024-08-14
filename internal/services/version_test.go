package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewVersionSvc(t *testing.T) {
	svc := NewVersionSvc()
	assert.NotNil(t, svc)
}

func Test_versionSvc_Version(t *testing.T) {
	svc := NewVersionSvc()
	res, err := svc.Version(nil, nil)
	assert.NotNil(t, res)
	assert.Nil(t, err)
}
