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

func TestEstablishmentTypeMapsKnownCodes(t *testing.T) {
	tests := []struct {
		code     string
		expected string
	}{
		{
			code:     "1",
			expected: "headquarters",
		},
		{
			code:     "2",
			expected: "branch",
		},
	}

	for _, test := range tests {
		t.Run(test.code, func(t *testing.T) {
			result := establishmentType(test.code)

			if result == nil {
				t.Fatal("expected establishment type, got nil")
			}

			if *result != test.expected {
				t.Fatalf(
					"expected establishment type %q, got %q",
					test.expected,
					*result,
				)
			}
		})
	}
}

func TestEstablishmentTypeReturnsNilForUnknownCode(t *testing.T) {
	result := establishmentType("9")

	if result != nil {
		t.Fatalf(
			"expected nil establishment type, got %q",
			*result,
		)
	}
}

func TestRegistrationStatusReturnsNilForMissingCode(t *testing.T) {
	status := registrationStatus(nil)

	if status != nil {
		t.Fatalf("expected nil registration status, got %#v", status)
	}
}

func TestRegistrationStatusMapsKnownCode(t *testing.T) {
	code := "02"

	status := registrationStatus(&code)

	if status == nil {
		t.Fatal("expected registration status, got nil")
	}

	if status.Code != code {
		t.Fatalf("expected code %q, got %q", code, status.Code)
	}

	expectedDescription := "ATIVA"

	if status.Description == nil {
		t.Fatal("expected registration status description, got nil")
	}

	if *status.Description != expectedDescription {
		t.Fatalf(
			"expected description %q, got %q",
			expectedDescription,
			*status.Description,
		)
	}
}

func TestRegistrationStatusPreservesUnknownCode(t *testing.T) {
	code := "99"

	status := registrationStatus(&code)

	if status == nil {
		t.Fatal("expected registration status, got nil")
	}

	if status.Code != code {
		t.Fatalf("expected code %q, got %q", code, status.Code)
	}

	if status.Description != nil {
		t.Fatalf(
			"expected nil description, got %q",
			*status.Description,
		)
	}
}

func TestCodeDescriptionReturnsNilForMissingCode(t *testing.T) {
	result := codeDescription(nil, nil)

	if result != nil {
		t.Fatalf("expected nil code description, got %#v", result)
	}
}

func TestCodeDescriptionPreservesCodeAndDescription(t *testing.T) {
	code := "01"
	description := "SEM MOTIVO"

	result := codeDescription(&code, &description)

	if result == nil {
		t.Fatal("expected code description, got nil")
	}

	if result.Code != code {
		t.Fatalf("expected code %q, got %q", code, result.Code)
	}

	if result.Description == nil {
		t.Fatal("expected description, got nil")
	}

	if *result.Description != description {
		t.Fatalf(
			"expected description %q, got %q",
			description,
			*result.Description,
		)
	}
}

func TestCodeDescriptionPreservesCodeWithoutDescription(t *testing.T) {
	code := "99"

	result := codeDescription(&code, nil)

	if result == nil {
		t.Fatal("expected code description, got nil")
	}

	if result.Code != code {
		t.Fatalf("expected code %q, got %q", code, result.Code)
	}

	if result.Description != nil {
		t.Fatalf(
			"expected nil description, got %q",
			*result.Description,
		)
	}
}

func TestPhoneReturnsNilForMissingNumber(t *testing.T) {
	areaCode := "11"

	result := phone(&areaCode, nil)

	if result != nil {
		t.Fatalf("expected nil phone, got %#v", result)
	}
}

func TestPhonePreservesNumberWithoutAreaCode(t *testing.T) {
	number := "49999999"

	result := phone(nil, &number)

	if result == nil {
		t.Fatal("expected phone, got nil")
	}

	if result.AreaCode != nil {
		t.Fatalf("expected nil area code, got %q", *result.AreaCode)
	}

	if result.Number != number {
		t.Fatalf("expected number %q, got %q", number, result.Number)
	}
}

func TestPhonePreservesAreaCodeAndNumber(t *testing.T) {
	areaCode := "11"
	number := "49999999"

	result := phone(&areaCode, &number)

	if result == nil {
		t.Fatal("expected phone, got nil")
	}

	if result.AreaCode == nil {
		t.Fatal("expected area code, got nil")
	}

	if *result.AreaCode != areaCode {
		t.Fatalf(
			"expected area code %q, got %q",
			areaCode,
			*result.AreaCode,
		)
	}

	if result.Number != number {
		t.Fatalf("expected number %q, got %q", number, result.Number)
	}
}

func TestSecondaryEconomicActivityCodesReturnsEmptyForMissingValue(
	t *testing.T,
) {
	result := secondaryEconomicActivityCodes(nil)

	if result == nil {
		t.Fatal("expected empty collection, got nil")
	}

	if len(result) != 0 {
		t.Fatalf("expected no codes, got %d", len(result))
	}
}

func TestSecondaryEconomicActivityCodesSplitsAndPreservesOrder(
	t *testing.T,
) {
	value := "6201501, 6201502,,6202300"

	result := secondaryEconomicActivityCodes(&value)

	expected := []string{
		"6201501",
		"6201502",
		"6202300",
	}

	if len(result) != len(expected) {
		t.Fatalf(
			"expected %d codes, got %d",
			len(expected),
			len(result),
		)
	}

	for index := range expected {
		if result[index] != expected[index] {
			t.Fatalf(
				"expected code %q at index %d, got %q",
				expected[index],
				index,
				result[index],
			)
		}
	}
}
