package manticore

import (
	"context"
	"database/sql"
	"fmt"
)

func validSearchTableName(table string) bool {
	return table == "cnpj_companies" ||
		indexTableNamePattern.MatchString(table) ||
		establishmentTableNamePattern.MatchString(table) ||
		partnerTableNamePattern.MatchString(table)
}

func ValidateDropTable(table string, activeTables ...string) error {
	if !validSearchTableName(table) {
		return fmt.Errorf("invalid Manticore table name: %q", table)
	}
	if len(activeTables) == 0 {
		return fmt.Errorf("active Manticore tables must be supplied")
	}
	for _, activeTable := range activeTables {
		if !validSearchTableName(activeTable) {
			return fmt.Errorf("invalid active Manticore table name: %q", activeTable)
		}
		if table == activeTable {
			return fmt.Errorf("cannot drop active Manticore table: %s", table)
		}
	}
	return nil
}

func DropTable(
	ctx context.Context,
	db *sql.DB,
	table string,
	activeTables ...string,
) error {
	if err := ValidateDropTable(table, activeTables...); err != nil {
		return err
	}

	rows, err := db.QueryContext(ctx, "SHOW TABLES")
	if err != nil {
		return fmt.Errorf("list Manticore tables: %w", err)
	}

	found := false
	for rows.Next() {
		var name, tableType string
		if err := rows.Scan(&name, &tableType); err != nil {
			rows.Close()
			return fmt.Errorf("scan Manticore table: %w", err)
		}
		if name == table {
			found = true
		}
	}

	readErr := rows.Err()
	closeErr := rows.Close()
	if readErr != nil {
		return fmt.Errorf("iterate Manticore tables: %w", readErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close Manticore table list: %w", closeErr)
	}

	if !found {
		return fmt.Errorf("Manticore table does not exist: %s", table)
	}

	if _, err := db.ExecContext(ctx, "DROP TABLE "+table); err != nil {
		return fmt.Errorf("drop Manticore table %s: %w", table, err)
	}

	return nil
}
