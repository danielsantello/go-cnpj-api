package manticore

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/danielsantello/go-cnpj-api/internal/config"
	"github.com/danielsantello/go-cnpj-api/internal/database"
)

func TestDropTableIntegration(t *testing.T) {
	if os.Getenv("CNPJ_API_RUN_INTEGRATION_TESTS") != "1" {
		t.Skip("integration tests are disabled")
	}

	configuration, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if !configuration.ManticoreEnabled {
		t.Fatal("Manticore must be enabled")
	}

	db, err := database.OpenManticore(database.ManticoreConfig{
		Host:           configuration.ManticoreHost,
		Port:           configuration.ManticorePort,
		ConnectTimeout: configuration.ManticoreConnectTimeout,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	table := fmt.Sprintf("cnpj_companies_drop_test_%d", time.Now().UnixNano())

	if _, err := db.ExecContext(ctx,
		"CREATE TABLE "+table+" (title text)",
	); err != nil {
		t.Fatalf("create disposable table: %v", err)
	}

	defer func() {
		cleanupContext, cleanupCancel := context.WithTimeout(
			context.Background(), 5*time.Second,
		)
		defer cleanupCancel()

		if _, err := db.ExecContext(
			cleanupContext, "DROP TABLE IF EXISTS "+table,
		); err != nil {
			t.Errorf("clean up disposable table: %v", err)
		}
	}()

	// Treat the disposable table as active to exercise the protection.
	if err := DropTable(ctx, db, table, table); err == nil {
		t.Fatal("expected active table to be blocked")
	}

	// Confirm that the blocked attempt preserved the table.
	var count int64
	if err := db.QueryRowContext(
		ctx, "SELECT COUNT(*) FROM "+table,
	).Scan(&count); err != nil {
		t.Fatalf("blocked attempt did not preserve table: %v", err)
	}

	// Remove only the disposable table, using the configured active name.
	if err := DropTable(
		ctx, db, table, configuration.ManticoreTable,
	); err != nil {
		t.Fatalf("drop disposable table: %v", err)
	}

	if err := db.QueryRowContext(
		ctx, "SELECT COUNT(*) FROM "+table,
	).Scan(&count); err == nil {
		t.Fatal("table still exists after drop")
	}

	if err := DropTable(
		ctx, db, table, configuration.ManticoreTable,
	); err == nil {
		t.Fatal("expected error when dropping a missing table")
	}

	t.Log("validated active table protection, real drop and missing table rejection")
}
