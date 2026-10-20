package search

import (
	"context"
	"fmt"
)

type IndexOptions struct {
	BatchSize      int
	Limit          int
	All            bool
	AfterBasicCNPJ string
}

type IndexProgress struct {
	Processed     int
	LastBasicCNPJ string
}

func BuildIndex(
	ctx context.Context,
	source Source,
	index Index,
	options IndexOptions,
	progress func(IndexProgress),
) (IndexProgress, error) {
	result := IndexProgress{
		LastBasicCNPJ: options.AfterBasicCNPJ,
	}

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

	if err := index.EnsureTable(ctx); err != nil {
		return result, fmt.Errorf("ensure search table: %w", err)
	}

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

		documents, err := source.ReadBatch(
			ctx,
			result.LastBasicCNPJ,
			batchSize,
		)
		if err != nil {
			return result, fmt.Errorf("read search batch: %w", err)
		}
		if len(documents) == 0 {
			break
		}
		if len(documents) > batchSize {
			return result, fmt.Errorf("source returned more documents than requested")
		}

		cursor := result.LastBasicCNPJ
		for _, document := range documents {
			if document.BasicCNPJ <= cursor {
				return result, fmt.Errorf("source returned an unordered or repeated basic CNPJ")
			}
			cursor = document.BasicCNPJ
		}

		if err := index.WriteBatch(ctx, documents); err != nil {
			return result, fmt.Errorf("write search batch: %w", err)
		}

		result.Processed += len(documents)
		result.LastBasicCNPJ = cursor

		if progress != nil {
			progress(result)
		}
	}

	return result, nil
}
