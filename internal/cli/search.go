package cli

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/danielsantello/go-cnpj-api/internal/config"
	"github.com/danielsantello/go-cnpj-api/internal/search"
	manticorestorage "github.com/danielsantello/go-cnpj-api/internal/storage/manticore"
	mysqlv1 "github.com/danielsantello/go-cnpj-api/internal/storage/mysql/v1"
)

func runSearch(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("search command is required")
	}
	if args[0] != "index" {
		return fmt.Errorf("unknown search command: %s", args[0])
	}

	flags := flag.NewFlagSet("search index", flag.ContinueOnError)
	limit := flags.Int("limit", 1000, "maximum number of companies to index")
	batchSize := flags.Int("batch-size", 100, "companies per batch")

	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("search index does not accept positional arguments")
	}
	if *limit <= 0 {
		return fmt.Errorf("limit must be greater than zero")
	}
	if *batchSize <= 0 || *batchSize > 1000 {
		return fmt.Errorf("batch-size must be between 1 and 1000")
	}

	configuration, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}
	if err := config.Validate(configuration); err != nil {
		return fmt.Errorf("validate configuration: %w", err)
	}
	if !configuration.ManticoreEnabled {
		return fmt.Errorf("Manticore must be enabled to index companies")
	}

	mysqlDatabase, formatVersion, err := initializeMySQL(configuration)
	if err != nil {
		return err
	}
	defer closeMySQL(mysqlDatabase)

	if formatVersion != 1 {
		return fmt.Errorf("unsupported indexing format version: %d", formatVersion)
	}

	manticore, err := initializeManticore(configuration)
	if err != nil {
		return err
	}
	defer closeManticore(manticore)

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	started := time.Now()

	slog.Info(
		"indexing companies",
		"limit", *limit,
		"batch_size", *batchSize,
	)

	result, err := search.BuildIndex(
		ctx,
		mysqlv1.NewSearchSource(mysqlDatabase),
		manticorestorage.NewIndex(manticore),
		search.IndexOptions{
			BatchSize: *batchSize,
			Limit:     *limit,
		},
		func(progress search.IndexProgress) {
			slog.Info(
				"search indexing progress",
				"processed", progress.Processed,
				"last_basic_cnpj", progress.LastBasicCNPJ,
				"elapsed", time.Since(started),
			)
		},
	)
	if err != nil {
		return fmt.Errorf(
			"index companies after %d documents (last basic CNPJ %q): %w",
			result.Processed, result.LastBasicCNPJ, err,
		)
	}

	slog.Info(
		"search indexing completed",
		"processed", result.Processed,
		"last_basic_cnpj", result.LastBasicCNPJ,
		"elapsed", time.Since(started),
	)
	return nil
}
