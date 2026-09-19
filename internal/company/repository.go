package company

import (
	"context"
	"errors"
)

var ErrCompanyNotFound = errors.New("company not found")

type Details struct {
	Company Company `json:"company"`
}

type Repository interface {
	FindByCNPJ(
		ctx context.Context,
		cnpj string,
	) (Details, error)
}
