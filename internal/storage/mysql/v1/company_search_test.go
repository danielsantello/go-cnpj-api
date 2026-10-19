package v1

import "testing"

func TestCorporateNamePrefix(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "adds prefix wildcard",
			input:    "SANTELLO",
			expected: "SANTELLO%",
		},
		{
			name:     "escapes percent sign",
			input:    "EMPRESA 100%",
			expected: "EMPRESA 100!%%",
		},
		{
			name:     "escapes underscore",
			input:    "EMPRESA_1",
			expected: "EMPRESA!_1%",
		},
		{
			name:     "escapes escape character",
			input:    "EMPRESA!1",
			expected: "EMPRESA!!1%",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := corporateNamePrefix(test.input)

			if result != test.expected {
				t.Fatalf(
					"expected prefix %q, got %q",
					test.expected,
					result,
				)
			}
		})
	}
}
