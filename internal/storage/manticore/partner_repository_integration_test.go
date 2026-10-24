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

func TestPartnerRepositoryIntegration(t *testing.T) {
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
		"cnpj_partners_contract_test_%d", time.Now().UnixNano(),
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

	index, err := NewPartnerIndex(db, table)
	if err != nil {
		t.Fatal(err)
	}
	if err := index.RecreateTable(ctx); err != nil {
		t.Fatal(err)
	}

	maskA := "***111111**"
	maskB := "***222222**"
	emptyDocument := ""

	makeDocument := func(
		id uint64,
		root string,
		cnpj string,
		name string,
		document *string,
	) search.PartnerDocument {
		return search.PartnerDocument{
			ID:            id,
			SourceKey:     fmt.Sprintf("%s:%020d", root, id),
			BasicCNPJ:     root,
			CNPJ:          cnpj,
			CorporateName: "EMPRESA DE TESTE",
			Name:          name,
			Document:      document,
		}
	}

	documents := []search.PartnerDocument{
		makeDocument(1, "12ABC345", "12ABC345000190", "DANIEL SANTELLO", &maskA),
		makeDocument(2, "12ABC345", "12ABC345000190", "DANIEL SANTELLO", &maskB),
		makeDocument(3, "98XYZ765", "98XYZ765000190", "DANIEL SANTELLO", &maskA),
		makeDocument(4, "12ABC345", "12ABC345000190", "DANIEL SANTELLO", nil),
		makeDocument(5, "12ABC345", "12ABC345000190", "DANIEL SANTELLO", &emptyDocument),
		makeDocument(6, "12ABC345", "12ABC345000190", "DANIEL SANTELLO", &maskA),
		makeDocument(7, "98XYZ765", "98XYZ765000190", "DANIEL", &maskA),
		makeDocument(8, "98XYZ765", "98XYZ765000190", "SANTELLO", &maskA),
		makeDocument(9, "98XYZ765", "98XYZ765000190", "OUTRO NOME", &maskA),
	}
	documents[8].CorporateName = "DANIEL SANTELLO PARTICIPACOES"

	if err := index.WriteBatch(ctx, documents); err != nil {
		t.Fatal(err)
	}

	// Reuse the enrichment fixture from the corporate-name integration test.
	enrichment := &corporateNameEnrichmentStub{}
	repository, err := NewPartnerRepository(db, table, enrichment)
	if err != nil {
		t.Fatal(err)
	}

	options := company.PartnerSearchOptions{
		Name: "daniel santello", Page: 1, PageSize: 20,
	}
	result, err := repository.SearchByPartnerName(ctx, options)
	if err != nil {
		t.Fatal(err)
	}
	if result.Pagination != company.NewPagination(1, 20, 6) ||
		len(result.Items) != 6 {
		t.Fatalf("expected six matching links, got %#v", result)
	}

	for position, item := range result.Items {
		expected := documents[position]
		if item.Name != expected.Name ||
			item.CNPJ != expected.CNPJ ||
			item.CorporateName != expected.CorporateName ||
			!reflect.DeepEqual(item.Document, expected.Document) {
			t.Fatalf("wrong link at position %d: %#v", position, item)
		}
		if item.State == nil || *item.State != "SP" ||
			item.Municipality == nil || item.Municipality.Code != "7107" ||
			item.RegistrationStatus == nil ||
			item.RegistrationStatus.Code != "02" {
			t.Fatalf("missing enrichment at position %d", position)
		}
	}

	if result.Items[3].Document != nil {
		t.Fatal("null document must remain null")
	}
	if result.Items[4].Document == nil || *result.Items[4].Document != "" {
		t.Fatal("stored empty document must remain distinct from null")
	}
	if !reflect.DeepEqual(result.Items[0], result.Items[5]) {
		t.Fatal("repeated links must remain separate matching results")
	}

	expectedCNPJs := make([]string, 6)
	for position := range expectedCNPJs {
		expectedCNPJs[position] = documents[position].CNPJ
	}
	if enrichment.calls != 1 ||
		!reflect.DeepEqual(enrichment.cnpjs, expectedCNPJs) {
		t.Fatalf("expected one enrichment batch for the page: %v", enrichment.cnpjs)
	}

	// Name terms AND literal masked document, with independent pagination.
	filteredOptions := options
	filteredOptions.Document = &maskA
	filteredOptions.PageSize = 2

	first, err := repository.SearchByPartnerName(ctx, filteredOptions)
	if err != nil {
		t.Fatal(err)
	}
	if first.Pagination != company.NewPagination(1, 2, 3) ||
		len(first.Items) != 2 ||
		first.Items[0].CNPJ != documents[0].CNPJ ||
		first.Items[1].CNPJ != documents[2].CNPJ {
		t.Fatalf("wrong filtered first page: %#v", first)
	}

	filteredOptions.Page = 2
	second, err := repository.SearchByPartnerName(ctx, filteredOptions)
	if err != nil {
		t.Fatal(err)
	}
	if second.Pagination != company.NewPagination(2, 2, 3) ||
		len(second.Items) != 1 ||
		!reflect.DeepEqual(second.Items[0], result.Items[5]) {
		t.Fatalf("wrong filtered second page: %#v", second)
	}

	callsBefore := enrichment.calls
	filteredOptions.Page = 3
	beyond, err := repository.SearchByPartnerName(ctx, filteredOptions)
	if err != nil || len(beyond.Items) != 0 ||
		beyond.Pagination.TotalItems != 3 {
		t.Fatalf("wrong filtered page beyond results: %#v, %v", beyond, err)
	}

	for _, document := range []string{"***999999**", "***111%**"} {
		query := options
		query.Document = &document
		empty, err := repository.SearchByPartnerName(ctx, query)
		if err != nil || len(empty.Items) != 0 ||
			empty.Pagination.TotalItems != 0 {
			t.Fatalf("document must be compared literally: %#v, %v", empty, err)
		}
	}
	if enrichment.calls != callsBefore {
		t.Fatal("empty pages must not query enrichment")
	}

	// Internal storage check: an empty string must not select the null record.
	// The public service separately rejects an empty document parameter.
	emptyOptions := options
	emptyOptions.Document = &emptyDocument
	emptyResult, err := repository.SearchByPartnerName(ctx, emptyOptions)
	if err != nil || len(emptyResult.Items) != 1 ||
		emptyResult.Items[0].Document == nil {
		t.Fatalf("empty and null documents were confused: %#v, %v", emptyResult, err)
	}

	expectedError := errors.New("MySQL enrichment failed")
	enrichment.err = expectedError
	_, err = repository.SearchByPartnerName(ctx, options)
	if !errors.Is(err, expectedError) ||
		errors.Is(err, company.ErrSearchUnavailable) {
		t.Fatalf("wrong enrichment error classification: %v", err)
	}
	enrichment.err = nil

	// Six original matches plus 994 = exactly 1000 matching links.
	for start := 0; start < 994; start += 100 {
		batch := make([]search.PartnerDocument, 0, 100)
		for offset := start; offset < start+100 && offset < 994; offset++ {
			batch = append(batch, makeDocument(
				uint64(100+offset),
				"12ABC345",
				"12ABC345000190",
				"DANIEL SANTELLO",
				&maskB,
			))
		}
		if err := index.WriteBatch(ctx, batch); err != nil {
			t.Fatal(err)
		}
	}

	last, err := repository.SearchByPartnerName(ctx, company.PartnerSearchOptions{
		Name: "daniel santello", Page: 10, PageSize: 100,
	})
	if err != nil || len(last.Items) != 100 ||
		last.Pagination != company.NewPagination(10, 100, 1000) {
		t.Fatalf("wrong page at exactly 1000 links: %#v, %v", last, err)
	}
	if len(enrichment.cnpjs) != 100 {
		t.Fatal("enrichment must receive only the current page")
	}

	if err := index.WriteBatch(ctx, []search.PartnerDocument{
		makeDocument(
			2000, "12ABC345", "12ABC345000190",
			"DANIEL SANTELLO", &maskB,
		),
	}); err != nil {
		t.Fatal(err)
	}

	callsBefore = enrichment.calls
	_, err = repository.SearchByPartnerName(ctx, options)
	if !errors.Is(err, company.ErrSearchTooBroad) ||
		errors.Is(err, company.ErrSearchUnavailable) {
		t.Fatalf("expected search limit without fallback: %v", err)
	}
	if enrichment.calls != callsBefore {
		t.Fatal("too broad must not query enrichment")
	}

	// The document filter must run before the 1000-result check.
	filteredOptions.Page = 1
	filteredOptions.PageSize = 20
	narrowed, err := repository.SearchByPartnerName(ctx, filteredOptions)
	if err != nil || len(narrowed.Items) != 3 ||
		narrowed.Pagination.TotalItems != 3 {
		t.Fatalf("document filter must narrow the count before the limit: %#v, %v",
			narrowed, err)
	}
	for _, item := range narrowed.Items {
		if item.Document == nil || *item.Document != maskA {
			t.Fatalf("unexpected document in narrowed result: %#v", item)
		}
	}

	t.Log("validated partner links, homonyms, literal documents, nulls and filtered 1000/1001")
}
