package company

type FindOptions struct {
	IncludeEstablishments bool
	Page                  int
	PageSize              int
}

type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalItems int64 `json:"total_items"`
	TotalPages int64 `json:"total_pages"`
}

type EstablishmentList struct {
	Items      []Establishment `json:"items"`
	Pagination Pagination      `json:"pagination"`
}

func NewPagination(
	page int,
	pageSize int,
	totalItems int64,
) Pagination {
	var totalPages int64

	if totalItems > 0 {
		totalPages =
			(totalItems + int64(pageSize) - 1) /
				int64(pageSize)
	}

	return Pagination{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}
}
