package company

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"
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
	options FindOptions,
) (Details, error) {
	cnpj, err := NormalizeCNPJ(input)
	if err != nil {
		return Details{}, fmt.Errorf("normalize CNPJ: %w", err)
	}

	details, err := service.repository.FindByCNPJ(
		ctx,
		cnpj,
		options,
	)
	if err != nil {
		return Details{}, fmt.Errorf("find company: %w", err)
	}

	return details, nil
}

func (service *Service) SearchByCorporateName(
	ctx context.Context,
	options SearchOptions,
) (SearchResult, error) {
	options.Name = strings.TrimSpace(options.Name)

	nameLength := utf8.RuneCountInString(options.Name)

	if nameLength < minCorporateNameSearchLength {
		return SearchResult{}, ErrCorporateNameTooShort
	}

	if nameLength > maxCorporateNameSearchLength {
		return SearchResult{}, ErrCorporateNameTooLong
	}

	result, err := service.repository.SearchByCorporateName(
		ctx,
		options,
	)
	if err != nil {
		return SearchResult{}, fmt.Errorf(
			"search companies by corporate name: %w",
			err,
		)
	}

	return result, nil
}
