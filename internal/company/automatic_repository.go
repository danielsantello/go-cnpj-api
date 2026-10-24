package company

import (
	"context"
	"errors"
)

var ErrSearchUnavailable = errors.New("search engine is unavailable")

type NameSearchRepository interface {
	SearchByCorporateName(context.Context, SearchOptions) (SearchResult, error)
	SearchByTradeName(context.Context, SearchOptions) (EstablishmentSearchResult, error)
	SearchByPartnerName(context.Context, PartnerSearchOptions) (PartnerSearchResult, error)
}

type automaticRepository struct {
	Repository
	primary NameSearchRepository
}

func NewAutomaticRepository(
	fallback Repository,
	primary NameSearchRepository,
) Repository {
	return &automaticRepository{
		Repository: fallback,
		primary:    primary,
	}
}

func (repository *automaticRepository) SearchByCorporateName(
	ctx context.Context,
	options SearchOptions,
) (SearchResult, error) {
	if repository.primary != nil {
		result, err := repository.primary.SearchByCorporateName(ctx, options)
		if err == nil || !errors.Is(err, ErrSearchUnavailable) {
			return result, err
		}
	}
	if err := ctx.Err(); err != nil {
		return SearchResult{}, err
	}
	return repository.Repository.SearchByCorporateName(ctx, options)
}

func (repository *automaticRepository) SearchByTradeName(
	ctx context.Context,
	options SearchOptions,
) (EstablishmentSearchResult, error) {
	if repository.primary != nil {
		result, err := repository.primary.SearchByTradeName(ctx, options)
		if err == nil || !errors.Is(err, ErrSearchUnavailable) {
			return result, err
		}
	}
	if err := ctx.Err(); err != nil {
		return EstablishmentSearchResult{}, err
	}
	return repository.Repository.SearchByTradeName(ctx, options)
}

func (repository *automaticRepository) SearchByPartnerName(
	ctx context.Context,
	options PartnerSearchOptions,
) (PartnerSearchResult, error) {
	if repository.primary != nil {
		result, err := repository.primary.SearchByPartnerName(ctx, options)
		if err == nil || !errors.Is(err, ErrSearchUnavailable) {
			return result, err
		}
	}
	if err := ctx.Err(); err != nil {
		return PartnerSearchResult{}, err
	}
	return repository.Repository.SearchByPartnerName(ctx, options)
}

var _ Repository = (*automaticRepository)(nil)
