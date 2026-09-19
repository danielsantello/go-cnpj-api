package httpapi

import (
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestIDMiddlewareAddsHeaderAndContext(t *testing.T) {
	var contextRequestID string

	next := http.HandlerFunc(
		func(response http.ResponseWriter, request *http.Request) {
			contextRequestID = requestIDFromContext(request.Context())
			response.WriteHeader(http.StatusNoContent)
		},
	)

	handler := requestIDMiddleware(next)

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	result := response.Result()
	defer result.Body.Close()

	headerRequestID := result.Header.Get("X-Request-ID")

	if headerRequestID == "" {
		t.Fatal("expected X-Request-ID header, got empty value")
	}

	if contextRequestID == "" {
		t.Fatal("expected request ID in context, got empty value")
	}

	if contextRequestID != headerRequestID {
		t.Fatalf(
			"expected context request ID %q, got %q",
			headerRequestID,
			contextRequestID,
		)
	}

	decoded, err := hex.DecodeString(headerRequestID)
	if err != nil {
		t.Fatalf("decode request ID: %v", err)
	}

	expectedBytes := 16

	if len(decoded) != expectedBytes {
		t.Fatalf(
			"expected request ID with %d bytes, got %d",
			expectedBytes,
			len(decoded),
		)
	}
}
