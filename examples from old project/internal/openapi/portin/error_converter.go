package portin

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"gitlab.services.mts.ru/salsa/go-base/application/diagnostics"
	"go.uber.org/zap"

	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/model"
)

// Обертка над openapi ErrorResponse,
// чтобы не пользоватся генеренными GetPortInOrdersXXXJSONResponse объектами
// с фиксированным статус-кодом и убрать ветвление под каждый из таких типов.
type errorResponseWrapper struct {
	StatusCode int
	Body       ErrorResponse
}

func (e *errorResponseWrapper) VisitCreatePortInOrderResponse(w http.ResponseWriter) error {
	return e.write(w)
}

func (e *errorResponseWrapper) VisitGetPortInOrderResponse(w http.ResponseWriter) error {
	return e.write(w)
}

func (e *errorResponseWrapper) VisitGetPortInOrdersResponse(w http.ResponseWriter) error {
	return e.write(w)
}

func (e *errorResponseWrapper) write(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.StatusCode)

	return json.NewEncoder(w).Encode(e.Body)
}

func errorResponse(ctx context.Context, err error) *errorResponseWrapper {
	apiErr := model.ToAPIError(err)

	if errors.Is(apiErr, model.ErrInternal) {
		log := diagnostics.LoggerFromContext(ctx).Named("http-server")
		log.Error("unexpected error occurred", zap.Error(err))
	}

	apiErrMessage := apiErr.CompleteMessage()

	return &errorResponseWrapper{
		StatusCode: apiErr.StatusCode,
		Body: ErrorResponse{
			Code:    &apiErr.Code,
			Message: &apiErrMessage,
		},
	}
}
