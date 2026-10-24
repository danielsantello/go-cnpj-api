package manticore

import "testing"

func TestValidateDropTable(t *testing.T) {
	tests := []struct {
		name    string
		table   string
		active  string
		wantErr bool
	}{
		{
			name:   "old publication",
			table:  "cnpj_companies_2026_07",
			active: "cnpj_companies_2026_08",
		},
		{
			name:   "old default table",
			table:  "cnpj_companies",
			active: "cnpj_companies_2026_08",
		},
		{
			name:    "active publication",
			table:   "cnpj_companies_2026_08",
			active:  "cnpj_companies_2026_08",
			wantErr: true,
		},
		{
			name:    "active default",
			table:   "cnpj_companies",
			active:  "cnpj_companies",
			wantErr: true,
		},
		{
			name:    "invalid destination",
			table:   "cnpj_companies_test; DROP TABLE x",
			active:  "cnpj_companies",
			wantErr: true,
		},
		{
			name:    "missing active configuration",
			table:   "cnpj_companies_2026_07",
			active:  "",
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateDropTable(test.table, test.active)
			if (err != nil) != test.wantErr {
				t.Fatalf("expected error=%t, got %v", test.wantErr, err)
			}
		})
	}
}

func TestDropProtectsEveryActiveTable(t *testing.T) {
	active := []string{
		"cnpj_companies_2026_08",
		"cnpj_establishments_2026_08",
		"cnpj_partners_2026_08",
	}
	for _, table := range active {
		if err := ValidateDropTable(table, active...); err == nil {
			t.Fatalf("active table %q must be protected", table)
		}
	}
	for _, table := range []string{
		"cnpj_establishments_2026_07", "cnpj_partners_2026_07",
		"cnpj_establishments", "cnpj_partners",
	} {
		if err := ValidateDropTable(table, active...); err != nil {
			t.Fatalf("inactive table %q must be accepted: %v", table, err)
		}
	}
	if err := ValidateDropTable("cnpj_partners_test"); err == nil {
		t.Fatal("missing active-table configuration must fail")
	}
	if err := ValidateDropTable("cnpj_partners_test; DROP TABLE x", active...); err == nil {
		t.Fatal("invalid identifier must fail")
	}
}
