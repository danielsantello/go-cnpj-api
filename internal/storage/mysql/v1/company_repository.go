package v1

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/danielsantello/go-cnpj-api/internal/company"
)

const findCompanyByCNPJQuery = `
	SELECT
		establishment.cnpj_root,
		company.legal_name,
		company.legal_nature_code,
		legal_nature.name,
		company.responsible_qualification_code,
		responsible_qualification.name,
		company.share_capital,
		company.company_size_code,
		company.responsible_federative_entity
	FROM establishments AS establishment
	JOIN companies AS company
		ON company.cnpj_root = establishment.cnpj_root
	LEFT JOIN legal_natures AS legal_nature
		ON legal_nature.code = company.legal_nature_code
	LEFT JOIN partner_qualifications AS responsible_qualification
		ON responsible_qualification.code =
			company.responsible_qualification_code
	WHERE establishment.cnpj = ?
	LIMIT 1
`

type CompanyRepository struct {
	db *sql.DB
}

func NewCompanyRepository(db *sql.DB) *CompanyRepository {
	return &CompanyRepository{
		db: db,
	}
}

func (repository *CompanyRepository) FindByCNPJ(
	ctx context.Context,
	cnpj string,
) (company.Details, error) {
	var details company.Details
	var legalNatureDescription sql.NullString
	var responsibleQualificationDescription sql.NullString
	var shareCapital sql.NullString
	var companySizeCode sql.NullString
	var responsibleFederativeEntity sql.NullString

	err := repository.db.QueryRowContext(
		ctx,
		findCompanyByCNPJQuery,
		cnpj,
	).Scan(
		&details.Company.BasicCNPJ,
		&details.Company.CorporateName,
		&details.Company.LegalNature.Code,
		&legalNatureDescription,
		&details.Company.ResponsibleQualification.Code,
		&responsibleQualificationDescription,
		&shareCapital,
		&companySizeCode,
		&responsibleFederativeEntity,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return company.Details{}, company.ErrCompanyNotFound
	}
	if err != nil {
		return company.Details{}, fmt.Errorf(
			"query company by CNPJ: %w",
			err,
		)
	}

	details.Company.LegalNature.Description =
		nullableString(legalNatureDescription)

	details.Company.ResponsibleQualification.Description =
		nullableString(responsibleQualificationDescription)

	details.Company.ShareCapital = nullableString(shareCapital)

	details.Company.CompanySize =
		companySize(nullableString(companySizeCode))

	details.Company.ResponsibleFederativeEntity =
		nullableString(responsibleFederativeEntity)

	return details, nil
}

func nullableString(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}

	result := value.String

	return &result
}

var _ company.Repository = (*CompanyRepository)(nil)
