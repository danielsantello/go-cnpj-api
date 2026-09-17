package config

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

const (
	httpAddressEnvironmentVariable     = "CNPJ_API_HTTP_ADDRESS"
	shutdownTimeoutEnvironmentVariable = "CNPJ_API_SHUTDOWN_TIMEOUT"
	defaultHTTPAddress                 = ":8080"
	defaultShutdownTimeout             = 10 * time.Second
)

type Config struct {
	HTTPAddress     string
	ShutdownTimeout time.Duration
}

func Default() Config {
	return Config{
		HTTPAddress:     defaultHTTPAddress,
		ShutdownTimeout: defaultShutdownTimeout,
	}
}

func Load() (Config, error) {
	config := Default()

	if value, exists := os.LookupEnv(httpAddressEnvironmentVariable); exists {
		config.HTTPAddress = value
	}

	if value, exists := os.LookupEnv(shutdownTimeoutEnvironmentVariable); exists {
		shutdownTimeout, err := time.ParseDuration(value)
		if err != nil {
			return Config{}, fmt.Errorf(
				"%s must be a valid duration: %w",
				shutdownTimeoutEnvironmentVariable,
				err,
			)
		}

		config.ShutdownTimeout = shutdownTimeout
	}

	return config, nil
}

func Validate(config Config) error {
	if strings.TrimSpace(config.HTTPAddress) == "" {
		return fmt.Errorf("%s must not be empty", httpAddressEnvironmentVariable)
	}

	if _, _, err := net.SplitHostPort(config.HTTPAddress); err != nil {
		return fmt.Errorf("%s must use host:port format: %w", httpAddressEnvironmentVariable, err)
	}

	if config.ShutdownTimeout <= 0 {
		return fmt.Errorf("%s must be greater than zero", shutdownTimeoutEnvironmentVariable)
	}

	return nil
}
