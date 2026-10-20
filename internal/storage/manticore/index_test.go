package manticore

import (
	"strings"
	"testing"
)

func TestNewIndexForTableValidatesName(t *testing.T) {
	for _, name := range []string{
		"cnpj_companies_2026_08",
		"cnpj_companies_2026_08_test",
	} {
		index, err := NewIndexForTable(nil, name)
		if err != nil {
			t.Fatalf("accept table %q: %v", name, err)
		}
		if index.table != name {
			t.Fatalf("expected table %q, got %q", name, index.table)
		}
	}

	for _, name := range []string{
		"",
		"cnpj_companies",
		"other_table",
		"cnpj_companies_",
		"cnpj_companies_TEST",
		"cnpj_companies_test; DROP TABLE x",
		"cnpj_companies_" + strings.Repeat("a", 49),
	} {
		if _, err := NewIndexForTable(nil, name); err == nil {
			t.Fatalf("expected invalid table name error for %q", name)
		}
	}
}
