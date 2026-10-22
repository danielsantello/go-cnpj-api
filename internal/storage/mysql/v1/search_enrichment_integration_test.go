package v1

import (
	"context"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/danielsantello/go-cnpj-api/internal/company"
	"github.com/danielsantello/go-cnpj-api/internal/config"
	"github.com/danielsantello/go-cnpj-api/internal/database"
	"github.com/danielsantello/go-cnpj-api/internal/search"
)

func TestFindEstablishmentSummariesIntegration(t *testing.T) {
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

	rows, err := db.QueryContext(ctx, `
		SELECT cnpj
		FROM establishments
		WHERE trade_name LIKE 'SANTELLO%'
		ORDER BY cnpj
		LIMIT 5
	`)
	if err != nil {
		t.Fatal(err)
	}

	var cnpjs []string
	for rows.Next() {
		var cnpj string
		if err := rows.Scan(&cnpj); err != nil {
			rows.Close()
			t.Fatal(err)
		}
		cnpjs = append(cnpjs, cnpj)
	}
	readErr := rows.Err()
	closeErr := rows.Close()
	if readErr != nil {
		t.Fatal(readErr)
	}
	if closeErr != nil {
		t.Fatal(closeErr)
	}
	if len(cnpjs) == 0 {
		t.Fatal("expected establishments in the integration sample")
	}

	repository := NewCompanyRepository(db)

	// Reverse the input and repeat one CNPJ.
	requested := make([]string, 0, len(cnpjs)+3)
	for index := len(cnpjs) - 1; index >= 0; index-- {
		requested = append(requested, cnpjs[index])
	}

	// This value cannot be a full CNPJ.
	const missingCNPJ = "missing-integration-cnpj"
	requested = append(requested, cnpjs[0], "", missingCNPJ)

	summaries, err := repository.FindEstablishmentSummaries(ctx, requested)
	if err != nil {
		t.Fatal(err)
	}

	expected := make(map[string]search.EstablishmentSummary)
	for _, cnpj := range cnpjs {
		details, err := repository.FindByCNPJ(
			ctx, cnpj, company.FindOptions{},
		)
		if err != nil {
			t.Fatal(err)
		}

		expected[cnpj] = search.EstablishmentSummary{
			Municipality:       details.Establishment.Address.Municipality,
			State:              details.Establishment.Address.State,
			RegistrationStatus: details.Establishment.Registration.Status,
		}
	}

	if !reflect.DeepEqual(summaries, expected) {
		t.Fatalf("batch summaries differ from CNPJ details: got %#v, expected %#v",
			summaries, expected)
	}

	for _, input := range [][]string{nil, {}, {"", ""}, {missingCNPJ}} {
		result, err := repository.FindEstablishmentSummaries(ctx, input)
		if err != nil {
			t.Fatal(err)
		}
		if result == nil || len(result) != 0 {
			t.Fatalf("expected an empty map for %q, got %#v", input, result)
		}
	}

	t.Logf("validated %d establishment summaries against full CNPJ details",
		len(expected))
}

func TestFindEstablishmentSummariesSkipsEmptyInput(t *testing.T) {
	// A nil database makes an accidental query fail this test.
	repository := NewCompanyRepository(nil)

	for _, input := range [][]string{nil, {}, {"", ""}} {
		result, err := repository.FindEstablishmentSummaries(
			context.Background(), input,
		)
		if err != nil {
			t.Fatal(err)
		}
		if result == nil || len(result) != 0 {
			t.Fatalf("expected an empty map, got %#v", result)
		}
	}
}
