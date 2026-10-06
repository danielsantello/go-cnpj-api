package company

import "testing"

func TestNewPaginationCalculatesTotalPages(t *testing.T) {
	pagination := NewPagination(1, 20, 57)

	if pagination.Page != 1 {
		t.Fatalf(
			"expected page 1, got %d",
			pagination.Page,
		)
	}

	if pagination.PageSize != 20 {
		t.Fatalf(
			"expected page size 20, got %d",
			pagination.PageSize,
		)
	}

	if pagination.TotalItems != 57 {
		t.Fatalf(
			"expected 57 total items, got %d",
			pagination.TotalItems,
		)
	}

	if pagination.TotalPages != 3 {
		t.Fatalf(
			"expected 3 total pages, got %d",
			pagination.TotalPages,
		)
	}
}

func TestNewPaginationReturnsZeroPagesForEmptyCollection(
	t *testing.T,
) {
	pagination := NewPagination(1, 20, 0)

	if pagination.TotalPages != 0 {
		t.Fatalf(
			"expected 0 total pages, got %d",
			pagination.TotalPages,
		)
	}
}
