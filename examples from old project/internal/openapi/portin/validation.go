package portin

import (
	"fmt"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"

	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/model"
)

func MapValidationErrToResponse(validationError *openapi3.SchemaError, param *openapi3.Parameter) ErrorResponse {
	code := model.InvalidParamValueErrCode

	switch validationError.SchemaField {
	case "required", "minItems", "minLength":
		code = model.MandatoryParamMissingErrCode
	case "format":
		code = model.InvalidDateFormatErrCode
		validationError.Origin = nil
		validationError.Reason = "дата должена быть в формате ISO 8601"
	case "pattern":
		switch validationError.Schema.Pattern {
		case "^[1-9][0-9]{15}$":
			validationError.Reason = "идентификатор должен быть в формате 1000000001442290"
		case "^[7][0-9]{10}$":
			code = model.InvalidPhoneNumberFormatErrCode
			validationError.Reason = "номер должен быть в формате 79111111111"
		}
	}

	msg := validationError.Reason

	msgArr := strings.Split(validationError.Error(), "\n")
	if len(msgArr) != 0 {
		msg = msgArr[0]
	}

	if param != nil {
		msg = fmt.Sprintf("%s: %s", param.Name, msg)
	}

	return ErrorResponse{
		Code:    &code,
		Message: &msg,
	}
}
