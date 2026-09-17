package v1

import "testing"

func TestValidateAvailableSchemaAcceptsCompatibleSchema(t *testing.T) {
	available := completeAvailableSchema()

	available["extra_table"] = map[string]struct{}{
		"extra_column": {},
	}
	available["companies"]["extra_column"] = struct{}{}

	if err := validateAvailableSchema(available); err != nil {
		t.Fatalf("expected compatible schema, got error: %v", err)
	}
}

func TestValidateAvailableSchemaRejectsMissingTable(t *testing.T) {
	available := completeAvailableSchema()

	delete(available, "companies")

	err := validateAvailableSchema(available)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	expected := `required table "companies" is missing`

	if err.Error() != expected {
		t.Fatalf("expected error %q, got %q", expected, err.Error())
	}
}

func TestValidateAvailableSchemaRejectsMissingColumn(t *testing.T) {
	available := completeAvailableSchema()

	delete(available["companies"], "legal_name")

	err := validateAvailableSchema(available)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	expected := `required column "legal_name" is missing from table "companies"`

	if err.Error() != expected {
		t.Fatalf("expected error %q, got %q", expected, err.Error())
	}
}

func completeAvailableSchema() availableSchema {
	available := make(availableSchema)

	for _, table := range schemaContract {
		available[table.name] = make(map[string]struct{})

		for _, column := range table.columns {
			available[table.name][column] = struct{}{}
		}
	}

	return available
}
