package company

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

type automaticRepositoryStub struct {
	Repository
	calls int
	total int64
	err   error
}

func (repository *automaticRepositoryStub) SearchByCorporateName(
	_ context.Context,
	_ SearchOptions,
) (SearchResult, error) {
	repository.calls++
	return SearchResult{
		Items:      []Summary{},
		Pagination: NewPagination(1, 20, repository.total),
	}, repository.err
}

func (repository *automaticRepositoryStub) SearchByTradeName(
	_ context.Context,
	_ SearchOptions,
) (EstablishmentSearchResult, error) {
	repository.calls++
	return EstablishmentSearchResult{
		Items:      []EstablishmentSummary{},
		Pagination: NewPagination(1, 20, repository.total),
	}, repository.err
}

func (repository *automaticRepositoryStub) SearchByPartnerName(
	_ context.Context,
	_ PartnerSearchOptions,
) (PartnerSearchResult, error) {
	repository.calls++
	return PartnerSearchResult{
		Items:      []PartnerSummary{},
		Pagination: NewPagination(1, 20, repository.total),
	}, repository.err
}

func TestAutomaticRepositorySelection(t *testing.T) {
	unexpected := errors.New("query failed")

	operations := []struct {
		name string
		run  func(context.Context, Repository) (int64, error)
	}{
		{
			name: "corporate name",
			run: func(ctx context.Context, repository Repository) (int64, error) {
				result, err := repository.SearchByCorporateName(
					ctx, SearchOptions{Name: "empresa", Page: 1, PageSize: 20},
				)
				return result.Pagination.TotalItems, err
			},
		},
		{
			name: "trade name",
			run: func(ctx context.Context, repository Repository) (int64, error) {
				result, err := repository.SearchByTradeName(
					ctx, SearchOptions{Name: "empresa", Page: 1, PageSize: 20},
				)
				return result.Pagination.TotalItems, err
			},
		},
		{
			name: "partner name",
			run: func(ctx context.Context, repository Repository) (int64, error) {
				result, err := repository.SearchByPartnerName(
					ctx, PartnerSearchOptions{Name: "socio", Page: 1, PageSize: 20},
				)
				return result.Pagination.TotalItems, err
			},
		},
	}

	tests := []struct {
		name          string
		disabled      bool
		primaryTotal  int64
		primaryErr    error
		fallbackErr   error
		expectedTotal int64
		expectedErr   error
		fallbackCalls int
	}{
		{
			name:         "available",
			primaryTotal: 10, expectedTotal: 10,
		},
		{
			name:          "empty result",
			expectedTotal: 0,
		},
		{
			name:     "disabled",
			disabled: true, expectedTotal: 20, fallbackCalls: 1,
		},
		{
			name:          "unavailable",
			primaryErr:    fmt.Errorf("connection: %w", ErrSearchUnavailable),
			expectedTotal: 20, fallbackCalls: 1,
		},
		{
			name:       "too broad",
			primaryErr: ErrSearchTooBroad, expectedErr: ErrSearchTooBroad,
		},
		{
			name:        "validation error",
			primaryErr:  ErrPartnerDocumentInvalid,
			expectedErr: ErrPartnerDocumentInvalid,
		},
		{
			name:       "unexpected error",
			primaryErr: unexpected, expectedErr: unexpected,
		},
		{
			name:        "fallback error",
			primaryErr:  ErrSearchUnavailable,
			fallbackErr: unexpected, expectedErr: unexpected,
			fallbackCalls: 1,
		},
	}

	for _, operation := range operations {
		t.Run(operation.name, func(t *testing.T) {
			for _, test := range tests {
				t.Run(test.name, func(t *testing.T) {
					primary := &automaticRepositoryStub{
						total: test.primaryTotal, err: test.primaryErr,
					}
					fallback := &automaticRepositoryStub{
						total: 20, err: test.fallbackErr,
					}

					var selected NameSearchRepository
					if !test.disabled {
						selected = primary
					}

					total, err := operation.run(
						context.Background(),
						NewAutomaticRepository(fallback, selected),
					)
					if !errors.Is(err, test.expectedErr) {
						t.Fatalf("expected error %v, got %v", test.expectedErr, err)
					}
					if err == nil && total != test.expectedTotal {
						t.Fatalf("expected total %d, got %d",
							test.expectedTotal, total)
					}
					if fallback.calls != test.fallbackCalls {
						t.Fatalf("expected %d fallback calls, got %d",
							test.fallbackCalls, fallback.calls)
					}

					expectedPrimaryCalls := 1
					if test.disabled {
						expectedPrimaryCalls = 0
					}
					if primary.calls != expectedPrimaryCalls {
						t.Fatalf("expected %d primary calls, got %d",
							expectedPrimaryCalls, primary.calls)
					}
				})
			}

			t.Run("canceled request does not fall back", func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()

				primary := &automaticRepositoryStub{err: ErrSearchUnavailable}
				fallback := &automaticRepositoryStub{}
				_, err := operation.run(
					ctx, NewAutomaticRepository(fallback, primary),
				)
				if !errors.Is(err, context.Canceled) {
					t.Fatalf("expected cancellation, got %v", err)
				}
				if fallback.calls != 0 {
					t.Fatal("canceled request must not query fallback")
				}
			})
		})
	}
}
