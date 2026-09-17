package mysqlstorage

import "testing"

func TestSchemaValidatorForFormatAcceptsVersion1(t *testing.T) {
	validator, err := schemaValidatorForFormat(1)
	if err != nil {
		t.Fatalf("expected supported format, got error: %v", err)
	}

	if validator == nil {
		t.Fatal("expected schema validator, got nil")
	}
}

func TestSchemaValidatorForFormatRejectsUnsupportedVersion(t *testing.T) {
	validator, err := schemaValidatorForFormat(2)

	if validator != nil {
		t.Fatal("expected nil validator for unsupported format")
	}

	if err == nil {
		t.Fatal("expected unsupported format error, got nil")
	}

	expected := "unsupported database format version: 2"

	if err.Error() != expected {
		t.Fatalf("expected error %q, got %q", expected, err.Error())
	}
}
