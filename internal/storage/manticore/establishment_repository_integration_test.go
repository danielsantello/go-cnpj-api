package manticore

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/danielsantello/go-cnpj-api/internal/company"
	"github.com/danielsantello/go-cnpj-api/internal/config"
	"github.com/danielsantello/go-cnpj-api/internal/database"
	"github.com/danielsantello/go-cnpj-api/internal/search"
)

type establishmentEnrichmentFixture struct {
	calls int
	cnpjs []string
	err   error
}

func (repository *establishmentEnrichmentFixture) FindEstablishmentSummaries(
	_ context.Context,
	cnpjs []string,
) (map[string]search.EstablishmentSummary, error) {
	repository.calls++
	repository.cnpjs = append([]string(nil), cnpjs...)
	if repository.err != nil {
		return nil, repository.err
	}

	result := make(map[string]search.EstablishmentSummary, len(cnpjs))
	for _, cnpj := range cnpjs {
		state := "SC"
		switch cnpj {
		case "12ABC345000271":
			state = "SP"
		case "12ABC345000352":
			state = "PR"
		}
		description := "MUNICIPIO DE TESTE"
		active := "ATIVA"
		result[cnpj] = search.EstablishmentSummary{
			Municipality: &company.CodeDescription{
				Code: "0001", Description: &description,
			},
			State: &state,
			RegistrationStatus: &company.CodeDescription{
				Code: "02", Description: &active,
			},
		}
	}
	return result, nil
}

