package v1

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/danielsantello/go-cnpj-api/internal/company"
)

const findPartnersByCNPJRootQuery = `
	SELECT
		partner.partner_type_code,
		partner.partner_name,
		partner.partner_document,
		partner.partner_qualification_code,
		(
			SELECT qualification.name
			FROM partner_qualifications AS qualification
			WHERE qualification.code =
				partner.partner_qualification_code
			LIMIT 1
		),
		partner.entry_date,
		partner.country_code,
		(
			SELECT country.name
			FROM countries AS country
			WHERE country.code = partner.country_code
			LIMIT 1
		),
		partner.legal_representative_document,
		partner.legal_representative_name,
		partner.legal_representative_qualification_code,
		(
			SELECT qualification.name
			FROM partner_qualifications AS qualification
			WHERE qualification.code =
				partner.legal_representative_qualification_code
			LIMIT 1
		),
		partner.age_range_code
	FROM partners AS partner
	WHERE partner.cnpj_root = ?
`

func (repository *CompanyRepository) findPartners(
	ctx context.Context,
	cnpjRoot string,
) ([]company.Partner, error) {
	result := make([]company.Partner, 0)

	rows, err := repository.db.QueryContext(
		ctx,
		findPartnersByCNPJRootQuery,
		cnpjRoot,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"query partners by CNPJ root: %w",
			err,
		)
	}
	defer rows.Close()

	for rows.Next() {
		var partner company.Partner
		var partnerDocument sql.NullString
		var partnerQualificationCode sql.NullString
		var partnerQualificationDescription sql.NullString
		var entryDate sql.NullTime
		var countryCode sql.NullString
		var countryDescription sql.NullString
		var legalRepresentativeDocument sql.NullString
		var legalRepresentativeName sql.NullString
		var legalRepresentativeQualificationCode sql.NullString
		var legalRepresentativeQualificationDescription sql.NullString
		var ageRangeCode sql.NullString

		var partnerTypeCode string

		err := rows.Scan(
			&partnerTypeCode,
			&partner.Name,
			&partnerDocument,
			&partnerQualificationCode,
			&partnerQualificationDescription,
			&entryDate,
			&countryCode,
			&countryDescription,
			&legalRepresentativeDocument,
			&legalRepresentativeName,
			&legalRepresentativeQualificationCode,
			&legalRepresentativeQualificationDescription,
			&ageRangeCode,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"scan partner: %w",
				err,
			)
		}

		partner.Type = partnerType(partnerTypeCode)
		partner.Document = nullableString(partnerDocument)
		partner.Qualification = codeDescription(
			nullableString(partnerQualificationCode),
			nullableString(partnerQualificationDescription),
		)
		partner.EntryDate = nullableDate(entryDate)
		partner.Country = codeDescription(
			nullableString(countryCode),
			nullableString(countryDescription),
		)
		partner.AgeRange = ageRange(
			nullableString(ageRangeCode),
		)

		legalRepresentativeQualification := codeDescription(
			nullableString(
				legalRepresentativeQualificationCode,
			),
			nullableString(
				legalRepresentativeQualificationDescription,
			),
		)

		legalRepresentativeDocumentValue :=
			nullableString(legalRepresentativeDocument)

		legalRepresentativeNameValue :=
			nullableString(legalRepresentativeName)

		if legalRepresentativeDocumentValue != nil ||
			legalRepresentativeNameValue != nil ||
			legalRepresentativeQualification != nil {
			partner.LegalRepresentative =
				&company.LegalRepresentative{
					Document:      legalRepresentativeDocumentValue,
					Name:          legalRepresentativeNameValue,
					Qualification: legalRepresentativeQualification,
				}
		}

		result = append(result, partner)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate partners: %w",
			err,
		)
	}

	return result, nil
}
