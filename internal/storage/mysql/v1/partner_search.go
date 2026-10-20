package v1

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/danielsantello/go-cnpj-api/internal/company"
)

const partnerSearchFrom = `
	FROM partners AS partner
	INNER JOIN companies AS company
		ON company.cnpj_root = partner.cnpj_root
	WHERE EXISTS (
		SELECT 1
		FROM establishments AS establishment
		WHERE establishment.cnpj_root = partner.cnpj_root
	)
`

const partnerSearchColumns = `
	SELECT
		partner.partner_name,
		partner.partner_document,
		(
			SELECT establishment.cnpj
			FROM establishments AS establishment
			WHERE establishment.cnpj_root = partner.cnpj_root
			ORDER BY
				establishment.head_office_branch_indicator = '1' DESC,
				establishment.cnpj
			LIMIT 1
		),
		company.legal_name
`

func partnerSearchFilter(
	options company.PartnerSearchOptions,
) (string, []any) {
	if options.Document != nil {
		return `
			AND partner.partner_name = ?
			AND partner.partner_document = ?
		`, []any{options.Name, *options.Document}
	}

	return `
		AND partner.partner_name LIKE ? ESCAPE '!'
	`, []any{corporateNamePrefix(options.Name)}
}

func (repository *CompanyRepository) SearchByPartnerName(
	ctx context.Context,
	options company.PartnerSearchOptions,
) (company.PartnerSearchResult, error) {
	filter, arguments := partnerSearchFilter(options)

	var totalItems int64

	err := repository.db.QueryRowContext(
		ctx,
		"SELECT COUNT(*)"+partnerSearchFrom+filter,
		arguments...,
	).Scan(&totalItems)
	if err != nil {
		return company.PartnerSearchResult{}, fmt.Errorf(
			"count companies by partner name: %w", err,
		)
	}

	offset := int64(options.Page-1) * int64(options.PageSize)
	queryArguments := append(
		arguments,
		options.PageSize,
		offset,
	)

	rows, err := repository.db.QueryContext(
		ctx,
		partnerSearchColumns+partnerSearchFrom+filter+`
			ORDER BY
				partner.partner_name,
				partner.partner_document,
				partner.cnpj_root
			LIMIT ?
			OFFSET ?
		`,
		queryArguments...,
	)
	if err != nil {
		return company.PartnerSearchResult{}, fmt.Errorf(
			"query companies by partner name: %w", err,
		)
	}
	defer rows.Close()

	items := make([]company.PartnerSummary, 0, options.PageSize)

	for rows.Next() {
		var item company.PartnerSummary
		var document sql.NullString

		if err := rows.Scan(
			&item.Name,
			&document,
			&item.CNPJ,
			&item.CorporateName,
		); err != nil {
			return company.PartnerSearchResult{}, fmt.Errorf(
				"scan partner summary: %w", err,
			)
		}

		item.Document = nullableString(document)
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return company.PartnerSearchResult{}, fmt.Errorf(
			"iterate partner summaries: %w", err,
		)
	}

	return company.PartnerSearchResult{
		Items: items,
		Pagination: company.NewPagination(
			options.Page,
			options.PageSize,
			totalItems,
		),
	}, nil
}
