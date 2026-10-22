package v1

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/danielsantello/go-cnpj-api/internal/company"
)

const countCompaniesByCorporateNameQuery = `
	SELECT COUNT(*)
	FROM (
		SELECT 1
		FROM companies AS company
		WHERE company.legal_name LIKE ? ESCAPE '!'
			AND EXISTS (
				SELECT 1
				FROM establishments AS establishment
				WHERE establishment.cnpj_root = company.cnpj_root
			)
		LIMIT 1001
	) AS matches
`

const searchCompaniesByCorporateNameQuery = `
	SELECT
		establishment.cnpj,
		company.legal_name,
		establishment.municipality_code,
		municipality.name,
		establishment.state_code,
		establishment.registration_status_code
	FROM companies AS company
	INNER JOIN LATERAL (
		SELECT
			candidate.cnpj,
			candidate.municipality_code,
			candidate.state_code,
			candidate.registration_status_code
		FROM establishments AS candidate
		WHERE candidate.cnpj_root = company.cnpj_root
		ORDER BY
			candidate.head_office_branch_indicator = '1' DESC,
			candidate.cnpj
		LIMIT 1
	) AS establishment ON TRUE
	LEFT JOIN municipalities AS municipality
		ON municipality.code = establishment.municipality_code
	WHERE company.legal_name LIKE ? ESCAPE '!'
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

	if totalItems > company.MaximumSearchResults {
		return company.SearchResult{}, company.ErrSearchTooBroad
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
		var municipalityCode sql.NullString
		var municipalityDescription sql.NullString
		var state sql.NullString
		var statusCode sql.NullString

		if err := rows.Scan(
			&item.CNPJ,
			&item.CorporateName,
			&municipalityCode,
			&municipalityDescription,
			&state,
			&statusCode,
		); err != nil {
			return company.SearchResult{}, fmt.Errorf(
				"scan company summary: %w",
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
