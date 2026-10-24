package search

import (
	"context"

	"github.com/danielsantello/go-cnpj-api/internal/company"
)

type EstablishmentSummary struct {
	TradeName          *string
	Municipality       *company.CodeDescription
	State              *string
	RegistrationStatus *company.CodeDescription
}

type EnrichmentRepository interface {
	FindEstablishmentSummaries(
		context.Context,
		[]string,
	) (map[string]EstablishmentSummary, error)
}
