package v1

const FormatVersion uint16 = 1

type tableContract struct {
	name    string
	columns []string
}

var schemaContract = []tableContract{
	{
		name: "companies",
		columns: []string{
			"cnpj_root",
			"legal_name",
			"legal_nature_code",
			"responsible_qualification_code",
			"share_capital",
			"company_size_code",
			"responsible_federative_entity",
		},
	},
	{
		name: "countries",
		columns: []string{
			"code",
			"name",
		},
	},
	{
		name: "economic_activities",
		columns: []string{
			"code",
			"name",
		},
	},
	{
		name: "establishments",
		columns: []string{
			"cnpj",
			"cnpj_root",
			"branch_number",
			"check_digits",
			"head_office_branch_indicator",
			"trade_name",
			"registration_status_code",
			"registration_status_date",
			"registration_status_reason_code",
			"foreign_city_name",
			"country_code",
			"activity_start_date",
			"main_economic_activity_code",
			"secondary_economic_activities",
			"street_type",
			"street_name",
			"street_number",
			"address_complement",
			"neighborhood",
			"postal_code",
			"state_code",
			"municipality_code",
			"phone_area_code_1",
			"phone_number_1",
			"phone_area_code_2",
			"phone_number_2",
			"fax_area_code",
			"fax_number",
			"email_address",
			"special_status",
			"special_status_date",
		},
	},
	{
		name: "legal_natures",
		columns: []string{
			"code",
			"name",
		},
	},
	{
		name: "municipalities",
		columns: []string{
			"code",
			"name",
		},
	},
	{
		name: "partner_qualifications",
		columns: []string{
			"code",
			"name",
		},
	},
	{
		name: "partners",
		columns: []string{
			"cnpj_root",
			"partner_type_code",
			"partner_name",
			"partner_document",
			"partner_qualification_code",
			"entry_date",
			"country_code",
			"legal_representative_document",
			"legal_representative_name",
			"legal_representative_qualification_code",
			"age_range_code",
		},
	},
	{
		name: "registration_status_reasons",
		columns: []string{
			"code",
			"name",
		},
	},
	{
		name: "schema_metadata",
		columns: []string{
			"id",
			"format_version",
			"reference_year",
			"reference_month",
			"created_at_utc",
		},
	},
	{
		name: "simple_tax_options",
		columns: []string{
			"cnpj_root",
			"simple_option_indicator",
			"simple_option_date",
			"simple_exclusion_date",
			"mei_option_indicator",
			"mei_option_date",
			"mei_exclusion_date",
		},
	},
}
