package search

import "context"

type IndexOptions struct {
	BatchSize int
	Limit     int
	All       bool
}

type IndexProgress struct {
	Processed int
}

func companyIndexProgress(value batchIndexProgress) IndexProgress {
	return IndexProgress{
		Processed: value.Processed,
	}
}

func BuildIndex(
	ctx context.Context,
	source Source,
	index Index,
	options IndexOptions,
	progress func(IndexProgress),
) (IndexProgress, error) {
	result, err := buildBatches[Document](
		ctx,
		source,
		index,
		batchIndexOptions{
			BatchSize: options.BatchSize,
			Limit:     options.Limit,
			All:       options.All,
		},
		func(document Document) string {
			return document.BasicCNPJ
		},
		func(document *Document, id uint64) {
			document.ID = id
		},
		func(value batchIndexProgress) {
			if progress != nil {
				progress(companyIndexProgress(value))
			}
		},
	)

	return companyIndexProgress(result), err
}
