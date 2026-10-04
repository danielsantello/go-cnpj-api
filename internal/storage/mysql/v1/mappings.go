package v1

import (
	"strings"

	"github.com/danielsantello/go-cnpj-api/internal/company"
)

var companySizeDescriptions = map[string]string{
	"00": "NÃO INFORMADO",
	"01": "MICRO EMPRESA",
	"03": "EMPRESA DE PEQUENO PORTE",
	"05": "DEMAIS",
}

var registrationStatusDescriptions = map[string]string{
	"01": "NULA",
	"02": "ATIVA",
	"03": "SUSPENSA",
	"04": "INAPTA",
	"08": "BAIXADA",
}

func companySize(code *string) *company.CodeDescription {
	if code == nil {
		return nil
	}

	result := &company.CodeDescription{
		Code: *code,
	}

	description, exists := companySizeDescriptions[*code]
	if exists {
		result.Description = &description
	}

	return result
}

func establishmentType(code string) *string {
	var description string

	switch code {
	case "1":
		description = "headquarters"
	case "2":
		description = "branch"
	default:
		return nil
	}

	return &description
}

func registrationStatus(code *string) *company.CodeDescription {
	if code == nil {
		return nil
	}

	result := &company.CodeDescription{
		Code: *code,
	}

	description, exists := registrationStatusDescriptions[*code]
	if exists {
		result.Description = &description
	}

	return result
}

func codeDescription(
	code *string,
	description *string,
) *company.CodeDescription {
	if code == nil {
		return nil
	}

	return &company.CodeDescription{
		Code:        *code,
		Description: description,
	}
}

func phone(areaCode *string, number *string) *company.Phone {
	if number == nil {
		return nil
	}

	return &company.Phone{
		AreaCode: areaCode,
		Number:   *number,
	}
}

func secondaryEconomicActivityCodes(value *string) []string {
	result := make([]string, 0)

	if value == nil {
		return result
	}

	for _, code := range strings.Split(*value, ",") {
		code = strings.TrimSpace(code)

		if code != "" {
			result = append(result, code)
		}
	}

	return result
}
