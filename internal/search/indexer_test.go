package search

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type sourceStub struct {
	documents []Document
	cursors   []string
	limits    []int
}

func (source *sourceStub) ReadBatch(
	_ context.Context,
	after string,
	limit int,
) ([]Document, error) {
	source.cursors = append(source.cursors, after)
	source.limits = append(source.limits, limit)

	var result []Document
	for _, document := range source.documents {
		if document.BasicCNPJ > after {
			result = append(result, document)
			if len(result) == limit {
				break
			}
		}
	}
	return result, nil
}

type indexStub struct {
	batches   [][]Document
	failBatch int
	err       error
}

func (index *indexStub) EnsureTable(context.Context) error {
	return nil
}

func (index *indexStub) WriteBatch(_ context.Context, documents []Document) error {
	if index.failBatch > 0 && len(index.batches)+1 == index.failBatch {
		return index.err
	}
	index.batches = append(index.batches, append([]Document(nil), documents...))
	return nil
}

func testDocuments() []Document {
	return []Document{
		{ID: 2, BasicCNPJ: "00000001"},
		{ID: 3, BasicCNPJ: "00000002"},
		{ID: 4, BasicCNPJ: "00000003"},
		{ID: 5, BasicCNPJ: "00000004"},
	}
}

func TestBuildIndexRespectsLimitAndCursor(t *testing.T) {
	source := &sourceStub{documents: testDocuments()}
	index := &indexStub{}
	var progress []IndexProgress

	result, err := BuildIndex(
		context.Background(), source, index,
		IndexOptions{BatchSize: 2, Limit: 3},
		func(value IndexProgress) { progress = append(progress, value) },
	)
	if err != nil {
		t.Fatal(err)
	}

	if result.Processed != 3 || result.LastBasicCNPJ != "00000003" {
		t.Fatalf("unexpected result: %#v", result)
	}
	if !reflect.DeepEqual(source.cursors, []string{"", "00000002"}) ||
		!reflect.DeepEqual(source.limits, []int{2, 1}) {
		t.Fatalf("unexpected reads: %v %v", source.cursors, source.limits)
	}
	if len(index.batches) != 2 || len(index.batches[1]) != 1 ||
		len(progress) != 2 || progress[1] != result {
		t.Fatal("unexpected batches or progress")
	}
}

func TestBuildIndexPreservesProgressOnWriteFailure(t *testing.T) {
	expected := errors.New("write failed")
	index := &indexStub{failBatch: 2, err: expected}
	var progress []IndexProgress

	result, err := BuildIndex(
		context.Background(),
		&sourceStub{documents: testDocuments()},
		index,
		IndexOptions{BatchSize: 2, Limit: 4},
		func(value IndexProgress) { progress = append(progress, value) },
	)

	if !errors.Is(err, expected) {
		t.Fatalf("expected write error, got %v", err)
	}
	if result.Processed != 2 || result.LastBasicCNPJ != "00000002" ||
		len(progress) != 1 {
		t.Fatalf("progress advanced past failed batch: %#v", result)
	}
}

func TestBuildIndexStopsAtEndOfSource(t *testing.T) {
	result, err := BuildIndex(
		context.Background(),
		&sourceStub{documents: testDocuments()},
		&indexStub{},
		IndexOptions{BatchSize: 3, Limit: 10},
		nil,
	)
	if err != nil || result.Processed != 4 {
		t.Fatalf("unexpected result: %#v, error: %v", result, err)
	}
}

func TestBuildIndexRejectsInvalidOptions(t *testing.T) {
	for _, options := range []IndexOptions{
		{BatchSize: 0, Limit: 1},
		{BatchSize: 1, Limit: 0},
		{BatchSize: -1, Limit: 1},
	} {
		_, err := BuildIndex(
			context.Background(), &sourceStub{}, &indexStub{}, options, nil,
		)
		if err == nil {
			t.Fatalf("expected invalid options error for %#v", options)
		}
	}
}

func TestBuildIndexHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	source := &sourceStub{}

	_, err := BuildIndex(
		ctx, source, &indexStub{},
		IndexOptions{BatchSize: 2, Limit: 4}, nil,
	)
	if !errors.Is(err, context.Canceled) || len(source.cursors) != 0 {
		t.Fatalf("expected cancellation before reading, got %v", err)
	}
}

func TestBuildIndexProcessesFullSource(t *testing.T) {
	source := &sourceStub{documents: testDocuments()}
	index := &indexStub{}

	result, err := BuildIndex(
		context.Background(), source, index,
		IndexOptions{BatchSize: 3, All: true},
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	if result.Processed != 4 || result.LastBasicCNPJ != "00000004" {
		t.Fatalf("unexpected full indexing result: %#v", result)
	}
	if len(index.batches) != 2 ||
		len(index.batches[0]) != 3 ||
		len(index.batches[1]) != 1 {
		t.Fatalf("unexpected batches: %#v", index.batches)
	}
}

func TestBuildIndexRejectsFullSourceWithLimit(t *testing.T) {
	_, err := BuildIndex(
		context.Background(),
		&sourceStub{},
		&indexStub{},
		IndexOptions{BatchSize: 2, Limit: 10, All: true},
		nil,
	)
	if err == nil {
		t.Fatal("expected error when full indexing specifies a limit")
	}
}

func TestBuildIndexResumesAfterCursor(t *testing.T) {
	source := &sourceStub{documents: testDocuments()}
	index := &indexStub{}

	result, err := BuildIndex(
		context.Background(), source, index,
		IndexOptions{
			BatchSize:      2,
			All:            true,
			AfterBasicCNPJ: "00000002",
		},
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	if result.Processed != 2 || result.LastBasicCNPJ != "00000004" {
		t.Fatalf("unexpected resumed result: %#v", result)
	}
	if len(index.batches) != 1 ||
		index.batches[0][0].BasicCNPJ != "00000003" {
		t.Fatalf("unexpected resumed batches: %#v", index.batches)
	}
}
