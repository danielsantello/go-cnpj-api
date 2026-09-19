package company

type CodeDescription struct {
	Code        string  `json:"code"`
	Description *string `json:"description"`
}

type Company struct {
	BasicCNPJ                   string           `json:"basic_cnpj"`
	CorporateName               string           `json:"corporate_name"`
	LegalNature                 CodeDescription  `json:"legal_nature"`
	ResponsibleQualification    CodeDescription  `json:"responsible_qualification"`
	ShareCapital                *string          `json:"share_capital"`
	CompanySize                 *CodeDescription `json:"company_size"`
	ResponsibleFederativeEntity *string          `json:"responsible_federative_entity"`
}
