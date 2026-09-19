package httpapi

import (
	"log/slog"
	"net/http"
)

type errorDetail struct {
	Field  string `json:"field"`
	Reason string `json:"reason"`
}

type apiError struct {
	Code      string        `json:"code"`
	Message   string        `json:"message"`
	Details   []errorDetail `json:"details"`
	RequestID string        `json:"request_id"`
}

type errorResponse struct {
	Error apiError `json:"error"`
}

func writeError(
	response http.ResponseWriter,
	request *http.Request,
	statusCode int,
	code string,
	message string,
	details []errorDetail,
) {
	if details == nil {
		details = make([]errorDetail, 0)
	}

	requestID := requestIDFromContext(request.Context())

	payload := errorResponse{
		Error: apiError{
			Code:      code,
			Message:   message,
			Details:   details,
			RequestID: requestID,
		},
	}

	if err := writeJSON(response, statusCode, payload); err != nil {
		slog.Error(
			"failed to encode error response",
			"request_id", requestID,
			"error", err,
		)
	}
}
