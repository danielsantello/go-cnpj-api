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

func TestEstablishmentSourceReadBatchIntegration(t *testing.T) {
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
	source := NewEstablishmentSource(repository)

	first, err := source.ReadBatch(ctx, "", 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(first) == 0 || len(first) > 5 {
		t.Fatalf("unexpected first batch size: %d", len(first))
	}

	previous := ""
	for _, document := range first {
		if document.CNPJ <= previous {
			t.Fatal("establishments must advance by full CNPJ")
		}
		previous = document.CNPJ

		details, err := repository.FindByCNPJ(
			ctx, document.CNPJ, company.FindOptions{},
		)
		if err != nil {
			t.Fatal(err)
		}
		if details.Establishment.TradeName == nil {
			t.Fatal("indexed establishment must have a non-null trade name")
		}

		expected := search.EstablishmentDocument{
			CNPJ:          details.Establishment.CNPJ,
			CorporateName: details.Company.CorporateName,
			TradeName:     *details.Establishment.TradeName,
		}
		if !reflect.DeepEqual(document, expected) {
			t.Fatalf("source altered establishment data: got %#v, expected %#v",
				document, expected)
		}
	}

	second, err := source.ReadBatch(ctx, previous, 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(second) > 5 {
		t.Fatalf("unexpected second batch size: %d", len(second))
	}
	for _, document := range second {
		if document.CNPJ <= previous {
			t.Fatal("second batch repeated or moved behind the cursor")
		}
		if document.ID != 0 {
			t.Fatal("the source must not assign document IDs")
		}
		previous = document.CNPJ
	}

	// Repeating the same read must preserve content and ordering.
	repeated, err := source.ReadBatch(ctx, first[len(first)-1].CNPJ, 5)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(second, repeated) {
		t.Fatal("repeated source read changed the batch")
	}

	t.Logf("validated %d establishments and %d in the next batch",
		len(first), len(second))
}

func TestEstablishmentSourceRejectsInvalidLimit(t *testing.T) {
	source := NewEstablishmentSource(NewCompanyRepository(nil))

	for _, limit := range []int{-1, 0, 1001} {
		if _, err := source.ReadBatch(
			context.Background(), "", limit,
		); err == nil {
			t.Fatalf("expected invalid limit %d to fail before querying", limit)
		}
	}
}
