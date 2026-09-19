package httpapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/danielsantello/go-cnpj-api/internal/company"
)

type companyFinder interface {
	FindByCNPJ(
		ctx context.Context,
		input string,
	) (company.Details, error)
}

type companyResponse struct {
	Data company.Details `json:"data"`
}

func newCompanyHandler(finder companyFinder) http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		details, err := finder.FindByCNPJ(
			request.Context(),
			request.PathValue("cnpj"),
		)
		if errors.Is(err, company.ErrCNPJTooLong) {
			writeError(
				response,
				request,
				http.StatusBadRequest,
				"INVALID_CNPJ",
				"The provided CNPJ is invalid.",
				[]errorDetail{
					{
						Field:  "cnpj",
						Reason: "CNPJ must contain at most 14 characters after formatting is removed.",
					},
				},
			)

			return
		}
		if errors.Is(err, company.ErrCompanyNotFound) {
			writeError(
				response,
				request,
				http.StatusNotFound,
				"COMPANY_NOT_FOUND",
				"No company was found for the provided CNPJ.",
				nil,
			)

			return
		}
		if err != nil {
			requestID := requestIDFromContext(request.Context())

			slog.Error(
				"failed to find company",
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

		payload := companyResponse{
			Data: details,
		}

		if err := writeJSON(
			response,
			http.StatusOK,
			payload,
		); err != nil {
			slog.Error(
				"failed to encode company response",
				"request_id",
				requestIDFromContext(request.Context()),
				"error",
				err,
			)
		}
	}
}