func TestEstablishmentRepositoryIntegration(t *testing.T) {
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
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("close Manticore connection: %v", err)
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	table := fmt.Sprintf(
		"cnpj_establishments_contract_test_%d", time.Now().UnixNano(),
	)
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(
			context.Background(), 5*time.Second,
		)
		defer cleanupCancel()
		if _, err := db.ExecContext(
			cleanupCtx, "DROP TABLE IF EXISTS "+table,
		); err != nil {
			t.Errorf("drop test table %s: %v", table, err)
		}
	})

	index, err := NewEstablishmentIndex(db, table)
	if err != nil {
		t.Fatal(err)
	}
	if err := index.RecreateTable(ctx); err != nil {
		t.Fatal(err)
	}

	documents := []search.EstablishmentDocument{
		{
			ID: 1, CNPJ: "12ABC345000190",
			CorporateName: "EMPRESA A", TradeName: "MERCADO CENTRAL",
		},
		{
			ID: 2, CNPJ: "12ABC345000271",
			CorporateName: "EMPRESA A", TradeName: "PADARIA AURORA",
		},
		{
			ID: 3, CNPJ: "12ABC345000352",
			CorporateName: "EMPRESA A", TradeName: "AURORA LESTE",
		},
		{
			ID: 4, CNPJ: "98XYZ765000190",
			CorporateName: "AURORA PARTICIPACOES", TradeName: "LOJA DISTANTE",
		},
	}
	if err := index.WriteBatch(ctx, documents); err != nil {
		t.Fatal(err)
	}

	enrichment := &establishmentEnrichmentFixture{}
	repository, err := NewEstablishmentRepository(db, table, enrichment)
	if err != nil {
		t.Fatal(err)
	}

	options := company.SearchOptions{Name: "aurora", Page: 1, PageSize: 1}
	first, err := repository.SearchByTradeName(ctx, options)
	if err != nil {
		t.Fatal(err)
	}
	if first.Pagination != company.NewPagination(1, 1, 2) ||
		len(first.Items) != 1 ||
		first.Items[0].CNPJ != documents[1].CNPJ ||
		first.Items[0].TradeName != documents[1].TradeName ||
		first.Items[0].CorporateName != "EMPRESA A" ||
		first.Items[0].State == nil || *first.Items[0].State != "SP" {
		t.Fatalf("wrong first establishment: %#v", first)
	}

	options.Page = 2
	second, err := repository.SearchByTradeName(ctx, options)
	if err != nil {
		t.Fatal(err)
	}
	if second.Pagination != company.NewPagination(2, 1, 2) ||
		len(second.Items) != 1 ||
		second.Items[0].CNPJ != documents[2].CNPJ ||
		second.Items[0].TradeName != documents[2].TradeName ||
		second.Items[0].State == nil || *second.Items[0].State != "PR" {
		t.Fatalf("wrong second establishment: %#v", second)
	}

	options.Page = 1
	options.PageSize = 2
	combined, err := repository.SearchByTradeName(ctx, options)
	if err != nil {
		t.Fatal(err)
	}
	expectedItems := append(
		[]company.EstablishmentSummary{first.Items[0]},
		second.Items[0],
	)
	if !reflect.DeepEqual(combined.Items, expectedItems) {
		t.Fatal("pagination changed establishment order or content")
	}
	if !reflect.DeepEqual(enrichment.cnpjs, []string{
		documents[1].CNPJ, documents[2].CNPJ,
	}) {
		t.Fatalf("wrong CNPJs sent for enrichment: %v", enrichment.cnpjs)
	}
	for _, item := range combined.Items {
		if item.Municipality == nil ||
			item.Municipality.Description == nil ||
			item.RegistrationStatus == nil ||
			item.RegistrationStatus.Code != "02" ||
			item.RegistrationStatus.Description == nil {
			t.Fatalf("missing enrichment: %#v", item)
		}
	}

	callsBefore := enrichment.calls
	for _, options := range []company.SearchOptions{
		{Name: "aurora", Page: 3, PageSize: 1},
		{Name: "aurora central", Page: 1, PageSize: 20},
		{Name: "participacoes", Page: 1, PageSize: 20},
		{Name: "!!!", Page: 1, PageSize: 20},
	} {
		result, err := repository.SearchByTradeName(ctx, options)
		if err != nil || len(result.Items) != 0 {
			t.Fatalf("expected empty result for %#v: %#v, %v",
				options, result, err)
		}
		if options.Name == "aurora" && result.Pagination.TotalItems != 2 {
			t.Fatal("page beyond results lost the total")
		}
	}
	if enrichment.calls != callsBefore {
		t.Fatal("empty pages must not query enrichment")
	}

	result, err := repository.SearchByTradeName(ctx, company.SearchOptions{
		Name: "padaria aurora", Page: 1, PageSize: 20,
	})
	if err != nil || len(result.Items) != 1 ||
		result.Items[0].CNPJ != documents[1].CNPJ {
		t.Fatalf("terms must match the same trade name: %#v, %v", result, err)
	}

	expectedError := errors.New("MySQL enrichment failed")
	enrichment.err = expectedError
	_, err = repository.SearchByTradeName(ctx, options)
	if !errors.Is(err, expectedError) ||
		errors.Is(err, company.ErrSearchUnavailable) {
		t.Fatalf("wrong enrichment error classification: %v", err)
	}
	enrichment.err = nil

	// Add 998 matching establishments of the same company: total = 1000.
	for start := 0; start < 998; start += 100 {
		batch := make([]search.EstablishmentDocument, 0, 100)
		for offset := start; offset < start+100 && offset < 998; offset++ {
			number := 100 + offset
			batch = append(batch, search.EstablishmentDocument{
				ID:            uint64(number),
				CNPJ:          fmt.Sprintf("12ABC345%04d00", number),
				CorporateName: "EMPRESA A",
				TradeName:     "AURORA TESTE",
			})
		}
		if err := index.WriteBatch(ctx, batch); err != nil {
			t.Fatal(err)
		}
	}

	result, err = repository.SearchByTradeName(ctx, company.SearchOptions{
		Name: "aurora", Page: 10, PageSize: 100,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Pagination != company.NewPagination(10, 100, 1000) ||
		len(result.Items) != 100 ||
		result.Items[0].CNPJ != "12ABC345099800" ||
		result.Items[99].CNPJ != "12ABC345109700" {
		t.Fatalf("unexpected last page at 1000 results: %#v", result)
	}
	if len(enrichment.cnpjs) != 100 {
		t.Fatal("enrichment must receive only the page's 100 CNPJs")
	}

	if err := index.WriteBatch(ctx, []search.EstablishmentDocument{{
		ID: 2000, CNPJ: "12ABC345200000",
		CorporateName: "EMPRESA A", TradeName: "AURORA TESTE",
	}}); err != nil {
		t.Fatal(err)
	}

	callsBefore = enrichment.calls
	_, err = repository.SearchByTradeName(ctx, company.SearchOptions{
		Name: "aurora", Page: 1, PageSize: 20,
	})
	if !errors.Is(err, company.ErrSearchTooBroad) ||
		errors.Is(err, company.ErrSearchUnavailable) {
		t.Fatalf("expected search limit without fallback: %v", err)
	}
	if enrichment.calls != callsBefore {
		t.Fatal("too broad must not query enrichment")
	}

	t.Log("validated establishment granularity, field restriction, enrichment and 1000/1001")
}
