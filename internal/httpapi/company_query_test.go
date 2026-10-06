package httpapi

import (
	"net/url"
	"testing"
)

func TestParseCompanyQueryUsesDefaults(t *testing.T) {
	query, details := parseCompanyQuery(url.Values{})

	if len(details) != 0 {
		t.Fatalf("expected no error details, got %d", len(details))
	}

	if query.IncludeEstablishments {
		t.Fatal("expected establishments not to be included")
	}

	if query.Page != 1 {
		t.Fatalf("expected page 1, got %d", query.Page)
	}

	if query.PageSize != 20 {
		t.Fatalf("expected page size 20, got %d", query.PageSize)
	}
}

func TestParseCompanyQueryAcceptsEstablishmentsAndPagination(
	t *testing.T,
) {
	values := url.Values{
		"include":   {"establishments"},
		"page":      {"3"},
		"page_size": {"50"},
	}

	query, details := parseCompanyQuery(values)

	if len(details) != 0 {
		t.Fatalf("expected no error details, got %d", len(details))
	}

	if !query.IncludeEstablishments {
		t.Fatal("expected establishments to be included")
	}

	if query.Page != 3 {
		t.Fatalf("expected page 3, got %d", query.Page)
	}

	if query.PageSize != 50 {
		t.Fatalf("expected page size 50, got %d", query.PageSize)
	}
}

func TestParseCompanyQueryRejectsInvalidParameters(t *testing.T) {
	values := url.Values{
		"include":   {"partners"},
		"page":      {"0"},
		"page_size": {"101"},
	}

	_, details := parseCompanyQuery(values)

	if len(details) != 3 {
		t.Fatalf(
			"expected 3 error details, got %d",
			len(details),
		)
	}

	expectedFields := []string{
		"include",
		"page",
		"page_size",
	}

	for index, expectedField := range expectedFields {
		if details[index].Field != expectedField {
			t.Fatalf(
				"expected field %q at index %d, got %q",
				expectedField,
				index,
				details[index].Field,
			)
		}
	}
}

func TestParseCompanyQueryRejectsNonNumericPagination(t *testing.T) {
	values := url.Values{
		"page":      {"first"},
		"page_size": {"large"},
	}

	_, details := parseCompanyQuery(values)

	if len(details) != 2 {
		t.Fatalf(
			"expected 2 error details, got %d",
			len(details),
		)
	}

	if details[0].Field != "page" {
		t.Fatalf(
			"expected page error, got %q",
			details[0].Field,
		)
	}

	if details[1].Field != "page_size" {
		t.Fatalf(
			"expected page_size error, got %q",
			details[1].Field,
		)
	}
}
