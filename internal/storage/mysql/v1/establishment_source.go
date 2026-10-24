package v1

import (
	"context"
	"fmt"

	"github.com/danielsantello/go-cnpj-api/internal/search"
)

type EstablishmentSource struct {
	repository *CompanyRepository
}

func NewEstablishmentSource(repository *CompanyRepository) *EstablishmentSource {
	return &EstablishmentSource{repository: repository}
}

func (source *EstablishmentSource) ReadBatch(
	ctx context.Context,
	afterCNPJ string,
	limit int,
) ([]search.EstablishmentDocument, error) {
	if limit <= 0 || limit > 1000 {
		return nil, fmt.Errorf("establishment batch limit must be between 1 and 1000")
	}

	// One extra row detects a repeated CNPJ at the batch boundary.
	rows, err := source.repository.db.QueryContext(ctx, `
		SELECT
			establishment.cnpj,
			company.legal_name,
			establishment.trade_name
		FROM establishments AS establishment
		INNER JOIN companies AS company
			ON company.cnpj_root = establishment.cnpj_root
		WHERE establishment.cnpj > ?
			AND establishment.trade_name IS NOT NULL
		ORDER BY establishment.cnpj ASC
		LIMIT ?
	`, afterCNPJ, limit+1)
	if err != nil {
		return nil, fmt.Errorf("query indexing establishments: %w", err)
	}
	defer rows.Close()

	documents := make([]search.EstablishmentDocument, 0, limit+1)
	previous := afterCNPJ

	for rows.Next() {
		var document search.EstablishmentDocument
		if err := rows.Scan(
			&document.CNPJ,
			&document.CorporateName,
			&document.TradeName,
		); err != nil {
			return nil, fmt.Errorf("scan indexing establishment: %w", err)
		}

		if document.CNPJ <= previous {
			return nil, fmt.Errorf(
				"unordered or repeated establishment CNPJ: previous=%q current=%q",
				previous, document.CNPJ,
			)
		}

		previous = document.CNPJ
		documents = append(documents, document)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate indexing establishments: %w", err)
	}
	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("close indexing establishments: %w", err)
	}

	if len(documents) > limit {
		documents = documents[:limit]
	}

	return documents, nil
}
