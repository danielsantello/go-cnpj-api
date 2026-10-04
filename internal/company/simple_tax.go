package company

type MEI struct {
	IsOpted       *bool   `json:"is_opted"`
	OptionDate    *string `json:"option_date"`
	ExclusionDate *string `json:"exclusion_date"`
}

type SimpleTax struct {
	IsOpted       *bool   `json:"is_opted"`
	OptionDate    *string `json:"option_date"`
	ExclusionDate *string `json:"exclusion_date"`
	MEI           MEI     `json:"mei"`
}
