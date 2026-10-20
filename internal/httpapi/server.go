package httpapi

import (
	"net/http"
	"time"
)

type companyService interface {
	companyFinder
	companySearcher
	establishmentSearcher
	partnerSearcher
}

func NewServer(
	address string,
	mysql MySQLPinger,
	healthCheckTimeout time.Duration,
	companies companyService,
	search MySQLPinger,
	searchHealthCheckTimeout time.Duration,
	broadSearch broadSearcher,
) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc(
		"GET /health",
		newHealthHandler(
			mysql,
			healthCheckTimeout,
			search,
			searchHealthCheckTimeout,
		),
	)

	mux.HandleFunc(
		"GET /v1/companies",
		newCompanySearchHandler(companies),
	)

	mux.HandleFunc(
		"GET /v1/companies/{cnpj...}",
		newCompanyHandler(companies),
	)

	mux.HandleFunc(
		"GET /v1/establishments",
		newEstablishmentSearchHandler(companies),
	)

	mux.HandleFunc(
		"GET /v1/partners",
		newPartnerSearchHandler(companies),
	)

	mux.HandleFunc(
		"GET /v1/search",
		newBroadSearchHandler(broadSearch),
	)

	return &http.Server{
		Addr:              address,
		Handler:           requestIDMiddleware(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}
}
