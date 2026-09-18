package mysqlstorage

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"

	mysqlv1 "github.com/danielsantello/go-cnpj-api/internal/storage/mysql/v1"
)

type indexDefinition struct {
	name    string
	table   string
	columns []string
}

type IndexCreationResult struct {
	Created  []string
	Existing []string
}

type indexGroup struct {
	table   string
	indexes []indexDefinition
}

func indexDefinitionsForFormat(formatVersion uint16) ([]indexDefinition, error) {
	switch formatVersion {
	case 1:
		versionIndexes := mysqlv1.RequiredIndexes()
		indexes := make([]indexDefinition, 0, len(versionIndexes))

		for _, index := range versionIndexes {
			indexes = append(indexes, indexDefinition{
				name:    index.Name,
				table:   index.Table,
				columns: index.Columns,
			})
		}

		return indexes, nil

	default:
		return nil, fmt.Errorf(
			"unsupported database format version: %d",
			formatVersion,
		)
	}
}

func readExistingIndexes(
	ctx context.Context,
	db *sql.DB,
) (map[string]struct{}, error) {
	rows, err := db.QueryContext(
		ctx,
		`
			SELECT
				TABLE_NAME,
				INDEX_NAME
			FROM information_schema.statistics
			WHERE TABLE_SCHEMA = DATABASE()
		`,
	)
	if err != nil {
		return nil, fmt.Errorf("query existing MySQL indexes: %w", err)
	}
	defer rows.Close()

	indexes := make(map[string]struct{})

	for rows.Next() {
		var table string
		var name string

		if err := rows.Scan(&table, &name); err != nil {
			return nil, fmt.Errorf("scan existing MySQL index: %w", err)
		}

		indexes[indexKey(table, name)] = struct{}{}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate existing MySQL indexes: %w", err)
	}

	return indexes, nil
}

func indexKey(table string, name string) string {
	return table + "." + name
}

func quoteIdentifier(identifier string) string {
	escaped := strings.ReplaceAll(identifier, "`", "``")

	return "`" + escaped + "`"
}

func CreateRequiredIndexes(
	ctx context.Context,
	db *sql.DB,
	formatVersion uint16,
) (IndexCreationResult, error) {
	definitions, err := indexDefinitionsForFormat(formatVersion)
	if err != nil {
		return IndexCreationResult{}, fmt.Errorf(
			"get required indexes for database format %d: %w",
			formatVersion,
			err,
		)
	}

	existingIndexes, err := readExistingIndexes(ctx, db)
	if err != nil {
		return IndexCreationResult{}, err
	}

	result := IndexCreationResult{
		Created:  make([]string, 0),
		Existing: make([]string, 0),
	}

	for _, index := range definitions {
		key := indexKey(index.table, index.name)

		if _, exists := existingIndexes[key]; exists {
			result.Existing = append(result.Existing, key)
		}
	}

	groups := groupMissingIndexes(definitions, existingIndexes)

	for _, group := range groups {
		indexNames := make([]string, 0, len(group.indexes))

		for _, index := range group.indexes {
			indexNames = append(indexNames, index.name)
		}

		statement := createIndexGroupStatement(group)

		slog.Info(
			"creating MySQL indexes",
			"table", group.table,
			"indexes", indexNames,
		)

		if _, err := db.ExecContext(ctx, statement); err != nil {
			return result, fmt.Errorf(
				"create MySQL indexes on table %q: %w",
				group.table,
				err,
			)
		}

		slog.Info(
			"MySQL indexes created",
			"table", group.table,
			"indexes", indexNames,
		)

		for _, index := range group.indexes {
			key := indexKey(index.table, index.name)

			result.Created = append(result.Created, key)
			existingIndexes[key] = struct{}{}
		}
	}

	return result, nil
}

func groupMissingIndexes(
	definitions []indexDefinition,
	existingIndexes map[string]struct{},
) []indexGroup {
	groups := make([]indexGroup, 0)
	groupPositions := make(map[string]int)

	for _, index := range definitions {
		key := indexKey(index.table, index.name)

		if _, exists := existingIndexes[key]; exists {
			continue
		}

		position, exists := groupPositions[index.table]
		if !exists {
			position = len(groups)
			groupPositions[index.table] = position

			groups = append(groups, indexGroup{
				table:   index.table,
				indexes: make([]indexDefinition, 0),
			})
		}

		groups[position].indexes = append(
			groups[position].indexes,
			index,
		)
	}

	return groups
}

func createIndexGroupStatement(group indexGroup) string {
	additions := make([]string, 0, len(group.indexes))

	for _, index := range group.indexes {
		columns := make([]string, 0, len(index.columns))

		for _, column := range index.columns {
			columns = append(columns, quoteIdentifier(column))
		}

		additions = append(
			additions,
			fmt.Sprintf(
				"ADD INDEX %s (%s)",
				quoteIdentifier(index.name),
				strings.Join(columns, ", "),
			),
		)
	}

	return fmt.Sprintf(
		"ALTER TABLE %s %s",
		quoteIdentifier(group.table),
		strings.Join(additions, ", "),
	)
}
