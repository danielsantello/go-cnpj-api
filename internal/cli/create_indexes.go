package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/danielsantello/go-cnpj-api/internal/config"

	mysqlstorage "github.com/danielsantello/go-cnpj-api/internal/storage/mysql"
)

func createIndexes() error {
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

	operationContext, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	slog.Info(
		"creating required database indexes",
		"format_version", formatVersion,
	)

	result, err := mysqlstorage.CreateRequiredIndexes(
		operationContext,
		mysqlDatabase,
		formatVersion,
	)
	if err != nil {
		return fmt.Errorf("create required database indexes: %w", err)
	}

	slog.Info(
		"database indexes ready",
		"created", len(result.Created),
		"existing", len(result.Existing),
	)

	return nil
}
