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
)

func TestCompanyRepositorySearchByTradeNameIntegration(t *testing.T) {
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

	var expected company.EstablishmentSummary

	err = mysqlDatabase.QueryRowContext(
		ctx,
		`
			SELECT
				establishment.cnpj,
				establishment.trade_name,
				company.legal_name
			FROM establishments AS establishment
			INNER JOIN companies AS company
				ON company.cnpj_root = establishment.cnpj_root
			WHERE establishment.trade_name LIKE 'SANTELLO%'
			ORDER BY
				establishment.trade_name,
				establishment.cnpj
			LIMIT 1
		`,
	).Scan(
		&expected.CNPJ,
		&expected.TradeName,
		&expected.CorporateName,
	)
	if err != nil {
		t.Fatalf("select integration test establishment: %v", err)
	}

	repository := NewCompanyRepository(mysqlDatabase)

	details, err := repository.FindByCNPJ(
		ctx,
		expected.CNPJ,
		company.FindOptions{},
	)
	if err != nil {
		t.Fatalf("find expected establishment details: %v", err)
	}

	expected.Municipality = details.Establishment.Address.Municipality
	expected.State = details.Establishment.Address.State
	expected.RegistrationStatus = details.Establishment.Registration.Status

	result, err := repository.SearchByTradeName(
		ctx,
		company.SearchOptions{
			Name:     expected.TradeName,
			Page:     1,
			PageSize: 2,
		},
	)
	if err != nil {
		t.Fatalf("search establishments by trade name: %v", err)
	}

	if len(result.Items) == 0 {
		t.Fatal("expected establishments, got empty collection")
	}

	if !reflect.DeepEqual(result.Items[0], expected) {
		t.Fatalf(
			"expected first establishment %#v, got %#v",
			expected,
			result.Items[0],
		)
	}

	if result.Pagination.TotalItems < 1 {
		t.Fatalf(
			"expected positive total items, got %d",
			result.Pagination.TotalItems,
		)
	}

	expectedPagination := company.NewPagination(
		1,
		2,
		result.Pagination.TotalItems,
	)
	if result.Pagination != expectedPagination {
		t.Fatalf(
			"expected pagination %#v, got %#v",
			expectedPagination,
			result.Pagination,
		)
	}

	secondPage, err := repository.SearchByTradeName(
		ctx,
		company.SearchOptions{
			Name:     expected.TradeName,
			Page:     2,
			PageSize: 1,
		},
	)
	if err != nil {
		t.Fatalf("search second page: %v", err)
	}

	expectedItems := result.Items[1:]
	if !reflect.DeepEqual(secondPage.Items, expectedItems) {
		t.Fatalf(
			"expected second page items %#v, got %#v",
			expectedItems,
			secondPage.Items,
		)
	}

	expectedSecondPagination := company.NewPagination(
		2,
		1,
		result.Pagination.TotalItems,
	)
	if secondPage.Pagination != expectedSecondPagination {
		t.Fatalf(
			"expected second page pagination %#v, got %#v",
			expectedSecondPagination,
			secondPage.Pagination,
		)
	}
}
