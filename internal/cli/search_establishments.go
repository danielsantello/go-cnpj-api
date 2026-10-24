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

func runEstablishmentIndex(args []string) error {
	flags := flag.NewFlagSet("search index-establishments", flag.ContinueOnError)
	limit := flags.Int("limit", 1000, "maximum number of establishments to index")
	batchSize := flags.Int("batch-size", 100, "establishments per batch")
	all := flags.Bool("all", false, "index all establishments with a non-null trade name")
	table := flags.String("table", "", "explicit destination establishment table")

	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("search index-establishments does not accept positional arguments")
	}

	limitProvided := false
	flags.Visit(func(option *flag.Flag) {
		if option.Name == "limit" {
			limitProvided = true
		}
	})

	if *all {
		if limitProvided {
			return fmt.Errorf("all and limit must not be used together")
		}
		*limit = 0
	} else if *limit <= 0 {
		return fmt.Errorf("limit must be greater than zero")
	}
	if *batchSize <= 0 || *batchSize > 1000 {
		return fmt.Errorf("batch-size must be between 1 and 1000")
	}
	if *table == "" {
		return fmt.Errorf("indexing requires an explicit destination table")
	}

	// Validate the identifier before opening database connections.
	if _, err := manticorestorage.NewEstablishmentIndex(nil, *table); err != nil {
		return err
	}

	configuration, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}
	if err := config.Validate(configuration); err != nil {
		return fmt.Errorf("validate configuration: %w", err)
	}
	if configuration.IsActiveSearchTable(*table) {
		return fmt.Errorf(
			"refusing to index into the configured active table %q",
			*table,
		)
	}
	if !configuration.ManticoreEnabled {
		return fmt.Errorf("Manticore must be enabled to index establishments")
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

	destination, err := manticorestorage.NewEstablishmentIndex(manticore, *table)
	if err != nil {
		return err
	}
	source := mysqlv1.NewEstablishmentSource(
		mysqlv1.NewCompanyRepository(mysqlDatabase),
	)

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	started := time.Now()
	slog.Info(
		"indexing establishments",
		"limit", *limit,
		"batch_size", *batchSize,
		"all", *all,
		"destination_table", *table,
		"recreate_destination", true,
	)

	result, err := search.BuildEstablishmentIndex(
		ctx,
		source,
		destination,
		search.EstablishmentIndexOptions{
			BatchSize: *batchSize,
			Limit:     *limit,
			All:       *all,
		},
		func(progress search.EstablishmentIndexProgress) {
			slog.Info(
				"establishment indexing progress",
				"processed", progress.Processed,
				"elapsed", time.Since(started),
			)
		},
	)
	if err != nil {
		return fmt.Errorf("index establishments after %d documents; rerun to reload from zero: %w", result.Processed, err)
	}

	slog.Info(
		"establishment indexing completed",
		"processed", result.Processed,
		"elapsed", time.Since(started),
	)
	return nil
}
