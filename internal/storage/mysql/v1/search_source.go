package v1

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/danielsantello/go-cnpj-api/internal/search"
)

type SearchSource struct {
	db *sql.DB
}

func NewSearchSource(db *sql.DB) *SearchSource {
	return &SearchSource{db: db}
}

func (source *SearchSource) ReadBatch(
	ctx context.Context,
	afterBasicCNPJ string,
	limit int,
) ([]search.Document, error) {
	if limit <= 0 {
		return nil, fmt.Errorf("batch limit must be greater than zero")
	}

	rows, err := source.db.QueryContext(ctx, `
		SELECT cnpj_root, legal_name
		FROM companies
		WHERE cnpj_root > ?
		ORDER BY cnpj_root
		LIMIT ?
	`, afterBasicCNPJ, limit)
	if err != nil {
		return nil, fmt.Errorf("query indexing companies: %w", err)
	}

	documents := make([]search.Document, 0, limit)

	for rows.Next() {
		var document search.Document
		if err := rows.Scan(&document.BasicCNPJ, &document.CorporateName); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan indexing company: %w", err)
		}

		root, err := strconv.ParseUint(document.BasicCNPJ, 10, 64)
		if err != nil {
			rows.Close()
			return nil, fmt.Errorf("parse basic CNPJ %q: %w", document.BasicCNPJ, err)
		}

		// Manticore document IDs must be nonzero.
		document.ID = root + 1
		documents = append(documents, document)
	}

	readErr := rows.Err()
	closeErr := rows.Close()
	if readErr != nil {
		return nil, fmt.Errorf("iterate indexing companies: %w", readErr)
	}
	if closeErr != nil {
		return nil, fmt.Errorf("close indexing companies: %w", closeErr)
	}
	if len(documents) == 0 {
		return documents, nil
	}

	positions := make(map[string]int, len(documents))
	arguments := make([]any, len(documents))
	placeholders := make([]string, len(documents))

	for index := range documents {
		root := documents[index].BasicCNPJ
		positions[root] = index
		arguments[index] = root
		placeholders[index] = "?"
	}

	in := strings.Join(placeholders, ",")

	if err := source.readEstablishments(ctx, documents, positions, in, arguments); err != nil {
		return nil, err
	}
	if err := source.readPartners(ctx, documents, positions, in, arguments); err != nil {
		return nil, err
	}

	return documents, nil
}

func (source *SearchSource) readEstablishments(
	ctx context.Context,
	documents []search.Document,
	positions map[string]int,
	in string,
	arguments []any,
) error {
	rows, err := source.db.QueryContext(ctx, `
		SELECT cnpj_root, cnpj, head_office_branch_indicator, trade_name
		FROM establishments
		WHERE cnpj_root IN (`+in+`)
	`, arguments...)
	if err != nil {
		return fmt.Errorf("query indexing establishments: %w", err)
	}
	defer rows.Close()

	hasHeadOffice := make(map[string]bool)
	seenNames := make(map[string]map[string]struct{})

	for rows.Next() {
		var root, cnpj string
		var indicator, tradeName sql.NullString

		if err := rows.Scan(&root, &cnpj, &indicator, &tradeName); err != nil {
			return fmt.Errorf("scan indexing establishment: %w", err)
		}

		document := &documents[positions[root]]
		isHeadOffice := indicator.Valid && indicator.String == "1"

		if document.CNPJ == "" ||
			(isHeadOffice && !hasHeadOffice[root]) ||
			(isHeadOffice == hasHeadOffice[root] && cnpj < document.CNPJ) {
			document.CNPJ = cnpj
			hasHeadOffice[root] = isHeadOffice
		}

		if tradeName.Valid {
			document.TradeNames = appendUniqueName(
				document.TradeNames, seenNames, root, tradeName.String,
			)
		}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate indexing establishments: %w", err)
	}
	return nil
}

func (source *SearchSource) readPartners(
	ctx context.Context,
	documents []search.Document,
	positions map[string]int,
	in string,
	arguments []any,
) error {
	rows, err := source.db.QueryContext(ctx, `
		SELECT cnpj_root, partner_name
		FROM partners
		WHERE cnpj_root IN (`+in+`)
	`, arguments...)
	if err != nil {
		return fmt.Errorf("query indexing partners: %w", err)
	}
	defer rows.Close()

	seenNames := make(map[string]map[string]struct{})

	for rows.Next() {
		var root, name string
		if err := rows.Scan(&root, &name); err != nil {
			return fmt.Errorf("scan indexing partner: %w", err)
		}

		document := &documents[positions[root]]
		document.PartnerNames = appendUniqueName(
			document.PartnerNames, seenNames, root, name,
		)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate indexing partners: %w", err)
	}
	return nil
}

func appendUniqueName(
	names []string,
	seen map[string]map[string]struct{},
	root string,
	name string,
) []string {
	name = strings.TrimSpace(name)
	if name == "" {
		return names
	}
	if seen[root] == nil {
		seen[root] = make(map[string]struct{})
	}
	if _, exists := seen[root][name]; exists {
		return names
	}
	seen[root][name] = struct{}{}
	return append(names, name)
}

var _ search.Source = (*SearchSource)(nil)
