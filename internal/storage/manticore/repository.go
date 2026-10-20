package manticore

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/danielsantello/go-cnpj-api/internal/company"
	"github.com/danielsantello/go-cnpj-api/internal/search"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
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

	var total int64
	err := repository.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM cnpj_companies
		WHERE MATCH(?)
	`, expression).Scan(&total)
	if err != nil {
		return search.Result{}, fmt.Errorf("count Manticore results: %w", err)
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
		FROM cnpj_companies
		WHERE MATCH(?)
		ORDER BY id ASC
		LIMIT ? OFFSET ?
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
