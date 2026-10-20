package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/danielsantello/go-cnpj-api/internal/company"
)

type establishmentSearcherStub struct {
	receivedOptions company.SearchOptions
	calls           int
	result          company.EstablishmentSearchResult
	err             error
}

func (searcher *establishmentSearcherStub) SearchByTradeName(
	_ context.Context,
	options company.SearchOptions,
) (company.EstablishmentSearchResult, error) {
	searcher.calls++
	searcher.receivedOptions = options

	return searcher.result, searcher.err
}

func TestEstablishmentSearchHandlerReturnsEstablishments(t *testing.T) {
	expectedResult := company.EstablishmentSearchResult{
		Items: []company.EstablishmentSummary{
			{
				CNPJ:          "12345678000190",
				TradeName:     "SANTELLO SISTEMAS",
				CorporateName: "SANTELLO SISTEMAS LTDA",
			},
		},
		Pagination: company.NewPagination(2, 10, 15),
	}

	searcher := &establishmentSearcherStub{
		result: expectedResult,
	}
	handler := newEstablishmentSearchHandler(searcher)

	request := httptest.NewRequest(
		http.MethodGet,
		"/v1/establishments?name=SANTELLO&page=2&page_size=10",
		nil,
	)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	expectedOptions := company.SearchOptions{
		Name:     "SANTELLO",
		Page:     2,
		PageSize: 10,
	}

	if searcher.receivedOptions != expectedOptions {
		t.Fatalf(
			"expected options %#v, got %#v",
			expectedOptions,
			searcher.receivedOptions,
		)
	}

	if searcher.calls != 1 {
		t.Fatalf("expected 1 searcher call, got %d", searcher.calls)
	}

	var payload establishmentSearchResponse

	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if !reflect.DeepEqual(payload.Data.Establishments, expectedResult) {
		t.Fatalf(
			"expected result %#v, got %#v",
			expectedResult,
			payload.Data.Establishments,
		)
	}
}

func TestEstablishmentSearchHandlerReturnsStandardErrors(t *testing.T) {
	tests := []struct {
		name           string
		searchError    error
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "trade name too short",
			searchError:    company.ErrTradeNameTooShort,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_QUERY_PARAMETERS",
		},
		{
			name:           "trade name too long",
			searchError:    company.ErrTradeNameTooLong,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_QUERY_PARAMETERS",
		},
		{
			name:           "search too broad",
			searchError:    company.ErrSearchTooBroad,
			expectedStatus: http.StatusUnprocessableEntity,
			expectedCode:   "SEARCH_TOO_BROAD",
		},
		{
			name:           "unexpected error",
			searchError:    errors.New("database failure"),
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "INTERNAL_ERROR",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			searcher := &establishmentSearcherStub{
				err: test.searchError,
			}
			handler := newEstablishmentSearchHandler(searcher)

			request := httptest.NewRequest(
				http.MethodGet,
				"/v1/establishments?name=SANTELLO",
				nil,
			)
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if response.Code != test.expectedStatus {
				t.Fatalf(
					"expected status %d, got %d",
					test.expectedStatus,
					response.Code,
				)
			}

			var payload struct {
				Error struct {
					Code string `json:"code"`
				} `json:"error"`
			}

			if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
				t.Fatalf("decode response: %v", err)
			}

			if payload.Error.Code != test.expectedCode {
				t.Fatalf(
					"expected error code %q, got %q",
					test.expectedCode,
					payload.Error.Code,
				)
			}
		})
	}
}

func TestEstablishmentSearchHandlerRejectsInvalidPagination(t *testing.T) {
	tests := []string{
		"page=0",
		"page=abc",
		"page_size=0",
		"page_size=101",
	}

	for _, query := range tests {
		t.Run(query, func(t *testing.T) {
			searcher := &establishmentSearcherStub{}
			handler := newEstablishmentSearchHandler(searcher)

			request := httptest.NewRequest(
				http.MethodGet,
				"/v1/establishments?name=SANTELLO&"+query,
				nil,
			)
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if response.Code != http.StatusBadRequest {
				t.Fatalf("expected status 400, got %d", response.Code)
			}

			if searcher.calls != 0 {
				t.Fatalf(
					"expected no searcher calls, got %d",
					searcher.calls,
				)
			}
		})
	}
}
