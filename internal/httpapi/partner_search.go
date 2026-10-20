package httpapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/danielsantello/go-cnpj-api/internal/company"
)

type partnerSearcher interface {
	SearchByPartnerName(
		context.Context,
		company.PartnerSearchOptions,
	) (company.PartnerSearchResult, error)
}

type partnerSearchData struct {
	Partners company.PartnerSearchResult `json:"partners"`
}

type partnerSearchResponse struct {
	Data partnerSearchData `json:"data"`
}

func newPartnerSearchHandler(searcher partnerSearcher) http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		values := request.URL.Query()
		query, details := parseCompanySearchQuery(values)

		if len(details) > 0 {
			writeError(
				response, request, http.StatusBadRequest,
				"INVALID_QUERY_PARAMETERS",
				"One or more query parameters are invalid.",
				details,
			)
			return
		}

		options := company.PartnerSearchOptions{
			Name:     query.Name,
			Page:     query.Page,
			PageSize: query.PageSize,
		}

		if values.Has("document") {
			document := values.Get("document")
			options.Document = &document
		}

		result, err := searcher.SearchByPartnerName(
			request.Context(),
			options,
		)

		var detail errorDetail

		switch {
		case errors.Is(err, company.ErrPartnerNameTooShort):
			detail = errorDetail{
				Field: "name", Reason: "name must contain at least 3 characters.",
			}
		case errors.Is(err, company.ErrPartnerNameTooLong):
			detail = errorDetail{
				Field: "name", Reason: "name must contain at most 255 characters.",
			}
		case errors.Is(err, company.ErrPartnerDocumentInvalid):
			detail = errorDetail{
				Field:  "document",
				Reason: "document must contain between 1 and 20 characters.",
			}
		}

		if detail.Field != "" {
			writeError(
				response, request, http.StatusBadRequest,
				"INVALID_QUERY_PARAMETERS",
				"One or more query parameters are invalid.",
				[]errorDetail{detail},
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
			slog.Error(
				"failed to search companies by partner name",
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

		payload := partnerSearchResponse{
			Data: partnerSearchData{Partners: result},
		}

		if err := writeJSON(response, http.StatusOK, payload); err != nil {
			slog.Error(
				"failed to encode partner search response",
				"request_id", requestIDFromContext(request.Context()),
				"error", err,
			)
		}
	}
}
