package database

import (
	"database/sql"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/go-sql-driver/mysql"
)

type MySQLConfig struct {
	Host           string
	Port           uint16
	User           string
	Password       string
	ConnectTimeout time.Duration
}

func OpenMySQL(config MySQLConfig) (*sql.DB, error) {
	driverConfig := mysql.NewConfig()

	driverConfig.Net = "tcp"
	driverConfig.Addr = net.JoinHostPort(
		config.Host,
		strconv.FormatUint(uint64(config.Port), 10),
	)
	driverConfig.User = config.User
	driverConfig.Passwd = config.Password
	driverConfig.ParseTime = true
	driverConfig.Loc = time.UTC
	driverConfig.Timeout = config.ConnectTimeout

	database, err := sql.Open("mysql", driverConfig.FormatDSN())
	if err != nil {
		return nil, fmt.Errorf("open MySQL pool: %w", err)
	}

	return database, nil
}
