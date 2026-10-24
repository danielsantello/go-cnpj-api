package config

import (
	"strings"
	"testing"
)

func TestEntitySearchTableDefaults(t *testing.T) {
	configuration := Default()
	if configuration.ManticoreEstablishmentsTable != "cnpj_establishments" {
		t.Fatal("unexpected establishment table default")
	}
	if configuration.ManticorePartnersTable != "cnpj_partners" {
		t.Fatal("unexpected partner table default")
	}
}

func TestLoadEntitySearchTables(t *testing.T) {
	t.Setenv(
		manticoreEstablishmentsTableEnvironmentVariable,
		"cnpj_establishments_2026_08",
	)
	t.Setenv(
		manticorePartnersTableEnvironmentVariable,
		"cnpj_partners_2026_08",
	)

	configuration, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if configuration.ManticoreEstablishmentsTable != "cnpj_establishments_2026_08" {
		t.Fatal("establishment table environment override was not loaded")
	}
	if configuration.ManticorePartnersTable != "cnpj_partners_2026_08" {
		t.Fatal("partner table environment override was not loaded")
	}
}

func TestValidateEntitySearchTables(t *testing.T) {
	tests := []struct {
		name           string
		establishments string
		partners       string
		wantErr        bool
	}{
		{
			name:           "defaults",
			establishments: "cnpj_establishments",
			partners:       "cnpj_partners",
		},
		{
			name:           "publication tables",
			establishments: "cnpj_establishments_2026_08",
			partners:       "cnpj_partners_2026_08",
		},
		{
			name:           "wrong establishment family",
			establishments: "cnpj_companies_2026_08",
			partners:       "cnpj_partners",
			wantErr:        true,
		},
		{
			name:           "wrong partner family",
			establishments: "cnpj_establishments",
			partners:       "cnpj_establishments_2026_08",
			wantErr:        true,
		},
		{
			name:     "empty establishment table",
			partners: "cnpj_partners",
			wantErr:  true,
		},
		{
			name:           "empty partner table",
			establishments: "cnpj_establishments",
			wantErr:        true,
		},
		{
			name:           "unsafe name",
			establishments: "cnpj_establishments_test; DROP TABLE x",
			partners:       "cnpj_partners",
			wantErr:        true,
		},
		{
			name:           "oversized suffix",
			establishments: "cnpj_establishments",
			partners:       "cnpj_partners_" + strings.Repeat("a", 49),
			wantErr:        true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			configuration := validConfig()
			configuration.ManticoreEnabled = true
			configuration.ManticoreEstablishmentsTable = test.establishments
			configuration.ManticorePartnersTable = test.partners

			err := Validate(configuration)
			if (err != nil) != test.wantErr {
				t.Fatalf("unexpected validation result: %v", err)
			}
		})
	}
}

func TestDisabledManticoreIgnoresEntityTableNames(t *testing.T) {
	configuration := validConfig()
	configuration.ManticoreEnabled = false
	configuration.ManticoreEstablishmentsTable = ""
	configuration.ManticorePartnersTable = ""

	if err := Validate(configuration); err != nil {
		t.Fatal(err)
	}
}

func TestIsActiveSearchTable(t *testing.T) {
	configuration := Default()
	configuration.ManticoreTable = "cnpj_companies_2026_08"
	configuration.ManticoreEstablishmentsTable = "cnpj_establishments_2026_08"
	configuration.ManticorePartnersTable = "cnpj_partners_2026_08"

	for _, table := range []string{
		configuration.ManticoreTable,
		configuration.ManticoreEstablishmentsTable,
		configuration.ManticorePartnersTable,
	} {
		if !configuration.IsActiveSearchTable(table) {
			t.Fatalf("configured table %q must be protected", table)
		}
	}
	for _, table := range []string{
		"",
		"cnpj_companies_other",
		"cnpj_establishments_2026_08_test",
		"cnpj_partners_2026_08_test",
	} {
		if configuration.IsActiveSearchTable(table) {
			t.Fatalf("unconfigured table %q was marked active", table)
		}
	}
}
