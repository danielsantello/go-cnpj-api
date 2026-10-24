package manticore

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"github.com/danielsantello/go-cnpj-api/internal/search"
)

var partnerTableNamePattern = regexp.MustCompile(
	`^cnpj_partners(_[a-z0-9_]{1,48})?$`,
)

type PartnerIndex struct {
	db    *sql.DB
	table string
	ready bool
}

func NewPartnerIndex(
	db *sql.DB,
	table string,
) (*PartnerIndex, error) {
	if !partnerTableNamePattern.MatchString(table) {
		return nil, fmt.Errorf("invalid partner index table name: %q", table)
	}
	return &PartnerIndex{db: db, table: table}, nil
}

func (index *PartnerIndex) RecreateTable(ctx context.Context) error {
	index.ready = false
	if _, err := index.db.ExecContext(ctx, "DROP TABLE IF EXISTS "+index.table); err != nil {
		return fmt.Errorf("clear index destination %q: %w", index.table, err)
	}

	_, err := index.db.ExecContext(ctx, `
		CREATE TABLE `+index.table+` (
			basic_cnpj string,
			cnpj string,
			corporate_name string,
			partner_name text,
			partner_document string,
			has_document integer
		) charset_table='non_cjk' rt_mem_limit='128M'
	`)
	if err != nil {
		return fmt.Errorf("create partner search table: %w", err)
	}

	index.ready = true
	return nil
}

func (index *PartnerIndex) WriteBatch(
	ctx context.Context,
	documents []search.PartnerDocument,
) error {
	if len(documents) == 0 {
		return nil
	}
	if !index.ready {
		return fmt.Errorf("index destination must pass RecreateTable before writing")
	}

	values := make([]string, 0, len(documents))
	arguments := make([]any, 0, len(documents)*7)
	seenIDs := make(map[uint64]bool, len(documents))

	for _, document := range documents {
		if document.ID == 0 {
			return fmt.Errorf("partner document ID must not be zero")
		}
		if seenIDs[document.ID] {
			return fmt.Errorf("duplicate partner document ID in batch: %d",
				document.ID)
		}
		seenIDs[document.ID] = true

		documentValue := ""
		hasDocument := 0
		if document.Document != nil {
			documentValue = *document.Document
			hasDocument = 1
		}

		values = append(values, "(?, ?, ?, ?, ?, ?, ?)")
		arguments = append(arguments,
			document.ID,
			document.BasicCNPJ,
			document.CNPJ,
			document.CorporateName,
			document.Name,
			documentValue,
			hasDocument,
		)
	}

	_, err := index.db.ExecContext(ctx, `
		INSERT INTO `+index.table+` (
			id,
			basic_cnpj,
			cnpj,
			corporate_name,
			partner_name,
			partner_document,
			has_document
		) VALUES `+strings.Join(values, ","),
		arguments...,
	)
	if err != nil {
		return fmt.Errorf("write partner search batch: %w", err)
	}
	return nil
}

var _ search.PartnerIndex = (*PartnerIndex)(nil)
