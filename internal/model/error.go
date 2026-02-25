package model

import (
	"errors"
	"fmt"
	"net/http"
)

const (
	InvalidPhoneNumberFormatErrCode = "MNP-4002.INVALID_PHONE_NUMBER_FORMAT"
	InvalidDateFormatErrCode        = "MNP-4003.INVALID_DATE_FORMAT"
	MandatoryParamMissingErrCode    = "MNP-4004.MANDATORY_PARAMETER_MISSING"
	InvalidParamValueErrCode        = "MNP-4005.INVALID_PARAMETER_VALUE"
	PortationExistsErrCode          = "MNP-4007.PORTATION_REQUEST_EXISTS"
	OrderNotFoundErrCode            = "MNP-4010.ORDER_ID_NOT_FOUND"
	DocNotAllowedErrCode            = "MNP-4012.DOC_NOT_ALLOWED"
	DocobjectNotFoundErrCode        = "MNP-4013.DOC_OBJECT_NOT_FOUND"
	InternalErrCode                 = "MNP-5001.INTERNAL_ERROR"
)

type APIError struct {
	Code        string
	Message     string
	StatusCode  int
	Description string
}

var (
	ErrInvalidPhoneNumberFormat = &APIError{
		Code:        InvalidPhoneNumberFormatErrCode,
		Message:     "Некорректный формат номера.",
		StatusCode:  http.StatusBadRequest,
		Description: "msisdn должен быть в формате 79111111111",
	}

	ErrPortationRequestExists = &APIError{
		Code:       PortationExistsErrCode,
		Message:    "Заявка на портацию для данного номера уже существует.",
		StatusCode: http.StatusConflict,
	}

	ErrOrderIDNotFound = &APIError{
		Code:       OrderNotFoundErrCode,
		Message:    "Не найдена заявка.",
		StatusCode: http.StatusNotFound,
	}

	ErrDocNotAllowed = &APIError{
		Code:       DocNotAllowedErrCode,
		Message:    "Нет доступа к документу заявления по указанному пути.",
		StatusCode: http.StatusBadRequest,
	}

	ErrDocObjectNotFound = &APIError{
		Code:       DocNotAllowedErrCode,
		Message:    "Документ заявления отсутствует по указанному пути.",
		StatusCode: http.StatusBadRequest,
	}

	ErrInternal = &APIError{
		Code: InternalErrCode,
	}
)

func (e *APIError) Is(other error) bool {
	var oAPIError *APIError
	if !errors.As(other, &oAPIError) {
		return false
	}

	return e.Code == oAPIError.Code
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.CompleteMessage())
}

func (e *APIError) CompleteMessage() string {
	if e.Description == "" {
		return e.Message
	}

	return fmt.Sprintf("%s : %s", e.Message, e.Description)
}

func ErrMandatoryParameterMissing(parameter string) *APIError {
	return &APIError{
		Code:        MandatoryParamMissingErrCode,
		Message:     "Отсутствует обязательный параметр.",
		StatusCode:  http.StatusBadRequest,
		Description: fmt.Sprintf("Параметр '%s' является обязательным", parameter),
	}
}

func ErrInvalidParameterValue(parameter string) *APIError {
	return &APIError{
		Code:        InvalidParamValueErrCode,
		Message:     "Некорректное значение параметра.",
		StatusCode:  http.StatusBadRequest,
		Description: fmt.Sprintf("Некорректное значение параметра %s", parameter),
	}
}

func ErrInvalidDateFormat(field string) *APIError {
	return &APIError{
		Code:        InvalidDateFormatErrCode,
		Message:     "Некорректный формат даты.",
		StatusCode:  http.StatusBadRequest,
		Description: fmt.Sprintf("%s должен быть в формате ISO 8601", field),
	}
}
func InternalError(err error) *APIError {
	return &APIError{
		Code:        InternalErrCode,
		Message:     "Внутренная ошибка сервера.",
		Description: err.Error(),
		StatusCode:  http.StatusInternalServerError,
	}
}

func ToAPIError(err error) *APIError {
	var apiErr *APIError
	if err != nil && errors.As(err, &apiErr) {
		return apiErr
	}

	return InternalError(err)
}
