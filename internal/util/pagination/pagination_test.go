package pagination

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPageOffset(t *testing.T) {
	t.Run("ReturnsCorrectOffsetForPositivePageAndSize", func(t *testing.T) {
		offset := GetPageOffset(2, 10)
		assert.Equal(t, 10, offset)
	})

	t.Run("ReturnsZeroForFirstPageRegardlessOfSize", func(t *testing.T) {
		offset := GetPageOffset(1, 10)
		assert.Equal(t, 0, offset)
	})
}

func TestInitByDefault(t *testing.T) {
	t.Run("ReturnsDefaultValuesWhenInputIsNil", func(t *testing.T) {
		page, pageSize := InitByDefault(new(int), new(int))
		assert.Equal(t, 1, page)
		assert.Equal(t, 15, pageSize)
	})

	t.Run("ReturnsInputValuesWhenInputIsNotNil", func(t *testing.T) {
		pageInput := 2
		pageSizeInput := 20
		page, pageSize := InitByDefault(&pageInput, &pageSizeInput)
		assert.Equal(t, pageInput, page)
		assert.Equal(t, pageSizeInput, pageSize)
	})
}

func TestNewPagination(t *testing.T) {
	t.Run("CreatesNewPaginationWithGivenValues", func(t *testing.T) {
		page := 2
		pageSize := 20
		count := 100
		p := NewPagination(page, pageSize, count)
		assert.Equal(t, int32(page), p.Page)
		assert.Equal(t, int32(pageSize), p.PageSize)
		assert.Equal(t, int32(count), p.Count)
	})
}
