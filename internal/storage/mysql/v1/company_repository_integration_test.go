package v1

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/danielsantello/go-cnpj-api/internal/config"
	"github.com/danielsantello/go-cnpj-api/internal/database"
)

func TestCompanyRepositoryFindByCNPJIntegration(t *testing.T) {
	if os.Getenv("CNPJ_API_RUN_INTEGRATION_TESTS") != "1" {
		t.Skip("integration tests are disabled")
	}

	configuration, err := config.Load()
	if err != nil {
		t.Fatalf("load configuration: %v", err)
	}

	if err := config.Validate(configuration); err != nil {
		t.Fatalf("validate configuration: %v", err)
	}

	mysqlDatabase, err := database.OpenMySQL(database.MySQLConfig{
		Host:           configuration.MySQLHost,
		Port:           configuration.MySQLPort,
		Database:       configuration.MySQLDatabase,
		User:           configuration.MySQLUser,
		Password:       configuration.MySQLPassword,
		ConnectTimeout: configuration.MySQLConnectTimeout,
	})
	if err != nil {
		t.Fatalf("open MySQL: %v", err)
	}
	defer mysqlDatabase.Close()

	ctx, cancel := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)
	defer cancel()

	var cnpj string

	err = mysqlDatabase.QueryRowContext(
		ctx,
		`
			SELECT establishment.cnpj
			FROM establishments AS establishment
			JOIN companies AS company
				ON company.cnpj_root = establishment.cnpj_root
			LIMIT 1
		`,
	).Scan(&cnpj)
	if err != nil {
		t.Fatalf("select integration test CNPJ: %v", err)
	}

	repository := NewCompanyRepository(mysqlDatabase)

	details, err := repository.FindByCNPJ(ctx, cnpj)
	if err != nil {
		t.Fatalf("find company by CNPJ: %v", err)
	}

	expectedBasicCNPJ := cnpj[:8]

	if details.Company.BasicCNPJ != expectedBasicCNPJ {
		t.Fatalf(
			"expected basic CNPJ %q, got %q",
			expectedBasicCNPJ,
			details.Company.BasicCNPJ,
		)
	}

	if details.Company.CorporateName == "" {
		t.Fatal("expected corporate name, got empty value")
	}
}
