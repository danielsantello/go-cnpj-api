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

type companyFinderStub struct {
	receivedInput string
	details       company.Details
	err           error
}

func (finder *companyFinderStub) FindByCNPJ(
	_ context.Context,
	input string,
) (company.Details, error) {
	finder.receivedInput = input

	return finder.details, finder.err
}

func TestCompanyHandlerReturnsCompany(t *testing.T) {
	expectedDetails := company.Details{
		Company: company.Company{
			BasicCNPJ:     "12345678",
			CorporateName: "EMPRESA EXEMPLO LTDA",
			LegalNature: company.CodeDescription{
				Code: "2062",
			},
			ResponsibleQualification: company.CodeDescription{
				Code: "49",
			},
		},
	}

	finder := &companyFinderStub{
		details: expectedDetails,
	}

	handler := requestIDMiddleware(
		newCompanyHandler(finder),
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/v1/companies/12.345.678/0001-90",
		nil,
	)
	request.SetPathValue(
		"cnpj",
		"12.345.678/0001-90",
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

	var payload companyResponse

	if err := json.NewDecoder(result.Body).Decode(&payload); err != nil {
		t.Fatalf("decode company response: %v", err)
	}

	expectedInput := "12.345.678/0001-90"

	if finder.receivedInput != expectedInput {
		t.Fatalf(
			"expected finder input %q, got %q",
			expectedInput,
			finder.receivedInput,
		)
	}

	if payload.Data.Company.CorporateName !=
		expectedDetails.Company.CorporateName {
		t.Fatalf(
			"expected corporate name %q, got %q",
			expectedDetails.Company.CorporateName,
			payload.Data.Company.CorporateName,
		)
	}

	if result.Header.Get("X-Request-ID") == "" {
		t.Fatal("expected X-Request-ID header")
	}
}

func TestCompanyHandlerReturnsStandardErrors(t *testing.T) {
	tests := []struct {
		name            string
		finderError     error
		expectedStatus  int
		expectedCode    string
		expectedDetails int
	}{
		{
			name:            "invalid CNPJ",
			finderError:     company.ErrCNPJTooLong,
			expectedStatus:  http.StatusBadRequest,
			expectedCode:    "INVALID_CNPJ",
			expectedDetails: 1,
		},
		{
			name:            "company not found",
			finderError:     company.ErrCompanyNotFound,
			expectedStatus:  http.StatusNotFound,
			expectedCode:    "COMPANY_NOT_FOUND",
			expectedDetails: 0,
		},
		{
			name:            "unexpected error",
			finderError:     errors.New("database failure"),
			expectedStatus:  http.StatusInternalServerError,
			expectedCode:    "INTERNAL_ERROR",
			expectedDetails: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			finder := &companyFinderStub{
				err: test.finderError,
			}

			handler := requestIDMiddleware(
				newCompanyHandler(finder),
			)

			request := httptest.NewRequest(
				http.MethodGet,
				"/v1/companies/12345678000190",
				nil,
			)
			request.SetPathValue(
				"cnpj",
				"12345678000190",
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

			var payload errorResponse

			if err := json.NewDecoder(result.Body).Decode(&payload); err != nil {
				t.Fatalf("decode error response: %v", err)
			}

			if payload.Error.Code != test.expectedCode {
				t.Fatalf(
					"expected error code %q, got %q",
					test.expectedCode,
					payload.Error.Code,
				)
			}

			if len(payload.Error.Details) != test.expectedDetails {
				t.Fatalf(
					"expected %d error details, got %d",
					test.expectedDetails,
					len(payload.Error.Details),
				)
			}

			if payload.Error.RequestID == "" {
				t.Fatal("expected request ID in error response")
			}
		})
	}
}
