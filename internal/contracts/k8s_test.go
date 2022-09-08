package contracts

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func TestNewObj(t *testing.T) {
	p1 := &v1.Pod{}
	p2 := &v1.Pod{}
	newObj := NewObj(p1, p2, Add)
	assert.Equal(t, Add, newObj.Type())
	assert.Same(t, p1, newObj.Old())
	assert.Same(t, p2, newObj.Current())
}
