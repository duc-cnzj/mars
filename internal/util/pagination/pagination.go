package pagination

func GetPageOffset[T ~int | ~int64](page, pageSize T) int {
	return int((page - 1) * pageSize)
}

func InitByDefault(page *int64, pageSize *int64) {
	if *page == 0 {
		*page = 1
	}
	if *pageSize == 0 {
		*pageSize = 15
	}
}

type Pagination struct {
	Page     int64
	PageSize int64
	Count    int64
}
