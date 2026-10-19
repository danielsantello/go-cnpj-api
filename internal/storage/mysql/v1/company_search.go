package v1

import (
	"context"
	"fmt"
	"strings"

	"github.com/danielsantello/go-cnpj-api/internal/company"
)

const countCompaniesByCorporateNameQuery = `
	SELECT COUNT(*)
	FROM companies AS company
	WHERE company.legal_name LIKE ? ESCAPE '!'
		AND EXISTS (
			SELECT 1
			FROM establishments AS establishment
			WHERE establishment.cnpj_root = company.cnpj_root
		)
`

const searchCompaniesByCorporateNameQuery = `
	SELECT
		(
			SELECT establishment.cnpj
			FROM establishments AS establishment
			WHERE establishment.cnpj_root = company.cnpj_root
			ORDER BY
				establishment.head_office_branch_indicator = '1' DESC,
				establishment.cnpj
			LIMIT 1
		) AS cnpj,
		company.legal_name
	FROM companies AS company
	WHERE company.legal_name LIKE ? ESCAPE '!'
		AND EXISTS (
			SELECT 1
			FROM establishments AS establishment
			WHERE establishment.cnpj_root = company.cnpj_root
		)
	ORDER BY
		company.legal_name,
		company.cnpj_root
	LIMIT ?
	OFFSET ?
`

func (repository *CompanyRepository) SearchByCorporateName(
	ctx context.Context,
	options company.SearchOptions,
) (company.SearchResult, error) {
	prefix := corporateNamePrefix(options.Name)

	var totalItems int64

	err := repository.db.QueryRowContext(
		ctx,
		countCompaniesByCorporateNameQuery,
		prefix,
	).Scan(&totalItems)
	if err != nil {
		return company.SearchResult{}, fmt.Errorf(
			"count companies by corporate name: %w",
			err,
		)
	}

	offset := int64(options.Page-1) * int64(options.PageSize)

	rows, err := repository.db.QueryContext(
		ctx,
		searchCompaniesByCorporateNameQuery,
		prefix,
		options.PageSize,
		offset,
	)
	if err != nil {
		return company.SearchResult{}, fmt.Errorf(
			"query companies by corporate name: %w",
			err,
		)
	}
	defer rows.Close()

	items := make([]company.Summary, 0, options.PageSize)

	for rows.Next() {
		var item company.Summary

		if err := rows.Scan(
			&item.CNPJ,
			&item.CorporateName,
		); err != nil {
			return company.SearchResult{}, fmt.Errorf(
				"scan company summary: %w",
				err,
			)
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return company.SearchResult{}, fmt.Errorf(
			"iterate company summaries: %w",
			err,
		)
	}

	return company.SearchResult{
		Items: items,
		Pagination: company.NewPagination(
			options.Page,
			options.PageSize,
			totalItems,
		),
	}, nil
}

func corporateNamePrefix(value string) string {
	replacer := strings.NewReplacer(
		"!", "!!",
		"%", "!%",
		"_", "!_",
	)

	return replacer.Replace(value) + "%"
}
