package v1

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/danielsantello/go-cnpj-api/internal/search"
)

type PartnerSource struct {
	repository *CompanyRepository
}

func NewPartnerSource(repository *CompanyRepository) *PartnerSource {
	return &PartnerSource{repository: repository}
}

func (source *PartnerSource) ReadBatch(
	ctx context.Context,
	afterKey string,
	limit int,
) ([]search.PartnerDocument, error) {
	if limit <= 0 || limit > 1000 {
		return nil, fmt.Errorf("partner batch limit must be between 1 and 1000")
	}

	afterRoot, afterOccurrence, err := search.ParsePartnerSourceKey(afterKey)
	if err != nil {
		return nil, err
	}

	rows, err := source.repository.db.QueryContext(ctx, `
		WITH roots AS (
			SELECT STRAIGHT_JOIN
				company.cnpj_root,
				company.legal_name,
				(
					SELECT establishment.cnpj
					FROM establishments AS establishment
					FORCE INDEX (idx_establishments_root_head_office_cnpj)
					WHERE establishment.cnpj_root = company.cnpj_root
					ORDER BY
						establishment.head_office_branch_indicator = '1' DESC,
						establishment.cnpj ASC
					LIMIT 1
				) AS cnpj,
				(
					SELECT COUNT(*)
					FROM companies AS same_company
					WHERE same_company.cnpj_root = company.cnpj_root
				) AS company_rows
			FROM companies AS company FORCE INDEX (idx_companies_cnpj_root)
			WHERE company.cnpj_root >= ?
				AND EXISTS (
					SELECT 1 FROM establishments AS establishment
					WHERE establishment.cnpj_root = company.cnpj_root
				)
				AND EXISTS (
					SELECT 1 FROM partners AS partner
					WHERE partner.cnpj_root = company.cnpj_root
				)
				AND (
					company.cnpj_root > ?
					OR (
						SELECT COUNT(*)
						FROM partners AS remaining_partner
						WHERE remaining_partner.cnpj_root = company.cnpj_root
					) > ?
				)
			ORDER BY company.cnpj_root ASC
			LIMIT ?
		),
		ranked AS (
			SELECT
				roots.cnpj_root,
				roots.cnpj,
				roots.legal_name,
				roots.company_rows,
				partner.partner_name,
				partner.partner_document,
				ROW_NUMBER() OVER (
					PARTITION BY roots.cnpj_root
					ORDER BY
						CAST(partner.partner_name AS BINARY),
						partner.partner_document IS NOT NULL,
						CAST(partner.partner_document AS BINARY)
				) AS occurrence
			FROM roots
			INNER JOIN partners AS partner
				ON partner.cnpj_root = roots.cnpj_root
		)
		SELECT
			cnpj_root,
			cnpj,
			legal_name,
			partner_name,
			partner_document,
			occurrence,
			company_rows
		FROM ranked
		WHERE cnpj_root > ? OR occurrence > ?
		ORDER BY cnpj_root ASC, occurrence ASC
		LIMIT ?
	`,
		afterRoot,
		afterRoot,
		afterOccurrence,
		limit+1,
		afterRoot,
		afterOccurrence,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("query indexing partners: %w", err)
	}
	defer rows.Close()

	documents := make([]search.PartnerDocument, 0, limit)
	previous := afterKey

	for rows.Next() {
		var document search.PartnerDocument
		var partnerDocument sql.NullString
		var occurrence uint64
		var companyRows int64

		if err := rows.Scan(
			&document.BasicCNPJ,
			&document.CNPJ,
			&document.CorporateName,
			&document.Name,
			&partnerDocument,
			&occurrence,
			&companyRows,
		); err != nil {
			return nil, fmt.Errorf("scan indexing partner: %w", err)
		}
		if companyRows != 1 {
			return nil, fmt.Errorf(
				"partner indexing requires one company row for basic CNPJ %q; found %d",
				document.BasicCNPJ, companyRows,
			)
		}

		key, err := search.MakePartnerSourceKey(document.BasicCNPJ, occurrence)
		if err != nil {
			return nil, err
		}
		if key <= previous {
			return nil, fmt.Errorf(
				"unordered or repeated partner source key: previous=%q current=%q",
				previous, key,
			)
		}

		document.SourceKey = key
		document.Document = nullableString(partnerDocument)
		documents = append(documents, document)
		previous = key
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate indexing partners: %w", err)
	}
	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("close indexing partners: %w", err)
	}

	return documents, nil
}

var _ search.PartnerSource = (*PartnerSource)(nil)
