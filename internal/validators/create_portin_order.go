package validators

import (
	"context"
	"fmt"
	"strings"

	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/model"
)

type CreatePortInOrderValidator struct {
	model *model.CreatePortinOrder
}

func NewCreatePortInOrderValidator(cOrdModel *model.CreatePortinOrder) *CreatePortInOrderValidator {
	return &CreatePortInOrderValidator{
		model: cOrdModel,
	}
}

func (v *CreatePortInOrderValidator) Validate(_ context.Context) *model.APIError {
	if err := v.validatePortationNumbers(v.model.PortationNumbers); err != nil {
		return err
	}

	if err := v.validateContract(&v.model.Contract); err != nil {
		return err
	}

	if err := v.validateProcessType(v.model.ProcessType); err != nil {
		return err
	}

	if err := v.validateCustomerInfo(v.model); err != nil {
		return err
	}

	return nil
}

func (v *CreatePortInOrderValidator) validatePortationNumbers(portationNumbers []model.PortationNumber) *model.APIError {
	for i, portNum := range portationNumbers {
		if err := v.validateSinglePortationNumber(portNum, i); err != nil {
			return err
		}
	}

	return nil
}

func (v *CreatePortInOrderValidator) validateSinglePortationNumber(portNum model.PortationNumber, index int) *model.APIError {
	if (portNum.TelcoAccount.ID == nil || strings.TrimSpace(*portNum.TelcoAccount.ID) == "") &&
		(portNum.TelcoAccount.Msisdn == nil || strings.TrimSpace(*portNum.TelcoAccount.Msisdn) == "") {
		return model.ErrMandatoryParameterMissing(
			fmt.Sprintf("portationNumbers[%d].telcoAccount.id или portationNumbers[%d].telcoAccount.msisdn", index, index),
		)
	}

	return nil
}

func (v *CreatePortInOrderValidator) validateContract(contract *model.Contract) *model.APIError {
	if contract.DocumentURL == "" {
		return model.ErrMandatoryParameterMissing("contract.documentUrl")
	}

	return nil
}

func (v *CreatePortInOrderValidator) validateCustomerInfo(createOrder *model.CreatePortinOrder) *model.APIError {
	filledCount := v.countFilledCustomerTypes(createOrder)

	if filledCount != 1 {
		return model.ErrInvalidParameterValue("должно быть заполнено одно и только одно из полей: person, company, government, individual")
	}

	return nil
}

func (v *CreatePortInOrderValidator) validateProcessType(processType *string) *model.APIError {
	if processType != nil {
		switch *processType {
		case "ShortTimePort", "LongTimePort", "GOS":
			// Valid values
		default:
			return &model.APIError{
				Code:        model.InvalidParamValueErrCode,
				Message:     "processType должно быть одним из: ShortTimePort, LongTimePort, GOS",
				StatusCode:  400,
				Description: "processType должно быть одним из: ShortTimePort, LongTimePort, GOS",
			}
		}
	}

	return nil
}

func (v *CreatePortInOrderValidator) countFilledCustomerTypes(createOrder *model.CreatePortinOrder) int {
	count := 0
	if createOrder.Person != nil {
		count++
	}

	if createOrder.Company != nil {
		count++
	}

	if createOrder.Government != nil {
		count++
	}

	if createOrder.Individual != nil {
		count++
	}

	return count
}
