package manticore

import (
	"context"
	"database/sql"
	"errors"

	"github.com/danielsantello/go-cnpj-api/internal/company"
	"github.com/danielsantello/go-cnpj-api/internal/search"
)

type NameRepository struct {
	companies search.Repository
}

func NewNameRepository(
	db *sql.DB,
	companyTable string,
	enrichment search.EnrichmentRepository,
) (*NameRepository, error) {
	repository, err := NewRepositoryForTable(db, companyTable)
	if err != nil {
		return nil, err
	}

	return &NameRepository{
		companies: search.NewEnrichedRepository(
			&corporateNameRepository{repository: repository},
			enrichment,
		),
	}, nil
}

type corporateNameRepository struct {
	repository *Repository
}

func (repository *corporateNameRepository) Search(
	ctx context.Context,
	options search.Options,
) (search.Result, error) {
	result, err := repository.repository.searchWithExpression(
		ctx,
		options,
		corporateNameSearchExpression(options.Name),
		false,
	)
	if errors.Is(err, search.ErrTooBroad) {
		return search.Result{}, company.ErrSearchTooBroad
	}
	if err != nil {
		return search.Result{}, nameSearchQueryError(
			ctx, "search corporate names in Manticore", err,
		)
	}

	return result, nil
}

func (repository *NameRepository) SearchByCorporateName(
	ctx context.Context,
	options company.SearchOptions,
) (company.SearchResult, error) {
	result, err := repository.companies.Search(ctx, search.Options{
		Name:     options.Name,
		Page:     options.Page,
		PageSize: options.PageSize,
	})
	if err != nil {
		return company.SearchResult{}, err
	}

	items := make([]company.Summary, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, company.Summary{
			CNPJ:               item.CNPJ,
			CorporateName:      item.CorporateName,
			TradeName:          item.TradeName,
			Municipality:       item.Municipality,
			State:              item.State,
			RegistrationStatus: item.RegistrationStatus,
		})
	}

	return company.SearchResult{
		Items:      items,
		Pagination: result.Pagination,
	}, nil
}
