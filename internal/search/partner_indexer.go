package search

import "context"

type PartnerIndexOptions struct {
	BatchSize int
	Limit     int
	All       bool
}

type PartnerIndexProgress struct {
	Processed int
}

func partnerIndexProgress(value batchIndexProgress) PartnerIndexProgress {
	return PartnerIndexProgress{
		Processed: value.Processed,
	}
}

func BuildPartnerIndex(
	ctx context.Context,
	source PartnerSource,
	index PartnerIndex,
	options PartnerIndexOptions,
	progress func(PartnerIndexProgress),
) (PartnerIndexProgress, error) {
	result, err := buildBatches[PartnerDocument](
		ctx,
		source,
		index,
		batchIndexOptions{
			BatchSize: options.BatchSize,
			Limit:     options.Limit,
			All:       options.All,
		},
		func(document PartnerDocument) string {
			return document.SourceKey
		},
		func(document *PartnerDocument, id uint64) {
			document.ID = id
		},
		func(value batchIndexProgress) {
			if progress != nil {
				progress(partnerIndexProgress(value))
			}
		},
	)

	return partnerIndexProgress(result), err
}
