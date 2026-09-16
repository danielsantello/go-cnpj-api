package config

import (
	"fmt"
	"net"
	"os"
	"strings"
)

const (
	httpAddressEnvironmentVariable = "CNPJ_API_HTTP_ADDRESS"
	defaultHTTPAddress             = ":8080"
)

type Config struct {
	HTTPAddress string
}

func Default() Config {
	return Config{
		HTTPAddress: defaultHTTPAddress,
	}
}

func Load() Config {
	config := Default()

	if value, exists := os.LookupEnv(httpAddressEnvironmentVariable); exists {
		config.HTTPAddress = value
	}

	return config
}

func Validate(config Config) error {
	if strings.TrimSpace(config.HTTPAddress) == "" {
		return fmt.Errorf("%s must not be empty", httpAddressEnvironmentVariable)
	}

	if _, _, err := net.SplitHostPort(config.HTTPAddress); err != nil {
		return fmt.Errorf("%s must use host:port format: %w", httpAddressEnvironmentVariable, err)
	}

	return nil
}
