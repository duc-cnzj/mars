package socket

import (
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewTimeOrderedSetString(t *testing.T) {
	assert.IsType(t, (*timeOrderedSetString)(nil), newTimeOrderedSetString(time.Now))
}

func Test_imageUsedPipelineVars(t *testing.T) {
	var tests = []struct {
		pipelineVars pipelineVars
		wants        bool
		manifest     string
	}{
		{
			pipelineVars: pipelineVars{
				Pipeline: "p",
				Commit:   "c",
				Branch:   "b",
			},
			wants:    false,
			manifest: "image: xxx:v1",
		},
		{
			pipelineVars: pipelineVars{
				Pipeline: "p",
				Commit:   "c",
				Branch:   "b",
			},
			wants: true,
			manifest: `
image: xxx:v1
image: xxx:v2
image: xxx:p
image: xxx:c
image: xxx:b
`,
		},
	}
	for _, test := range tests {
		assert.Equal(t, test.wants, imageUsedPipelineVars(test.pipelineVars, test.manifest))
	}
}

func Test_matchDockerImage(t *testing.T) {
	var tests = []struct {
		pipelineVars pipelineVars
		wants        string
		manifest     string
	}{
		{
			pipelineVars: pipelineVars{
				Pipeline: "p",
				Commit:   "c",
				Branch:   "b",
			},
			wants:    "xxx:v1",
			manifest: "image: xxx:v1",
		},
		{
			pipelineVars: pipelineVars{
				Pipeline: "p",
				Commit:   "c",
				Branch:   "b",
			},
			wants: "xxx:v1 xxx:v2",
			manifest: `
image: xxx:v1
image: xxx:v2
image: xxx:v1
`,
		},
		{
			pipelineVars: pipelineVars{
				Pipeline: "p",
				Commit:   "c",
				Branch:   "b",
			},
			wants: "xxx:p xxx:c xxx:b",
			manifest: `
image: xxx:v1
image: xxx:v2
image: xxx:p
image: xxx:c
image: xxx:b
`,
		},
	}
	for _, test := range tests {
		assert.Equal(t, test.wants, matchDockerImage(test.pipelineVars, test.manifest))
	}
}

func Test_orderedItemList(t *testing.T) {
	var o = orderedItemList{
		{
			t:    time.Now().Add(1 * time.Second),
			data: "ccc",
		},
		{
			t:    time.Now(),
			data: "bbb",
		},
		{
			t:    time.Now().Add(-1 * time.Second),
			data: "aaa",
		},
	}
	sort.Sort(o)
	assert.Equal(t, "aaa", o[0].data)
	assert.Equal(t, "bbb", o[1].data)
	assert.Equal(t, "ccc", o[2].data)
}

func Test_timeOrderedSetString_add(t *testing.T) {
	var called int
	fn := func() time.Time {
		called++
		parse, _ := time.Parse("2006-01-02 00:00:00", "2022-05-09 00:00:00")

		return parse.Add(-10 * time.Duration(called) * time.Second)
	}
	o := newTimeOrderedSetString(fn)
	o.add("a")
	o.add("b")
	o.add("b")
	s := o.sortedItems()
	assert.Equal(t, "b", s[0])
	assert.Equal(t, "a", s[1])
}

func Test_timeOrderedSetString_has(t *testing.T) {
	o := newTimeOrderedSetString(time.Now)
	o.add("a")
	assert.True(t, o.has("a"))
	assert.False(t, o.has("c"))
}
