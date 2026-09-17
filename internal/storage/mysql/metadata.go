package mysqlstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Metadata struct {
	FormatVersion  uint16
	ReferenceYear  uint16
	ReferenceMonth uint8
	CreatedAtUTC   time.Time
}

func ReadMetadata(ctx context.Context, database *sql.DB) (Metadata, error) {
	const query = `
		SELECT
			format_version,
			reference_year,
			reference_month,
			created_at_utc,
			(
				SELECT COUNT(*)
				FROM schema_metadata
			) AS row_count
		FROM schema_metadata
		WHERE id = 1
	`

	var metadata Metadata

	var rowCount uint64

	err := database.QueryRowContext(ctx, query).Scan(
		&metadata.FormatVersion,
		&metadata.ReferenceYear,
		&metadata.ReferenceMonth,
		&metadata.CreatedAtUTC,
		&rowCount,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Metadata{}, fmt.Errorf("schema metadata row with id 1 was not found")
	}
	if err != nil {
		return Metadata{}, fmt.Errorf("read schema metadata: %w", err)
	}

	if rowCount != 1 {
		return Metadata{}, fmt.Errorf(
			"schema_metadata must contain exactly one row, found %d",
			rowCount,
		)
	}

	return metadata, nil
}
