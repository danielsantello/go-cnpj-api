package mysqlstorage

import (
	"database/sql"
	"fmt"

	"github.com/danielsantello/go-cnpj-api/internal/company"

	mysqlv1 "github.com/danielsantello/go-cnpj-api/internal/storage/mysql/v1"
)

func NewCompanyRepository(
	db *sql.DB,
	formatVersion uint16,
) (company.Repository, error) {
	switch formatVersion {
	case 1:
		return mysqlv1.NewCompanyRepository(db), nil

	default:
		return nil, fmt.Errorf(
			"unsupported database format version: %d",
			formatVersion,
		)
	}
}
