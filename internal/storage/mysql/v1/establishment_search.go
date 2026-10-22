package v1

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/danielsantello/go-cnpj-api/internal/company"
)

const countEstablishmentsByTradeNameQuery = `
	SELECT COUNT(*)
	FROM (
		SELECT 1
		FROM establishments AS establishment
		INNER JOIN companies AS company
			ON company.cnpj_root = establishment.cnpj_root
		WHERE establishment.trade_name LIKE ? ESCAPE '!'
		LIMIT 1001
	) AS matches
`

const searchEstablishmentsByTradeNameQuery = `
	SELECT
		establishment.cnpj,
		establishment.trade_name,
		company.legal_name,
		establishment.municipality_code,
		municipality.name,
		establishment.state_code,
		establishment.registration_status_code
	FROM establishments AS establishment
	INNER JOIN companies AS company
		ON company.cnpj_root = establishment.cnpj_root
	LEFT JOIN municipalities AS municipality
		ON municipality.code = establishment.municipality_code
	WHERE establishment.trade_name LIKE ? ESCAPE '!'
	ORDER BY
		establishment.trade_name,
		establishment.cnpj
	LIMIT ?
	OFFSET ?
`

func (repository *CompanyRepository) SearchByTradeName(
	ctx context.Context,
	options company.SearchOptions,
) (company.EstablishmentSearchResult, error) {
	prefix := corporateNamePrefix(options.Name)

	var totalItems int64

	err := repository.db.QueryRowContext(
		ctx,
		countEstablishmentsByTradeNameQuery,
		prefix,
	).Scan(&totalItems)
	if err != nil {
		return company.EstablishmentSearchResult{}, fmt.Errorf(
			"count establishments by trade name: %w",
			err,
		)
	}

	if totalItems > company.MaximumSearchResults {
		return company.EstablishmentSearchResult{}, company.ErrSearchTooBroad
	}

	offset := int64(options.Page-1) * int64(options.PageSize)

	rows, err := repository.db.QueryContext(
		ctx,
		searchEstablishmentsByTradeNameQuery,
		prefix,
		options.PageSize,
		offset,
	)
	if err != nil {
		return company.EstablishmentSearchResult{}, fmt.Errorf(
			"query establishments by trade name: %w",
			err,
		)
	}
	defer rows.Close()

	items := make([]company.EstablishmentSummary, 0, options.PageSize)

	for rows.Next() {
		var item company.EstablishmentSummary
		var municipalityCode sql.NullString
		var municipalityDescription sql.NullString
		var state sql.NullString
		var statusCode sql.NullString

		if err := rows.Scan(
			&item.CNPJ,
			&item.TradeName,
			&item.CorporateName,
			&municipalityCode,
			&municipalityDescription,
			&state,
			&statusCode,
		); err != nil {
			return company.EstablishmentSearchResult{}, fmt.Errorf(
				"scan establishment summary: %w",
				err,
			)
		}

		item.Municipality = codeDescription(
			nullableString(municipalityCode),
			nullableString(municipalityDescription),
		)
		item.State = nullableString(state)
		item.RegistrationStatus = registrationStatus(
			nullableString(statusCode),
		)

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return company.EstablishmentSearchResult{}, fmt.Errorf(
			"iterate establishment summaries: %w",
			err,
		)
	}

	return company.EstablishmentSearchResult{
		Items: items,
		Pagination: company.NewPagination(
			options.Page,
			options.PageSize,
			totalItems,
		),
	}, nil
}
