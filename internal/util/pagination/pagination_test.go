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
		// 真 nil 指针：整个 inPage != nil 守卫短路，page/pageSize 必须回落默认。
		page, pageSize := InitByDefault((*int)(nil), (*int)(nil))
		assert.Equal(t, 1, page)
		assert.Equal(t, 15, pageSize)
	})

	t.Run("ReturnsDefaultValuesWhenInputIsZeroValue", func(t *testing.T) {
		// 零值指针：inPage != nil 为真，但 *inPage > 0 为假，page/pageSize 必须回落默认。
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

	t.Run("NegativePageFallsBackToDefault", func(t *testing.T) {
		// 守卫 *inPage > 0：负数 page 必须回落到默认 1，绝不传给 DB 做负 offset。
		pageInput := -1
		pageSizeInput := 20
		page, pageSize := InitByDefault(&pageInput, &pageSizeInput)
		assert.Equal(t, 1, page)
		assert.Equal(t, pageSizeInput, pageSize)
	})

	t.Run("NegativePageSizeFallsBackToDefault", func(t *testing.T) {
		// 守卫 *inPageSize > 0：负数 pageSize 必须回落到默认 15。
		pageInput := 2
		pageSizeInput := -1
		page, pageSize := InitByDefault(&pageInput, &pageSizeInput)
		assert.Equal(t, pageInput, page)
		assert.Equal(t, 15, pageSize)
	})

	t.Run("ZeroPageSizeFallsBackToDefault", func(t *testing.T) {
		// 守卫 *inPageSize != zero：0 值 pageSize 必须回落到默认 15，
		// 否则 Limit(0) 会让部分 DB 返回空集而非全量。
		pageInput := 2
		pageSizeInput := 0
		page, pageSize := InitByDefault(&pageInput, &pageSizeInput)
		assert.Equal(t, pageInput, page)
		assert.Equal(t, 15, pageSize)
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
