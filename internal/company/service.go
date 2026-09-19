package company

import (
	"context"
	"fmt"
)

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (service *Service) FindByCNPJ(
	ctx context.Context,
	input string,
) (Details, error) {
	cnpj, err := NormalizeCNPJ(input)
	if err != nil {
		return Details{}, fmt.Errorf("normalize CNPJ: %w", err)
	}

	details, err := service.repository.FindByCNPJ(ctx, cnpj)
	if err != nil {
		return Details{}, fmt.Errorf("find company: %w", err)
	}

	return details, nil
}
