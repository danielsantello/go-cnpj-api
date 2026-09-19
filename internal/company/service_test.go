package company

import (
	"context"
	"errors"
	"testing"
)

type companyRepositoryStub struct {
	receivedCNPJ string
	calls        int
	details      Details
	err          error
}

func (repository *companyRepositoryStub) FindByCNPJ(
	_ context.Context,
	cnpj string,
) (Details, error) {
	repository.calls++
	repository.receivedCNPJ = cnpj

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

	details, err := service.FindByCNPJ(
		context.Background(),
		"12.345.678/0001-90",
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
	)

	if !errors.Is(err, ErrCompanyNotFound) {
		t.Fatalf("expected ErrCompanyNotFound, got %v", err)
	}
}
