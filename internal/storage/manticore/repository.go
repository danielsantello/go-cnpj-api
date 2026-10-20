package manticore

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/danielsantello/go-cnpj-api/internal/company"
	"github.com/danielsantello/go-cnpj-api/internal/search"
)

type Repository struct {
	db    *sql.DB
	table string
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db:    db,
		table: "cnpj_companies",
	}
}

func NewRepositoryForTable(db *sql.DB, table string) (*Repository, error) {
	if table != "cnpj_companies" &&
		!indexTableNamePattern.MatchString(table) {
		return nil, fmt.Errorf("invalid Manticore search table name: %q", table)
	}

	return &Repository{
		db:    db,
		table: table,
	}, nil
}

func (repository *Repository) Search(
	ctx context.Context,
	options search.Options,
) (search.Result, error) {
	result := search.Result{
		Items: []search.Item{},
		Pagination: company.NewPagination(
			options.Page, options.PageSize, 0,
		),
	}

	expression := searchExpression(options.Name)
	if expression == "" {
		return result, nil
	}

	total, err := repository.countMatches(ctx, expression)
	if err != nil {
		return search.Result{}, err
	}
	if total > int64(search.MaximumResults) {
		return search.Result{}, search.ErrTooBroad
	}

	result.Pagination = company.NewPagination(
		options.Page, options.PageSize, total,
	)

	// Avoid multiplying an arbitrarily large page number.
	if total == 0 || int64(options.Page) > result.Pagination.TotalPages {
		return result, nil
	}

	offset := int64(options.Page-1) * int64(options.PageSize)

	rows, err := repository.db.QueryContext(ctx, `
		SELECT id, basic_cnpj, cnpj, corporate_name
		FROM `+repository.table+`
		WHERE MATCH(?)
		ORDER BY id ASC
		LIMIT ? OFFSET ?
		OPTION max_matches=1000, cutoff=0
	`, expression, options.PageSize, offset)
	if err != nil {
		return search.Result{}, fmt.Errorf("query Manticore results: %w", err)
	}
	defer rows.Close()

	ids := make([]uint64, 0, options.PageSize)

	for rows.Next() {
		var id uint64
		var item search.Item
		item.MatchedFields = []string{}

		if err := rows.Scan(
			&id,
			&item.BasicCNPJ,
			&item.CNPJ,
			&item.CorporateName,
		); err != nil {
			return search.Result{}, fmt.Errorf("scan Manticore result: %w", err)
		}

		ids = append(ids, id)
		result.Items = append(result.Items, item)
	}

	if err := rows.Err(); err != nil {
		return search.Result{}, fmt.Errorf("iterate Manticore results: %w", err)
	}

	if err := rows.Close(); err != nil {
		return search.Result{}, fmt.Errorf("close Manticore results: %w", err)
	}

	if err := repository.fillMatchedFields(
		ctx, options.Name, ids, result.Items,
	); err != nil {
		return search.Result{}, err
	}

	return result, nil
}

var _ search.Repository = (*Repository)(nil)

func (repository *Repository) countMatches(
	ctx context.Context,
	expression string,
) (int64, error) {
	rows, err := repository.db.QueryContext(ctx, `
		SELECT id
		FROM `+repository.table+`
		WHERE MATCH(?)
		LIMIT 1001
		OPTION max_matches=1001, cutoff=0
	`, expression)
	if err != nil {
		return 0, fmt.Errorf("check Manticore result limit: %w", err)
	}
	defer rows.Close()

	var total int64
	for rows.Next() {
		var id uint64
		if err := rows.Scan(&id); err != nil {
			return 0, fmt.Errorf("scan Manticore match ID: %w", err)
		}
		total++
	}

	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("iterate Manticore match IDs: %w", err)
	}
	return total, nil
}

func (repository *Repository) fillMatchedFields(
	ctx context.Context,
	name string,
	ids []uint64,
	items []search.Item,
) error {
	if len(ids) == 0 {
		return nil
	}

	positions := make(map[uint64]int, len(ids))
	placeholders := make([]string, len(ids))
	for position, id := range ids {
		positions[id] = position
		placeholders[position] = "?"
	}

	for _, field := range []string{
		"corporate_name",
		"trade_names",
		"partner_names",
	} {
		arguments := make([]any, 0, len(ids)+1)
		arguments = append(arguments, fieldSearchExpression(field, name))
		for _, id := range ids {
			arguments = append(arguments, id)
		}

		rows, err := repository.db.QueryContext(ctx, `
			SELECT id
			FROM `+repository.table+`
			WHERE MATCH(?)
				AND id IN (`+strings.Join(placeholders, ",")+`)
			LIMIT 100
			OPTION max_matches=100, cutoff=0
		`, arguments...)
		if err != nil {
			return fmt.Errorf("query matching field %s: %w", field, err)
		}

		for rows.Next() {
			var id uint64
			if err := rows.Scan(&id); err != nil {
				rows.Close()
				return fmt.Errorf("scan matching field %s: %w", field, err)
			}

			if position, exists := positions[id]; exists {
				items[position].MatchedFields = append(
					items[position].MatchedFields, field,
				)
			}
		}

		readErr := rows.Err()
		closeErr := rows.Close()
		if readErr != nil {
			return fmt.Errorf("iterate matching field %s: %w", field, readErr)
		}
		if closeErr != nil {
			return fmt.Errorf("close matching field %s: %w", field, closeErr)
		}
	}

	return nil
}
