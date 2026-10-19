package company

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type companyRepositoryStub struct {
	receivedCNPJ                   string
	calls                          int
	details                        Details
	receivedOptions                FindOptions
	err                            error
	receivedSearchOptions          SearchOptions
	searchCalls                    int
	searchResult                   SearchResult
	searchErr                      error
	receivedTradeNameSearchOptions SearchOptions
	tradeNameSearchCalls           int
	tradeNameSearchResult          EstablishmentSearchResult
	tradeNameSearchErr             error
}

func (repository *companyRepositoryStub) FindByCNPJ(
	_ context.Context,
	cnpj string,
	options FindOptions,
) (Details, error) {
	repository.calls++
	repository.receivedCNPJ = cnpj
	repository.receivedOptions = options

	return repository.details, repository.err
}

func (repository *companyRepositoryStub) SearchByCorporateName(
	_ context.Context,
	options SearchOptions,
) (SearchResult, error) {
	repository.searchCalls++
	repository.receivedSearchOptions = options

	return repository.searchResult, repository.searchErr
}

func (repository *companyRepositoryStub) SearchByTradeName(
	_ context.Context,
	options SearchOptions,
) (EstablishmentSearchResult, error) {
	repository.tradeNameSearchCalls++
	repository.receivedTradeNameSearchOptions = options

	return repository.tradeNameSearchResult, repository.tradeNameSearchErr
}

func TestServiceFindByCNPJNormalizesInput(t *testing.T) {
	expectedDetails := Details{
		Company: Company{
			BasicCNPJ:     "12345678",
			CorporateName: "EMPRESA EXEMPLO",
		},
	}

	repository := &companyRepositoryStub{
		details: expectedDetails,
	}

	service := NewService(repository)

	expectedOptions := FindOptions{
		IncludeEstablishments: true,
		Page:                  3,
		PageSize:              50,
	}

	details, err := service.FindByCNPJ(
		context.Background(),
		"12.345.678/0001-90",
		expectedOptions,
	)
	if err != nil {
		t.Fatalf("find company by CNPJ: %v", err)
	}

	expectedCNPJ := "12345678000190"

	if repository.receivedCNPJ != expectedCNPJ {
		t.Fatalf(
			"expected repository CNPJ %q, got %q",
			expectedCNPJ,
			repository.receivedCNPJ,
		)
	}

	if repository.receivedOptions != expectedOptions {
		t.Fatalf(
			"expected repository options %#v, got %#v",
			expectedOptions,
			repository.receivedOptions,
		)
	}

	if details.Company.CorporateName !=
		expectedDetails.Company.CorporateName {
		t.Fatalf(
			"expected corporate name %q, got %q",
			expectedDetails.Company.CorporateName,
			details.Company.CorporateName,
		)
	}
}

func TestServiceFindByCNPJDoesNotQueryRepositoryForLongInput(
	t *testing.T,
) {
	repository := &companyRepositoryStub{}
	service := NewService(repository)

	_, err := service.FindByCNPJ(
		context.Background(),
		"123456789012345",
		FindOptions{},
	)

	if !errors.Is(err, ErrCNPJTooLong) {
		t.Fatalf("expected ErrCNPJTooLong, got %v", err)
	}

	if repository.calls != 0 {
		t.Fatalf(
			"expected repository not to be called, got %d calls",
			repository.calls,
		)
	}
}

func TestServiceFindByCNPJPreservesNotFoundError(t *testing.T) {
	repository := &companyRepositoryStub{
		err: ErrCompanyNotFound,
	}

	service := NewService(repository)

	_, err := service.FindByCNPJ(
		context.Background(),
		"12345678000190",
		FindOptions{},
	)

	if !errors.Is(err, ErrCompanyNotFound) {
		t.Fatalf("expected ErrCompanyNotFound, got %v", err)
	}
}

func TestServiceSearchByCorporateNameNormalizesInput(t *testing.T) {
	expectedResult := SearchResult{
		Items: []Summary{
			{
				CNPJ:          "12345678000190",
				CorporateName: "SANTELLO SISTEMAS LTDA",
			},
		},
		Pagination: NewPagination(2, 20, 25),
	}

	repository := &companyRepositoryStub{
		searchResult: expectedResult,
	}

	service := NewService(repository)

	result, err := service.SearchByCorporateName(
		context.Background(),
		SearchOptions{
			Name:     "  Santello  ",
			Page:     2,
			PageSize: 20,
		},
	)
	if err != nil {
		t.Fatalf("search companies by corporate name: %v", err)
	}

	expectedOptions := SearchOptions{
		Name:     "Santello",
		Page:     2,
		PageSize: 20,
	}

	if repository.receivedSearchOptions != expectedOptions {
		t.Fatalf(
			"expected repository options %#v, got %#v",
			expectedOptions,
			repository.receivedSearchOptions,
		)
	}

	if len(result.Items) != 1 {
		t.Fatalf(
			"expected 1 company summary, got %d",
			len(result.Items),
		)
	}

	if result.Items[0].CNPJ != expectedResult.Items[0].CNPJ {
		t.Fatalf(
			"expected CNPJ %q, got %q",
			expectedResult.Items[0].CNPJ,
			result.Items[0].CNPJ,
		)
	}
}

func TestServiceSearchByCorporateNameRejectsShortInput(t *testing.T) {
	repository := &companyRepositoryStub{}
	service := NewService(repository)

	_, err := service.SearchByCorporateName(
		context.Background(),
		SearchOptions{
			Name:     "  AB  ",
			Page:     1,
			PageSize: 20,
		},
	)

	if !errors.Is(err, ErrCorporateNameTooShort) {
		t.Fatalf(
			"expected ErrCorporateNameTooShort, got %v",
			err,
		)
	}

	if repository.searchCalls != 0 {
		t.Fatalf(
			"expected repository not to be called, got %d calls",
			repository.searchCalls,
		)
	}
}

func TestServiceSearchByCorporateNameRejectsLongInput(t *testing.T) {
	repository := &companyRepositoryStub{}
	service := NewService(repository)

	_, err := service.SearchByCorporateName(
		context.Background(),
		SearchOptions{
			Name:     strings.Repeat("A", 256),
			Page:     1,
			PageSize: 20,
		},
	)

	if !errors.Is(err, ErrCorporateNameTooLong) {
		t.Fatalf(
			"expected ErrCorporateNameTooLong, got %v",
			err,
		)
	}

	if repository.searchCalls != 0 {
		t.Fatalf(
			"expected repository not to be called, got %d calls",
			repository.searchCalls,
		)
	}
}
