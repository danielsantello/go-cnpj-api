package manticore

import "testing"

func TestNewRepositoryForTable(t *testing.T) {
	for _, name := range []string{
		"cnpj_companies",
		"cnpj_companies_2026_08",
	} {
		repository, err := NewRepositoryForTable(nil, name)
		if err != nil {
			t.Fatalf("accept table %q: %v", name, err)
		}
		if repository.table != name {
			t.Fatalf("expected table %q, got %q", name, repository.table)
		}
	}

	for _, name := range []string{
		"", "other_table", "cnpj_companies_test; DROP TABLE x",
	} {
		if _, err := NewRepositoryForTable(nil, name); err == nil {
			t.Fatalf("expected invalid table error for %q", name)
		}
	}
}
