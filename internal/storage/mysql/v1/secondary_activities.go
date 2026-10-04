package v1

import (
	"context"
	"fmt"
	"strings"

	"github.com/danielsantello/go-cnpj-api/internal/company"
)

const findSecondaryEconomicActivitiesQuery = `
	SELECT
		activity.code,
		activity.name
	FROM economic_activities AS activity
	WHERE FIND_IN_SET(activity.code, ?) > 0
`

func (repository *CompanyRepository) findSecondaryEconomicActivities(
	ctx context.Context,
	codes []string,
) ([]company.CodeDescription, error) {
	result := make([]company.CodeDescription, 0, len(codes))

	if len(codes) == 0 {
		return result, nil
	}

	rows, err := repository.db.QueryContext(
		ctx,
		findSecondaryEconomicActivitiesQuery,
		strings.Join(codes, ","),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"query secondary economic activities: %w",
			err,
		)
	}
	defer rows.Close()

	descriptions := make(map[string]string, len(codes))

	for rows.Next() {
		var code string
		var description string

		if err := rows.Scan(&code, &description); err != nil {
			return nil, fmt.Errorf(
				"scan secondary economic activity: %w",
				err,
			)
		}

		if _, exists := descriptions[code]; !exists {
			descriptions[code] = description
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate secondary economic activities: %w",
			err,
		)
	}

	for _, code := range codes {
		activity := company.CodeDescription{
			Code: code,
		}

		description, exists := descriptions[code]
		if exists {
			activity.Description = &description
		}

		result = append(result, activity)
	}

	return result, nil
}
