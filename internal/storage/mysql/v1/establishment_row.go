package v1

import (
	"database/sql"

	"github.com/danielsantello/go-cnpj-api/internal/company"
)

type establishmentRow struct {
	cnpj                                string
	typeCode                            string
	tradeName                           sql.NullString
	registrationStatusCode              sql.NullString
	registrationStatusDate              sql.NullTime
	registrationStatusReasonCode        sql.NullString
	registrationStatusReasonDescription sql.NullString
	specialStatus                       sql.NullString
	specialStatusDate                   sql.NullTime
	activityStartDate                   sql.NullTime
	mainEconomicActivityCode            sql.NullString
	mainEconomicActivityDescription     sql.NullString
	secondaryEconomicActivities         sql.NullString
	streetType                          sql.NullString
	streetName                          sql.NullString
	streetNumber                        sql.NullString
	addressComplement                   sql.NullString
	neighborhood                        sql.NullString
	postalCode                          sql.NullString
	stateCode                           sql.NullString
	municipalityCode                    sql.NullString
	municipalityDescription             sql.NullString
	countryCode                         sql.NullString
	countryDescription                  sql.NullString
	foreignCityName                     sql.NullString
	phoneAreaCode1                      sql.NullString
	phoneNumber1                        sql.NullString
	phoneAreaCode2                      sql.NullString
	phoneNumber2                        sql.NullString
	faxAreaCode                         sql.NullString
	faxNumber                           sql.NullString
	emailAddress                        sql.NullString
}

func (row *establishmentRow) destinations() []any {
	return []any{
		&row.cnpj,
		&row.typeCode,
		&row.tradeName,
		&row.registrationStatusCode,
		&row.registrationStatusDate,
		&row.registrationStatusReasonCode,
		&row.registrationStatusReasonDescription,
		&row.specialStatus,
		&row.specialStatusDate,
		&row.activityStartDate,
		&row.mainEconomicActivityCode,
		&row.mainEconomicActivityDescription,
		&row.secondaryEconomicActivities,
		&row.streetType,
		&row.streetName,
		&row.streetNumber,
		&row.addressComplement,
		&row.neighborhood,
		&row.postalCode,
		&row.stateCode,
		&row.municipalityCode,
		&row.municipalityDescription,
		&row.countryCode,
		&row.countryDescription,
		&row.foreignCityName,
		&row.phoneAreaCode1,
		&row.phoneNumber1,
		&row.phoneAreaCode2,
		&row.phoneNumber2,
		&row.faxAreaCode,
		&row.faxNumber,
		&row.emailAddress,
	}
}

func (row establishmentRow) establishment() company.Establishment {
	result := company.Establishment{
		CNPJ: row.cnpj,
		Type: establishmentType(
			row.typeCode,
		),
		TradeName: nullableString(
			row.tradeName,
		),
		Registration: company.Registration{
			Status: registrationStatus(
				nullableString(row.registrationStatusCode),
			),
			StatusDate: nullableDate(
				row.registrationStatusDate,
			),
			StatusReason: codeDescription(
				nullableString(row.registrationStatusReasonCode),
				nullableString(
					row.registrationStatusReasonDescription,
				),
			),
			SpecialStatus: nullableString(
				row.specialStatus,
			),
			SpecialStatusDate: nullableDate(
				row.specialStatusDate,
			),
		},
		ActivityStartDate: nullableDate(
			row.activityStartDate,
		),
		EconomicActivities: company.EconomicActivities{
			Primary: codeDescription(
				nullableString(row.mainEconomicActivityCode),
				nullableString(
					row.mainEconomicActivityDescription,
				),
			),
			Secondary: make([]company.CodeDescription, 0),
		},
		Address: company.Address{
			StreetType: nullableString(
				row.streetType,
			),
			Street: nullableString(
				row.streetName,
			),
			Number: nullableString(
				row.streetNumber,
			),
			Complement: nullableString(
				row.addressComplement,
			),
			Neighborhood: nullableString(
				row.neighborhood,
			),
			PostalCode: nullableString(
				row.postalCode,
			),
			State: nullableString(
				row.stateCode,
			),
			Municipality: codeDescription(
				nullableString(row.municipalityCode),
				nullableString(row.municipalityDescription),
			),
			Country: codeDescription(
				nullableString(row.countryCode),
				nullableString(row.countryDescription),
			),
			ForeignCityName: nullableString(
				row.foreignCityName,
			),
		},
		Contacts: company.Contacts{
			Phones: make([]company.Phone, 0, 2),
			Fax: phone(
				nullableString(row.faxAreaCode),
				nullableString(row.faxNumber),
			),
			Email: nullableString(
				row.emailAddress,
			),
		},
	}

	firstPhone := phone(
		nullableString(row.phoneAreaCode1),
		nullableString(row.phoneNumber1),
	)
	if firstPhone != nil {
		result.Contacts.Phones = append(
			result.Contacts.Phones,
			*firstPhone,
		)
	}

	secondPhone := phone(
		nullableString(row.phoneAreaCode2),
		nullableString(row.phoneNumber2),
	)
	if secondPhone != nil {
		result.Contacts.Phones = append(
			result.Contacts.Phones,
			*secondPhone,
		)
	}

	return result
}

func (row establishmentRow) secondaryCodes() []string {
	return secondaryEconomicActivityCodes(
		nullableString(row.secondaryEconomicActivities),
	)
}
