package xsort

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPriority(t *testing.T) {
	assert.NotNil(t, NewPriority[string]())
}

func TestPrioritySort_Add(t *testing.T) {
	p := NewPriority[string]()
	p.Add(1, "a")
	p.Add(3, "b")
	p.Add(2, "c")
	assert.Equal(t, []string{"b", "c", "a"}, p.Sort())
}

func TestPrioritySort_Sort(t *testing.T) {
	p := NewPriority[string]()
	assert.Equal(t, []string{}, p.Sort())
}
