package company

import "errors"

const (
	minCorporateNameSearchLength = 3
	maxCorporateNameSearchLength = 255
)

var (
	ErrCorporateNameTooShort = errors.New(
		"corporate name must contain at least 3 characters",
	)
	ErrCorporateNameTooLong = errors.New(
		"corporate name must contain at most 255 characters",
	)
)

type SearchOptions struct {
	Name     string
	Page     int
	PageSize int
}

type Summary struct {
	CNPJ               string           `json:"cnpj"`
	CorporateName      string           `json:"corporate_name"`
	Municipality       *CodeDescription `json:"municipality"`
	State              *string          `json:"state"`
	RegistrationStatus *CodeDescription `json:"registration_status"`
}

type SearchResult struct {
	Items      []Summary  `json:"items"`
	Pagination Pagination `json:"pagination"`
}
