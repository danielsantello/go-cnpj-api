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
	var secondaryEconomicActivities string

	err = mysqlDatabase.QueryRowContext(
		ctx,
		`
			SELECT
				establishment.cnpj,
				establishment.secondary_economic_activities
			FROM establishments AS establishment
			JOIN companies AS company
				ON company.cnpj_root = establishment.cnpj_root
			JOIN simple_tax_options AS simple_tax
				ON simple_tax.cnpj_root = establishment.cnpj_root
			WHERE establishment.secondary_economic_activities IS NOT NULL
				AND establishment.secondary_economic_activities <> ''
			LIMIT 1
		`,
	).Scan(
		&cnpj,
		&secondaryEconomicActivities,
	)
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

	if details.SimpleTax == nil {
		t.Fatal("expected simple tax information, got nil")
	}

	simpleTaxDates := []*string{
		details.SimpleTax.OptionDate,
		details.SimpleTax.ExclusionDate,
		details.SimpleTax.MEI.OptionDate,
		details.SimpleTax.MEI.ExclusionDate,
	}

	for _, date := range simpleTaxDates {
		if date == nil {
			continue
		}

		if _, err := time.Parse(time.DateOnly, *date); err != nil {
			t.Fatalf(
				"expected simple tax date in YYYY-MM-DD format, got %q",
				*date,
			)
		}
	}

	if details.Establishment.CNPJ != cnpj {
		t.Fatalf(
			"expected establishment CNPJ %q, got %q",
			cnpj,
			details.Establishment.CNPJ,
		)
	}

	if details.Establishment.Registration.StatusDate != nil {
		_, err := time.Parse(
			time.DateOnly,
			*details.Establishment.Registration.StatusDate,
		)
		if err != nil {
			t.Fatalf(
				"expected registration status date in YYYY-MM-DD format, got %q",
				*details.Establishment.Registration.StatusDate,
			)
		}
	}

	if details.Establishment.ActivityStartDate != nil {
		_, err := time.Parse(
			time.DateOnly,
			*details.Establishment.ActivityStartDate,
		)
		if err != nil {
			t.Fatalf(
				"expected activity start date in YYYY-MM-DD format, got %q",
				*details.Establishment.ActivityStartDate,
			)
		}
	}

	if details.Establishment.EconomicActivities.Secondary == nil {
		t.Fatal("expected secondary economic activities collection, got nil")
	}

	if len(details.Establishment.EconomicActivities.Secondary) == 0 {
		t.Fatal("expected secondary economic activities, got empty collection")
	}

	expectedSecondaryCodes := secondaryEconomicActivityCodes(
		&secondaryEconomicActivities,
	)

	if len(details.Establishment.EconomicActivities.Secondary) !=
		len(expectedSecondaryCodes) {
		t.Fatalf(
			"expected %d secondary economic activities, got %d",
			len(expectedSecondaryCodes),
			len(details.Establishment.EconomicActivities.Secondary),
		)
	}

	for index, expectedCode := range expectedSecondaryCodes {
		actualCode :=
			details.Establishment.EconomicActivities.Secondary[index].Code

		if actualCode != expectedCode {
			t.Fatalf(
				"expected secondary economic activity code %q at index %d, got %q",
				expectedCode,
				index,
				actualCode,
			)
		}
	}
}
