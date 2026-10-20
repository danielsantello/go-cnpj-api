package httpapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/danielsantello/go-cnpj-api/internal/search"
)

type broadSearcher interface {
	Search(context.Context, search.Options) (search.Result, error)
}

type broadSearchData struct {
	Companies search.Result `json:"companies"`
}

type broadSearchResponse struct {
	Data broadSearchData `json:"data"`
}

func newBroadSearchHandler(searcher broadSearcher) http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		query, details := parseCompanySearchQuery(request.URL.Query())
		if len(details) > 0 {
			writeError(
				response, request, http.StatusBadRequest,
				"INVALID_QUERY_PARAMETERS",
				"One or more query parameters are invalid.",
				details,
			)
			return
		}

		result, err := searcher.Search(request.Context(), search.Options{
			Name:     query.Name,
			Page:     query.Page,
			PageSize: query.PageSize,
		})

		switch {
		case errors.Is(err, search.ErrInvalidName):
			writeError(
				response, request, http.StatusBadRequest,
				"INVALID_QUERY_PARAMETERS",
				"One or more query parameters are invalid.",
				[]errorDetail{{
					Field:  "name",
					Reason: "name must contain between 3 and 255 characters.",
				}},
			)
			return

		case errors.Is(err, search.ErrInvalidPagination):
			writeError(
				response, request, http.StatusBadRequest,
				"INVALID_QUERY_PARAMETERS",
				"One or more query parameters are invalid.",
				[]errorDetail{{
					Field:  "page",
					Reason: "Invalid search pagination.",
				}},
			)
			return

		case errors.Is(err, search.ErrDisabled):
			writeError(
				response, request, http.StatusServiceUnavailable,
				"SEARCH_DISABLED",
				"Broad search is disabled.",
				nil,
			)
			return

		case err != nil:
			slog.Error(
				"broad search failed",
				"request_id", requestIDFromContext(request.Context()),
				"error", err,
			)
			writeError(
				response, request, http.StatusInternalServerError,
				"INTERNAL_ERROR",
				"An unexpected error occurred.",
				nil,
			)
			return
		}

		payload := broadSearchResponse{
			Data: broadSearchData{Companies: result},
		}
		if err := writeJSON(response, http.StatusOK, payload); err != nil {
			slog.Error(
				"failed to encode broad search response",
				"request_id", requestIDFromContext(request.Context()),
				"error", err,
			)
		}
	}
}
