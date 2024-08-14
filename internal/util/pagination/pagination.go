package pagination

func GetPageOffset[T ~int | ~int32](page, pageSize T) int {
	return int((page - 1) * pageSize)
}

func InitByDefault[T ~int32 | ~int64 | ~int](inPage *T, inPageSize *T) (page, pageSize T) {
	page, pageSize = 1, 15

	var zero T
	if inPage != nil && *inPage != zero && *inPage > 0 {
		page = *inPage
	}

	if inPageSize != nil && *inPageSize != zero && *inPageSize > 0 {
		pageSize = *inPageSize
	}

	return
}

type Pagination struct {
	Page     int32
	PageSize int32
	Count    int32
}

func NewPagination[T ~int | ~int64 | ~int32, V ~int | ~int64 | ~int32](page T, pageSize T, count V) *Pagination {
	return &Pagination{Page: int32(page), PageSize: int32(pageSize), Count: int32(count)}
}
