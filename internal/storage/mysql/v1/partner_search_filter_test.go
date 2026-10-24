package v1

import (
	"reflect"
	"strings"
	"testing"

	"github.com/danielsantello/go-cnpj-api/internal/company"
)

func TestPartnerSearchFilter(t *testing.T) {
	document := "***111769**"

	tests := []struct {
		name     string
		options  company.PartnerSearchOptions
		wantSQL  string
		wantArgs []any
	}{
		{
			name: "name prefix",
			options: company.PartnerSearchOptions{
				Name: "DANIEL",
			},
			wantSQL:  "AND partner.partner_name LIKE ? ESCAPE '!'",
			wantArgs: []any{"DANIEL%"},
		},
		{
			name: "name prefix and literal masked document",
			options: company.PartnerSearchOptions{
				Name: "DANIEL", Document: &document,
			},
			wantSQL: "AND partner.partner_name LIKE ? ESCAPE '!' " +
				"AND partner.partner_document = ?",
			wantArgs: []any{"DANIEL%", "***111769**"},
		},
		{
			name: "escaped prefix with document",
			options: company.PartnerSearchOptions{
				Name: "NOME_100%!", Document: &document,
			},
			wantSQL: "AND partner.partner_name LIKE ? ESCAPE '!' " +
				"AND partner.partner_document = ?",
			wantArgs: []any{"NOME!_100!%!!%", "***111769**"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			filter, arguments := partnerSearchFilter(test.options)
			normalized := strings.Join(strings.Fields(filter), " ")

			if normalized != test.wantSQL {
				t.Fatalf("expected SQL %q, got %q", test.wantSQL, normalized)
			}
			if !reflect.DeepEqual(arguments, test.wantArgs) {
				t.Fatalf("expected arguments %#v, got %#v",
					test.wantArgs, arguments)
			}
		})
	}
}
