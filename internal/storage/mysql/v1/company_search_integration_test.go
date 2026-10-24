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

func TestCompanyRepositorySearchByCorporateNameIntegration(
	t *testing.T,
) {
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

	var expectedCNPJ string
	var corporateName string

	err = mysqlDatabase.QueryRowContext(
		ctx,
		`
			SELECT
				(
					SELECT establishment.cnpj
					FROM establishments AS establishment
					WHERE establishment.cnpj_root =
						company.cnpj_root
					ORDER BY
						establishment.head_office_branch_indicator =
							'1' DESC,
						establishment.cnpj
					LIMIT 1
				),
				company.legal_name
			FROM companies AS company
			WHERE CHAR_LENGTH(TRIM(company.legal_name)) >= 3
				AND EXISTS (
					SELECT 1
					FROM establishments AS establishment
					WHERE establishment.cnpj_root =
						company.cnpj_root
				)
			ORDER BY
				company.legal_name,
				company.cnpj_root
			LIMIT 1
		`,
	).Scan(
		&expectedCNPJ,
		&corporateName,
	)
	if err != nil {
		t.Fatalf(
			"select integration test company: %v",
			err,
		)
	}

	repository := NewCompanyRepository(mysqlDatabase)

	result, err := repository.SearchByCorporateName(
		ctx,
		company.SearchOptions{
			Name:     corporateName,
			Page:     1,
			PageSize: 10,
		},
	)
	if err != nil {
		t.Fatalf(
			"search companies by corporate name: %v",
			err,
		)
	}

	if len(result.Items) == 0 {
		t.Fatal("expected companies, got empty collection")
	}

	if result.Items[0].CNPJ != expectedCNPJ {
		t.Fatalf(
			"expected first CNPJ %q, got %q",
			expectedCNPJ,
			result.Items[0].CNPJ,
		)
	}

	if result.Items[0].CorporateName != corporateName {
		t.Fatalf(
			"expected first corporate name %q, got %q",
			corporateName,
			result.Items[0].CorporateName,
		)
	}

	for index, item := range result.Items {
		details, err := repository.FindByCNPJ(
			ctx,
			item.CNPJ,
			company.FindOptions{},
		)
		if err != nil {
			t.Fatalf("find details at index %d: %v", index, err)
		}

		if !reflect.DeepEqual(
			item.TradeName,
			details.Establishment.TradeName,
		) {
			t.Fatalf(
				"trade name differs from CNPJ details at index %d",
				index,
			)
		}

		if !reflect.DeepEqual(
			item.Municipality,
			details.Establishment.Address.Municipality,
		) {
			t.Fatalf("municipality differs from CNPJ details at index %d", index)
		}

		if !reflect.DeepEqual(
			item.State,
			details.Establishment.Address.State,
		) {
			t.Fatalf("state differs from CNPJ details at index %d", index)
		}

		if !reflect.DeepEqual(
			item.RegistrationStatus,
			details.Establishment.Registration.Status,
		) {
			t.Fatalf("registration status differs from CNPJ details at index %d", index)
		}

		if len(item.CNPJ) != 14 {
			t.Fatalf(
				"expected 14-character CNPJ at index %d, got %q",
				index,
				item.CNPJ,
			)
		}

		if !strings.HasPrefix(item.CorporateName, corporateName) {
			t.Fatalf(
				"expected corporate name with prefix %q at index %d, got %q",
				corporateName,
				index,
				item.CorporateName,
			)
		}
	}

	if result.Pagination.Page != 1 {
		t.Fatalf(
			"expected page 1, got %d",
			result.Pagination.Page,
		)
	}

	if result.Pagination.PageSize != 10 {
		t.Fatalf(
			"expected page size 10, got %d",
			result.Pagination.PageSize,
		)
	}

	if result.Pagination.TotalItems < int64(len(result.Items)) {
		t.Fatalf(
			"expected total items at least %d, got %d",
			len(result.Items),
			result.Pagination.TotalItems,
		)
	}
}
