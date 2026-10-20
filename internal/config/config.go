package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	httpAddressEnvironmentVariable     = "CNPJ_API_HTTP_ADDRESS"
	shutdownTimeoutEnvironmentVariable = "CNPJ_API_SHUTDOWN_TIMEOUT"
	defaultHTTPAddress                 = ":8080"
	defaultShutdownTimeout             = 10 * time.Second

	mysqlHostEnvironmentVariable           = "CNPJ_API_MYSQL_HOST"
	mysqlPortEnvironmentVariable           = "CNPJ_API_MYSQL_PORT"
	mysqlUserEnvironmentVariable           = "CNPJ_API_MYSQL_USER"
	mysqlPasswordEnvironmentVariable       = "CNPJ_API_MYSQL_PASSWORD"
	mysqlConnectTimeoutEnvironmentVariable = "CNPJ_API_MYSQL_CONNECT_TIMEOUT"
	mysqlDatabaseEnvironmentVariable       = "CNPJ_API_MYSQL_DATABASE"

	defaultMySQLHost           = "127.0.0.1"
	defaultMySQLPort           = uint16(3306)
	defaultMySQLConnectTimeout = 5 * time.Second

	manticoreEnabledEnvironmentVariable        = "CNPJ_API_MANTICORE_ENABLED"
	manticoreHostEnvironmentVariable           = "CNPJ_API_MANTICORE_HOST"
	manticorePortEnvironmentVariable           = "CNPJ_API_MANTICORE_PORT"
	manticoreConnectTimeoutEnvironmentVariable = "CNPJ_API_MANTICORE_CONNECT_TIMEOUT"

	defaultManticoreHost           = "127.0.0.1"
	defaultManticorePort           = uint16(9306)
	defaultManticoreConnectTimeout = 5 * time.Second
)

type Config struct {
	HTTPAddress             string
	ShutdownTimeout         time.Duration
	MySQLHost               string
	MySQLPort               uint16
	MySQLDatabase           string
	MySQLUser               string
	MySQLPassword           string
	MySQLConnectTimeout     time.Duration
	ManticoreEnabled        bool
	ManticoreHost           string
	ManticorePort           uint16
	ManticoreConnectTimeout time.Duration
}

func Default() Config {
	return Config{
		HTTPAddress:             defaultHTTPAddress,
		ShutdownTimeout:         defaultShutdownTimeout,
		MySQLHost:               defaultMySQLHost,
		MySQLPort:               defaultMySQLPort,
		MySQLConnectTimeout:     defaultMySQLConnectTimeout,
		ManticoreHost:           defaultManticoreHost,
		ManticorePort:           defaultManticorePort,
		ManticoreConnectTimeout: defaultManticoreConnectTimeout,
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

	if value, exists := os.LookupEnv(mysqlHostEnvironmentVariable); exists {
		config.MySQLHost = value
	}

	if value, exists := os.LookupEnv(mysqlPortEnvironmentVariable); exists {
		port, err := strconv.ParseUint(value, 10, 16)
		if err != nil {
			return Config{}, fmt.Errorf(
				"%s must be a valid TCP port: %w",
				mysqlPortEnvironmentVariable,
				err,
			)
		}

		config.MySQLPort = uint16(port)
	}

	if value, exists := os.LookupEnv(mysqlDatabaseEnvironmentVariable); exists {
		config.MySQLDatabase = value
	}

	if value, exists := os.LookupEnv(mysqlUserEnvironmentVariable); exists {
		config.MySQLUser = value
	}

	if value, exists := os.LookupEnv(mysqlPasswordEnvironmentVariable); exists {
		config.MySQLPassword = value
	}

	if value, exists := os.LookupEnv(mysqlConnectTimeoutEnvironmentVariable); exists {
		connectTimeout, err := time.ParseDuration(value)
		if err != nil {
			return Config{}, fmt.Errorf(
				"%s must be a valid duration: %w",
				mysqlConnectTimeoutEnvironmentVariable,
				err,
			)
		}

		config.MySQLConnectTimeout = connectTimeout
	}

	if value, exists := os.LookupEnv(manticoreEnabledEnvironmentVariable); exists {
		enabled, err := strconv.ParseBool(value)
		if err != nil {
			return Config{}, fmt.Errorf(
				"%s must be a valid boolean: %w",
				manticoreEnabledEnvironmentVariable, err,
			)
		}
		config.ManticoreEnabled = enabled
	}

	if value, exists := os.LookupEnv(manticoreHostEnvironmentVariable); exists {
		config.ManticoreHost = value
	}

	if value, exists := os.LookupEnv(manticorePortEnvironmentVariable); exists {
		port, err := strconv.ParseUint(value, 10, 16)
		if err != nil {
			return Config{}, fmt.Errorf(
				"%s must be a valid TCP port: %w",
				manticorePortEnvironmentVariable, err,
			)
		}
		config.ManticorePort = uint16(port)
	}

	if value, exists := os.LookupEnv(manticoreConnectTimeoutEnvironmentVariable); exists {
		timeout, err := time.ParseDuration(value)
		if err != nil {
			return Config{}, fmt.Errorf(
				"%s must be a valid duration: %w",
				manticoreConnectTimeoutEnvironmentVariable, err,
			)
		}
		config.ManticoreConnectTimeout = timeout
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

	if strings.TrimSpace(config.MySQLHost) == "" {
		return fmt.Errorf("%s must not be empty", mysqlHostEnvironmentVariable)
	}

	if config.MySQLPort == 0 {
		return fmt.Errorf("%s must be greater than zero", mysqlPortEnvironmentVariable)
	}

	if strings.TrimSpace(config.MySQLDatabase) == "" {
		return fmt.Errorf("%s must not be empty", mysqlDatabaseEnvironmentVariable)
	}

	if strings.TrimSpace(config.MySQLUser) == "" {
		return fmt.Errorf("%s must not be empty", mysqlUserEnvironmentVariable)
	}

	if config.MySQLPassword == "" {
		return fmt.Errorf("%s must not be empty", mysqlPasswordEnvironmentVariable)
	}

	if config.MySQLConnectTimeout <= 0 {
		return fmt.Errorf(
			"%s must be greater than zero",
			mysqlConnectTimeoutEnvironmentVariable,
		)
	}

	if config.ManticoreEnabled {
		if strings.TrimSpace(config.ManticoreHost) == "" {
			return fmt.Errorf(
				"%s must not be empty",
				manticoreHostEnvironmentVariable,
			)
		}

		if config.ManticorePort == 0 {
			return fmt.Errorf(
				"%s must be greater than zero",
				manticorePortEnvironmentVariable,
			)
		}

		if config.ManticoreConnectTimeout <= 0 {
			return fmt.Errorf(
				"%s must be greater than zero",
				manticoreConnectTimeoutEnvironmentVariable,
			)
		}
	}

	return nil
}
