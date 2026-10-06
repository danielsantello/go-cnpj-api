package company

import (
	"context"
	"errors"
	"testing"
)

type companyRepositoryStub struct {
	receivedCNPJ    string
	calls           int
	details         Details
	receivedOptions FindOptions
	err             error
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
