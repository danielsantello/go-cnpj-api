package database

import (
	"context"
	"database/sql"
	"fmt"
)

func Ping(ctx context.Context, database *sql.DB) error {
	if err := database.PingContext(ctx); err != nil {
		return fmt.Errorf("ping MySQL: %w", err)
	}

	return nil
}
