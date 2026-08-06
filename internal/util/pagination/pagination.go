package pagination

// GetPageOffset 计算 page/pageSize 对应的偏移量 (page-1)*pageSize。
func GetPageOffset[T ~int | ~int32](page, pageSize T) int {
	return int((page - 1) * pageSize)
}

// InitByDefault 返回分页参数，nil、零值或负值一律回退为默认 page=1、pageSize=15。
func InitByDefault[T ~int32 | ~int64 | ~int](inPage *T, inPageSize *T) (page, pageSize T) {
	page, pageSize = 1, 15

	// 数值类型下 *inPage > 0 已蕴含 *inPage != 0，无需重复判零。
	if inPage != nil && *inPage > 0 {
		page = *inPage
	}

	if inPageSize != nil && *inPageSize > 0 {
		pageSize = *inPageSize
	}

	return
}

// Pagination 描述一次分页查询的结果元信息。
type Pagination struct {
	Page     int32
	PageSize int32
	Count    int32
}

// NewPagination 以 page/pageSize/count 构造 Pagination，统一转为 int32。
func NewPagination[T ~int | ~int64 | ~int32, V ~int | ~int64 | ~int32](page T, pageSize T, count V) *Pagination {
	return &Pagination{Page: int32(page), PageSize: int32(pageSize), Count: int32(count)}
}
