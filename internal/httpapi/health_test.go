package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type mysqlPingerStub struct {
	err error
}

func (stub mysqlPingerStub) PingContext(_ context.Context) error {
	return stub.err
}

func TestHealthReturnsHealthyWhenMySQLIsAvailable(t *testing.T) {
	mysql := mysqlPingerStub{}

	server := NewServer(":0", mysql, time.Second, nil, nil, time.Second)

	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	response := httptest.NewRecorder()

	server.Handler.ServeHTTP(response, request)

	result := response.Result()
	defer result.Body.Close()

	if result.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, result.StatusCode)
	}

	if contentType := result.Header.Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", contentType)
	}

	var payload healthResponse

	if err := json.NewDecoder(result.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload.Status != "healthy" {
		t.Fatalf("expected status healthy, got %q", payload.Status)
	}

	if payload.Services.MySQL.Status != "available" {
		t.Fatalf(
			"expected MySQL status available, got %q",
			payload.Services.MySQL.Status,
		)
	}

	if payload.Services.Search.Status != "disabled" {
		t.Fatalf(
			"expected search status disabled, got %q",
			payload.Services.Search.Status,
		)
	}
}

func TestHealthReturnsUnhealthyWhenMySQLIsUnavailable(t *testing.T) {
	mysql := mysqlPingerStub{
		err: errors.New("MySQL unavailable"),
	}

	server := NewServer(":0", mysql, time.Second, nil, nil, time.Second)

	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	response := httptest.NewRecorder()

	server.Handler.ServeHTTP(response, request)

	result := response.Result()
	defer result.Body.Close()

	if result.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusServiceUnavailable,
			result.StatusCode,
		)
	}

	var payload healthResponse

	if err := json.NewDecoder(result.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload.Status != "unhealthy" {
		t.Fatalf("expected status unhealthy, got %q", payload.Status)
	}

	if payload.Services.MySQL.Status != "unavailable" {
		t.Fatalf(
			"expected MySQL status unavailable, got %q",
			payload.Services.MySQL.Status,
		)
	}

	if payload.Services.Search.Status != "disabled" {
		t.Fatalf(
			"expected search status disabled, got %q",
			payload.Services.Search.Status,
		)
	}
}

func TestHealthChecksEnabledSearch(t *testing.T) {
	tests := []struct {
		name         string
		searchError  error
		statusCode   int
		status       string
		searchStatus string
	}{
		{
			name:         "available",
			statusCode:   http.StatusOK,
			status:       "healthy",
			searchStatus: "available",
		},
		{
			name:         "unavailable",
			searchError:  errors.New("Manticore unavailable"),
			statusCode:   http.StatusServiceUnavailable,
			status:       "unhealthy",
			searchStatus: "unavailable",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := NewServer(
				":0",
				mysqlPingerStub{},
				time.Second,
				nil,
				mysqlPingerStub{err: test.searchError},
				time.Second,
			)

			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/health", nil)
			server.Handler.ServeHTTP(response, request)

			if response.Code != test.statusCode {
				t.Fatalf("expected status %d, got %d", test.statusCode, response.Code)
			}

			var payload healthResponse
			if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
				t.Fatalf("decode response: %v", err)
			}

			if payload.Status != test.status ||
				payload.Services.MySQL.Status != "available" ||
				payload.Services.Search.Status != test.searchStatus {
				t.Fatalf("unexpected health response: %#v", payload)
			}
		})
	}
}
