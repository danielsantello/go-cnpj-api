package cli

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/danielsantello/go-cnpj-api/internal/config"
	"github.com/danielsantello/go-cnpj-api/internal/database"
)

func initializeManticore(
	configuration config.Config,
) (*sql.DB, error) {
	if !configuration.ManticoreEnabled {
		return nil, nil
	}

	manticore, err := database.OpenManticore(database.ManticoreConfig{
		Host:           configuration.ManticoreHost,
		Port:           configuration.ManticorePort,
		ConnectTimeout: configuration.ManticoreConnectTimeout,
	})
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		configuration.ManticoreConnectTimeout,
	)
	defer cancel()

	if err := manticore.PingContext(ctx); err != nil {
		closeManticore(manticore)
		return nil, fmt.Errorf("verify Manticore connection: %w", err)
	}

	slog.Info(
		"Manticore connection established",
		"host", configuration.ManticoreHost,
		"port", configuration.ManticorePort,
	)

	return manticore, nil
}

func closeManticore(manticore *sql.DB) {
	if manticore == nil {
		return
	}

	if err := manticore.Close(); err != nil {
		slog.Error("failed to close Manticore pool", "error", err)
	}
}
