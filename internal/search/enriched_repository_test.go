package search

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/danielsantello/go-cnpj-api/internal/company"
)

type enrichmentRepositoryStub struct {
	calls  int
	cnpjs  []string
	result map[string]EstablishmentSummary
	err    error
}

func (repository *enrichmentRepositoryStub) FindEstablishmentSummaries(
	_ context.Context,
	cnpjs []string,
) (map[string]EstablishmentSummary, error) {
	repository.calls++
	repository.cnpjs = append([]string(nil), cnpjs...)
	return repository.result, repository.err
}

func TestEnrichedRepositoryPreservesSearchPage(t *testing.T) {
	sp := "SP"
	pr := "PR"
	city := "SAO PAULO"
	active := "ATIVA"

	municipality := &company.CodeDescription{
		Code: "7107", Description: &city,
	}
	status := &company.CodeDescription{
		Code: "02", Description: &active,
	}

	page := company.NewPagination(2, 2, 5)
	source := &searchRepositoryStub{
		result: Result{
			Items: []Item{
				{
					BasicCNPJ:     "22222222",
					CNPJ:          "22222222000200",
					CorporateName: "EMPRESA B",
					MatchedFields: []string{"trade_names"},
				},
				{
					BasicCNPJ:     "11111111",
					CNPJ:          "11111111000100",
					CorporateName: "EMPRESA A",
					MatchedFields: []string{"partner_names", "corporate_name"},
				},
			},
			Pagination: page,
		},
	}
	enrichment := &enrichmentRepositoryStub{
		result: map[string]EstablishmentSummary{
			"11111111000100": {State: &pr},
			"22222222000200": {
				Municipality:       municipality,
				State:              &sp,
				RegistrationStatus: status,
			},
		},
	}

	options := Options{Name: "empresa", Page: 2, PageSize: 2}
	result, err := NewService(
		NewEnrichedRepository(source, enrichment),
	).Search(context.Background(), options)
	if err != nil {
		t.Fatal(err)
	}

	expected := Result{
		Items: []Item{
			{
				BasicCNPJ:          "22222222",
				CNPJ:               "22222222000200",
				CorporateName:      "EMPRESA B",
				MatchedFields:      []string{"trade_names"},
				Municipality:       municipality,
				State:              &sp,
				RegistrationStatus: status,
			},
			{
				BasicCNPJ:     "11111111",
				CNPJ:          "11111111000100",
				CorporateName: "EMPRESA A",
				MatchedFields: []string{"partner_names", "corporate_name"},
				State:         &pr,
			},
		},
		Pagination: page,
	}
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("unexpected enriched page: %#v", result)
	}
	if source.calls != 1 || source.options != options {
		t.Fatalf("unexpected search calls or options: %#v", source)
	}
	if enrichment.calls != 1 || !reflect.DeepEqual(
		enrichment.cnpjs,
		[]string{"22222222000200", "11111111000100"},
	) {
		t.Fatalf("expected one batch with full page CNPJs: %#v", enrichment)
	}
	if source.result.Items[0].State != nil {
		t.Fatal("enrichment mutated the source collection")
	}
}

func TestEnrichedRepositoryPreservesMissingAndRepeatedCNPJs(t *testing.T) {
	sp := "SP"
	source := &searchRepositoryStub{
		result: Result{
			Items: []Item{
				{CNPJ: "11111111000100"},
				{CNPJ: "99999999000100"},
				{CNPJ: "11111111000100"},
				{CNPJ: ""},
			},
			Pagination: company.NewPagination(1, 20, 4),
		},
	}
	enrichment := &enrichmentRepositoryStub{
		result: map[string]EstablishmentSummary{
			"11111111000100": {State: &sp},
		},
	}

	result, err := NewEnrichedRepository(source, enrichment).Search(
		context.Background(), Options{},
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Items) != 4 || result.Pagination != source.result.Pagination {
		t.Fatal("enrichment changed result count or pagination")
	}
	for index, original := range source.result.Items {
		if result.Items[index].CNPJ != original.CNPJ {
			t.Fatalf("CNPJ or order changed at index %d", index)
		}
	}
	if result.Items[0].State == nil || *result.Items[0].State != sp ||
		result.Items[2].State == nil || *result.Items[2].State != sp {
		t.Fatal("repeated CNPJs must receive the same enrichment")
	}
	for _, index := range []int{1, 3} {
		item := result.Items[index]
		if item.State != nil || item.Municipality != nil ||
			item.RegistrationStatus != nil {
			t.Fatalf("expected null fields at index %d", index)
		}
	}
	if enrichment.calls != 1 {
		t.Fatal("expected one enrichment call")
	}
}

func TestEnrichedRepositorySkipsPagesWithoutCNPJs(t *testing.T) {
	for _, items := range [][]Item{nil, {}, {{CNPJ: ""}}} {
		source := &searchRepositoryStub{
			result: Result{
				Items:      items,
				Pagination: company.NewPagination(10, 20, 5),
			},
		}
		enrichment := &enrichmentRepositoryStub{}

		result, err := NewEnrichedRepository(source, enrichment).Search(
			context.Background(), Options{},
		)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(result, source.result) {
			t.Fatal("page without CNPJs changed")
		}
		if enrichment.calls != 0 {
			t.Fatal("page without CNPJs must not query MySQL")
		}
	}
}

func TestEnrichedRepositoryPreservesErrors(t *testing.T) {
	t.Run("search", func(t *testing.T) {
		source := &searchRepositoryStub{err: ErrTooBroad}
		enrichment := &enrichmentRepositoryStub{}

		_, err := NewEnrichedRepository(source, enrichment).Search(
			context.Background(), Options{},
		)
		if !errors.Is(err, ErrTooBroad) {
			t.Fatalf("expected search error, got %v", err)
		}
		if enrichment.calls != 0 {
			t.Fatal("failed search must not query MySQL")
		}
	})

	t.Run("enrichment", func(t *testing.T) {
		expected := errors.New("MySQL unavailable")
		source := &searchRepositoryStub{
			result: Result{Items: []Item{{CNPJ: "11111111000100"}}},
		}
		enrichment := &enrichmentRepositoryStub{err: expected}

		_, err := NewEnrichedRepository(source, enrichment).Search(
			context.Background(), Options{},
		)
		if !errors.Is(err, expected) {
			t.Fatalf("expected wrapped enrichment error, got %v", err)
		}
	})
}
