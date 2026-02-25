package validators

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/model"
	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/orchestrator/kafka/mnpevent"
)

var msisdnPattern = regexp.MustCompile(`^7\d{10}$`)

type PortInPatchEventValidator struct {
	event *mnpevent.PortInPatch
}

func NewPortInPatchEventValidator(event *mnpevent.PortInPatch) *PortInPatchEventValidator {
	return &PortInPatchEventValidator{
		event: event,
	}
}

func (v *PortInPatchEventValidator) Validate(_ context.Context) *model.APIError {
	if v.event == nil {
		return model.ErrMandatoryParameterMissing("portInPatch")
	}

	if strings.TrimSpace(v.event.Data.OrderID) == "" && strings.TrimSpace(v.event.Data.CDBProcessID) == "" {
		return model.ErrMandatoryParameterMissing("orderId|cdbProcessId")
	}

	hasState := v.event.Data.OrderState != nil && strings.TrimSpace(v.event.Data.OrderState.Code) != ""
	hasDueDate := strings.TrimSpace(v.event.Data.DueDate) != ""
	hasPN := len(v.event.Data.PortationNumbers) > 0

	if !hasState && !hasDueDate && !hasPN {
		return model.ErrMandatoryParameterMissing("orderState|dueDate|portationNumbers")
	}

	if err := v.validateRequired(hasState, hasDueDate, hasPN); err != nil {
		return err
	}

	if v.event.Date.IsZero() {
		return model.ErrMandatoryParameterMissing("date")
	}

	return nil
}

func (v *PortInPatchEventValidator) validateRequired(hasState, hasDueDate, hasPN bool) *model.APIError {
	if hasState {
		if v.event.Data.OrderState.StatusDate != nil {
			if _, err := time.Parse(time.RFC3339, *v.event.Data.OrderState.StatusDate); err != nil {
				return model.ErrInvalidDateFormat("orderState.statusDate")
			}
		}
	}

	if hasDueDate {
		if _, err := time.Parse(time.RFC3339, v.event.Data.DueDate); err != nil {
			return model.ErrInvalidParameterValue("dueDate")
		}
	}

	if hasPN {
		if err := v.validatePortationNumbers(v.event.Data.PortationNumbers); err != nil {
			return err
		}
	}

	return nil
}

func (v *PortInPatchEventValidator) validatePortationNumbers(portationNumbers []mnpevent.PortationNumberDTO) *model.APIError {
	if len(portationNumbers) == 0 {
		return nil
	}

	for i, portNum := range portationNumbers {
		if err := v.validateSinglePortationNumber(portNum, i); err != nil {
			return err
		}
	}

	return nil
}

func (v *PortInPatchEventValidator) validateSinglePortationNumber(portNum mnpevent.PortationNumberDTO, index int) *model.APIError {
	if portNum.MSISDN == "" {
		return model.ErrMandatoryParameterMissing(fmt.Sprintf("portationNumbers[%d].msisdn", index))
	}

	if !msisdnPattern.MatchString(strings.TrimSpace(portNum.MSISDN)) {
		return model.ErrInvalidPhoneNumberFormat
	}

	if portNum.Status == nil {
		return model.ErrMandatoryParameterMissing(fmt.Sprintf("portationNumbers[%d].status", index))
	}

	if strings.TrimSpace(portNum.Status.Code) == "" {
		return model.ErrMandatoryParameterMissing(fmt.Sprintf("portationNumbers[%d].status", index))
	}

	if portNum.Status.StatusDate != nil {
		if _, err := time.Parse(time.RFC3339, *portNum.Status.StatusDate); err != nil {
			return model.ErrInvalidDateFormat(fmt.Sprintf("portationNumbers[%d].status.statusDate", index))
		}
	}

	if portNum.TelcoAccount.MSISDN != nil && *portNum.TelcoAccount.MSISDN != "" {
		if !msisdnPattern.MatchString(*portNum.TelcoAccount.MSISDN) {
			return model.ErrInvalidPhoneNumberFormat
		}
	}

	return nil
}
