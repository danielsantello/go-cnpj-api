package manticore

import (
	"reflect"
	"testing"
)

func TestPartnerMatchFilter(t *testing.T) {
	masked := "***111769**"
	literal := "' OR 1=1"

	tests := []struct {
		name     string
		document *string
		wantSQL  string
		wantArgs []any
	}{
		{
			name:     "name only",
			wantSQL:  "WHERE MATCH(?)",
			wantArgs: []any{`"daniel" "santello"`},
		},
		{
			name:     "name and masked document",
			document: &masked,
			wantSQL: "WHERE MATCH(?) AND has_document = 1 " +
				"AND partner_document = ?",
			wantArgs: []any{`"daniel" "santello"`, "***111769**"},
		},
		{
			name:     "document remains a parameter",
			document: &literal,
			wantSQL: "WHERE MATCH(?) AND has_document = 1 " +
				"AND partner_document = ?",
			wantArgs: []any{`"daniel" "santello"`, "' OR 1=1"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			filter, arguments := partnerMatchFilter(
				`"daniel" "santello"`, test.document,
			)
			if filter != test.wantSQL {
				t.Fatalf("expected filter %q, got %q", test.wantSQL, filter)
			}
			if !reflect.DeepEqual(arguments, test.wantArgs) {
				t.Fatalf("expected arguments %#v, got %#v",
					test.wantArgs, arguments)
			}
		})
	}
}
