package search

import (
	"context"
	"fmt"
)

type batchIndexOptions struct {
	BatchSize int
	Limit     int
	All       bool
}

type batchIndexProgress struct {
	Processed int
}

type batchSource[T any] interface {
	ReadBatch(context.Context, string, int) ([]T, error)
}

type batchDestination[T any] interface {
	RecreateTable(context.Context) error
	WriteBatch(context.Context, []T) error
}

func buildBatches[T any](
	ctx context.Context,
	source batchSource[T],
	destination batchDestination[T],
	options batchIndexOptions,
	key func(T) string,
	assignID func(*T, uint64),
	progress func(batchIndexProgress),
) (batchIndexProgress, error) {
	result := batchIndexProgress{}

	if options.BatchSize <= 0 {
		return result, fmt.Errorf("batch size must be greater than zero")
	}
	if options.All {
		if options.Limit != 0 {
			return result, fmt.Errorf("full indexing must not specify a limit")
		}
	} else if options.Limit <= 0 {
		return result, fmt.Errorf("indexing limit must be greater than zero")
	}

	if err := ctx.Err(); err != nil {
		return result, err
	}
	if err := destination.RecreateTable(ctx); err != nil {
		return result, fmt.Errorf("recreate search table: %w", err)
	}
	cursor := ""
	var lastID uint64
	for options.All || result.Processed < options.Limit {
		if err := ctx.Err(); err != nil {
			return result, err
		}

		batchSize := options.BatchSize
		if !options.All {
			if remaining := options.Limit - result.Processed; remaining < batchSize {
				batchSize = remaining
			}
		}

		documents, err := source.ReadBatch(ctx, cursor, batchSize)
		if err != nil {
			return result, fmt.Errorf("read search batch: %w", err)
		}
		if len(documents) == 0 {
			break
		}
		if len(documents) > batchSize {
			return result, fmt.Errorf("source returned more documents than requested")
		}

		for position := range documents {
			current := key(documents[position])
			if current <= cursor {
				return result, fmt.Errorf(
					"source returned an unordered or repeated indexing key: previous=%q current=%q",
					cursor, current,
				)
			}
			if lastID == ^uint64(0) {
				return result, fmt.Errorf("document ID sequence exhausted")
			}

			lastID++
			assignID(&documents[position], lastID)
			cursor = current
		}

		if err := ctx.Err(); err != nil {
			return result, err
		}
		if err := destination.WriteBatch(ctx, documents); err != nil {
			return result, fmt.Errorf("write search batch: %w", err)
		}

		result.Processed += len(documents)

		if progress != nil {
			progress(result)
		}
	}

	return result, nil
}
