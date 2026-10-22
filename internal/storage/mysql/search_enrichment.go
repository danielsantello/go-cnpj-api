package mysqlstorage

import (
	"database/sql"
	"fmt"

	"github.com/danielsantello/go-cnpj-api/internal/search"
	mysqlv1 "github.com/danielsantello/go-cnpj-api/internal/storage/mysql/v1"
)

func NewSearchEnrichmentRepository(
	db *sql.DB,
	formatVersion uint16,
) (search.EnrichmentRepository, error) {
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
