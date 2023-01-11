package pipeline

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMyPipeline_Send(t *testing.T) {
	var result []string
	p := NewPipeline[string]()
	p.Send("xxx").Through(func(f func(msg string)) func(msg string) {
		return func(msg string) {
			result = append(result, "1")
			result = append(result, "2")
			f(msg)
			result = append(result, "3")
		}
	}, func(f func(msg string)) func(msg string) {
		return func(msg string) {
			result = append(result, "4")
			result = append(result, "5")
			f(msg)
			result = append(result, "6")
		}
	}).Then(func(msg string) {
		result = append(result, msg)
	})

	assert.Equal(t, []string{"1", "2", "4", "5", "xxx", "6", "3"}, result)
}
