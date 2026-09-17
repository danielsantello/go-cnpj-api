package cli

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/danielsantello/go-cnpj-api/internal/config"
	"github.com/danielsantello/go-cnpj-api/internal/database"
	"github.com/danielsantello/go-cnpj-api/internal/httpapi"

	mysqlstorage "github.com/danielsantello/go-cnpj-api/internal/storage/mysql"
)

func serve() error {
	configuration, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	if err := config.Validate(configuration); err != nil {
		return fmt.Errorf("validate configuration: %w", err)
	}

	mysqlDatabase, err := database.OpenMySQL(database.MySQLConfig{
		Host:           configuration.MySQLHost,
		Port:           configuration.MySQLPort,
		Database:       configuration.MySQLDatabase,
		User:           configuration.MySQLUser,
		Password:       configuration.MySQLPassword,
		ConnectTimeout: configuration.MySQLConnectTimeout,
	})
	if err != nil {
		return fmt.Errorf("open MySQL: %w", err)
	}

	defer func() {
		if err := mysqlDatabase.Close(); err != nil {
			slog.Error("failed to close MySQL pool", "error", err)
		}
	}()

	pingContext, cancelPing := context.WithTimeout(
		context.Background(),
		configuration.MySQLConnectTimeout,
	)

	err = database.Ping(pingContext, mysqlDatabase)
	cancelPing()

	if err != nil {
		return fmt.Errorf("verify MySQL connection: %w", err)
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
		return fmt.Errorf("initialize database schema: %w", err)
	}

	slog.Info(
		"database schema validated",
		"format_version", metadata.FormatVersion,
		"reference_year", metadata.ReferenceYear,
		"reference_month", metadata.ReferenceMonth,
		"created_at_utc", metadata.CreatedAtUTC,
	)

	server := httpapi.NewServer(
		configuration.HTTPAddress,
		mysqlDatabase,
		configuration.MySQLConnectTimeout,
	)

	// signalContext is canceled when the process receives SIGINT or SIGTERM.
	signalContext, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	serverError := make(chan error, 1)

	slog.Info("HTTP server listening", "address", server.Addr)

	go func() {
		serverError <- server.ListenAndServe()
	}()

	select {
	case err := <-serverError:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("start HTTP server: %w", err)
		}

		return nil

	case <-signalContext.Done():
		stop()
		slog.Info("shutting down HTTP server")
	}

	shutdownContext, cancel := context.WithTimeout(
		context.Background(),
		configuration.ShutdownTimeout,
	)
	defer cancel()

	if err := server.Shutdown(shutdownContext); err != nil {
		return fmt.Errorf("shutdown HTTP server: %w", err)
	}

	if err := <-serverError; err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("stop HTTP server: %w", err)
	}

	slog.Info("HTTP server stopped")

	return nil
}
