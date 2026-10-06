package httpapi

import (
	"net/url"
	"strconv"
)

const (
	defaultCompanyPage     = 1
	defaultCompanyPageSize = 20
	maximumCompanyPageSize = 100
)

type companyQuery struct {
	IncludeEstablishments bool
	Page                  int
	PageSize              int
}

func parseCompanyQuery(values url.Values) (companyQuery, []errorDetail) {
	query := companyQuery{
		Page:     defaultCompanyPage,
		PageSize: defaultCompanyPageSize,
	}

	details := make([]errorDetail, 0)

	if values.Has("include") {
		include := values.Get("include")

		if include == "establishments" {
			query.IncludeEstablishments = true
		} else {
			details = append(details, errorDetail{
				Field:  "include",
				Reason: "include must be establishments.",
			})
		}
	}

	page, detail := parsePositiveIntegerParameter(
		values,
		"page",
		defaultCompanyPage,
		0,
	)
	if detail != nil {
		details = append(details, *detail)
	} else {
		query.Page = page
	}

	pageSize, detail := parsePositiveIntegerParameter(
		values,
		"page_size",
		defaultCompanyPageSize,
		maximumCompanyPageSize,
	)
	if detail != nil {
		details = append(details, *detail)
	} else {
		query.PageSize = pageSize
	}

	return query, details
}

func parsePositiveIntegerParameter(
	values url.Values,
	name string,
	defaultValue int,
	maximumValue int,
) (int, *errorDetail) {
	if !values.Has(name) {
		return defaultValue, nil
	}

	value, err := strconv.Atoi(values.Get(name))
	if err != nil || value <= 0 {
		return 0, &errorDetail{
			Field:  name,
			Reason: name + " must be a positive integer.",
		}
	}

	if maximumValue > 0 && value > maximumValue {
		return 0, &errorDetail{
			Field: name,
			Reason: name +
				" must not be greater than " +
				strconv.Itoa(maximumValue) +
				".",
		}
	}

	return value, nil
}
