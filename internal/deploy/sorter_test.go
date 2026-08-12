package deploy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPriority_Add(t *testing.T) {
	p := priority[string]{}
	p.Add(1, "a")
	p.Add(3, "b")
	p.Add(2, "c")
	assert.Equal(t, []string{"b", "c", "a"}, p.Sort())
}

func TestPriority_Sort(t *testing.T) {
	p := priority[string]{}
	assert.Equal(t, []string{}, p.Sort())
}

func TestPriority_Sort_DoesNotMutateInternalList(t *testing.T) {
	p := priority[string]{}
	p.Add(1, "a")
	p.Add(3, "b")
	p.Add(2, "c")

	first := p.Sort()
	second := p.Sort()

	// 两次 Sort 结果一致，且内部列表仍为 3 条，说明 Sort 返回副本、不改动内部列表。
	assert.Equal(t, []string{"b", "c", "a"}, first)
	assert.Equal(t, []string{"b", "c", "a"}, second)
	assert.Len(t, p.list, 3)
}
