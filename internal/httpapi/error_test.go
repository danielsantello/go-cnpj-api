package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteErrorReturnsStandardErrorResponse(t *testing.T) {
	handler := requestIDMiddleware(
		http.HandlerFunc(
			func(response http.ResponseWriter, request *http.Request) {
				writeError(
					response,
					request,
					http.StatusBadRequest,
					"INVALID_CNPJ",
					"The provided CNPJ is invalid.",
					nil,
				)
			},
		),
	)

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	result := response.Result()
	defer result.Body.Close()

	if result.StatusCode != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			result.StatusCode,
		)
	}

	if contentType := result.Header.Get("Content-Type"); contentType != "application/json" {
		t.Fatalf(
			"expected Content-Type application/json, got %q",
			contentType,
		)
	}

	var payload errorResponse

	if err := json.NewDecoder(result.Body).Decode(&payload); err != nil {
		t.Fatalf("decode error response: %v", err)
	}

	if payload.Error.Code != "INVALID_CNPJ" {
		t.Fatalf(
			"expected error code INVALID_CNPJ, got %q",
			payload.Error.Code,
		)
	}

	if payload.Error.Message != "The provided CNPJ is invalid." {
		t.Fatalf(
			"unexpected error message %q",
			payload.Error.Message,
		)
	}

	if len(payload.Error.Details) != 0 {
		t.Fatalf(
			"expected no error details, got %d",
			len(payload.Error.Details),
		)
	}

	headerRequestID := result.Header.Get("X-Request-ID")

	if payload.Error.RequestID == "" {
		t.Fatal("expected request ID in error response")
	}

	if payload.Error.RequestID != headerRequestID {
		t.Fatalf(
			"expected response request ID %q, got %q",
			headerRequestID,
			payload.Error.RequestID,
		)
	}
}
