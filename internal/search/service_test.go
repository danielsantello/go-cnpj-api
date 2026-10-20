package search

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type searchRepositoryStub struct {
	options Options
	calls   int
	result  Result
	err     error
}

func (repository *searchRepositoryStub) Search(
	_ context.Context,
	options Options,
) (Result, error) {
	repository.calls++
	repository.options = options
	return repository.result, repository.err
}

func TestSearchServiceNormalizesName(t *testing.T) {
	repository := &searchRepositoryStub{
		result: Result{Items: []Item{{CorporateName: "BANCO DO BRASIL SA"}}},
	}

	result, err := NewService(repository).Search(
		context.Background(),
		Options{Name: "  banco brasil  ", Page: 2, PageSize: 5},
	)
	if err != nil {
		t.Fatal(err)
	}
	if repository.options != (Options{
		Name: "banco brasil", Page: 2, PageSize: 5,
	}) {
		t.Fatalf("unexpected options: %#v", repository.options)
	}
	if repository.calls != 1 || len(result.Items) != 1 {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestSearchServiceRejectsInvalidOptions(t *testing.T) {
	tests := []struct {
		name     string
		options  Options
		expected error
	}{
		{"short name", Options{" AB ", 1, 20}, ErrInvalidName},
		{"long name", Options{strings.Repeat("Á", 256), 1, 20}, ErrInvalidName},
		{"zero page", Options{"banco", 0, 20}, ErrInvalidPagination},
		{"zero size", Options{"banco", 1, 0}, ErrInvalidPagination},
		{"large size", Options{"banco", 1, 101}, ErrInvalidPagination},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &searchRepositoryStub{}
			_, err := NewService(repository).Search(
				context.Background(), test.options,
			)
			if !errors.Is(err, test.expected) {
				t.Fatalf("expected %v, got %v", test.expected, err)
			}
			if repository.calls != 0 {
				t.Fatal("invalid options must not reach repository")
			}
		})
	}
}

func TestSearchServiceReturnsDisabled(t *testing.T) {
	_, err := NewService(nil).Search(
		context.Background(),
		Options{Name: "banco", Page: 1, PageSize: 20},
	)
	if !errors.Is(err, ErrDisabled) {
		t.Fatalf("expected ErrDisabled, got %v", err)
	}
}

func TestSearchServicePreservesRepositoryError(t *testing.T) {
	expected := errors.New("connection failed")
	repository := &searchRepositoryStub{err: expected}

	_, err := NewService(repository).Search(
		context.Background(),
		Options{Name: "banco", Page: 1, PageSize: 20},
	)
	if !errors.Is(err, expected) {
		t.Fatalf("expected wrapped repository error, got %v", err)
	}
}
