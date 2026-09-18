package company

import (
	"errors"
	"testing"
)

func TestNormalizeCNPJ(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "keeps numeric CNPJ",
			input:    "12345678000190",
			expected: "12345678000190",
		},
		{
			name:     "removes numeric CNPJ formatting",
			input:    "12.345.678/0001-90",
			expected: "12345678000190",
		},
		{
			name:     "keeps alphanumeric CNPJ",
			input:    "12ABC34501DE35",
			expected: "12ABC34501DE35",
		},
		{
			name:     "removes formatting and converts letters to uppercase",
			input:    "12.abc.345/01de-35",
			expected: "12ABC34501DE35",
		},
		{
			name:     "adds zeros to the left",
			input:    "123.456",
			expected: "00000000123456",
		},
		{
			name:     "removes spaces",
			input:    "  123.456  ",
			expected: "00000000123456",
		},
		{
			name:     "normalizes input without alphanumeric characters",
			input:    "---",
			expected: "00000000000000",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cnpj, err := NormalizeCNPJ(test.input)
			if err != nil {
				t.Fatalf("normalize CNPJ: %v", err)
			}

			if cnpj != test.expected {
				t.Fatalf("expected CNPJ %q, got %q", test.expected, cnpj)
			}
		})
	}
}

func TestNormalizeCNPJRejectsMoreThanFourteenCharacters(t *testing.T) {
	_, err := NormalizeCNPJ("123456789012345")

	if !errors.Is(err, ErrCNPJTooLong) {
		t.Fatalf("expected ErrCNPJTooLong, got %v", err)
	}
}
