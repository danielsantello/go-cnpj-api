package company

import (
	"errors"
	"strings"
)

const cnpjLength = 14

var ErrCNPJTooLong = errors.New("CNPJ must contain at most 14 characters")

func NormalizeCNPJ(value string) (string, error) {
	var normalized strings.Builder

	for _, character := range strings.ToUpper(value) {
		if isCNPJCharacter(character) {
			normalized.WriteRune(character)
		}
	}

	cnpj := normalized.String()

	if len(cnpj) > cnpjLength {
		return "", ErrCNPJTooLong
	}

	return strings.Repeat("0", cnpjLength-len(cnpj)) + cnpj, nil
}

func isCNPJCharacter(character rune) bool {
	isNumber := character >= '0' && character <= '9'
	isLetter := character >= 'A' && character <= 'Z'

	return isNumber || isLetter
}
