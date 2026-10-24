package v1

import (
	"context"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/danielsantello/go-cnpj-api/internal/config"
	"github.com/danielsantello/go-cnpj-api/internal/database"
	"github.com/danielsantello/go-cnpj-api/internal/search"
)

func TestPartnerSourceReadBatchIntegration(t *testing.T) {
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

	source := NewPartnerSource(NewCompanyRepository(db))

	batch, err := source.ReadBatch(ctx, "", 8)
	if err != nil {
		t.Fatal(err)
	}
	if len(batch) == 0 || len(batch) > 8 {
		t.Fatalf("unexpected batch size: %d", len(batch))
	}

	cursor := ""
	withinCompany := false

	for position, document := range batch {
		if document.ID != 0 {
			t.Fatal("the source must not assign document IDs")
		}
		if document.SourceKey <= cursor {
			t.Fatal("partner source keys must advance")
		}

		root, occurrence, err := search.ParsePartnerSourceKey(document.SourceKey)
		if err != nil {
			t.Fatal(err)
		}
		if root != document.BasicCNPJ || occurrence == 0 {
			t.Fatal("cursor does not identify the returned partner link")
		}

		single, err := source.ReadBatch(ctx, cursor, 1)
		if err != nil {
			t.Fatal(err)
		}
		if len(single) != 1 || !reflect.DeepEqual(single[0], document) {
			t.Fatalf("reading one item changed batch position %d: %#v",
				position, single)
		}

		var corporateName, representativeCNPJ string
		err = db.QueryRowContext(ctx, `
			SELECT
				company.legal_name,
				(
					SELECT establishment.cnpj
					FROM establishments AS establishment
					WHERE establishment.cnpj_root = company.cnpj_root
					ORDER BY
						establishment.head_office_branch_indicator = '1' DESC,
						establishment.cnpj ASC
					LIMIT 1
				)
			FROM companies AS company
			WHERE company.cnpj_root = ?
		`, document.BasicCNPJ).Scan(&corporateName, &representativeCNPJ)
		if err != nil {
			t.Fatal(err)
		}
		if document.CorporateName != corporateName ||
			document.CNPJ != representativeCNPJ {
			t.Fatal("source changed company data or representative CNPJ")
		}

		var expectedDocument any
		if document.Document != nil {
			expectedDocument = *document.Document
		}

		var matchingLinks int
		err = db.QueryRowContext(ctx, `
			SELECT COUNT(*)
			FROM partners
			WHERE cnpj_root = ?
				AND CAST(partner_name AS BINARY) = CAST(? AS BINARY)
				AND CAST(partner_document AS BINARY) <=> CAST(? AS BINARY)
		`, document.BasicCNPJ, document.Name, expectedDocument).
			Scan(&matchingLinks)
		if err != nil {
			t.Fatal(err)
		}
		if matchingLinks == 0 {
			t.Fatal("returned name and document do not match a source link")
		}

		if position > 0 &&
			batch[position-1].BasicCNPJ == document.BasicCNPJ {
			withinCompany = true
		}
		cursor = document.SourceKey
	}

	next, err := source.ReadBatch(ctx, cursor, 8)
	if err != nil {
		t.Fatal(err)
	}
	repeated, err := source.ReadBatch(ctx, cursor, 8)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(next, repeated) {
		t.Fatal("repeating an internal cursor changed the returned batch")
	}

	previous := cursor
	for _, document := range next {
		if document.SourceKey <= previous || document.ID != 0 {
			t.Fatal("next batch repeated a cursor or assigned an ID")
		}
		previous = document.SourceKey
	}

	if !withinCompany {
		t.Log("sample has no consecutive links from the same company")
	}

	t.Logf(
		"validated %d links, single-item pagination, source values and repeatable resume; within_company=%t",
		len(batch), withinCompany,
	)
}

func TestPartnerSourceRejectsInvalidInput(t *testing.T) {
	source := NewPartnerSource(NewCompanyRepository(nil))
	ctx := context.Background()

	for _, limit := range []int{-1, 0, 1001} {
		if _, err := source.ReadBatch(ctx, "", limit); err == nil {
			t.Fatalf("expected invalid limit %d to fail before querying", limit)
		}
	}

	if _, err := source.ReadBatch(ctx, "invalid-cursor", 1); err == nil {
		t.Fatal("expected invalid cursor to fail before querying")
	}
}
