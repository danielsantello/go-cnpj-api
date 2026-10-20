package manticore

import (
	"strings"
	"unicode"
)

func searchExpression(value string) string {
	terms := strings.FieldsFunc(value, func(character rune) bool {
		return !unicode.IsLetter(character) && !unicode.IsNumber(character)
	})

	quoted := make([]string, 0, len(terms))
	for _, term := range terms {
		quoted = append(quoted, `"`+term+`"`)
	}

	return strings.Join(quoted, " ")
}
