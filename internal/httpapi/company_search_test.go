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

type companySearcherStub struct {
	receivedOptions company.SearchOptions
	calls           int
	result          company.SearchResult
	err             error
}

func (searcher *companySearcherStub) SearchByCorporateName(
	_ context.Context,
	options company.SearchOptions,
) (company.SearchResult, error) {
	searcher.calls++
	searcher.receivedOptions = options

	return searcher.result, searcher.err
}

func TestCompanySearchHandlerReturnsCompanies(t *testing.T) {
	expectedResult := company.SearchResult{
		Items: []company.Summary{
			{
				CNPJ:          "12345678000190",
				CorporateName: "SANTELLO SISTEMAS LTDA",
			},
		},
		Pagination: company.NewPagination(2, 10, 15),
	}

	searcher := &companySearcherStub{
		result: expectedResult,
	}

	handler := newCompanySearchHandler(searcher)

	request := httptest.NewRequest(
		http.MethodGet,
		"/v1/companies?name=SANTELLO&page=2&page_size=10",
		nil,
	)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	result := response.Result()
	defer result.Body.Close()

	if result.StatusCode != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			result.StatusCode,
		)
	}

	expectedOptions := company.SearchOptions{
		Name:     "SANTELLO",
		Page:     2,
		PageSize: 10,
	}

	if searcher.receivedOptions != expectedOptions {
		t.Fatalf(
			"expected search options %#v, got %#v",
			expectedOptions,
			searcher.receivedOptions,
		)
	}

	var payload companySearchResponse

	if err := json.NewDecoder(result.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if len(payload.Data.Companies.Items) != 1 {
		t.Fatalf(
			"expected 1 company, got %d",
			len(payload.Data.Companies.Items),
		)
	}

	if payload.Data.Companies.Items[0].CNPJ !=
		expectedResult.Items[0].CNPJ {
		t.Fatalf(
			"expected CNPJ %q, got %q",
			expectedResult.Items[0].CNPJ,
			payload.Data.Companies.Items[0].CNPJ,
		)
	}

	if payload.Data.Companies.Pagination !=
		expectedResult.Pagination {
		t.Fatalf(
			"expected pagination %#v, got %#v",
			expectedResult.Pagination,
			payload.Data.Companies.Pagination,
		)
	}
}

func TestCompanySearchHandlerReturnsStandardErrors(t *testing.T) {
	tests := []struct {
		name           string
		searchError    error
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "corporate name too short",
			searchError:    company.ErrCorporateNameTooShort,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_QUERY_PARAMETERS",
		},
		{
			name:           "corporate name too long",
			searchError:    company.ErrCorporateNameTooLong,
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
			searcher := &companySearcherStub{
				err: test.searchError,
			}

			handler := newCompanySearchHandler(searcher)

			request := httptest.NewRequest(
				http.MethodGet,
				"/v1/companies?name=SANTELLO",
				nil,
			)
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			result := response.Result()
			defer result.Body.Close()

			if result.StatusCode != test.expectedStatus {
				t.Fatalf(
					"expected status %d, got %d",
					test.expectedStatus,
					result.StatusCode,
				)
			}

			var payload struct {
				Error struct {
					Code string `json:"code"`
				} `json:"error"`
			}

			if err := json.NewDecoder(result.Body).Decode(
				&payload,
			); err != nil {
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

func TestCompanySearchHandlerRejectsInvalidPagination(t *testing.T) {
	searcher := &companySearcherStub{}
	handler := newCompanySearchHandler(searcher)

	request := httptest.NewRequest(
		http.MethodGet,
		"/v1/companies?name=SANTELLO&page=0",
		nil,
	)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	result := response.Result()
	defer result.Body.Close()

	if result.StatusCode != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			result.StatusCode,
		)
	}

	if searcher.calls != 0 {
		t.Fatalf(
			"expected searcher not to be called, got %d calls",
			searcher.calls,
		)
	}
}

func TestCompanySearchHandlerTradeNameJSON(t *testing.T) {
	name := "FANTASIA FILIAL"
	empty := ""

	tests := []struct {
		name     string
		value    *string
		expected string
	}{
		{"filled", &name, `"FANTASIA FILIAL"`},
		{"empty", &empty, `""`},
		{"null", nil, `null`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			searcher := &companySearcherStub{
				result: company.SearchResult{
					Items: []company.Summary{
						{
							CNPJ:      "12345678000200",
							TradeName: test.value,
						},
					},
					Pagination: company.NewPagination(1, 20, 1),
				},
			}

			request := httptest.NewRequest(
				http.MethodGet,
				"/v1/companies?name=EMPRESA",
				nil,
			)
			response := httptest.NewRecorder()

			newCompanySearchHandler(searcher).ServeHTTP(response, request)

			if response.Code != http.StatusOK {
				t.Fatalf("unexpected status: %d", response.Code)
			}

			var payload struct {
				Data struct {
					Companies struct {
						Items []map[string]json.RawMessage `json:"items"`
					} `json:"companies"`
				} `json:"data"`
			}

			if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
				t.Fatal(err)
			}

			if len(payload.Data.Companies.Items) != 1 {
				t.Fatal("expected one company")
			}

			got := string(payload.Data.Companies.Items[0]["trade_name"])
			if got != test.expected {
				t.Fatalf("trade_name: got %s, want %s", got, test.expected)
			}
		})
	}
}
