package manticore

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"github.com/danielsantello/go-cnpj-api/internal/search"
)

const createTableQuery = `
	CREATE TABLE %s (
		basic_cnpj string,
		cnpj string,
		corporate_name text,
		trade_names text,
		partner_names text
	) charset_table='non_cjk' rt_mem_limit='128M'
`

type Index struct {
	db    *sql.DB
	table string
	ready bool
}

func NewIndex(db *sql.DB) *Index {
	return &Index{db: db, table: "cnpj_companies"}
}

var indexTableNamePattern = regexp.MustCompile(
	`^cnpj_companies_[a-z0-9_]{1,48}$`,
)

func NewIndexForTable(db *sql.DB, table string) (*Index, error) {
	if !indexTableNamePattern.MatchString(table) {
		return nil, fmt.Errorf(
			"invalid index table name: use cnpj_companies_ followed by 1 to 48 lowercase letters, digits or underscores",
		)
	}

	return &Index{db: db, table: table}, nil
}

func (index *Index) RecreateTable(ctx context.Context) error {
	index.ready = false
	if _, err := index.db.ExecContext(ctx, "DROP TABLE IF EXISTS "+index.table); err != nil {
		return fmt.Errorf("clear index destination %q: %w", index.table, err)
	}

	if _, err := index.db.ExecContext(
		ctx, fmt.Sprintf(createTableQuery, index.table),
	); err != nil {
		return fmt.Errorf("create Manticore search table: %w", err)
	}

	index.ready = true
	return nil
}

func (index *Index) WriteBatch(
	ctx context.Context,
	documents []search.Document,
) error {
	if len(documents) == 0 {
		return nil
	}
	if !index.ready {
		return fmt.Errorf("index destination must pass RecreateTable before writing")
	}

	values := make([]string, 0, len(documents))
	arguments := make([]any, 0, len(documents)*6)
	seenIDs := make(map[uint64]bool, len(documents))

	for _, document := range documents {
		if document.ID == 0 {
			return fmt.Errorf("Manticore document ID must not be zero")
		}
		if seenIDs[document.ID] {
			return fmt.Errorf("duplicate Manticore document ID in batch: %d", document.ID)
		}
		seenIDs[document.ID] = true

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
		INSERT INTO ` + index.table + ` (
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
