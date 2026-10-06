package v1

import (
	"context"
	"fmt"

	"github.com/danielsantello/go-cnpj-api/internal/company"
)

const countEstablishmentsByCNPJRootQuery = `
	SELECT COUNT(*)
	FROM establishments
	WHERE cnpj_root = ?
`

const findEstablishmentsByCNPJRootQuery = `
	SELECT
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
	WHERE establishment.cnpj_root = ?
	ORDER BY establishment.cnpj
	LIMIT ?
	OFFSET ?
`

func (repository *CompanyRepository) findEstablishments(
	ctx context.Context,
	cnpjRoot string,
	page int,
	pageSize int,
) (company.EstablishmentList, error) {
	var totalItems int64

	err := repository.db.QueryRowContext(
		ctx,
		countEstablishmentsByCNPJRootQuery,
		cnpjRoot,
	).Scan(&totalItems)
	if err != nil {
		return company.EstablishmentList{}, fmt.Errorf(
			"count establishments by CNPJ root: %w",
			err,
		)
	}

	offset := int64(page-1) * int64(pageSize)

	rows, err := repository.db.QueryContext(
		ctx,
		findEstablishmentsByCNPJRootQuery,
		cnpjRoot,
		pageSize,
		offset,
	)
	if err != nil {
		return company.EstablishmentList{}, fmt.Errorf(
			"query establishments by CNPJ root: %w",
			err,
		)
	}
	defer rows.Close()

	records := make([]establishmentRow, 0, pageSize)

	for rows.Next() {
		var record establishmentRow

		if err := rows.Scan(record.destinations()...); err != nil {
			return company.EstablishmentList{}, fmt.Errorf(
				"scan establishment: %w",
				err,
			)
		}

		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return company.EstablishmentList{}, fmt.Errorf(
			"iterate establishments: %w",
			err,
		)
	}

	descriptions, err :=
		repository.findSecondaryEconomicActivityDescriptions(
			ctx,
			records,
		)
	if err != nil {
		return company.EstablishmentList{}, fmt.Errorf(
			"find secondary economic activity descriptions: %w",
			err,
		)
	}

	items := make([]company.Establishment, 0, len(records))

	for _, record := range records {
		establishment := record.establishment()

		for _, code := range record.secondaryCodes() {
			activity := company.CodeDescription{
				Code: code,
			}

			description, exists := descriptions[code]
			if exists {
				activity.Description = description
			}

			establishment.EconomicActivities.Secondary = append(
				establishment.EconomicActivities.Secondary,
				activity,
			)
		}

		items = append(items, establishment)
	}

	return company.EstablishmentList{
		Items: items,
		Pagination: company.NewPagination(
			page,
			pageSize,
			totalItems,
		),
	}, nil
}

func (
	repository *CompanyRepository,
) findSecondaryEconomicActivityDescriptions(
	ctx context.Context,
	records []establishmentRow,
) (map[string]*string, error) {
	uniqueCodes := make([]string, 0)
	seenCodes := make(map[string]struct{})

	for _, record := range records {
		for _, code := range record.secondaryCodes() {
			if _, exists := seenCodes[code]; exists {
				continue
			}

			seenCodes[code] = struct{}{}
			uniqueCodes = append(uniqueCodes, code)
		}
	}

	descriptions := make(map[string]*string, len(uniqueCodes))

	if len(uniqueCodes) == 0 {
		return descriptions, nil
	}

	activities, err :=
		repository.findSecondaryEconomicActivities(
			ctx,
			uniqueCodes,
		)
	if err != nil {
		return nil, err
	}

	for _, activity := range activities {
		if _, exists := descriptions[activity.Code]; exists {
			continue
		}

		descriptions[activity.Code] = activity.Description
	}

	return descriptions, nil
}
