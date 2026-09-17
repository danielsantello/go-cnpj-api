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

	if config.MySQLHost != "127.0.0.1" {
		t.Fatalf(
			"expected MySQL host %q, got %q",
			"127.0.0.1",
			config.MySQLHost,
		)
	}

	if config.MySQLPort != 3306 {
		t.Fatalf(
			"expected MySQL port %d, got %d",
			3306,
			config.MySQLPort,
		)
	}

	if config.MySQLUser != "" {
		t.Fatal("expected empty default MySQL user")
	}

	if config.MySQLPassword != "" {
		t.Fatal("expected empty default MySQL password")
	}

	expectedMySQLConnectTimeout := 5 * time.Second

	if config.MySQLConnectTimeout != expectedMySQLConnectTimeout {
		t.Fatalf(
			"expected MySQL connect timeout %s, got %s",
			expectedMySQLConnectTimeout,
			config.MySQLConnectTimeout,
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

func TestValidateAcceptsValidConfiguration(t *testing.T) {
	config := validConfig()
	config.HTTPAddress = "127.0.0.1:9000"

	if err := Validate(config); err != nil {
		t.Fatalf("expected valid configuration, got error: %v", err)
	}
}

func TestValidateRejectsEmptyHTTPAddress(t *testing.T) {
	config := validConfig()
	config.HTTPAddress = ""

	if err := Validate(config); err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestValidateRejectsHTTPAddressWithoutPort(t *testing.T) {
	config := validConfig()
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
	config := validConfig()
	config.ShutdownTimeout = 0

	if err := Validate(config); err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestLoadOverridesMySQLConfigurationFromEnvironment(t *testing.T) {
	t.Setenv(mysqlHostEnvironmentVariable, "mysql")
	t.Setenv(mysqlPortEnvironmentVariable, "3307")
	t.Setenv(mysqlUserEnvironmentVariable, "cnpj_api")
	t.Setenv(mysqlPasswordEnvironmentVariable, "secret")
	t.Setenv(mysqlConnectTimeoutEnvironmentVariable, "8s")

	config, err := Load()
	if err != nil {
		t.Fatalf("load configuration: %v", err)
	}

	if config.MySQLHost != "mysql" {
		t.Fatalf(
			"expected MySQL host %q, got %q",
			"mysql",
			config.MySQLHost,
		)
	}

	if config.MySQLPort != 3307 {
		t.Fatalf(
			"expected MySQL port %d, got %d",
			3307,
			config.MySQLPort,
		)
	}

	if config.MySQLUser != "cnpj_api" {
		t.Fatalf(
			"expected MySQL user %q, got %q",
			"cnpj_api",
			config.MySQLUser,
		)
	}

	if config.MySQLPassword != "secret" {
		t.Fatal("expected MySQL password to be loaded from environment")
	}

	expectedConnectTimeout := 8 * time.Second

	if config.MySQLConnectTimeout != expectedConnectTimeout {
		t.Fatalf(
			"expected MySQL connect timeout %s, got %s",
			expectedConnectTimeout,
			config.MySQLConnectTimeout,
		)
	}
}

func TestLoadRejectsInvalidMySQLPort(t *testing.T) {
	t.Setenv(mysqlPortEnvironmentVariable, "invalid")

	_, err := Load()
	if err == nil {
		t.Fatal("expected load error, got nil")
	}
}

func TestLoadRejectsInvalidMySQLConnectTimeout(t *testing.T) {
	t.Setenv(mysqlConnectTimeoutEnvironmentVariable, "invalid")

	_, err := Load()
	if err == nil {
		t.Fatal("expected load error, got nil")
	}
}

func validConfig() Config {
	config := Default()
	config.MySQLUser = "cnpj_api"
	config.MySQLPassword = "secret"

	return config
}

func TestValidateRejectsEmptyMySQLHost(t *testing.T) {
	config := validConfig()
	config.MySQLHost = ""

	if err := Validate(config); err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestValidateRejectsZeroMySQLPort(t *testing.T) {
	config := validConfig()
	config.MySQLPort = 0

	if err := Validate(config); err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestValidateRejectsEmptyMySQLUser(t *testing.T) {
	config := validConfig()
	config.MySQLUser = ""

	if err := Validate(config); err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestValidateRejectsEmptyMySQLPassword(t *testing.T) {
	config := validConfig()
	config.MySQLPassword = ""

	if err := Validate(config); err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestValidateRejectsNonPositiveMySQLConnectTimeout(t *testing.T) {
	config := validConfig()
	config.MySQLConnectTimeout = 0

	if err := Validate(config); err == nil {
		t.Fatal("expected validation error, got nil")
	}
}
