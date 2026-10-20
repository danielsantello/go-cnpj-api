package search

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/danielsantello/go-cnpj-api/internal/company"
)

var (
	ErrDisabled          = errors.New("search is disabled")
	ErrInvalidName       = errors.New("name must contain between 3 and 255 characters")
	ErrInvalidPagination = errors.New("invalid search pagination")
)

type Options struct {
	Name     string
	Page     int
	PageSize int
}

type Item struct {
	BasicCNPJ     string `json:"basic_cnpj"`
	CNPJ          string `json:"cnpj"`
	CorporateName string `json:"corporate_name"`
}

type Result struct {
	Items      []Item             `json:"items"`
	Pagination company.Pagination `json:"pagination"`
}

type Repository interface {
	Search(context.Context, Options) (Result, error)
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (service *Service) Search(
	ctx context.Context,
	options Options,
) (Result, error) {
	options.Name = strings.TrimSpace(options.Name)
	length := utf8.RuneCountInString(options.Name)

	if length < 3 || length > 255 {
		return Result{}, ErrInvalidName
	}
	if options.Page <= 0 || options.PageSize <= 0 || options.PageSize > 100 {
		return Result{}, ErrInvalidPagination
	}
	if service.repository == nil {
		return Result{}, ErrDisabled
	}

	result, err := service.repository.Search(ctx, options)
	if err != nil {
		return Result{}, fmt.Errorf("search companies: %w", err)
	}
	return result, nil
}
