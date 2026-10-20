package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danielsantello/go-cnpj-api/internal/company"
)

type partnerSearcherStub struct {
	options company.PartnerSearchOptions
	calls   int
	result  company.PartnerSearchResult
	err     error
}

func (stub *partnerSearcherStub) SearchByPartnerName(
	_ context.Context,
	options company.PartnerSearchOptions,
) (company.PartnerSearchResult, error) {
	stub.calls++
	stub.options = options
	return stub.result, stub.err
}

func TestPartnerSearchHandlerReturnsResult(t *testing.T) {
	document := "***123456**"
	stub := &partnerSearcherStub{
		result: company.PartnerSearchResult{
			Items: []company.PartnerSummary{
				{
					Name:          "DANIEL SANTELLO",
					Document:      &document,
					CNPJ:          "12345678000190",
					CorporateName: "EMPRESA EXEMPLO",
				},
			},
			Pagination: company.NewPagination(2, 5, 6),
		},
	}

	request := httptest.NewRequest(
		http.MethodGet,
		"/v1/partners?name=DANIEL%20SANTELLO&document=***123456**&page=2&page_size=5",
		nil,
	)
	response := httptest.NewRecorder()

	requestIDMiddleware(newPartnerSearchHandler(stub)).ServeHTTP(
		response, request,
	)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", response.Code, response.Body)
	}
	if response.Header().Get("Content-Type") != "application/json" {
		t.Fatal("expected JSON response")
	}
	if response.Header().Get("X-Request-ID") == "" {
		t.Fatal("expected request ID")
	}

	if stub.calls != 1 ||
		stub.options.Name != "DANIEL SANTELLO" ||
		stub.options.Document == nil ||
		*stub.options.Document != document ||
		stub.options.Page != 2 ||
		stub.options.PageSize != 5 {
		t.Fatalf("unexpected options: %#v", stub.options)
	}

	var payload partnerSearchResponse
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	result := payload.Data.Partners
	if len(result.Items) != 1 {
		t.Fatalf("expected one item, got %d", len(result.Items))
	}
	item := result.Items[0]
	if item.Document == nil || *item.Document != document ||
		item.Name != "DANIEL SANTELLO" ||
		item.CNPJ != "12345678000190" ||
		item.CorporateName != "EMPRESA EXEMPLO" ||
		result.Pagination.TotalItems != 6 {
		t.Fatalf("unexpected response: %#v", result)
	}
}

func TestPartnerSearchHandlerUsesDefaults(t *testing.T) {
	stub := &partnerSearcherStub{
		result: company.PartnerSearchResult{
			Items:      []company.PartnerSummary{},
			Pagination: company.NewPagination(1, 20, 0),
		},
	}
	response := httptest.NewRecorder()

	newPartnerSearchHandler(stub).ServeHTTP(
		response,
		httptest.NewRequest(http.MethodGet, "/v1/partners?name=SANTELLO", nil),
	)

	if response.Code != http.StatusOK ||
		stub.options.Document != nil ||
		stub.options.Page != 1 ||
		stub.options.PageSize != 20 {
		t.Fatalf("unexpected status or options: %d %#v", response.Code, stub.options)
	}

	var payload partnerSearchResponse
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Data.Partners.Items == nil {
		t.Fatal("expected empty array")
	}
}

func TestPartnerSearchHandlerRejectsInvalidPagination(t *testing.T) {
	for _, query := range []string{
		"page=0", "page=abc", "page_size=0", "page_size=101",
	} {
		t.Run(query, func(t *testing.T) {
			stub := &partnerSearcherStub{}
			response := httptest.NewRecorder()

			newPartnerSearchHandler(stub).ServeHTTP(
				response,
				httptest.NewRequest(
					http.MethodGet, "/v1/partners?name=SANTELLO&"+query, nil,
				),
			)

			if response.Code != http.StatusBadRequest || stub.calls != 0 {
				t.Fatalf("unexpected status or calls: %d %d", response.Code, stub.calls)
			}
		})
	}
}

func TestPartnerSearchHandlerMapsErrors(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
		code   string
		field  string
	}{
		{"short name", company.ErrPartnerNameTooShort, 400, "INVALID_QUERY_PARAMETERS", "name"},
		{"long name", company.ErrPartnerNameTooLong, 400, "INVALID_QUERY_PARAMETERS", "name"},
		{"document", company.ErrPartnerDocumentInvalid, 400, "INVALID_QUERY_PARAMETERS", "document"},
		{"too broad", company.ErrSearchTooBroad, 422, "SEARCH_TOO_BROAD", ""},
		{"internal", errors.New("database unavailable"), 500, "INTERNAL_ERROR", ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stub := &partnerSearcherStub{err: test.err}
			response := httptest.NewRecorder()

			requestIDMiddleware(newPartnerSearchHandler(stub)).ServeHTTP(
				response,
				httptest.NewRequest(http.MethodGet, "/v1/partners?name=SANTELLO", nil),
			)

			if response.Code != test.status {
				t.Fatalf("expected %d, got %d", test.status, response.Code)
			}

			var payload errorResponse
			if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
				t.Fatal(err)
			}
			if payload.Error.Code != test.code ||
				payload.Error.RequestID != response.Header().Get("X-Request-ID") {
				t.Fatalf("unexpected error: %#v", payload.Error)
			}
			if test.field != "" {
				if len(payload.Error.Details) != 1 ||
					payload.Error.Details[0].Field != test.field {
					t.Fatalf("unexpected details: %#v", payload.Error.Details)
				}
			}
		})
	}
}
