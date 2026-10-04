package company

type Registration struct {
	Status            *CodeDescription `json:"status"`
	StatusDate        *string          `json:"status_date"`
	StatusReason      *CodeDescription `json:"status_reason"`
	SpecialStatus     *string          `json:"special_status"`
	SpecialStatusDate *string          `json:"special_status_date"`
}

type EconomicActivities struct {
	Primary   *CodeDescription  `json:"primary"`
	Secondary []CodeDescription `json:"secondary"`
}

type Address struct {
	StreetType      *string          `json:"street_type"`
	Street          *string          `json:"street"`
	Number          *string          `json:"number"`
	Complement      *string          `json:"complement"`
	Neighborhood    *string          `json:"neighborhood"`
	PostalCode      *string          `json:"postal_code"`
	State           *string          `json:"state"`
	Municipality    *CodeDescription `json:"municipality"`
	Country         *CodeDescription `json:"country"`
	ForeignCityName *string          `json:"foreign_city_name"`
}

type Phone struct {
	AreaCode *string `json:"area_code"`
	Number   string  `json:"number"`
}

type Contacts struct {
	Phones []Phone `json:"phones"`
	Fax    *Phone  `json:"fax"`
	Email  *string `json:"email"`
}

type Establishment struct {
	CNPJ               string             `json:"cnpj"`
	Type               *string            `json:"type"`
	TradeName          *string            `json:"trade_name"`
	Registration       Registration       `json:"registration"`
	ActivityStartDate  *string            `json:"activity_start_date"`
	EconomicActivities EconomicActivities `json:"economic_activities"`
	Address            Address            `json:"address"`
	Contacts           Contacts           `json:"contacts"`
}
