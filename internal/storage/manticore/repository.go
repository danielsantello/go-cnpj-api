package manticore

import (
	"context"
	"database/sql"
	"fmt"

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
		SELECT basic_cnpj, cnpj, corporate_name
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

	for rows.Next() {
		var item search.Item
		if err := rows.Scan(
			&item.BasicCNPJ,
			&item.CNPJ,
			&item.CorporateName,
		); err != nil {
			return search.Result{}, fmt.Errorf("scan Manticore result: %w", err)
		}
		result.Items = append(result.Items, item)
	}

	if err := rows.Err(); err != nil {
		return search.Result{}, fmt.Errorf("iterate Manticore results: %w", err)
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
