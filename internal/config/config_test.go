package config

import "testing"

func TestDefault(t *testing.T) {
	config := Default()

	expected := ":8080"

	if config.HTTPAddress != expected {
		t.Fatalf("expected HTTP address %q, got %q", expected, config.HTTPAddress)
	}
}

func TestLoadOverridesHTTPAddressFromEnvironment(t *testing.T) {
	t.Setenv(httpAddressEnvironmentVariable, "127.0.0.1:9000")

	config := Load()

	expected := "127.0.0.1:9000"

	if config.HTTPAddress != expected {
		t.Fatalf("expected HTTP address %q, got %q", expected, config.HTTPAddress)
	}
}

func TestValidateAcceptsValidHTTPAddress(t *testing.T) {
	config := Config{
		HTTPAddress: "127.0.0.1:9000",
	}

	if err := Validate(config); err != nil {
		t.Fatalf("expected valid configuration, got error: %v", err)
	}
}

func TestValidateRejectsEmptyHTTPAddress(t *testing.T) {
	config := Config{
		HTTPAddress: "",
	}

	if err := Validate(config); err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestValidateRejectsHTTPAddressWithoutPort(t *testing.T) {
	config := Config{
		HTTPAddress: "localhost",
	}

	if err := Validate(config); err == nil {
		t.Fatal("expected validation error, got nil")
	}
}
