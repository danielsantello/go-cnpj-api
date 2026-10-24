package v1

import (
	"context"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/danielsantello/go-cnpj-api/internal/config"
	"github.com/danielsantello/go-cnpj-api/internal/database"
)

func TestSearchSourceReadBatchIntegration(t *testing.T) {
	if os.Getenv("CNPJ_API_RUN_INTEGRATION_TESTS") != "1" {
		t.Skip("integration tests are disabled")
	}

	configuration, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if err := config.Validate(configuration); err != nil {
		t.Fatal(err)
	}

	db, err := database.OpenMySQL(database.MySQLConfig{
		Host:           configuration.MySQLHost,
		Port:           configuration.MySQLPort,
		Database:       configuration.MySQLDatabase,
		User:           configuration.MySQLUser,
		Password:       configuration.MySQLPassword,
		ConnectTimeout: configuration.MySQLConnectTimeout,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	source := NewSearchSource(db)
	documents, err := source.ReadBatch(ctx, "", 5)
	if err != nil {
		t.Fatalf("read first batch: %v", err)
	}
	if len(documents) == 0 || len(documents) > 5 {
		t.Fatalf("unexpected batch size: %d", len(documents))
	}

	previous := ""
	for _, document := range documents {
		if document.BasicCNPJ <= previous {
			t.Fatal("batch must be ordered by basic CNPJ")
		}
		previous = document.BasicCNPJ

		var corporateName string
		err := db.QueryRowContext(ctx, `
			SELECT legal_name FROM companies WHERE cnpj_root = ?
		`, document.BasicCNPJ).Scan(&corporateName)
		if err != nil {
			t.Fatal(err)
		}
		if document.CorporateName != corporateName || document.ID != 0 {
			t.Fatalf("unexpected company document: %#v", document)
		}

		var expectedCNPJ string
		err = db.QueryRowContext(ctx, `
			SELECT COALESCE((
				SELECT cnpj
				FROM establishments
				WHERE cnpj_root = ?
				ORDER BY head_office_branch_indicator = '1' DESC, cnpj
				LIMIT 1
			), '')
		`, document.BasicCNPJ).Scan(&expectedCNPJ)
		if err != nil {
			t.Fatal(err)
		}
		if document.CNPJ != expectedCNPJ {
			t.Fatalf("expected CNPJ %q, got %q", expectedCNPJ, document.CNPJ)
		}

		for _, check := range []struct {
			query string
			names []string
		}{
			{
				`SELECT trade_name FROM establishments
				 WHERE cnpj_root = ? AND trade_name IS NOT NULL`,
				document.TradeNames,
			},
			{
				`SELECT partner_name FROM partners WHERE cnpj_root = ?`,
				document.PartnerNames,
			},
		} {
			rows, err := db.QueryContext(ctx, check.query, document.BasicCNPJ)
			if err != nil {
				t.Fatal(err)
			}

			expected := make(map[string]bool)
			for rows.Next() {
				var name string
				if err := rows.Scan(&name); err != nil {
					rows.Close()
					t.Fatal(err)
				}
				// Normalize whitespace independently of the source helper.
				name = strings.TrimSpace(name)
				if name != "" {
					expected[name] = true
				}
			}
			readErr := rows.Err()
			rows.Close()
			if readErr != nil {
				t.Fatal(readErr)
			}

			actual := make(map[string]bool)
			for _, name := range check.names {
				if actual[name] {
					t.Fatalf("duplicate indexed name: %q", name)
				}
				actual[name] = true
			}
			if !reflect.DeepEqual(expected, actual) {
				t.Fatalf("names differ: expected %v, got %v", expected, actual)
			}
		}
	}

	next, err := source.ReadBatch(ctx, previous, 5)
	if err != nil {
		t.Fatalf("read next batch: %v", err)
	}
	for _, document := range next {
		if document.BasicCNPJ <= previous {
			t.Fatal("next batch repeated or moved behind the cursor")
		}
		previous = document.BasicCNPJ
	}

	t.Logf("validated %d companies and %d in the next batch", len(documents), len(next))
}

func TestSearchSourceRejectsDuplicateAtBatchBoundaryIntegration(t *testing.T) {
	if os.Getenv("CNPJ_API_RUN_INTEGRATION_TESTS") != "1" {
		t.Skip("integration tests are disabled")
	}
	configuration, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	db, err := database.OpenMySQL(database.MySQLConfig{
		Host:           configuration.MySQLHost,
		Port:           configuration.MySQLPort,
		Database:       configuration.MySQLDatabase,
		User:           configuration.MySQLUser,
		Password:       configuration.MySQLPassword,
		ConnectTimeout: configuration.MySQLConnectTimeout,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	// The temporary table shadows companies only on this private connection.
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if _, err := db.ExecContext(ctx, `CREATE TEMPORARY TABLE companies (
		cnpj_root VARCHAR(8) NOT NULL,
		legal_name VARCHAR(255) NOT NULL,
		INDEX idx_companies_cnpj_root (cnpj_root)
	)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO companies (cnpj_root, legal_name)
		VALUES ('12ABC345', 'PRIMEIRA'), ('12ABC345', 'SEGUNDA')`); err != nil {
		t.Fatal(err)
	}
	_, err = NewSearchSource(db).ReadBatch(ctx, "", 1)
	if err == nil || !strings.Contains(err.Error(), "repeated company basic CNPJ") {
		t.Fatalf("duplicate crossing the batch boundary must stop indexing: %v", err)
	}
}
