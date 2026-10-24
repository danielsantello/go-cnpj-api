package search

import (
	"context"
	"fmt"
)

type enrichedRepository struct {
	repository Repository
	enrichment EnrichmentRepository
}

func NewEnrichedRepository(
	repository Repository,
	enrichment EnrichmentRepository,
) Repository {
	return &enrichedRepository{
		repository: repository,
		enrichment: enrichment,
	}
}

func (repository *enrichedRepository) Search(
	ctx context.Context,
	options Options,
) (Result, error) {
	if repository.repository == nil {
		return Result{}, ErrDisabled
	}

	result, err := repository.repository.Search(ctx, options)
	if err != nil {
		return Result{}, err
	}

	cnpjs := make([]string, 0, len(result.Items))
	for _, item := range result.Items {
		if item.CNPJ != "" {
			cnpjs = append(cnpjs, item.CNPJ)
		}
	}
	if len(cnpjs) == 0 {
		return result, nil
	}

	if repository.enrichment == nil {
		return Result{}, fmt.Errorf("search enrichment repository is not configured")
	}

	summaries, err := repository.enrichment.FindEstablishmentSummaries(
		ctx, cnpjs,
	)
	if err != nil {
		return Result{}, fmt.Errorf("enrich search results: %w", err)
	}

	// Copy items before filling fields, preserving the original collection.
	items := make([]Item, len(result.Items))
	copy(items, result.Items)

	for index := range items {
		summary := summaries[items[index].CNPJ]
		items[index].TradeName = summary.TradeName
		items[index].Municipality = summary.Municipality
		items[index].State = summary.State
		items[index].RegistrationStatus = summary.RegistrationStatus
	}

	result.Items = items
	return result, nil
}
