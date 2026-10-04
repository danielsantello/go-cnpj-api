package company

type LegalRepresentative struct {
	Document      *string          `json:"document"`
	Name          *string          `json:"name"`
	Qualification *CodeDescription `json:"qualification"`
}

type Partner struct {
	Type                CodeDescription      `json:"type"`
	Name                string               `json:"name"`
	Document            *string              `json:"document"`
	Qualification       *CodeDescription     `json:"qualification"`
	EntryDate           *string              `json:"entry_date"`
	Country             *CodeDescription     `json:"country"`
	LegalRepresentative *LegalRepresentative `json:"legal_representative"`
	AgeRange            *CodeDescription     `json:"age_range"`
}
