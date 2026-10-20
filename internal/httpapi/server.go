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
) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc(
		"GET /health",
		newHealthHandler(mysql, healthCheckTimeout),
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

	return &http.Server{
		Addr:              address,
		Handler:           requestIDMiddleware(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}
}
