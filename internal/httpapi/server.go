package httpapi

import (
	"net/http"
	"time"
)

func NewServer(
	address string,
	mysql MySQLPinger,
	healthCheckTimeout time.Duration,
) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc(
		"GET /health",
		newHealthHandler(mysql, healthCheckTimeout),
	)

	return &http.Server{
		Addr:              address,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
}
