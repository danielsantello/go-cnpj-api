package manticore

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"github.com/danielsantello/go-cnpj-api/internal/search"
)

var establishmentTableNamePattern = regexp.MustCompile(
	`^cnpj_establishments(_[a-z0-9_]{1,48})?$`,
)

type EstablishmentIndex struct {
	db    *sql.DB
	table string
	ready bool
}

func NewEstablishmentIndex(
	db *sql.DB,
	table string,
) (*EstablishmentIndex, error) {
	if !establishmentTableNamePattern.MatchString(table) {
		return nil, fmt.Errorf(
			"invalid establishment index table name: %q", table,
		)
	}
	return &EstablishmentIndex{db: db, table: table}, nil
}

func (index *EstablishmentIndex) RecreateTable(ctx context.Context) error {
	index.ready = false
	if _, err := index.db.ExecContext(ctx, "DROP TABLE IF EXISTS "+index.table); err != nil {
		return fmt.Errorf("clear index destination %q: %w", index.table, err)
	}

	_, err := index.db.ExecContext(ctx, `
		CREATE TABLE `+index.table+` (
			cnpj string,
			corporate_name string,
			trade_name text
		) charset_table='non_cjk' rt_mem_limit='128M'
	`)
	if err != nil {
		return fmt.Errorf("create establishment search table: %w", err)
	}

	index.ready = true
	return nil
}

func (index *EstablishmentIndex) WriteBatch(
	ctx context.Context,
	documents []search.EstablishmentDocument,
) error {
	if len(documents) == 0 {
		return nil
	}
	if !index.ready {
		return fmt.Errorf("index destination must pass RecreateTable before writing")
	}

	values := make([]string, 0, len(documents))
	arguments := make([]any, 0, len(documents)*4)
	seenIDs := make(map[uint64]bool, len(documents))
	seenCNPJs := make(map[string]bool, len(documents))

	for _, document := range documents {
		if document.ID == 0 {
			return fmt.Errorf("establishment document ID must not be zero")
		}
		if seenIDs[document.ID] {
			return fmt.Errorf(
				"duplicate establishment document ID in batch: %d",
				document.ID,
			)
		}
		if seenCNPJs[document.CNPJ] {
			return fmt.Errorf(
				"duplicate establishment CNPJ in index batch: %q",
				document.CNPJ,
			)
		}
		seenIDs[document.ID] = true
		seenCNPJs[document.CNPJ] = true

		values = append(values, "(?, ?, ?, ?)")
		arguments = append(arguments,
			document.ID,
			document.CNPJ,
			document.CorporateName,
			document.TradeName,
		)
	}

	_, err := index.db.ExecContext(ctx, `
		INSERT INTO `+index.table+` (
			id, cnpj, corporate_name, trade_name
		) VALUES `+strings.Join(values, ","),
		arguments...,
	)
	if err != nil {
		return fmt.Errorf("write establishment search batch: %w", err)
	}
	return nil
}

var _ search.EstablishmentIndex = (*EstablishmentIndex)(nil)
