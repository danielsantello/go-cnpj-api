package httpapi

import "net/url"

type companySearchQuery struct {
	Name     string
	Page     int
	PageSize int
}

func parseCompanySearchQuery(
	values url.Values,
) (companySearchQuery, []errorDetail) {
	query := companySearchQuery{
		Name:     values.Get("name"),
		Page:     defaultCompanyPage,
		PageSize: defaultCompanyPageSize,
	}

	details := make([]errorDetail, 0)

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
