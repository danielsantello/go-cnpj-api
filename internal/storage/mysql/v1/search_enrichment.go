package v1

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/danielsantello/go-cnpj-api/internal/search"
)

func (repository *CompanyRepository) FindEstablishmentSummaries(
	ctx context.Context,
	cnpjs []string,
) (map[string]search.EstablishmentSummary, error) {
	summaries := make(map[string]search.EstablishmentSummary)

	// Deduplicate full CNPJs without changing the search result collection.
	seen := make(map[string]bool, len(cnpjs))
	arguments := make([]any, 0, len(cnpjs))
	placeholders := make([]string, 0, len(cnpjs))

	for _, cnpj := range cnpjs {
		if cnpj == "" || seen[cnpj] {
			continue
		}
		seen[cnpj] = true
		arguments = append(arguments, cnpj)
		placeholders = append(placeholders, "?")
	}

	if len(arguments) == 0 {
		return summaries, nil
	}

	rows, err := repository.db.QueryContext(ctx, `
		SELECT
			establishment.cnpj,
			establishment.trade_name,
			establishment.municipality_code,
			municipality.name,
			establishment.state_code,
			establishment.registration_status_code
		FROM establishments AS establishment
		LEFT JOIN municipalities AS municipality
			ON municipality.code = establishment.municipality_code
		WHERE establishment.cnpj IN (`+
		strings.Join(placeholders, ",")+`)
	`, arguments...)
	if err != nil {
		return nil, fmt.Errorf("query establishment summaries: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var cnpj string
		var tradeName sql.NullString
		var municipalityCode sql.NullString
		var municipalityDescription sql.NullString
		var state sql.NullString
		var statusCode sql.NullString

		if err := rows.Scan(
			&cnpj,
			&tradeName,
			&municipalityCode,
			&municipalityDescription,
			&state,
			&statusCode,
		); err != nil {
			return nil, fmt.Errorf("scan establishment summary: %w", err)
		}

		if _, exists := summaries[cnpj]; exists {
			return nil, fmt.Errorf(
				"multiple establishment summary rows for CNPJ %q",
				cnpj,
			)
		}

		summaries[cnpj] = search.EstablishmentSummary{
			TradeName: nullableString(tradeName),
			Municipality: codeDescription(
				nullableString(municipalityCode),
				nullableString(municipalityDescription),
			),
			State: nullableString(state),
			RegistrationStatus: registrationStatus(
				nullableString(statusCode),
			),
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate establishment summaries: %w", err)
	}
	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("close establishment summaries: %w", err)
	}

	return summaries, nil
}

var _ search.EnrichmentRepository = (*CompanyRepository)(nil)
