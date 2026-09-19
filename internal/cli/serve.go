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

	"github.com/danielsantello/go-cnpj-api/internal/company"
	"github.com/danielsantello/go-cnpj-api/internal/config"
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

	mysqlDatabase, formatVersion, err := initializeMySQL(configuration)
	if err != nil {
		return err
	}
	defer closeMySQL(mysqlDatabase)

	companyRepository, err := mysqlstorage.NewCompanyRepository(
		mysqlDatabase,
		formatVersion,
	)
	if err != nil {
		return fmt.Errorf("initialize company repository: %w", err)
	}

	companyService := company.NewService(companyRepository)

	server := httpapi.NewServer(
		configuration.HTTPAddress,
		mysqlDatabase,
		configuration.MySQLConnectTimeout,
		companyService,
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
