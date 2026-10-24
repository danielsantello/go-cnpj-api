package search

import "context"

type PartnerDocument struct {
	ID            uint64
	SourceKey     string
	BasicCNPJ     string
	CNPJ          string
	CorporateName string
	Name          string
	Document      *string
}

type PartnerSource interface {
	ReadBatch(context.Context, string, int) ([]PartnerDocument, error)
}

type PartnerIndex interface {
	RecreateTable(context.Context) error
	WriteBatch(context.Context, []PartnerDocument) error
}
