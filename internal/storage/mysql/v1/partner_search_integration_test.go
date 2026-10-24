package v1

import (
	"context"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/danielsantello/go-cnpj-api/internal/company"
	"github.com/danielsantello/go-cnpj-api/internal/config"
	"github.com/danielsantello/go-cnpj-api/internal/database"
)

func TestCompanyRepositorySearchByPartnerNameIntegration(t *testing.T) {
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

	repository := NewCompanyRepository(db)
	options := company.PartnerSearchOptions{
		Name:     "SANTELLO",
		Page:     1,
		PageSize: 20,
	}

	check := func(t *testing.T, options company.PartnerSearchOptions) company.PartnerSearchResult {
		t.Helper()

		result, err := repository.SearchByPartnerName(ctx, options)
		if err != nil {
			t.Fatal(err)
		}
		if len(result.Items) == 0 {
			t.Fatal("expected partner results")
		}

		for index, item := range result.Items {
			if options.Document != nil &&
				(!strings.HasPrefix(item.Name, options.Name) ||
					item.Document == nil ||
					*item.Document != *options.Document) {
				t.Fatalf("unexpected exact match at index %d: %#v", index, item)
			}

			details, err := repository.FindByCNPJ(
				ctx, item.CNPJ, company.FindOptions{},
			)
			if err != nil {
				t.Fatal(err)
			}

			if item.CorporateName != details.Company.CorporateName {
				t.Fatalf("corporate name differs at index %d", index)
			}
			if !reflect.DeepEqual(
				item.Municipality, details.Establishment.Address.Municipality,
			) {
				t.Fatalf("municipality differs at index %d", index)
			}
			if !reflect.DeepEqual(
				item.State, details.Establishment.Address.State,
			) {
				t.Fatalf("state differs at index %d", index)
			}
			if !reflect.DeepEqual(
				item.RegistrationStatus, details.Establishment.Registration.Status,
			) {
				t.Fatalf("registration status differs at index %d", index)
			}
		}

		expectedPagination := company.NewPagination(
			options.Page, options.PageSize, result.Pagination.TotalItems,
		)
		if result.Pagination != expectedPagination {
			t.Fatalf("unexpected pagination: %#v", result.Pagination)
		}

		return result
	}

	var prefixResult company.PartnerSearchResult
	t.Run("prefix", func(t *testing.T) {
		prefixResult = check(t, options)
	})

	t.Run("name prefix and document", func(t *testing.T) {
		for _, item := range prefixResult.Items {
			if item.Document != nil {
				exact := options
				exact.Name = options.Name
				exact.Document = item.Document
				check(t, exact)
				return
			}
		}
		t.Fatal("expected a partner with a document to test exact search")
	})
}
