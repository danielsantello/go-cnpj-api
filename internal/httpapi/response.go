package httpapi

import (
	"encoding/json"
	"net/http"
)

func writeJSON(
	response http.ResponseWriter,
	statusCode int,
	payload any,
) error {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(statusCode)

	return json.NewEncoder(response).Encode(payload)
}
