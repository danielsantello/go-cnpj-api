package mysqlstorage

import (
	"context"
	"database/sql"
	"fmt"

	mysqlv1 "github.com/danielsantello/go-cnpj-api/internal/storage/mysql/v1"
)

type schemaValidator func(context.Context, *sql.DB) error

func ValidateSchema(
	ctx context.Context,
	database *sql.DB,
	formatVersion uint16,
) error {
	validator, err := schemaValidatorForFormat(formatVersion)
	if err != nil {
		return err
	}

	if err := validator(ctx, database); err != nil {
		return fmt.Errorf("validate database schema: %w", err)
	}

	return nil
}

func schemaValidatorForFormat(
	formatVersion uint16,
) (schemaValidator, error) {
	switch formatVersion {
	case mysqlv1.FormatVersion:
		return mysqlv1.ValidateSchema, nil
	default:
		return nil, fmt.Errorf(
			"unsupported database format version: %d",
			formatVersion,
		)
	}
}
