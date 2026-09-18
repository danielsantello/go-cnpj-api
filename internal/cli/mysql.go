package cli

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/danielsantello/go-cnpj-api/internal/config"
	"github.com/danielsantello/go-cnpj-api/internal/database"

	mysqlstorage "github.com/danielsantello/go-cnpj-api/internal/storage/mysql"
)

func initializeMySQL(
	configuration config.Config,
) (*sql.DB, uint16, error) {
	mysqlDatabase, err := database.OpenMySQL(database.MySQLConfig{
		Host:           configuration.MySQLHost,
		Port:           configuration.MySQLPort,
		Database:       configuration.MySQLDatabase,
		User:           configuration.MySQLUser,
		Password:       configuration.MySQLPassword,
		ConnectTimeout: configuration.MySQLConnectTimeout,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("open MySQL: %w", err)
	}

	pingContext, cancelPing := context.WithTimeout(
		context.Background(),
		configuration.MySQLConnectTimeout,
	)

	err = database.Ping(pingContext, mysqlDatabase)
	cancelPing()

	if err != nil {
		closeMySQL(mysqlDatabase)

		return nil, 0, fmt.Errorf("verify MySQL connection: %w", err)
	}

	slog.Info(
		"MySQL connection established",
		"host", configuration.MySQLHost,
		"port", configuration.MySQLPort,
	)

	schemaContext, cancelSchema := context.WithTimeout(
		context.Background(),
		configuration.MySQLConnectTimeout,
	)

	metadata, err := mysqlstorage.ReadMetadata(
		schemaContext,
		mysqlDatabase,
	)
	if err == nil {
		err = mysqlstorage.ValidateSchema(
			schemaContext,
			mysqlDatabase,
			metadata.FormatVersion,
		)
	}

	cancelSchema()

	if err != nil {
		closeMySQL(mysqlDatabase)

		return nil, 0, fmt.Errorf("initialize database schema: %w", err)
	}

	slog.Info(
		"database schema validated",
		"format_version", metadata.FormatVersion,
		"reference_year", metadata.ReferenceYear,
		"reference_month", metadata.ReferenceMonth,
		"created_at_utc", metadata.CreatedAtUTC,
	)

	return mysqlDatabase, metadata.FormatVersion, nil
}

func closeMySQL(mysqlDatabase *sql.DB) {
	if err := mysqlDatabase.Close(); err != nil {
		slog.Error("failed to close MySQL pool", "error", err)
	}
}
