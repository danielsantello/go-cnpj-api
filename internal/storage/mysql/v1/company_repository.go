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
) (company.Details, error) {
	var details company.Details
	var legalNatureDescription sql.NullString
	var responsibleQualificationDescription sql.NullString
	var shareCapital sql.NullString
	var companySizeCode sql.NullString
	var responsibleFederativeEntity sql.NullString
	var establishmentTypeCode string
	var tradeName sql.NullString
	var registrationStatusCode sql.NullString
	var registrationStatusDate sql.NullTime
	var registrationStatusReasonCode sql.NullString
	var registrationStatusReasonDescription sql.NullString
	var specialStatus sql.NullString
	var specialStatusDate sql.NullTime
	var activityStartDate sql.NullTime
	var mainEconomicActivityCode sql.NullString
	var mainEconomicActivityDescription sql.NullString
	var secondaryEconomicActivities sql.NullString
	var streetType sql.NullString
	var streetName sql.NullString
	var streetNumber sql.NullString
	var addressComplement sql.NullString
	var neighborhood sql.NullString
	var postalCode sql.NullString
	var stateCode sql.NullString
	var municipalityCode sql.NullString
	var municipalityDescription sql.NullString
	var countryCode sql.NullString
	var countryDescription sql.NullString
	var foreignCityName sql.NullString
	var phoneAreaCode1 sql.NullString
	var phoneNumber1 sql.NullString
	var phoneAreaCode2 sql.NullString
	var phoneNumber2 sql.NullString
	var faxAreaCode sql.NullString
	var faxNumber sql.NullString
	var emailAddress sql.NullString

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
		&details.Establishment.CNPJ,
		&establishmentTypeCode,
		&tradeName,
		&registrationStatusCode,
		&registrationStatusDate,
		&registrationStatusReasonCode,
		&registrationStatusReasonDescription,
		&specialStatus,
		&specialStatusDate,
		&activityStartDate,
		&mainEconomicActivityCode,
		&mainEconomicActivityDescription,
		&secondaryEconomicActivities,
		&streetType,
		&streetName,
		&streetNumber,
		&addressComplement,
		&neighborhood,
		&postalCode,
		&stateCode,
		&municipalityCode,
		&municipalityDescription,
		&countryCode,
		&countryDescription,
		&foreignCityName,
		&phoneAreaCode1,
		&phoneNumber1,
		&phoneAreaCode2,
		&phoneNumber2,
		&faxAreaCode,
		&faxNumber,
		&emailAddress,
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

	details.Company.ShareCapital =
		nullableString(shareCapital)

	details.Company.CompanySize =
		companySize(nullableString(companySizeCode))

	details.Company.ResponsibleFederativeEntity =
		nullableString(responsibleFederativeEntity)

	details.Establishment.Type =
		establishmentType(establishmentTypeCode)

	details.Establishment.TradeName =
		nullableString(tradeName)

	details.Establishment.Registration.Status =
		registrationStatus(nullableString(registrationStatusCode))

	details.Establishment.Registration.StatusDate =
		nullableDate(registrationStatusDate)

	details.Establishment.Registration.StatusReason =
		codeDescription(
			nullableString(registrationStatusReasonCode),
			nullableString(registrationStatusReasonDescription),
		)

	details.Establishment.Registration.SpecialStatus =
		nullableString(specialStatus)

	details.Establishment.Registration.SpecialStatusDate =
		nullableDate(specialStatusDate)

	details.Establishment.ActivityStartDate =
		nullableDate(activityStartDate)

	details.Establishment.EconomicActivities.Primary =
		codeDescription(
			nullableString(mainEconomicActivityCode),
			nullableString(mainEconomicActivityDescription),
		)

	secondaryCodes := secondaryEconomicActivityCodes(
		nullableString(secondaryEconomicActivities),
	)

	secondaryActivities, err :=
		repository.findSecondaryEconomicActivities(
			ctx,
			secondaryCodes,
		)
	if err != nil {
		return company.Details{}, fmt.Errorf(
			"find secondary economic activities: %w",
			err,
		)
	}

	details.Establishment.EconomicActivities.Secondary =
		secondaryActivities

	details.Establishment.Address.StreetType =
		nullableString(streetType)

	details.Establishment.Address.Street =
		nullableString(streetName)

	details.Establishment.Address.Number =
		nullableString(streetNumber)

	details.Establishment.Address.Complement =
		nullableString(addressComplement)

	details.Establishment.Address.Neighborhood =
		nullableString(neighborhood)

	details.Establishment.Address.PostalCode =
		nullableString(postalCode)

	details.Establishment.Address.State =
		nullableString(stateCode)

	details.Establishment.Address.Municipality =
		codeDescription(
			nullableString(municipalityCode),
			nullableString(municipalityDescription),
		)

	details.Establishment.Address.Country =
		codeDescription(
			nullableString(countryCode),
			nullableString(countryDescription),
		)

	details.Establishment.Address.ForeignCityName =
		nullableString(foreignCityName)

	details.Establishment.Contacts.Phones =
		make([]company.Phone, 0, 2)

	firstPhone := phone(
		nullableString(phoneAreaCode1),
		nullableString(phoneNumber1),
	)
	if firstPhone != nil {
		details.Establishment.Contacts.Phones = append(
			details.Establishment.Contacts.Phones,
			*firstPhone,
		)
	}

	secondPhone := phone(
		nullableString(phoneAreaCode2),
		nullableString(phoneNumber2),
	)
	if secondPhone != nil {
		details.Establishment.Contacts.Phones = append(
			details.Establishment.Contacts.Phones,
			*secondPhone,
		)
	}

	details.Establishment.Contacts.Fax =
		phone(
			nullableString(faxAreaCode),
			nullableString(faxNumber),
		)

	details.Establishment.Contacts.Email =
		nullableString(emailAddress)

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
