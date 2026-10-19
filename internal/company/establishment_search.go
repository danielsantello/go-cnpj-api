package company

import "errors"

const (
	minTradeNameSearchLength = 3
	maxTradeNameSearchLength = 255
)

var (
	ErrTradeNameTooShort = errors.New(
		"trade name must contain at least 3 characters",
	)
	ErrTradeNameTooLong = errors.New(
		"trade name must contain at most 255 characters",
	)
)

type EstablishmentSummary struct {
	CNPJ          string `json:"cnpj"`
	TradeName     string `json:"trade_name"`
	CorporateName string `json:"corporate_name"`
}

type EstablishmentSearchResult struct {
	Items      []EstablishmentSummary `json:"items"`
	Pagination Pagination             `json:"pagination"`
}
