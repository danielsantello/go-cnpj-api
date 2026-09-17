package v1

import (
	"context"
	"database/sql"
	"fmt"
)

type availableSchema map[string]map[string]struct{}

func ValidateSchema(ctx context.Context, database *sql.DB) error {
	available, err := readAvailableSchema(ctx, database)
	if err != nil {
		return err
	}

	if err := validateAvailableSchema(available); err != nil {
		return fmt.Errorf("validate MySQL format version 1: %w", err)
	}

	return nil
}

func readAvailableSchema(
	ctx context.Context,
	database *sql.DB,
) (availableSchema, error) {
	const query = `
		SELECT
			TABLE_NAME,
			COLUMN_NAME
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE()
	`

	rows, err := database.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("read database structure: %w", err)
	}
	defer rows.Close()

	available := make(availableSchema)

	for rows.Next() {
		var tableName string
		var columnName string

		if err := rows.Scan(&tableName, &columnName); err != nil {
			return nil, fmt.Errorf("scan database structure: %w", err)
		}

		if _, exists := available[tableName]; !exists {
			available[tableName] = make(map[string]struct{})
		}

		available[tableName][columnName] = struct{}{}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate database structure: %w", err)
	}

	return available, nil
}

func validateAvailableSchema(available availableSchema) error {
	for _, table := range schemaContract {
		availableColumns, exists := available[table.name]
		if !exists {
			return fmt.Errorf("required table %q is missing", table.name)
		}

		for _, column := range table.columns {
			if _, exists := availableColumns[column]; !exists {
				return fmt.Errorf(
					"required column %q is missing from table %q",
					column,
					table.name,
				)
			}
		}
	}

	return nil
}
