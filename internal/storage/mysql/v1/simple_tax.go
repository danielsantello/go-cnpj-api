package v1

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/danielsantello/go-cnpj-api/internal/company"
)

const findSimpleTaxByCNPJRootQuery = `
	SELECT
		simple_option_indicator,
		simple_option_date,
		simple_exclusion_date,
		mei_option_indicator,
		mei_option_date,
		mei_exclusion_date
	FROM simple_tax_options
	WHERE cnpj_root = ?
	LIMIT 1
`

func (repository *CompanyRepository) findSimpleTax(
	ctx context.Context,
	cnpjRoot string,
) (*company.SimpleTax, error) {
	var simpleOptionIndicator sql.NullString
	var simpleOptionDate sql.NullTime
	var simpleExclusionDate sql.NullTime
	var meiOptionIndicator sql.NullString
	var meiOptionDate sql.NullTime
	var meiExclusionDate sql.NullTime

	err := repository.db.QueryRowContext(
		ctx,
		findSimpleTaxByCNPJRootQuery,
		cnpjRoot,
	).Scan(
		&simpleOptionIndicator,
		&simpleOptionDate,
		&simpleExclusionDate,
		&meiOptionIndicator,
		&meiOptionDate,
		&meiExclusionDate,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf(
			"query simple tax by CNPJ root: %w",
			err,
		)
	}

	return &company.SimpleTax{
		IsOpted: taxOptionIndicator(
			nullableString(simpleOptionIndicator),
		),
		OptionDate: nullableDate(simpleOptionDate),
		ExclusionDate: nullableDate(
			simpleExclusionDate,
		),
		MEI: company.MEI{
			IsOpted: taxOptionIndicator(
				nullableString(meiOptionIndicator),
			),
			OptionDate: nullableDate(meiOptionDate),
			ExclusionDate: nullableDate(
				meiExclusionDate,
			),
		},
	}, nil
}
