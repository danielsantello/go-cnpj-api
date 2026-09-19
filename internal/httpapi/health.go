package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

type MySQLPinger interface {
	PingContext(context.Context) error
}

type serviceHealth struct {
	Status string `json:"status"`
}

type healthServices struct {
	MySQL  serviceHealth `json:"mysql"`
	Search serviceHealth `json:"search"`
}

type healthResponse struct {
	Status   string         `json:"status"`
	Services healthServices `json:"services"`
}

func newHealthHandler(mysql MySQLPinger, timeout time.Duration) http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		pingContext, cancelPing := context.WithTimeout(
			request.Context(),
			timeout,
		)
		defer cancelPing()

		statusCode := http.StatusOK

		payload := healthResponse{
			Status: "healthy",
			Services: healthServices{
				MySQL: serviceHealth{
					Status: "available",
				},
				Search: serviceHealth{
					Status: "disabled",
				},
			},
		}

		if err := mysql.PingContext(pingContext); err != nil {
			slog.Error("MySQL health check failed", "error", err)

			statusCode = http.StatusServiceUnavailable
			payload.Status = "unhealthy"
			payload.Services.MySQL.Status = "unavailable"
		}

		if err := writeJSON(response, statusCode, payload); err != nil {
			slog.Error("failed to encode health response", "error", err)
		}
	}
}
