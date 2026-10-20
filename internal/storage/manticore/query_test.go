package manticore

import "testing"

func TestSearchExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"DANIEL SANTELLO", `"DANIEL" "SANTELLO"`},
		{"  associação  São José ", `"associação" "São" "José"`},
		{"SANTELLO | SILVA", `"SANTELLO" "SILVA"`},
		{"@partner_names (SANTELLO*)", `"partner" "names" "SANTELLO"`},
		{`LOJA "EXEMPLO" -TESTE`, `"LOJA" "EXEMPLO" "TESTE"`},
		{"***", ""},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			got := searchExpression(test.input)
			if got != test.expected {
				t.Fatalf("expected %q, got %q", test.expected, got)
			}
		})
	}
}
