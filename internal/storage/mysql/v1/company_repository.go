package v1

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

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
		company.responsible_federative_entity,
		establishment.cnpj,
		establishment.head_office_branch_indicator,
		establishment.trade_name,
		establishment.registration_status_code,
		establishment.registration_status_date,
		establishment.registration_status_reason_code,
		registration_status_reason.name,
		establishment.special_status,
		establishment.special_status_date,
		establishment.activity_start_date,
		establishment.main_economic_activity_code,
		main_economic_activity.name,
		establishment.secondary_economic_activities,
		establishment.street_type,
		establishment.street_name,
		establishment.street_number,
		establishment.address_complement,
		establishment.neighborhood,
		establishment.postal_code,
		establishment.state_code,
		establishment.municipality_code,
		municipality.name,
		establishment.country_code,
		country.name,
		establishment.foreign_city_name,
		establishment.phone_area_code_1,
		establishment.phone_number_1,
		establishment.phone_area_code_2,
		establishment.phone_number_2,
		establishment.fax_area_code,
		establishment.fax_number,
		establishment.email_address
	FROM establishments AS establishment
	JOIN companies AS company
		ON company.cnpj_root = establishment.cnpj_root
	LEFT JOIN legal_natures AS legal_nature
		ON legal_nature.code = company.legal_nature_code
	LEFT JOIN partner_qualifications AS responsible_qualification
		ON responsible_qualification.code =
			company.responsible_qualification_code
	LEFT JOIN registration_status_reasons AS registration_status_reason
		ON registration_status_reason.code =
			establishment.registration_status_reason_code
	LEFT JOIN economic_activities AS main_economic_activity
		ON main_economic_activity.code =
			establishment.main_economic_activity_code
	LEFT JOIN municipalities AS municipality
		ON municipality.code = establishment.municipality_code
	LEFT JOIN countries AS country
		ON country.code = establishment.country_code
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
	options company.FindOptions,
) (company.Details, error) {
	var details company.Details
	var legalNatureDescription sql.NullString
	var responsibleQualificationDescription sql.NullString
	var shareCapital sql.NullString
	var companySizeCode sql.NullString
	var responsibleFederativeEntity sql.NullString
	var establishment establishmentRow

	destinations := []any{
		&details.Company.BasicCNPJ,
		&details.Company.CorporateName,
		&details.Company.LegalNature.Code,
		&legalNatureDescription,
		&details.Company.ResponsibleQualification.Code,
		&responsibleQualificationDescription,
		&shareCapital,
		&companySizeCode,
		&responsibleFederativeEntity,
	}

	destinations = append(
		destinations,
		establishment.destinations()...,
	)

	err := repository.db.QueryRowContext(
		ctx,
		findCompanyByCNPJQuery,
		cnpj,
	).Scan(destinations...)
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

	details.Company.ShareCapital =
		nullableString(shareCapital)

	details.Company.CompanySize =
		companySize(nullableString(companySizeCode))

	details.Company.ResponsibleFederativeEntity =
		nullableString(responsibleFederativeEntity)

	details.Establishment = establishment.establishment()

	secondaryActivities, err :=
		repository.findSecondaryEconomicActivities(
			ctx,
			establishment.secondaryCodes(),
		)
	if err != nil {
		return company.Details{}, fmt.Errorf(
			"find secondary economic activities: %w",
			err,
		)
	}

	details.Establishment.EconomicActivities.Secondary =
		secondaryActivities

	simpleTax, err := repository.findSimpleTax(
		ctx,
		details.Company.BasicCNPJ,
	)
	if err != nil {
		return company.Details{}, fmt.Errorf(
			"find simple tax: %w",
			err,
		)
	}

	details.SimpleTax = simpleTax

	partners, err := repository.findPartners(
		ctx,
		details.Company.BasicCNPJ,
	)
	if err != nil {
		return company.Details{}, fmt.Errorf(
			"find partners: %w",
			err,
		)
	}

	details.Partners = partners

	if options.IncludeEstablishments {
		establishments, err := repository.findEstablishments(
			ctx,
			details.Company.BasicCNPJ,
			options.Page,
			options.PageSize,
		)
		if err != nil {
			return company.Details{}, fmt.Errorf(
				"find establishments: %w",
				err,
			)
		}

		details.Establishments = &establishments
	}

	return details, nil
}

func nullableString(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}

	result := value.String

	return &result
}

func nullableDate(value sql.NullTime) *string {
	if !value.Valid {
		return nil
	}

	result := value.Time.Format(time.DateOnly)

	return &result
}

var _ company.Repository = (*CompanyRepository)(nil)
