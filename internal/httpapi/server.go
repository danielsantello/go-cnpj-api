package httpapi

import (
	"net/http"
	"time"
)

func NewServer(
	address string,
	mysql MySQLPinger,
	healthCheckTimeout time.Duration,
	companyService companyFinder,
) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc(
		"GET /health",
		newHealthHandler(mysql, healthCheckTimeout),
	)

	mux.HandleFunc(
		"GET /v1/companies/{cnpj...}",
		newCompanyHandler(companyService),
	)

	return &http.Server{
		Addr:              address,
		Handler:           requestIDMiddleware(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}
}
