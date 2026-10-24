package cli

import (
	"strings"
	"testing"
)

func TestEstablishmentIndexRejectsInvalidArguments(t *testing.T) {
	// Prevent database access even if an argument validation regresses.
	t.Setenv("CNPJ_API_MYSQL_DATABASE", "")

	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			"missing destination",
			nil,
			"explicit destination",
		},
		{
			"conflicting limits",
			[]string{"--all", "--limit=10"},
			"all and limit",
		},
		{
			"invalid batch size",
			[]string{"--batch-size=0"},
			"batch-size",
		},
		{
			"resume flag removed",
			[]string{"--after=12ABC345000190"},
			"flag provided but not defined",
		},
		{
			"wrong table family",
			[]string{"--table=cnpj_companies_test"},
			"invalid establishment index table",
		},
		{
			"unexpected positional argument",
			[]string{"unexpected"},
			"positional arguments",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := runEstablishmentIndex(test.args)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("expected error containing %q, got %v", test.want, err)
			}
		})
	}
}

func TestIndexCommandsRejectResumeFlags(t *testing.T) {
	for name, run := range map[string]func([]string) error{
		"companies":      func(args []string) error { return runSearch(append([]string{"index"}, args...)) },
		"establishments": runEstablishmentIndex,
		"partners":       runPartnerIndex,
	} {
		for _, flag := range []string{"--after=12ABC345", "--after-id=1"} {
			t.Run(name+flag, func(t *testing.T) {
				err := run([]string{flag})
				if err == nil || !strings.Contains(err.Error(), "flag provided but not defined") {
					t.Fatalf("removed resume flag must fail before database access: %v", err)
				}
			})
		}
	}
}

func TestIndexCommandsProtectActiveDestinations(t *testing.T) {
	for key, value := range map[string]string{
		"CNPJ_API_MYSQL_HOST":                     "127.0.0.1",
		"CNPJ_API_MYSQL_PORT":                     "1",
		"CNPJ_API_MYSQL_DATABASE":                 "test",
		"CNPJ_API_MYSQL_USER":                     "test",
		"CNPJ_API_MYSQL_PASSWORD":                 "test",
		"CNPJ_API_MANTICORE_ENABLED":              "true",
		"CNPJ_API_MANTICORE_TABLE":                "cnpj_companies_active_test",
		"CNPJ_API_MANTICORE_ESTABLISHMENTS_TABLE": "cnpj_establishments_active_test",
		"CNPJ_API_MANTICORE_PARTNERS_TABLE":       "cnpj_partners_active_test",
	} {
		t.Setenv(key, value)
	}
	for _, table := range []string{
		"cnpj_companies_active_test",
		"cnpj_establishments_active_test",
		"cnpj_partners_active_test",
	} {
		command := "index"
		if strings.HasPrefix(table, "cnpj_establishments") {
			command = "index-establishments"
		}
		if strings.HasPrefix(table, "cnpj_partners") {
			command = "index-partners"
		}
		err := runSearch([]string{command, "--table=" + table, "--all"})
		if err == nil || !strings.Contains(err.Error(), "configured active table") {
			t.Fatalf("active destination %q must fail before connecting to databases: %v", table, err)
		}
	}
}
