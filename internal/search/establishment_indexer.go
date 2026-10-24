package search

import "context"

type EstablishmentSource interface {
	ReadBatch(
		context.Context,
		string,
		int,
	) ([]EstablishmentDocument, error)
}

type EstablishmentIndex interface {
	RecreateTable(context.Context) error
	WriteBatch(context.Context, []EstablishmentDocument) error
}

type EstablishmentIndexOptions struct {
	BatchSize int
	Limit     int
	All       bool
}

type EstablishmentIndexProgress struct {
	Processed int
}

func establishmentIndexProgress(
	value batchIndexProgress,
) EstablishmentIndexProgress {
	return EstablishmentIndexProgress{
		Processed: value.Processed,
	}
}

func BuildEstablishmentIndex(
	ctx context.Context,
	source EstablishmentSource,
	index EstablishmentIndex,
	options EstablishmentIndexOptions,
	progress func(EstablishmentIndexProgress),
) (EstablishmentIndexProgress, error) {
	result, err := buildBatches[EstablishmentDocument](
		ctx,
		source,
		index,
		batchIndexOptions{
			BatchSize: options.BatchSize,
			Limit:     options.Limit,
			All:       options.All,
		},
		func(document EstablishmentDocument) string {
			return document.CNPJ
		},
		func(document *EstablishmentDocument, id uint64) {
			document.ID = id
		},
		func(value batchIndexProgress) {
			if progress != nil {
				progress(establishmentIndexProgress(value))
			}
		},
	)

	return establishmentIndexProgress(result), err
}
