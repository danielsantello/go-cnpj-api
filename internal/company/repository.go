package company

import (
	"context"
	"errors"
)

var ErrCompanyNotFound = errors.New("company not found")

type Details struct {
	Company       Company       `json:"company"`
	Establishment Establishment `json:"establishment"`
	SimpleTax     *SimpleTax    `json:"simple_tax"`
	Partners      []Partner     `json:"partners"`
}

type Repository interface {
	FindByCNPJ(
		ctx context.Context,
		cnpj string,
	) (Details, error)
}
