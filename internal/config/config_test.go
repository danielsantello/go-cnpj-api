package config

import (
	"testing"
	"time"
)

func TestDefault(t *testing.T) {
	config := Default()

	expected := ":8080"

	if config.HTTPAddress != expected {
		t.Fatalf(
			"expected HTTP address %q, got %q",
			expected,
			config.HTTPAddress,
		)
	}

	expectedShutdownTimeout := 10 * time.Second

	if config.ShutdownTimeout != expectedShutdownTimeout {
		t.Fatalf(
			"expected shutdown timeout %s, got %s",
			expectedShutdownTimeout,
			config.ShutdownTimeout,
		)
	}
}

func TestLoadOverridesHTTPAddressFromEnvironment(t *testing.T) {
	t.Setenv(httpAddressEnvironmentVariable, "127.0.0.1:9000")

	config, err := Load()
	if err != nil {
		t.Fatalf("load configuration: %v", err)
	}

	expected := "127.0.0.1:9000"

	if config.HTTPAddress != expected {
		t.Fatalf(
			"expected HTTP address %q, got %q",
			expected,
			config.HTTPAddress,
		)
	}
}

func TestValidateAcceptsValidHTTPAddress(t *testing.T) {
	config := Default()
	config.HTTPAddress = "127.0.0.1:9000"

	if err := Validate(config); err != nil {
		t.Fatalf("expected valid configuration, got error: %v", err)
	}
}

func TestValidateRejectsEmptyHTTPAddress(t *testing.T) {
	config := Default()
	config.HTTPAddress = ""

	if err := Validate(config); err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestValidateRejectsHTTPAddressWithoutPort(t *testing.T) {
	config := Default()
	config.HTTPAddress = "localhost"

	if err := Validate(config); err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestLoadOverridesShutdownTimeoutFromEnvironment(t *testing.T) {
	t.Setenv(shutdownTimeoutEnvironmentVariable, "25s")

	config, err := Load()
	if err != nil {
		t.Fatalf("load configuration: %v", err)
	}

	expected := 25 * time.Second

	if config.ShutdownTimeout != expected {
		t.Fatalf(
			"expected shutdown timeout %s, got %s",
			expected,
			config.ShutdownTimeout,
		)
	}
}

func TestLoadRejectsInvalidShutdownTimeout(t *testing.T) {
	t.Setenv(shutdownTimeoutEnvironmentVariable, "invalid")

	_, err := Load()
	if err == nil {
		t.Fatal("expected load error, got nil")
	}
}

func TestValidateRejectsNonPositiveShutdownTimeout(t *testing.T) {
	config := Default()
	config.ShutdownTimeout = 0

	if err := Validate(config); err == nil {
		t.Fatal("expected validation error, got nil")
	}
}
