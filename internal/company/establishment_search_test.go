package company

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestServiceSearchByTradeNameNormalizesInput(t *testing.T) {
	expectedResult := EstablishmentSearchResult{
		Items: []EstablishmentSummary{
			{
				CNPJ:          "12345678000190",
				TradeName:     "LOJA EXEMPLO",
				CorporateName: "EMPRESA EXEMPLO",
			},
		},
		Pagination: NewPagination(2, 10, 15),
	}

	repository := &companyRepositoryStub{
		tradeNameSearchResult: expectedResult,
	}
	service := NewService(repository)

	result, err := service.SearchByTradeName(
		context.Background(),
		SearchOptions{
			Name:     "  LOJA EXEMPLO  ",
			Page:     2,
			PageSize: 10,
		},
	)
	if err != nil {
		t.Fatalf("search by trade name: %v", err)
	}

	expectedOptions := SearchOptions{
		Name:     "LOJA EXEMPLO",
		Page:     2,
		PageSize: 10,
	}

	if repository.receivedTradeNameSearchOptions != expectedOptions {
		t.Fatalf(
			"expected options %#v, got %#v",
			expectedOptions,
			repository.receivedTradeNameSearchOptions,
		)
	}

	if repository.tradeNameSearchCalls != 1 {
		t.Fatalf(
			"expected 1 repository call, got %d",
			repository.tradeNameSearchCalls,
		)
	}

	if !reflect.DeepEqual(result, expectedResult) {
		t.Fatalf("expected result %#v, got %#v", expectedResult, result)
	}
}

func TestServiceSearchByTradeNameValidatesLength(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{name: "empty", input: "", wantErr: ErrTradeNameTooShort},
		{name: "spaces", input: "   ", wantErr: ErrTradeNameTooShort},
		{name: "short", input: "AB", wantErr: ErrTradeNameTooShort},
		{name: "short unicode", input: "ÁÉ", wantErr: ErrTradeNameTooShort},
		{name: "minimum", input: "ABC"},
		{name: "minimum unicode", input: "ÁÉÍ"},
		{name: "maximum", input: strings.Repeat("Á", 255)},
		{
			name:    "too long",
			input:   strings.Repeat("Á", 256),
			wantErr: ErrTradeNameTooLong,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &companyRepositoryStub{}
			service := NewService(repository)

			_, err := service.SearchByTradeName(
				context.Background(),
				SearchOptions{
					Name:     test.input,
					Page:     1,
					PageSize: 20,
				},
			)

			if !errors.Is(err, test.wantErr) {
				t.Fatalf("expected error %v, got %v", test.wantErr, err)
			}

			expectedCalls := 1
			if test.wantErr != nil {
				expectedCalls = 0
			}

			if repository.tradeNameSearchCalls != expectedCalls {
				t.Fatalf(
					"expected %d repository calls, got %d",
					expectedCalls,
					repository.tradeNameSearchCalls,
				)
			}
		})
	}
}

func TestServiceSearchByTradeNamePreservesRepositoryError(t *testing.T) {
	expectedErr := errors.New("database unavailable")
	repository := &companyRepositoryStub{
		tradeNameSearchErr: expectedErr,
	}
	service := NewService(repository)

	_, err := service.SearchByTradeName(
		context.Background(),
		SearchOptions{
			Name:     "LOJA",
			Page:     1,
			PageSize: 20,
		},
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected repository error %v, got %v", expectedErr, err)
	}
}
