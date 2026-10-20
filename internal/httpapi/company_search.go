package httpapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/danielsantello/go-cnpj-api/internal/company"
)

type companySearcher interface {
	SearchByCorporateName(
		ctx context.Context,
		options company.SearchOptions,
	) (company.SearchResult, error)
}

type companySearchData struct {
	Companies company.SearchResult `json:"companies"`
}

type companySearchResponse struct {
	Data companySearchData `json:"data"`
}

func newCompanySearchHandler(
	searcher companySearcher,
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

		result, err := searcher.SearchByCorporateName(
			request.Context(),
			company.SearchOptions{
				Name:     query.Name,
				Page:     query.Page,
				PageSize: query.PageSize,
			},
		)
		if errors.Is(err, company.ErrCorporateNameTooShort) {
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
		if errors.Is(err, company.ErrCorporateNameTooLong) {
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
		if errors.Is(err, company.ErrSearchTooBroad) {
			writeError(
				response,
				request,
				http.StatusUnprocessableEntity,
				"SEARCH_TOO_BROAD",
				"Too many results were found. Refine your search.",
				nil,
			)
			return
		}
		if err != nil {
			requestID := requestIDFromContext(request.Context())

			slog.Error(
				"failed to search companies by corporate name",
				"request_id", requestID,
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

		payload := companySearchResponse{
			Data: companySearchData{
				Companies: result,
			},
		}

		if err := writeJSON(
			response,
			http.StatusOK,
			payload,
		); err != nil {
			slog.Error(
				"failed to encode company search response",
				"request_id",
				requestIDFromContext(request.Context()),
				"error",
				err,
			)
		}
	}
}
