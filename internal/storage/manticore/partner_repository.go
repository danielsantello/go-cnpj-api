package manticore

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/danielsantello/go-cnpj-api/internal/company"
	"github.com/danielsantello/go-cnpj-api/internal/search"
)

type PartnerRepository struct {
	db         *sql.DB
	table      string
	enrichment search.EnrichmentRepository
}

func NewPartnerRepository(
	db *sql.DB,
	table string,
	enrichment search.EnrichmentRepository,
) (*PartnerRepository, error) {
	if !partnerTableNamePattern.MatchString(table) {
		return nil, fmt.Errorf("invalid partner search table name: %q", table)
	}

	return &PartnerRepository{
		db: db, table: table, enrichment: enrichment,
	}, nil
}

func partnerMatchFilter(
	expression string,
	document *string,
) (string, []any) {
	filter := "WHERE MATCH(?)"
	arguments := []any{expression}

	if document != nil {
		filter += " AND has_document = 1 AND partner_document = ?"
		arguments = append(arguments, *document)
	}
	return filter, arguments
}

func (repository *PartnerRepository) countMatches(
	ctx context.Context,
	filter string,
	arguments []any,
) (int64, error) {
	rows, err := repository.db.QueryContext(ctx, `
		SELECT id
		FROM `+repository.table+`
		`+filter+`
		LIMIT 1001
		OPTION max_matches=1001, cutoff=0
	`, arguments...)
	if err != nil {
		return 0, nameSearchQueryError(ctx, "count partner matches", err)
	}
	defer rows.Close()

	var total int64
	for rows.Next() {
		var id uint64
		if err := rows.Scan(&id); err != nil {
			return 0, nameSearchQueryError(ctx, "scan partner match ID", err)
		}
		total++
	}
	if err := rows.Err(); err != nil {
		return 0, nameSearchQueryError(ctx, "iterate partner match IDs", err)
	}
	if err := rows.Close(); err != nil {
		return 0, nameSearchQueryError(ctx, "close partner match IDs", err)
	}
	return total, nil
}

func (repository *PartnerRepository) SearchByPartnerName(
	ctx context.Context,
	options company.PartnerSearchOptions,
) (company.PartnerSearchResult, error) {
	if options.Page <= 0 || options.PageSize <= 0 || options.PageSize > 100 {
		return company.PartnerSearchResult{}, fmt.Errorf(
			"invalid partner search pagination",
		)
	}

	result := company.PartnerSearchResult{
		Items: []company.PartnerSummary{},
		Pagination: company.NewPagination(
			options.Page, options.PageSize, 0,
		),
	}
	expression := searchExpression(options.Name)
	if expression == "" {
		return result, nil
	}

	filter, arguments := partnerMatchFilter(expression, options.Document)
	total, err := repository.countMatches(ctx, filter, arguments)
	if err != nil {
		return company.PartnerSearchResult{}, err
	}
	if total > int64(company.MaximumSearchResults) {
		return company.PartnerSearchResult{}, company.ErrSearchTooBroad
	}

	result.Pagination = company.NewPagination(
		options.Page, options.PageSize, total,
	)
	if total == 0 || int64(options.Page) > result.Pagination.TotalPages {
		return result, nil
	}

	offset := int64(options.Page-1) * int64(options.PageSize)
	pageArguments := append([]any(nil), arguments...)
	pageArguments = append(pageArguments, options.PageSize, offset)

	rows, err := repository.db.QueryContext(ctx, `
		SELECT
			partner_name,
			partner_document,
			has_document,
			cnpj,
			corporate_name
		FROM `+repository.table+`
		`+filter+`
		ORDER BY id ASC
		LIMIT ? OFFSET ?
		OPTION max_matches=1000, cutoff=0
	`, pageArguments...)
	if err != nil {
		return company.PartnerSearchResult{}, nameSearchQueryError(
			ctx, "query partner matches", err,
		)
	}
	defer rows.Close()

	cnpjs := make([]string, 0, options.PageSize)
	for rows.Next() {
		var item company.PartnerSummary
		var documentValue string
		var hasDocument int

		if err := rows.Scan(
			&item.Name,
			&documentValue,
			&hasDocument,
			&item.CNPJ,
			&item.CorporateName,
		); err != nil {
			return company.PartnerSearchResult{}, nameSearchQueryError(
				ctx, "scan partner match", err,
			)
		}
		switch hasDocument {
		case 0:
			item.Document = nil
		case 1:
			item.Document = &documentValue
		default:
			return company.PartnerSearchResult{}, fmt.Errorf(
				"invalid document presence flag: %d", hasDocument,
			)
		}

		result.Items = append(result.Items, item)
		cnpjs = append(cnpjs, item.CNPJ)
	}
	if err := rows.Err(); err != nil {
		return company.PartnerSearchResult{}, nameSearchQueryError(
			ctx, "iterate partner matches", err,
		)
	}
	if err := rows.Close(); err != nil {
		return company.PartnerSearchResult{}, nameSearchQueryError(
			ctx, "close partner matches", err,
		)
	}

	if len(result.Items) == 0 {
		return result, nil
	}
	if repository.enrichment == nil {
		return company.PartnerSearchResult{}, fmt.Errorf(
			"partner enrichment repository is not configured",
		)
	}

	summaries, err := repository.enrichment.FindEstablishmentSummaries(ctx, cnpjs)
	if err != nil {
		return company.PartnerSearchResult{}, fmt.Errorf(
			"enrich partner matches: %w", err,
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
