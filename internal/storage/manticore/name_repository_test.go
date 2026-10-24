package manticore

import (
	"context"
	"reflect"
	"testing"

	"github.com/danielsantello/go-cnpj-api/internal/company"
	"github.com/danielsantello/go-cnpj-api/internal/search"
)

type corporateNamePageStub struct {
	result search.Result
}

func (stub corporateNamePageStub) Search(
	context.Context,
	search.Options,
) (search.Result, error) {
	return stub.result, nil
}

type corporateNameSummariesStub map[string]search.EstablishmentSummary

func (stub corporateNameSummariesStub) FindEstablishmentSummaries(
	context.Context,
	[]string,
) (map[string]search.EstablishmentSummary, error) {
	return stub, nil
}

func TestNameRepositoryTradeNameFromFullCNPJ(t *testing.T) {
	branch := "FANTASIA FILIAL"
	head := "FANTASIA MATRIZ"
	empty := ""
	page := company.NewPagination(2, 4, 8)

	source := corporateNamePageStub{
		result: search.Result{
			Items: []search.Item{
				{CNPJ: "12345678000200", CorporateName: "EMPRESA B"},
				{CNPJ: "23456789000100", CorporateName: "EMPRESA A"},
				{CNPJ: "34567890000100"},
				{CNPJ: "45678901000100"},
			},
			Pagination: page,
		},
	}

	summaries := corporateNameSummariesStub{
		"12345678000100": {TradeName: &head},
		"12345678000200": {TradeName: &branch},
		"23456789000100": {TradeName: &empty},
		"34567890000100": {},
		// O quarto CNPJ não foi encontrado no MySQL.
	}

	repository := &NameRepository{
		companies: search.NewEnrichedRepository(source, summaries),
	}

	result, err := repository.SearchByCorporateName(
		context.Background(),
		company.SearchOptions{
			Name:     "EMPRESA",
			Page:     2,
			PageSize: 4,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	expected := company.SearchResult{
		Items: []company.Summary{
			{
				CNPJ:          "12345678000200",
				CorporateName: "EMPRESA B",
				TradeName:     &branch,
			},
			{
				CNPJ:          "23456789000100",
				CorporateName: "EMPRESA A",
				TradeName:     &empty,
			},
			{CNPJ: "34567890000100"},
			{CNPJ: "45678901000100"},
		},
		Pagination: page,
	}

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("unexpected company page:\ngot: %#v\nwant: %#v",
			result, expected)
	}

	if source.result.Items[0].TradeName != nil {
		t.Fatal("enrichment mutated the source collection")
	}
}
