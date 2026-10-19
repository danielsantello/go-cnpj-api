package httpapi

import (
	"net/url"
	"testing"
)

func TestParseCompanySearchQueryUsesDefaults(t *testing.T) {
	query, details := parseCompanySearchQuery(url.Values{
		"name": []string{"SANTELLO"},
	})

	if len(details) != 0 {
		t.Fatalf("expected no error details, got %#v", details)
	}

	if query.Name != "SANTELLO" {
		t.Fatalf(
			"expected name %q, got %q",
			"SANTELLO",
			query.Name,
		)
	}

	if query.Page != defaultCompanyPage {
		t.Fatalf(
			"expected page %d, got %d",
			defaultCompanyPage,
			query.Page,
		)
	}

	if query.PageSize != defaultCompanyPageSize {
		t.Fatalf(
			"expected page size %d, got %d",
			defaultCompanyPageSize,
			query.PageSize,
		)
	}
}

func TestParseCompanySearchQueryAcceptsPagination(t *testing.T) {
	query, details := parseCompanySearchQuery(url.Values{
		"name":      []string{"SANTELLO"},
		"page":      []string{"3"},
		"page_size": []string{"50"},
	})

	if len(details) != 0 {
		t.Fatalf("expected no error details, got %#v", details)
	}

	if query.Page != 3 {
		t.Fatalf("expected page 3, got %d", query.Page)
	}

	if query.PageSize != 50 {
		t.Fatalf("expected page size 50, got %d", query.PageSize)
	}
}

func TestParseCompanySearchQueryRejectsInvalidPagination(
	t *testing.T,
) {
	_, details := parseCompanySearchQuery(url.Values{
		"name":      []string{"SANTELLO"},
		"page":      []string{"0"},
		"page_size": []string{"101"},
	})

	if len(details) != 2 {
		t.Fatalf(
			"expected 2 error details, got %d",
			len(details),
		)
	}

	if details[0].Field != "page" {
		t.Fatalf(
			"expected first error field page, got %q",
			details[0].Field,
		)
	}

	if details[1].Field != "page_size" {
		t.Fatalf(
			"expected second error field page_size, got %q",
			details[1].Field,
		)
	}
}
