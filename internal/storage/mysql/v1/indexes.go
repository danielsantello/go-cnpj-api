package v1

type IndexDefinition struct {
	Name    string
	Table   string
	Columns []string
}

func RequiredIndexes() []IndexDefinition {
	return []IndexDefinition{
		{
			Name:    "idx_establishments_cnpj",
			Table:   "establishments",
			Columns: []string{"cnpj"},
		},
		{
			Name:    "idx_establishments_cnpj_root",
			Table:   "establishments",
			Columns: []string{"cnpj_root"},
		},
		{
			Name:    "idx_companies_cnpj_root",
			Table:   "companies",
			Columns: []string{"cnpj_root"},
		},
		{
			Name:    "idx_legal_natures_code",
			Table:   "legal_natures",
			Columns: []string{"code"},
		},
		{
			Name:    "idx_partner_qualifications_code",
			Table:   "partner_qualifications",
			Columns: []string{"code"},
		},
		{
			Name:    "idx_registration_status_reasons_code",
			Table:   "registration_status_reasons",
			Columns: []string{"code"},
		},
		{
			Name:    "idx_countries_code",
			Table:   "countries",
			Columns: []string{"code"},
		},
		{
			Name:    "idx_municipalities_code",
			Table:   "municipalities",
			Columns: []string{"code"},
		},
		{
			Name:    "idx_economic_activities_code",
			Table:   "economic_activities",
			Columns: []string{"code"},
		},
		{
			Name:    "idx_simple_tax_options_cnpj_root",
			Table:   "simple_tax_options",
			Columns: []string{"cnpj_root"},
		},
		{
			Name:    "idx_partners_cnpj_root",
			Table:   "partners",
			Columns: []string{"cnpj_root"},
		},
	}
}
