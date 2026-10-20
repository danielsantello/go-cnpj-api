package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danielsantello/go-cnpj-api/internal/company"
	"github.com/danielsantello/go-cnpj-api/internal/search"
)

type broadSearcherStub struct {
	options search.Options
	calls   int
	result  search.Result
	err     error
}

func (stub *broadSearcherStub) Search(
	_ context.Context,
	options search.Options,
) (search.Result, error) {
	stub.calls++
	stub.options = options
	return stub.result, stub.err
}

func TestBroadSearchHandlerReturnsResult(t *testing.T) {
	stub := &broadSearcherStub{
		result: search.Result{
			Items: []search.Item{{
				BasicCNPJ:     "00000208",
				CNPJ:          "00000208000100",
				CorporateName: "BRB BANCO DE BRASILIA SA",
			}},
			Pagination: company.NewPagination(2, 1, 2),
		},
	}
	response := httptest.NewRecorder()

	requestIDMiddleware(newBroadSearchHandler(stub)).ServeHTTP(
		response,
		httptest.NewRequest(
			http.MethodGet,
			"/v1/search?name=banco&page=2&page_size=1",
			nil,
		),
	)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", response.Code, response.Body)
	}
	if response.Header().Get("Content-Type") != "application/json" ||
		response.Header().Get("X-Request-ID") == "" {
		t.Fatal("missing JSON content type or request ID")
	}
	if stub.calls != 1 ||
		stub.options != (search.Options{Name: "banco", Page: 2, PageSize: 1}) {
		t.Fatalf("unexpected options: %#v", stub.options)
	}

	var payload broadSearchResponse
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	result := payload.Data.Companies
	if len(result.Items) != 1 ||
		result.Items[0].CNPJ != "00000208000100" ||
		result.Pagination.Page != 2 ||
		result.Pagination.TotalItems != 2 {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestBroadSearchHandlerRejectsPaginationBeforeCallingService(t *testing.T) {
	for _, query := range []string{
		"page=0", "page=abc", "page_size=0", "page_size=101",
	} {
		t.Run(query, func(t *testing.T) {
			stub := &broadSearcherStub{}
			response := httptest.NewRecorder()

			newBroadSearchHandler(stub).ServeHTTP(
				response,
				httptest.NewRequest(
					http.MethodGet, "/v1/search?name=banco&"+query, nil,
				),
			)

			if response.Code != http.StatusBadRequest || stub.calls != 0 {
				t.Fatalf("unexpected status or calls: %d %d", response.Code, stub.calls)
			}
		})
	}
}

func TestBroadSearchHandlerMapsErrors(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
		code   string
	}{
		{"name", search.ErrInvalidName, 400, "INVALID_QUERY_PARAMETERS"},
		{"pagination", search.ErrInvalidPagination, 400, "INVALID_QUERY_PARAMETERS"},
		{"disabled", search.ErrDisabled, 503, "SEARCH_DISABLED"},
		{"internal", errors.New("connection failed"), 500, "INTERNAL_ERROR"},
		{"too broad", search.ErrTooBroad, 422, "SEARCH_TOO_BROAD"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stub := &broadSearcherStub{err: test.err}
			response := httptest.NewRecorder()

			requestIDMiddleware(newBroadSearchHandler(stub)).ServeHTTP(
				response,
				httptest.NewRequest(http.MethodGet, "/v1/search?name=banco", nil),
			)

			if response.Code != test.status {
				t.Fatalf("expected %d, got %d", test.status, response.Code)
			}

			var payload errorResponse
			if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
				t.Fatal(err)
			}
			if payload.Error.Code != test.code ||
				payload.Error.RequestID == "" ||
				payload.Error.RequestID != response.Header().Get("X-Request-ID") {
				t.Fatalf("unexpected error response: %#v", payload)
			}
		})
	}
}

func TestBroadSearchHandlerWithDisabledService(t *testing.T) {
	response := httptest.NewRecorder()
	newBroadSearchHandler(search.NewService(nil)).ServeHTTP(
		response,
		httptest.NewRequest(http.MethodGet, "/v1/search?name=banco", nil),
	)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", response.Code)
	}
}
