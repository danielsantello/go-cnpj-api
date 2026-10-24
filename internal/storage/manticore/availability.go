package manticore

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"net"
	"regexp"

	"github.com/danielsantello/go-cnpj-api/internal/company"
	mysql "github.com/go-sql-driver/mysql"
)

var missingSearchTablePattern = regexp.MustCompile(
	`^unknown local table\(s\) '[a-z0-9_]+' in search request$`,
)

func isMissingSearchTable(err error) bool {
	var mysqlError *mysql.MySQLError
	return errors.As(err, &mysqlError) &&
		mysqlError.Number == 1064 &&
		missingSearchTablePattern.MatchString(mysqlError.Message)
}

func nameSearchQueryError(
	ctx context.Context,
	operation string,
	err error,
) error {
	if err == nil {
		return nil
	}

	if contextErr := ctx.Err(); contextErr != nil {
		return fmt.Errorf("%s: %w", operation, contextErr)
	}

	var networkError net.Error
	unavailable := errors.As(err, &networkError) ||
		errors.Is(err, driver.ErrBadConn) ||
		errors.Is(err, mysql.ErrInvalidConn) ||
		errors.Is(err, io.EOF) ||
		errors.Is(err, io.ErrUnexpectedEOF) ||
		errors.Is(err, context.DeadlineExceeded) ||
		isMissingSearchTable(err)

	if unavailable {
		return fmt.Errorf(
			"%s: %w: %w",
			operation,
			company.ErrSearchUnavailable,
			err,
		)
	}

	return fmt.Errorf("%s: %w", operation, err)
}
