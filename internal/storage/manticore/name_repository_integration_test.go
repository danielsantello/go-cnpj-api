package manticore

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/danielsantello/go-cnpj-api/internal/company"
	"github.com/danielsantello/go-cnpj-api/internal/config"
	"github.com/danielsantello/go-cnpj-api/internal/database"
	"github.com/danielsantello/go-cnpj-api/internal/search"
)

type corporateNameEnrichmentStub struct {
	calls int
	cnpjs []string
	err   error
}

func (repository *corporateNameEnrichmentStub) FindEstablishmentSummaries(
	_ context.Context,
	cnpjs []string,
) (map[string]search.EstablishmentSummary, error) {
	repository.calls++
	repository.cnpjs = append([]string(nil), cnpjs...)

	if repository.err != nil {
		return nil, repository.err
	}

	state := "SP"
	city := "SAO PAULO"
	active := "ATIVA"
	result := make(map[string]search.EstablishmentSummary, len(cnpjs))
	for _, cnpj := range cnpjs {
		result[cnpj] = search.EstablishmentSummary{
			Municipality: &company.CodeDescription{
				Code: "7107", Description: &city,
			},
			State: &state,
			RegistrationStatus: &company.CodeDescription{
				Code: "02", Description: &active,
			},
		}
	}
	return result, nil
}

func TestNameRepositoryCorporateNameIntegration(t *testing.T) {
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

	table := fmt.Sprintf(
		"cnpj_companies_names_test_%d", time.Now().UnixNano(),
	)
	index, err := NewIndexForTable(db, table)
	if err != nil {
		t.Fatal(err)
	}
	if err := index.RecreateTable(ctx); err != nil {
		t.Fatal(err)
	}
	defer func() {
		cleanupContext, cleanupCancel := context.WithTimeout(
			context.Background(), 5*time.Second,
		)
		defer cleanupCancel()
		if _, err := db.ExecContext(
			cleanupContext, "DROP TABLE "+table,
		); err != nil {
			t.Errorf("drop test table %s: %v", table, err)
		}
	}()

	document := func(id uint64, corporateName string) search.Document {
		return search.Document{
			ID:            id,
			BasicCNPJ:     fmt.Sprintf("%08d", id),
			CNPJ:          fmt.Sprintf("%08d000100", id),
			CorporateName: corporateName,
		}
	}

	for first := uint64(1); first <= 1000; first += 100 {
		batch := make([]search.Document, 0, 100)
		for id := first; id < first+100; id++ {
			batch = append(batch, document(id, "ALVOTERMO COMPOSTO"))
		}
		if err := index.WriteBatch(ctx, batch); err != nil {
			t.Fatal(err)
		}
	}

	// These documents must not match a corporate-name-only search.
	tradeOnly := document(1001, "OUTRA EMPRESA")
	tradeOnly.TradeNames = []string{"ALVOTERMO COMPOSTO"}

	partnerOnly := document(1002, "OUTRA EMPRESA")
	partnerOnly.PartnerNames = []string{"ALVOTERMO COMPOSTO"}

	splitFields := document(1003, "ALVOTERMO")
	splitFields.TradeNames = []string{"COMPOSTO"}

	if err := index.WriteBatch(ctx, []search.Document{
		tradeOnly, partnerOnly, splitFields,
	}); err != nil {
		t.Fatal(err)
	}

	enrichment := &corporateNameEnrichmentStub{}
	repository, err := NewNameRepository(db, table, enrichment)
	if err != nil {
		t.Fatal(err)
	}

	options := company.SearchOptions{
		Name: "alvotermo composto", Page: 10, PageSize: 100,
	}
	result, err := repository.SearchByCorporateName(ctx, options)
	if err != nil {
		t.Fatal(err)
	}
	if result.Pagination != company.NewPagination(10, 100, 1000) ||
		len(result.Items) != 100 {
		t.Fatalf("unexpected page: %#v; items=%d",
			result.Pagination, len(result.Items))
	}
	if enrichment.calls != 1 || len(enrichment.cnpjs) != 100 {
		t.Fatalf("expected one enrichment batch of 100 CNPJs")
	}

	for position, item := range result.Items {
		expectedCNPJ := document(uint64(901+position), "").CNPJ
		if item.CNPJ != expectedCNPJ ||
			item.CorporateName != "ALVOTERMO COMPOSTO" {
			t.Fatalf("unexpected item at position %d: %#v", position, item)
		}
		if enrichment.cnpjs[position] != expectedCNPJ {
			t.Fatalf("wrong enrichment CNPJ at position %d", position)
		}
		if item.State == nil || *item.State != "SP" ||
			item.Municipality == nil || item.Municipality.Code != "7107" ||
			item.Municipality.Description == nil ||
			*item.Municipality.Description != "SAO PAULO" ||
			item.RegistrationStatus == nil ||
			item.RegistrationStatus.Code != "02" ||
			item.RegistrationStatus.Description == nil ||
			*item.RegistrationStatus.Description != "ATIVA" {
			t.Fatalf("missing enrichment at position %d: %#v", position, item)
		}
	}

	options.Page = 11
	result, err = repository.SearchByCorporateName(ctx, options)
	if err != nil || len(result.Items) != 0 ||
		result.Pagination != company.NewPagination(11, 100, 1000) {
		t.Fatalf("unexpected page beyond results: %#v, %v", result, err)
	}
	if enrichment.calls != 1 {
		t.Fatal("empty page must not query enrichment")
	}

	options.Page = 1
	options.Name = "!!!"
	result, err = repository.SearchByCorporateName(ctx, options)
	if err != nil || len(result.Items) != 0 ||
		result.Pagination.TotalItems != 0 {
		t.Fatalf("unexpected punctuation result: %#v, %v", result, err)
	}
	if enrichment.calls != 1 {
		t.Fatal("empty expression must not query enrichment")
	}

	// An enrichment failure must remain an error, without fallback marking.
	options.Name = "alvotermo composto"
	expectedError := errors.New("MySQL enrichment failed")
	enrichment.err = expectedError
	_, err = repository.SearchByCorporateName(ctx, options)
	if !errors.Is(err, expectedError) ||
		errors.Is(err, company.ErrSearchUnavailable) {
		t.Fatalf("unexpected enrichment error classification: %v", err)
	}
	enrichment.err = nil

	if err := index.WriteBatch(ctx, []search.Document{
		document(1004, "ALVOTERMO COMPOSTO"),
	}); err != nil {
		t.Fatal(err)
	}

	callsBefore := enrichment.calls
	_, err = repository.SearchByCorporateName(ctx, options)
	if !errors.Is(err, company.ErrSearchTooBroad) {
		t.Fatalf("expected company.ErrSearchTooBroad, got %v", err)
	}
	if errors.Is(err, company.ErrSearchUnavailable) {
		t.Fatal("too broad must not trigger fallback")
	}
	if enrichment.calls != callsBefore {
		t.Fatal("too broad must not query enrichment")
	}

	// A missing expression cannot be replaced by an unrestricted search.
	if expression := corporateNameSearchExpression(
		strings.Repeat("!", 3),
	); expression != "" {
		t.Fatalf("unexpected expression: %q", expression)
	}

	t.Log("validated field restriction, required terms, enrichment, pagination and 1000/1001")
}
