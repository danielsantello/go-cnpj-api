package company

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

var (
	ErrPartnerNameTooShort    = errors.New("partner name must contain at least 3 characters")
	ErrPartnerNameTooLong     = errors.New("partner name must contain at most 255 characters")
	ErrPartnerDocumentInvalid = errors.New("partner document must contain between 1 and 20 characters")
)

type PartnerSearchOptions struct {
	Name     string
	Document *string
	Page     int
	PageSize int
}

type PartnerSummary struct {
	Name               string           `json:"name"`
	Document           *string          `json:"document"`
	CNPJ               string           `json:"cnpj"`
	CorporateName      string           `json:"corporate_name"`
	Municipality       *CodeDescription `json:"municipality"`
	State              *string          `json:"state"`
	RegistrationStatus *CodeDescription `json:"registration_status"`
}

type PartnerSearchResult struct {
	Items      []PartnerSummary `json:"items"`
	Pagination Pagination       `json:"pagination"`
}

type partnerSearchRepository interface {
	SearchByPartnerName(
		context.Context,
		PartnerSearchOptions,
	) (PartnerSearchResult, error)
}

func searchByPartnerName(
	ctx context.Context,
	repository partnerSearchRepository,
	options PartnerSearchOptions,
) (PartnerSearchResult, error) {
	options.Name = strings.TrimSpace(options.Name)

	length := utf8.RuneCountInString(options.Name)
	if length < 3 {
		return PartnerSearchResult{}, ErrPartnerNameTooShort
	}
	if length > 255 {
		return PartnerSearchResult{}, ErrPartnerNameTooLong
	}

	if options.Document != nil {
		document := strings.TrimSpace(*options.Document)
		length := utf8.RuneCountInString(document)
		if length == 0 || length > 20 {
			return PartnerSearchResult{}, ErrPartnerDocumentInvalid
		}
		options.Document = &document
	}

	result, err := repository.SearchByPartnerName(ctx, options)
	if err != nil {
		return PartnerSearchResult{}, fmt.Errorf("search companies by partner name: %w", err)
	}

	return result, nil
}

func (service *Service) SearchByPartnerName(
	ctx context.Context,
	options PartnerSearchOptions,
) (PartnerSearchResult, error) {
	return searchByPartnerName(ctx, service.repository, options)
}
