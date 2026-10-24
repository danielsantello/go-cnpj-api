package manticore

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/danielsantello/go-cnpj-api/internal/config"
	"github.com/danielsantello/go-cnpj-api/internal/database"
	"github.com/danielsantello/go-cnpj-api/internal/search"
)

type rebuildSource[T any] struct {
	documents []T
	key       func(T) string
}

func (source rebuildSource[T]) ReadBatch(_ context.Context, after string, limit int) ([]T, error) {
	result := make([]T, 0, limit)
	for _, document := range source.documents {
		if source.key(document) > after {
			result = append(result, document)
			if len(result) == limit {
				break
			}
		}
	}
	return result, nil
}

func TestIndexRebuildIntegration(t *testing.T) {
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

	mask, empty := "***123456**", ""
	cases := []struct {
		prefix string
		cnpjs  []string
		load   func(string, int) error
	}{
		{
			"cnpj_companies",
			[]string{"00000001000100", "12ABC345000190"},
			func(table string, limit int) error {
				index, err := NewIndexForTable(db, table)
				if err != nil {
					return err
				}
				_, err = search.BuildIndex(ctx, rebuildSource[search.Document]{
					documents: []search.Document{
						{BasicCNPJ: "00000001", CNPJ: "00000001000100", CorporateName: "NUMERICA"},
						{BasicCNPJ: "12ABC345", CNPJ: "12ABC345000190", CorporateName: "ALFA"},
					},
					key: func(document search.Document) string { return document.BasicCNPJ },
				}, index, search.IndexOptions{BatchSize: 1, Limit: limit}, nil)
				return err
			},
		},
		{
			"cnpj_establishments",
			[]string{"00000001000100", "12ABC345000190", "12ABC345000271"},
			func(table string, limit int) error {
				index, err := NewEstablishmentIndex(db, table)
				if err != nil {
					return err
				}
				_, err = search.BuildEstablishmentIndex(ctx, rebuildSource[search.EstablishmentDocument]{
					documents: []search.EstablishmentDocument{
						{CNPJ: "00000001000100", CorporateName: "NUMERICA", TradeName: "LOJA"},
						{CNPJ: "12ABC345000190", CorporateName: "ALFA", TradeName: "MATRIZ"},
						{CNPJ: "12ABC345000271", CorporateName: "ALFA", TradeName: "FILIAL"},
					},
					key: func(document search.EstablishmentDocument) string { return document.CNPJ },
				}, index, search.EstablishmentIndexOptions{BatchSize: 1, Limit: limit}, nil)
				return err
			},
		},
		{
			"cnpj_partners",
			[]string{"12ABC345000190", "12ABC345000190", "12ABC345000190", "12ABC345000190"},
			func(table string, limit int) error {
				index, err := NewPartnerIndex(db, table)
				if err != nil {
					return err
				}
				documents := make([]search.PartnerDocument, 0, 4)
				for position, document := range []*string{&mask, &mask, nil, &empty} {
					key, err := search.MakePartnerSourceKey("12ABC345", uint64(position+1))
					if err != nil {
						return err
					}
					documents = append(documents, search.PartnerDocument{
						SourceKey: key, BasicCNPJ: "12ABC345", CNPJ: "12ABC345000190",
						CorporateName: "ALFA", Name: "MESMO SOCIO", Document: document,
					})
				}
				_, err = search.BuildPartnerIndex(ctx, rebuildSource[search.PartnerDocument]{
					documents: documents,
					key:       func(document search.PartnerDocument) string { return document.SourceKey },
				}, index, search.PartnerIndexOptions{BatchSize: 1, Limit: limit}, nil)
				return err
			},
		},
	}

	for _, test := range cases {
		t.Run(test.prefix, func(t *testing.T) {
			table := fmt.Sprintf("%s_rebuild_test_%d", test.prefix, time.Now().UnixNano())
			defer func() {
				cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cleanupCancel()
				if _, err := db.ExecContext(cleanupCtx, "DROP TABLE IF EXISTS "+table); err != nil {
					t.Errorf("drop test table: %v", err)
				}
			}()
			// A pre-existing schema and its records must be replaced from zero.
			if _, err := db.ExecContext(ctx, "CREATE TABLE "+table+" (legacy text)"); err != nil {
				t.Fatal(err)
			}
			if _, err := db.ExecContext(ctx, "INSERT INTO "+table+" (id, legacy) VALUES (999, 'old')"); err != nil {
				t.Fatal(err)
			}
			for _, limit := range []int{len(test.cnpjs), 1} {
				if err := test.load(table, limit); err != nil {
					t.Fatal(err)
				}
				rows, err := db.QueryContext(ctx, "SELECT id, cnpj FROM "+table+" ORDER BY id ASC LIMIT 100")
				if err != nil {
					t.Fatal(err)
				}
				position := 0
				for rows.Next() {
					var id uint64
					var cnpj string
					if err := rows.Scan(&id, &cnpj); err != nil {
						rows.Close()
						t.Fatal(err)
					}
					if position >= limit || id != uint64(position+1) || cnpj != test.cnpjs[position] {
						rows.Close()
						t.Fatalf("rebuild changed a textual CNPJ, kept stale data or failed to reset IDs: %d %q", id, cnpj)
					}
					position++
				}
				readErr := rows.Err()
				closeErr := rows.Close()
				if readErr != nil {
					t.Fatal(readErr)
				}
				if closeErr != nil {
					t.Fatal(closeErr)
				}
				if position != limit {
					t.Fatalf("expected %d rows, got %d", limit, position)
				}
			}
		})
	}
}
