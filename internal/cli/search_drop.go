package cli

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/danielsantello/go-cnpj-api/internal/config"
	manticorestorage "github.com/danielsantello/go-cnpj-api/internal/storage/manticore"
)

func runSearchDrop(args []string) error {
	flags := flag.NewFlagSet("search drop", flag.ContinueOnError)
	table := flags.String("table", "", "Manticore table to permanently remove")

	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("search drop does not accept positional arguments")
	}
	if *table == "" {
		return fmt.Errorf("table is required")
	}

	configuration, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}
	if err := config.Validate(configuration); err != nil {
		return fmt.Errorf("validate configuration: %w", err)
	}
	if !configuration.ManticoreEnabled {
		return fmt.Errorf("Manticore must be enabled to drop a search table")
	}

	if err := manticorestorage.ValidateDropTable(
		*table,
		configuration.ManticoreTable,
	); err != nil {
		return err
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

	slog.Info(
		"dropping Manticore search table",
		"table", *table,
		"active_table", configuration.ManticoreTable,
	)

	if err := manticorestorage.DropTable(
		ctx,
		manticore,
		*table,
		configuration.ManticoreTable,
	); err != nil {
		return err
	}

	slog.Info("Manticore search table dropped", "table", *table)
	return nil
}
