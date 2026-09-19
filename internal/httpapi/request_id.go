package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
)

type requestIDContextKey struct{}

func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(response http.ResponseWriter, request *http.Request) {
			requestID, err := newRequestID()
			if err != nil {
				http.Error(
					response,
					"Internal Server Error",
					http.StatusInternalServerError,
				)

				return
			}

			response.Header().Set("X-Request-ID", requestID)

			ctx := context.WithValue(
				request.Context(),
				requestIDContextKey{},
				requestID,
			)

			next.ServeHTTP(
				response,
				request.WithContext(ctx),
			)
		},
	)
}

func newRequestID() (string, error) {
	bytes := make([]byte, 16)

	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate request ID: %w", err)
	}

	return hex.EncodeToString(bytes), nil
}

func requestIDFromContext(ctx context.Context) string {
	requestID, _ := ctx.Value(requestIDContextKey{}).(string)

	return requestID
}
