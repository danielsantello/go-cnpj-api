package search

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

func MakePartnerSourceKey(
	basicCNPJ string,
	occurrence uint64,
) (string, error) {
	if utf8.RuneCountInString(basicCNPJ) != 8 ||
		strings.Contains(basicCNPJ, ":") {
		return "", fmt.Errorf("partner source key requires an 8-character basic CNPJ")
	}
	if occurrence == 0 {
		return "", fmt.Errorf("partner occurrence must be greater than zero")
	}

	return fmt.Sprintf("%s:%020d", basicCNPJ, occurrence), nil
}

func ParsePartnerSourceKey(key string) (string, uint64, error) {
	if key == "" {
		return "", 0, nil
	}

	root, number, found := strings.Cut(key, ":")
	if !found || utf8.RuneCountInString(root) != 8 || len(number) != 20 {
		return "", 0, fmt.Errorf("invalid partner source key: %q", key)
	}
	for _, character := range number {
		if character < '0' || character > '9' {
			return "", 0, fmt.Errorf("invalid partner occurrence in source key: %q", key)
		}
	}

	// Only the occurrence number is parsed; the CNPJ remains textual.
	occurrence, err := strconv.ParseUint(number, 10, 64)
	if err != nil || occurrence == 0 {
		return "", 0, fmt.Errorf("invalid partner occurrence in source key: %q", key)
	}

	return root, occurrence, nil
}
