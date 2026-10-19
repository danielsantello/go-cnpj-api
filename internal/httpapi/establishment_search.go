package httpapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/danielsantello/go-cnpj-api/internal/company"
)

type establishmentSearcher interface {
	SearchByTradeName(
		ctx context.Context,
		options company.SearchOptions,
	) (company.EstablishmentSearchResult, error)
}

type establishmentSearchData struct {
	Establishments company.EstablishmentSearchResult `json:"establishments"`
}

type establishmentSearchResponse struct {
	Data establishmentSearchData `json:"data"`
}

func newEstablishmentSearchHandler(
	searcher establishmentSearcher,
) http.HandlerFunc {
	return func(
		response http.ResponseWriter,
		request *http.Request,
	) {
		query, queryErrorDetails := parseCompanySearchQuery(
			request.URL.Query(),
		)
		if len(queryErrorDetails) > 0 {
			writeError(
				response,
				request,
				http.StatusBadRequest,
				"INVALID_QUERY_PARAMETERS",
				"One or more query parameters are invalid.",
				queryErrorDetails,
			)

			return
		}

		result, err := searcher.SearchByTradeName(
			request.Context(),
			company.SearchOptions{
				Name:     query.Name,
				Page:     query.Page,
				PageSize: query.PageSize,
			},
		)
		if errors.Is(err, company.ErrTradeNameTooShort) {
			writeError(
				response,
				request,
				http.StatusBadRequest,
				"INVALID_QUERY_PARAMETERS",
				"One or more query parameters are invalid.",
				[]errorDetail{
					{
						Field: "name",
						Reason: "name must contain at least " +
							"3 characters.",
					},
				},
			)

			return
		}
		if errors.Is(err, company.ErrTradeNameTooLong) {
			writeError(
				response,
				request,
				http.StatusBadRequest,
				"INVALID_QUERY_PARAMETERS",
				"One or more query parameters are invalid.",
				[]errorDetail{
					{
						Field: "name",
						Reason: "name must contain at most " +
							"255 characters.",
					},
				},
			)

			return
		}
		if err != nil {
			slog.Error(
				"failed to search establishments by trade name",
				"request_id", requestIDFromContext(request.Context()),
				"error", err,
			)

			writeError(
				response,
				request,
				http.StatusInternalServerError,
				"INTERNAL_ERROR",
				"An unexpected error occurred.",
				nil,
			)

			return
		}

		payload := establishmentSearchResponse{
			Data: establishmentSearchData{
				Establishments: result,
			},
		}

		if err := writeJSON(
			response,
			http.StatusOK,
			payload,
		); err != nil {
			slog.Error(
				"failed to encode establishment search response",
				"request_id", requestIDFromContext(request.Context()),
				"error", err,
			)
		}
	}
}
