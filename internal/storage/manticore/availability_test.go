package manticore

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"testing"
	"time"

	"github.com/danielsantello/go-cnpj-api/internal/company"
	"github.com/danielsantello/go-cnpj-api/internal/config"
	"github.com/danielsantello/go-cnpj-api/internal/database"
	mysql "github.com/go-sql-driver/mysql"
)

func TestNameSearchQueryError(t *testing.T) {
	queryError := errors.New("invalid query")
	enrichmentError := errors.New("MySQL enrichment failed")

	tests := []struct {
		name        string
		err         error
		unavailable bool
	}{
		{"success", nil, false},
		{
			"connection failure",
			&net.OpError{
				Op:  "dial",
				Net: "tcp",
				Err: errors.New("connection refused"),
			},
			true,
		},
		{"broken connection", driver.ErrBadConn, true},
		{"invalid driver connection", mysql.ErrInvalidConn, true},
		{"closed connection", io.EOF, true},
		{"truncated response", io.ErrUnexpectedEOF, true},
		{
			"wrapped connection failure",
			fmt.Errorf("read results: %w", io.EOF),
			true,
		},
		{"operation timeout", context.DeadlineExceeded, true},
		{"query error", queryError, false},
		{"too broad", company.ErrSearchTooBroad, false},
		{"validation error", company.ErrPartnerDocumentInvalid, false},
		{"enrichment error", enrichmentError, false},
		{"cancellation", context.Canceled, false},
		{
			"missing table",
			&mysql.MySQLError{
				Number:  1064,
				Message: "unknown local table(s) 'cnpj_partners_test' in search request",
			},
			true,
		},
		{
			"wrapped missing table",
			fmt.Errorf("count partners: %w", &mysql.MySQLError{
				Number:  1064,
				Message: "unknown local table(s) 'cnpj_partners_test' in search request",
			}),
			true,
		},
		{
			"SQL syntax error",
			&mysql.MySQLError{
				Number:  1064,
				Message: "syntax error near 'SELECT'",
			},
			false,
		},
		{
			"unexpected SQL error code",
			&mysql.MySQLError{
				Number:  1105,
				Message: "unknown local table(s) 'cnpj_partners_test' in search request",
			},
			false,
		},
		{
			"unexpected message suffix",
			&mysql.MySQLError{
				Number:  1064,
				Message: "unknown local table(s) 'cnpj_partners_test' in search request: unexpected failure",
			},
			false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := nameSearchQueryError(
				context.Background(), "search Manticore", test.err,
			)

			if test.err == nil {
				if err != nil {
					t.Fatalf("expected nil, got %v", err)
				}
				return
			}

			if !errors.Is(err, test.err) {
				t.Fatalf("original error was lost: %v", err)
			}
			if errors.Is(err, company.ErrSearchUnavailable) != test.unavailable {
				t.Fatalf("unexpected availability classification: %v", err)
			}
		})
	}
}

func TestNameSearchQueryErrorPreservesRequestCancellation(t *testing.T) {
	tests := []struct {
		name string
		ctx  func() (context.Context, context.CancelFunc)
		want error
	}{
		{
			name: "canceled",
			ctx: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx, cancel
			},
			want: context.Canceled,
		},
		{
			name: "expired",
			ctx: func() (context.Context, context.CancelFunc) {
				return context.WithDeadline(
					context.Background(), time.Now().Add(-time.Second),
				)
			},
			want: context.DeadlineExceeded,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := test.ctx()
			defer cancel()

			err := nameSearchQueryError(ctx, "search Manticore", io.EOF)
			if !errors.Is(err, test.want) {
				t.Fatalf("expected %v, got %v", test.want, err)
			}
			if errors.Is(err, company.ErrSearchUnavailable) {
				t.Fatal("canceled request must not trigger fallback")
			}
		})
	}
}

// A TCP peer can accept connections without sending a MySQL greeting.
// A socket read timeout must allow fallback while the request remains alive.
func TestUnresponsiveManticoreAllowsFallback(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	stop := make(chan struct{})
	defer close(stop)
	defer listener.Close()
	go func() {
		for {
			connection, err := listener.Accept()
			if err != nil {
				return
			}
			go func(connection net.Conn) {
				defer connection.Close()
				<-stop
			}(connection)
		}
	}()
	pool, err := database.OpenManticore(database.ManticoreConfig{
		Host:           "127.0.0.1",
		Port:           uint16(listener.Addr().(*net.TCPAddr).Port),
		ConnectTimeout: time.Second,
		IOTimeout:      25 * time.Millisecond,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = pool.PingContext(ctx)
	if err == nil {
		t.Fatal("unresponsive peer must fail")
	}
	if ctx.Err() != nil {
		t.Fatal("socket timeout must precede request cancellation")
	}
	classified := nameSearchQueryError(ctx, "unresponsive Manticore", err)
	if !errors.Is(classified, company.ErrSearchUnavailable) {
		t.Fatalf("read timeout must allow fallback: %v", classified)
	}
}

func TestMissingTableAllowsFallbackIntegration(t *testing.T) {
	if os.Getenv("CNPJ_API_RUN_INTEGRATION_TESTS") != "1" {
		t.Skip("integration tests are disabled")
	}
	configuration, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if !configuration.ManticoreEnabled {
		t.Fatal("Manticore must be enabled")
	}
	db, err := database.OpenManticore(database.ManticoreConfig{
		Host:           configuration.ManticoreHost,
		Port:           configuration.ManticorePort,
		ConnectTimeout: configuration.ManticoreConnectTimeout,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	missing := fmt.Sprintf("cnpj_partners_missing_probe_%d", time.Now().UnixNano())
	var id uint64
	err = db.QueryRowContext(ctx, "SELECT id FROM "+missing+" LIMIT 1").Scan(&id)
	if err == nil {
		t.Fatal("missing table must return an error")
	}
	classified := nameSearchQueryError(ctx, "missing table", err)
	if !errors.Is(classified, company.ErrSearchUnavailable) || !errors.Is(classified, err) {
		t.Fatalf("missing table must allow fallback and preserve its error: %v", classified)
	}
}
