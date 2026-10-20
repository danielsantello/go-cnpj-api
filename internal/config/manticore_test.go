package config

import (
	"strings"
	"testing"
	"time"
)

func TestManticoreDefaults(t *testing.T) {
	got := Default()

	if got.ManticoreEnabled ||
		got.ManticoreHost != "127.0.0.1" ||
		got.ManticorePort != 9306 ||
		got.ManticoreConnectTimeout != 5*time.Second {
		t.Fatalf("unexpected Manticore defaults: %#v", got)
	}
}

func TestLoadManticoreConfiguration(t *testing.T) {
	t.Setenv(manticoreEnabledEnvironmentVariable, "true")
	t.Setenv(manticoreHostEnvironmentVariable, "localhost")
	t.Setenv(manticorePortEnvironmentVariable, "19306")
	t.Setenv(manticoreConnectTimeoutEnvironmentVariable, "3s")

	got, err := Load()
	if err != nil {
		t.Fatalf("load configuration: %v", err)
	}

	if !got.ManticoreEnabled ||
		got.ManticoreHost != "localhost" ||
		got.ManticorePort != 19306 ||
		got.ManticoreConnectTimeout != 3*time.Second {
		t.Fatalf("unexpected Manticore configuration: %#v", got)
	}
}

func TestLoadManticoreRejectsMalformedValues(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{manticoreEnabledEnvironmentVariable, "invalid"},
		{manticorePortEnvironmentVariable, "abc"},
		{manticorePortEnvironmentVariable, "65536"},
		{manticoreConnectTimeoutEnvironmentVariable, "invalid"},
	}

	for _, test := range tests {
		t.Run(test.name+"/"+test.value, func(t *testing.T) {
			t.Setenv(manticoreEnabledEnvironmentVariable, "false")
			t.Setenv(manticoreHostEnvironmentVariable, "127.0.0.1")
			t.Setenv(manticorePortEnvironmentVariable, "9306")
			t.Setenv(manticoreConnectTimeoutEnvironmentVariable, "5s")
			t.Setenv(test.name, test.value)

			_, err := Load()
			if err == nil || !strings.Contains(err.Error(), test.name) {
				t.Fatalf("expected error naming %s, got %v", test.name, err)
			}
		})
	}
}

func TestValidateManticoreConfiguration(t *testing.T) {
	valid := Default()
	valid.MySQLDatabase = "cnpj_test"
	valid.MySQLUser = "test"
	valid.MySQLPassword = "test"
	valid.ManticoreEnabled = true

	if err := Validate(valid); err != nil {
		t.Fatalf("validate enabled Manticore: %v", err)
	}

	tests := []struct {
		name   string
		change func(*Config)
	}{
		{
			manticoreHostEnvironmentVariable,
			func(c *Config) { c.ManticoreHost = " " },
		},
		{
			manticorePortEnvironmentVariable,
			func(c *Config) { c.ManticorePort = 0 },
		},
		{
			manticoreConnectTimeoutEnvironmentVariable,
			func(c *Config) { c.ManticoreConnectTimeout = 0 },
		},
		{
			manticoreConnectTimeoutEnvironmentVariable,
			func(c *Config) { c.ManticoreConnectTimeout = -time.Second },
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := valid
			test.change(&got)

			err := Validate(got)
			if err == nil || !strings.Contains(err.Error(), test.name) {
				t.Fatalf("expected error naming %s, got %v", test.name, err)
			}

			got.ManticoreEnabled = false
			if err := Validate(got); err != nil {
				t.Fatalf("disabled Manticore should not require connection settings: %v", err)
			}
		})
	}
}
