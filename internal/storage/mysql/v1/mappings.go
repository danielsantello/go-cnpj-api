package v1

import "github.com/danielsantello/go-cnpj-api/internal/company"

var companySizeDescriptions = map[string]string{
	"00": "NÃO INFORMADO",
	"01": "MICRO EMPRESA",
	"03": "EMPRESA DE PEQUENO PORTE",
	"05": "DEMAIS",
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
