package company

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestServiceSearchByPartnerNameNormalizesInput(t *testing.T) {
	document := "  ***123456**  "
	repository := &companyRepositoryStub{
		partnerSearchResult: PartnerSearchResult{
			Items: []PartnerSummary{
				{Name: "DANIEL SANTELLO"},
			},
			Pagination: NewPagination(2, 5, 6),
		},
	}

	result, err := NewService(repository).SearchByPartnerName(
		context.Background(),
		PartnerSearchOptions{
			Name:     "  DANIEL SANTELLO  ",
			Document: &document,
			Page:     2,
			PageSize: 5,
		},
	)
	if err != nil {
		t.Fatalf("search partners: %v", err)
	}

	got := repository.receivedPartnerSearchOptions
	if got.Name != "DANIEL SANTELLO" ||
		got.Document == nil ||
		*got.Document != "***123456**" ||
		got.Page != 2 ||
		got.PageSize != 5 {
		t.Fatalf("unexpected options: %#v", got)
	}

	if document != "  ***123456**  " {
		t.Fatal("caller document must not be modified")
	}

	if repository.partnerSearchCalls != 1 ||
		len(result.Items) != 1 ||
		result.Pagination.TotalItems != 6 {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestServiceSearchByPartnerNameWithoutDocument(t *testing.T) {
	repository := &companyRepositoryStub{}

	_, err := NewService(repository).SearchByPartnerName(
		context.Background(),
		PartnerSearchOptions{Name: "SANTELLO", Page: 1, PageSize: 20},
	)
	if err != nil {
		t.Fatalf("search partners: %v", err)
	}

	if repository.partnerSearchCalls != 1 ||
		repository.receivedPartnerSearchOptions.Document != nil {
		t.Fatal("expected search without document filter")
	}
}

func TestServiceSearchByPartnerNameRejectsInvalidInput(t *testing.T) {
	emptyDocument := "   "
	longDocument := strings.Repeat("1", 21)

	tests := []struct {
		name     string
		options  PartnerSearchOptions
		expected error
	}{
		{
			name:     "short name",
			options:  PartnerSearchOptions{Name: " AB "},
			expected: ErrPartnerNameTooShort,
		},
		{
			name:     "long name",
			options:  PartnerSearchOptions{Name: strings.Repeat("Á", 256)},
			expected: ErrPartnerNameTooLong,
		},
		{
			name: "empty document",
			options: PartnerSearchOptions{
				Name: "DANIEL SANTELLO", Document: &emptyDocument,
			},
			expected: ErrPartnerDocumentInvalid,
		},
		{
			name: "long document",
			options: PartnerSearchOptions{
				Name: "DANIEL SANTELLO", Document: &longDocument,
			},
			expected: ErrPartnerDocumentInvalid,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &companyRepositoryStub{}

			_, err := NewService(repository).SearchByPartnerName(
				context.Background(), test.options,
			)

			if !errors.Is(err, test.expected) {
				t.Fatalf("expected %v, got %v", test.expected, err)
			}
			if repository.partnerSearchCalls != 0 {
				t.Fatal("repository must not be called for invalid input")
			}
		})
	}
}

func TestServiceSearchByPartnerNamePreservesRepositoryError(t *testing.T) {
	expected := errors.New("database unavailable")
	repository := &companyRepositoryStub{partnerSearchErr: expected}

	_, err := NewService(repository).SearchByPartnerName(
		context.Background(),
		PartnerSearchOptions{Name: "SANTELLO", Page: 1, PageSize: 20},
	)

	if !errors.Is(err, expected) {
		t.Fatalf("expected wrapped repository error, got %v", err)
	}
}
