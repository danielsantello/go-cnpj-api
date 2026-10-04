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

func taxOptionIndicator(value *string) *bool {
	if value == nil {
		return nil
	}

	var result bool

	switch *value {
	case "S":
		result = true
	case "N":
		result = false
	default:
		return nil
	}

	return &result
}

var partnerTypeDescriptions = map[string]string{
	"1": "PESSOA JURÍDICA",
	"2": "PESSOA FÍSICA",
	"3": "ESTRANGEIRO",
}

var ageRangeDescriptions = map[string]string{
	"0": "NÃO SE APLICA",
	"1": "ENTRE 0 E 12 ANOS",
	"2": "ENTRE 13 E 20 ANOS",
	"3": "ENTRE 21 E 30 ANOS",
	"4": "ENTRE 31 E 40 ANOS",
	"5": "ENTRE 41 E 50 ANOS",
	"6": "ENTRE 51 E 60 ANOS",
	"7": "ENTRE 61 E 70 ANOS",
	"8": "ENTRE 71 E 80 ANOS",
	"9": "MAIOR DE 80 ANOS",
}

func partnerType(code string) company.CodeDescription {
	result := company.CodeDescription{
		Code: code,
	}

	description, exists := partnerTypeDescriptions[code]
	if exists {
		result.Description = &description
	}

	return result
}

func ageRange(code *string) *company.CodeDescription {
	if code == nil {
		return nil
	}

	result := &company.CodeDescription{
		Code: *code,
	}

	description, exists := ageRangeDescriptions[*code]
	if exists {
		result.Description = &description
	}

	return result
}
