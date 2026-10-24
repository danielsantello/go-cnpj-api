package search

import "testing"

func TestPartnerSourceKeyRoundTrip(t *testing.T) {
	for _, root := range []string{"00000001", "12ABC345", "12abc345"} {
		for _, occurrence := range []uint64{1, 2, 10, ^uint64(0)} {
			key, err := MakePartnerSourceKey(root, occurrence)
			if err != nil {
				t.Fatal(err)
			}
			actualRoot, actualOccurrence, err := ParsePartnerSourceKey(key)
			if err != nil || actualRoot != root || actualOccurrence != occurrence {
				t.Fatalf("cursor changed its values: %q, %d, %v",
					actualRoot, actualOccurrence, err)
			}
		}
	}

	first, _ := MakePartnerSourceKey("12ABC345", 2)
	second, _ := MakePartnerSourceKey("12ABC345", 10)
	if first >= second {
		t.Fatal("source keys must preserve occurrence ordering")
	}
}

func TestPartnerSourceKeyRejectsInvalidValues(t *testing.T) {
	for _, key := range []string{
		"invalid",
		"12ABC345:1",
		"12ABC345:00000000000000000000",
		"12ABC345:0000000000000000000A",
		"12ABC345:99999999999999999999",
		"SHORT:00000000000000000001",
	} {
		if _, _, err := ParsePartnerSourceKey(key); err == nil {
			t.Fatalf("expected rejection of %q", key)
		}
	}
	if _, err := MakePartnerSourceKey("SHORT", 1); err == nil {
		t.Fatal("short basic CNPJ must fail")
	}
	if _, err := MakePartnerSourceKey("12ABC345", 0); err == nil {
		t.Fatal("zero occurrence must fail")
	}

	root, occurrence, err := ParsePartnerSourceKey("")
	if err != nil || root != "" || occurrence != 0 {
		t.Fatal("empty cursor must represent the beginning")
	}
}
