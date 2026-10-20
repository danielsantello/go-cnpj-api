package manticore

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/danielsantello/go-cnpj-api/internal/search"
)

const createTableQuery = `
	CREATE TABLE IF NOT EXISTS cnpj_companies (
		basic_cnpj string,
		cnpj string,
		corporate_name text,
		trade_names text,
		partner_names text
	) charset_table='non_cjk' rt_mem_limit='128M'
`

type Index struct {
	db *sql.DB
}

func NewIndex(db *sql.DB) *Index {
	return &Index{db: db}
}

func (index *Index) EnsureTable(ctx context.Context) error {
	if _, err := index.db.ExecContext(ctx, createTableQuery); err != nil {
		return fmt.Errorf("create Manticore search table: %w", err)
	}
	return nil
}

func (index *Index) WriteBatch(
	ctx context.Context,
	documents []search.Document,
) error {
	if len(documents) == 0 {
		return nil
	}

	values := make([]string, 0, len(documents))
	arguments := make([]any, 0, len(documents)*6)

	for _, document := range documents {
		if document.ID == 0 {
			return fmt.Errorf("Manticore document ID must not be zero")
		}

		values = append(values, "(?, ?, ?, ?, ?, ?)")
		arguments = append(arguments,
			document.ID,
			document.BasicCNPJ,
			document.CNPJ,
			document.CorporateName,
			strings.Join(document.TradeNames, "\n"),
			strings.Join(document.PartnerNames, "\n"),
		)
	}

	query := `
		REPLACE INTO cnpj_companies (
			id,
			basic_cnpj,
			cnpj,
			corporate_name,
			trade_names,
			partner_names
		) VALUES ` + strings.Join(values, ",")

	if _, err := index.db.ExecContext(ctx, query, arguments...); err != nil {
		return fmt.Errorf("write Manticore search batch: %w", err)
	}
	return nil
}

var _ search.Index = (*Index)(nil)
