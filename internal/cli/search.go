package cli

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"regexp"
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
		if args[0] == "drop" {
			return runSearchDrop(args[1:])
		}
		if args[0] != "index" {
			return fmt.Errorf("unknown search command: %s", args[0])
		}
	}

	flags := flag.NewFlagSet("search index", flag.ContinueOnError)
	limit := flags.Int("limit", 1000, "maximum number of companies to index")
	batchSize := flags.Int("batch-size", 100, "companies per batch")
	all := flags.Bool("all", false, "index all companies")
	table := flags.String("table", "", "destination table, such as cnpj_companies_2026_08")
	after := flags.String("after", "", "resume after this 8-digit basic CNPJ")

	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("search index does not accept positional arguments")
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
		if *table == "" {
			return fmt.Errorf("full indexing requires an explicit destination table")
		}
		*limit = 0
	} else if *limit <= 0 {
		return fmt.Errorf("limit must be greater than zero")
	}
	if *batchSize <= 0 || *batchSize > 1000 {
		return fmt.Errorf("batch-size must be between 1 and 1000")
	}

	if *after != "" {
		if !regexp.MustCompile(`^[0-9]{8}$`).MatchString(*after) {
			return fmt.Errorf("after must contain exactly 8 digits")
		}
		if *table == "" {
			return fmt.Errorf("resuming requires an explicit destination table")
		}
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

	destination := manticorestorage.NewIndex(manticore)
	if *table != "" {
		destination, err = manticorestorage.NewIndexForTable(manticore, *table)
		if err != nil {
			return err
		}
	}

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
		"all", *all,
		"destination_table", *table,
		"after_basic_cnpj", *after,
	)

	result, err := search.BuildIndex(
		ctx,
		mysqlv1.NewSearchSource(mysqlDatabase),
		destination,
		search.IndexOptions{
			BatchSize:      *batchSize,
			Limit:          *limit,
			All:            *all,
			AfterBasicCNPJ: *after,
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
