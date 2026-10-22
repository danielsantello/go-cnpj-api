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
		establishment.cnpj,
		company.legal_name,
		establishment.municipality_code,
		municipality.name,
		establishment.state_code,
		establishment.registration_status_code
`

const partnerSearchResultsFrom = `
	FROM partners AS partner
	INNER JOIN companies AS company
		ON company.cnpj_root = partner.cnpj_root
	INNER JOIN LATERAL (
		SELECT
			candidate.cnpj,
			candidate.municipality_code,
			candidate.state_code,
			candidate.registration_status_code
		FROM establishments AS candidate
		WHERE candidate.cnpj_root = partner.cnpj_root
		ORDER BY
			candidate.head_office_branch_indicator = '1' DESC,
			candidate.cnpj
		LIMIT 1
	) AS establishment ON TRUE
	LEFT JOIN municipalities AS municipality
		ON municipality.code = establishment.municipality_code
	WHERE TRUE
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
		"SELECT COUNT(*) FROM (SELECT 1"+
			partnerSearchFrom+filter+
			" LIMIT 1001) AS matches",
		arguments...,
	).Scan(&totalItems)
	if err != nil {
		return company.PartnerSearchResult{}, fmt.Errorf(
			"count companies by partner name: %w", err,
		)
	}

	if totalItems > company.MaximumSearchResults {
		return company.PartnerSearchResult{}, company.ErrSearchTooBroad
	}

	offset := int64(options.Page-1) * int64(options.PageSize)
	queryArguments := append(
		arguments,
		options.PageSize,
		offset,
	)

	rows, err := repository.db.QueryContext(
		ctx,
		partnerSearchColumns+partnerSearchResultsFrom+filter+`
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
		var municipalityCode sql.NullString
		var municipalityDescription sql.NullString
		var state sql.NullString
		var statusCode sql.NullString

		if err := rows.Scan(
			&item.Name,
			&document,
			&item.CNPJ,
			&item.CorporateName,
			&municipalityCode,
			&municipalityDescription,
			&state,
			&statusCode,
		); err != nil {
			return company.PartnerSearchResult{}, fmt.Errorf(
				"scan partner summary: %w", err,
			)
		}

		item.Document = nullableString(document)
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
