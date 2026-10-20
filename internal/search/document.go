package search

import "context"

type Document struct {
	ID            uint64
	BasicCNPJ     string
	CNPJ          string
	CorporateName string
	TradeNames    []string
	PartnerNames  []string
}

type Source interface {
	ReadBatch(
		ctx context.Context,
		afterBasicCNPJ string,
		limit int,
	) ([]Document, error)
}

type Index interface {
	EnsureTable(ctx context.Context) error
	WriteBatch(ctx context.Context, documents []Document) error
}
