package v1

import "testing"

func TestCompanySizeReturnsNilForMissingCode(t *testing.T) {
	size := companySize(nil)

	if size != nil {
		t.Fatalf("expected nil company size, got %#v", size)
	}
}

func TestCompanySizeMapsKnownCode(t *testing.T) {
	code := "03"

	size := companySize(&code)

	if size == nil {
		t.Fatal("expected company size, got nil")
	}

	if size.Code != code {
		t.Fatalf("expected code %q, got %q", code, size.Code)
	}

	expectedDescription := "EMPRESA DE PEQUENO PORTE"

	if size.Description == nil {
		t.Fatal("expected company size description, got nil")
	}

	if *size.Description != expectedDescription {
		t.Fatalf(
			"expected description %q, got %q",
			expectedDescription,
			*size.Description,
		)
	}
}

func TestCompanySizePreservesUnknownCode(t *testing.T) {
	code := "99"

	size := companySize(&code)

	if size == nil {
		t.Fatal("expected company size, got nil")
	}

	if size.Code != code {
		t.Fatalf("expected code %q, got %q", code, size.Code)
	}

	if size.Description != nil {
		t.Fatalf(
			"expected nil description, got %q",
			*size.Description,
		)
	}
}
