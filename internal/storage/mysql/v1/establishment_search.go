package v1

import (
	"context"
	"fmt"

	"github.com/danielsantello/go-cnpj-api/internal/company"
)

const countEstablishmentsByTradeNameQuery = `
	SELECT COUNT(*)
	FROM establishments AS establishment
	INNER JOIN companies AS company
		ON company.cnpj_root = establishment.cnpj_root
	WHERE establishment.trade_name LIKE ? ESCAPE '!'
`

const searchEstablishmentsByTradeNameQuery = `
	SELECT
		establishment.cnpj,
		establishment.trade_name,
		company.legal_name
	FROM establishments AS establishment
	INNER JOIN companies AS company
		ON company.cnpj_root = establishment.cnpj_root
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

		if err := rows.Scan(
			&item.CNPJ,
			&item.TradeName,
			&item.CorporateName,
		); err != nil {
			return company.EstablishmentSearchResult{}, fmt.Errorf(
				"scan establishment summary: %w",
				err,
			)
		}

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
