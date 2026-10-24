package manticore

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/danielsantello/go-cnpj-api/internal/company"
	"github.com/danielsantello/go-cnpj-api/internal/search"
)

type EstablishmentRepository struct {
	db         *sql.DB
	table      string
	enrichment search.EnrichmentRepository
}

func NewEstablishmentRepository(
	db *sql.DB,
	table string,
	enrichment search.EnrichmentRepository,
) (*EstablishmentRepository, error) {
	if !establishmentTableNamePattern.MatchString(table) {
		return nil, fmt.Errorf(
			"invalid establishment search table name: %q", table,
		)
	}

	return &EstablishmentRepository{
		db:         db,
		table:      table,
		enrichment: enrichment,
	}, nil
}

func (repository *EstablishmentRepository) SearchByTradeName(
	ctx context.Context,
	options company.SearchOptions,
) (company.EstablishmentSearchResult, error) {
	if options.Page <= 0 || options.PageSize <= 0 || options.PageSize > 100 {
		return company.EstablishmentSearchResult{}, fmt.Errorf(
			"invalid establishment search pagination",
		)
	}

	result := company.EstablishmentSearchResult{
		Items: []company.EstablishmentSummary{},
		Pagination: company.NewPagination(
			options.Page, options.PageSize, 0,
		),
	}

	// Only trade_name is full-text indexed in this table.
	expression := searchExpression(options.Name)
	if expression == "" {
		return result, nil
	}

	counter := &Repository{db: repository.db, table: repository.table}
	total, err := counter.countMatches(ctx, expression)
	if err != nil {
		return company.EstablishmentSearchResult{}, nameSearchQueryError(
			ctx, "count establishment matches", err,
		)
	}
	if total > int64(company.MaximumSearchResults) {
		return company.EstablishmentSearchResult{}, company.ErrSearchTooBroad
	}

	result.Pagination = company.NewPagination(
		options.Page, options.PageSize, total,
	)
	if total == 0 || int64(options.Page) > result.Pagination.TotalPages {
		return result, nil
	}

	offset := int64(options.Page-1) * int64(options.PageSize)
	rows, err := repository.db.QueryContext(ctx, `
		SELECT cnpj, trade_name, corporate_name
		FROM `+repository.table+`
		WHERE MATCH(?)
		ORDER BY id ASC
		LIMIT ? OFFSET ?
		OPTION max_matches=1000, cutoff=0
	`, expression, options.PageSize, offset)
	if err != nil {
		return company.EstablishmentSearchResult{}, nameSearchQueryError(
			ctx, "query establishment matches", err,
		)
	}
	defer rows.Close()

	cnpjs := make([]string, 0, options.PageSize)
	for rows.Next() {
		var item company.EstablishmentSummary
		if err := rows.Scan(
			&item.CNPJ,
			&item.TradeName,
			&item.CorporateName,
		); err != nil {
			return company.EstablishmentSearchResult{}, nameSearchQueryError(
				ctx, "scan establishment match", err,
			)
		}

		cnpjs = append(cnpjs, item.CNPJ)
		result.Items = append(result.Items, item)
	}
	if err := rows.Err(); err != nil {
		return company.EstablishmentSearchResult{}, nameSearchQueryError(
			ctx, "iterate establishment matches", err,
		)
	}
	if err := rows.Close(); err != nil {
		return company.EstablishmentSearchResult{}, nameSearchQueryError(
			ctx, "close establishment matches", err,
		)
	}

	if len(result.Items) == 0 {
		return result, nil
	}
	if repository.enrichment == nil {
		return company.EstablishmentSearchResult{}, fmt.Errorf(
			"establishment enrichment repository is not configured",
		)
	}

	summaries, err := repository.enrichment.FindEstablishmentSummaries(
		ctx, cnpjs,
	)
	if err != nil {
		return company.EstablishmentSearchResult{}, fmt.Errorf(
			"enrich establishment matches: %w", err,
		)
	}

	for position := range result.Items {
		item := &result.Items[position]
		summary := summaries[item.CNPJ]
		item.Municipality = summary.Municipality
		item.State = summary.State
		item.RegistrationStatus = summary.RegistrationStatus
	}

	return result, nil
}
