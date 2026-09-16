package httpapi

import (
	"encoding/json"
	"net/http"
)

type healthResponse struct {
	Status string `json:"status"`
}

func healthHandler(response http.ResponseWriter, _ *http.Request) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)

	payload := healthResponse{
		Status: "healthy",
	}

	encoder := json.NewEncoder(response)

	if err := encoder.Encode(payload); err != nil {
		return
	}
}
