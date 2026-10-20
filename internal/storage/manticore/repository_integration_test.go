package manticore

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/danielsantello/go-cnpj-api/internal/config"
	"github.com/danielsantello/go-cnpj-api/internal/database"
	"github.com/danielsantello/go-cnpj-api/internal/search"
)

func TestRepositorySearchResultLimitIntegration(t *testing.T) {
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

	table := fmt.Sprintf("cnpj_limit_test_%d", time.Now().UnixNano())

	_, err = db.ExecContext(ctx, `
		CREATE TABLE `+table+` (
			basic_cnpj string,
			cnpj string,
			corporate_name text
		)
	`)
	if err != nil {
		t.Fatalf("create test table: %v", err)
	}
	defer func() {
		cleanupContext, cleanupCancel := context.WithTimeout(
			context.Background(), 5*time.Second,
		)
		defer cleanupCancel()

		if _, err := db.ExecContext(
			cleanupContext, "DROP TABLE "+table,
		); err != nil {
			t.Errorf("drop test table %s: %v", table, err)
		}
	}()

	for first := 1; first <= 1000; first += 100 {
		values := make([]string, 0, 100)
		arguments := make([]any, 0, 400)

		for id := first; id < first+100; id++ {
			values = append(values, "(?, ?, ?, ?)")
			arguments = append(arguments,
				id,
				fmt.Sprintf("%08d", id),
				fmt.Sprintf("%08d000100", id),
				"EMPRESA LIMITETESTE",
			)
		}

		_, err := db.ExecContext(ctx,
			"INSERT INTO "+table+
				" (id, basic_cnpj, cnpj, corporate_name) VALUES "+
				strings.Join(values, ","),
			arguments...,
		)
		if err != nil {
			t.Fatalf("insert test documents: %v", err)
		}
	}

	repository := &Repository{db: db, table: table}
	service := search.NewService(repository)

	result, err := service.Search(ctx, search.Options{
		Name: "limiteteste", Page: 10, PageSize: 100,
	})
	if err != nil {
		t.Fatalf("search exactly 1000 matches: %v", err)
	}
	if result.Pagination.TotalItems != 1000 ||
		result.Pagination.TotalPages != 10 ||
		len(result.Items) != 100 {
		t.Fatalf("unexpected pagination: %#v; items=%d",
			result.Pagination, len(result.Items))
	}
	if result.Items[0].BasicCNPJ != "00000901" ||
		result.Items[99].BasicCNPJ != "00001000" {
		t.Fatal("last page returned unexpected documents")
	}

	result, err = service.Search(ctx, search.Options{
		Name: "limiteteste", Page: 11, PageSize: 100,
	})
	if err != nil || len(result.Items) != 0 ||
		result.Pagination.TotalItems != 1000 {
		t.Fatalf("unexpected page beyond results: %#v, %v", result, err)
	}

	_, err = db.ExecContext(ctx,
		"INSERT INTO "+table+
			" (id, basic_cnpj, cnpj, corporate_name) VALUES (?, ?, ?, ?)",
		1001, "00001001", "00001001000100", "EMPRESA LIMITETESTE",
	)
	if err != nil {
		t.Fatalf("insert match 1001: %v", err)
	}

	_, err = service.Search(ctx, search.Options{
		Name: "limiteteste", Page: 1, PageSize: 20,
	})
	if !errors.Is(err, search.ErrTooBroad) {
		t.Fatalf("expected ErrTooBroad for 1001 matches, got %v", err)
	}

	t.Log("validated 1000 matches, last page, empty page and rejection of 1001 matches")
}
