package httpapi

import (
	"net/http"
	"time"
)

func NewServer(address string) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)

	return &http.Server{
		Addr:              address,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
}
