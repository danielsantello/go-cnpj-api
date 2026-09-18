package v1

import "testing"

func TestRequiredIndexesAreValid(t *testing.T) {
	indexes := RequiredIndexes()

	expectedCount := 11

	if len(indexes) != expectedCount {
		t.Fatalf(
			"expected %d required indexes, got %d",
			expectedCount,
			len(indexes),
		)
	}

	seen := make(map[string]struct{})

	for _, index := range indexes {
		if index.Name == "" {
			t.Fatal("index name must not be empty")
		}

		if index.Table == "" {
			t.Fatalf("table must not be empty for index %q", index.Name)
		}

		if len(index.Columns) == 0 {
			t.Fatalf("columns must not be empty for index %q", index.Name)
		}

		key := index.Table + "." + index.Name

		if _, exists := seen[key]; exists {
			t.Fatalf("duplicate index definition %q", key)
		}

		seen[key] = struct{}{}

		for _, column := range index.Columns {
			if column == "" {
				t.Fatalf("column must not be empty for index %q", index.Name)
			}
		}
	}
}
