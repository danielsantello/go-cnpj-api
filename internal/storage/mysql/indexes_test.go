package mysqlstorage

import "testing"

func TestIndexDefinitionsForFormatAcceptsVersion1(t *testing.T) {
	indexes, err := indexDefinitionsForFormat(1)
	if err != nil {
		t.Fatalf("get index definitions: %v", err)
	}

	expectedCount := 11

	if len(indexes) != expectedCount {
		t.Fatalf(
			"expected %d index definitions, got %d",
			expectedCount,
			len(indexes),
		)
	}
}

func TestIndexDefinitionsForFormatRejectsUnsupportedVersion(t *testing.T) {
	_, err := indexDefinitionsForFormat(999)

	if err == nil {
		t.Fatal("expected unsupported version error, got nil")
	}
}

func TestQuoteIdentifierEscapesBackticks(t *testing.T) {
	quoted := quoteIdentifier("index`name")
	expected := "`index``name`"

	if quoted != expected {
		t.Fatalf("expected identifier %q, got %q", expected, quoted)
	}
}

func TestGroupMissingIndexesGroupsByTableAndSkipsExisting(t *testing.T) {
	definitions := []indexDefinition{
		{
			name:    "idx_establishments_cnpj",
			table:   "establishments",
			columns: []string{"cnpj"},
		},
		{
			name:    "idx_establishments_cnpj_root",
			table:   "establishments",
			columns: []string{"cnpj_root"},
		},
		{
			name:    "idx_companies_cnpj_root",
			table:   "companies",
			columns: []string{"cnpj_root"},
		},
		{
			name:    "idx_countries_code",
			table:   "countries",
			columns: []string{"code"},
		},
	}

	existingIndexes := map[string]struct{}{
		indexKey("companies", "idx_companies_cnpj_root"): {},
	}

	groups := groupMissingIndexes(definitions, existingIndexes)

	if len(groups) != 2 {
		t.Fatalf("expected 2 index groups, got %d", len(groups))
	}

	establishmentsGroup := groups[0]

	if establishmentsGroup.table != "establishments" {
		t.Fatalf(
			"expected first group for establishments, got %q",
			establishmentsGroup.table,
		)
	}

	if len(establishmentsGroup.indexes) != 2 {
		t.Fatalf(
			"expected 2 establishment indexes, got %d",
			len(establishmentsGroup.indexes),
		)
	}

	countriesGroup := groups[1]

	if countriesGroup.table != "countries" {
		t.Fatalf(
			"expected second group for countries, got %q",
			countriesGroup.table,
		)
	}

	if len(countriesGroup.indexes) != 1 {
		t.Fatalf(
			"expected 1 country index, got %d",
			len(countriesGroup.indexes),
		)
	}
}

func TestCreateIndexGroupStatement(t *testing.T) {
	group := indexGroup{
		table: "establishments",
		indexes: []indexDefinition{
			{
				name:    "idx_establishments_cnpj",
				table:   "establishments",
				columns: []string{"cnpj"},
			},
			{
				name:    "idx_establishments_cnpj_root",
				table:   "establishments",
				columns: []string{"cnpj_root"},
			},
		},
	}

	statement := createIndexGroupStatement(group)

	expected := "ALTER TABLE `establishments` " +
		"ADD INDEX `idx_establishments_cnpj` (`cnpj`), " +
		"ADD INDEX `idx_establishments_cnpj_root` (`cnpj_root`)"

	if statement != expected {
		t.Fatalf(
			"expected statement %q, got %q",
			expected,
			statement,
		)
	}
}
