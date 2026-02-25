package service

import (
	"context"
	"strconv"
	"strings"

	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/config"
	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/converters"
	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/dal/containers"
	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/dal/entities"
	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/model"
	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/orchestrator/kafka/mnpevent"
	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/validators"
)

type IValidationService interface {
	ValidateAndPrepareCreatePortInOrder(ctx context.Context, createOrder *model.CreatePortinOrder) *model.APIError
	ValidatePortInPatchEvent(ctx context.Context, patch *mnpevent.PortInPatch) (*containers.OrderGetByContainer, error)
	ValidatePortInPatchDate(event *mnpevent.PortInPatch, order *entities.PortInOrderEntity) *model.APIError
}

type ValidationService struct {
	ftpService IFTPService
	mtsConfig  *config.MTSConfig
}

func NewValidationService(
	ftpService IFTPService,
	mtsConfig *config.MTSConfig) *ValidationService {
	return &ValidationService{
		ftpService: ftpService,
		mtsConfig:  mtsConfig,
	}
}

func (s *ValidationService) ValidateAndPrepareCreatePortInOrder(
	ctx context.Context, createOrder *model.CreatePortinOrder) *model.APIError {
	validator := validators.NewCreatePortInOrderValidator(createOrder)
	if err := validator.Validate(ctx); err != nil {
		return err
	}

	mtsOperator := &model.Operator{
		Rn:      s.mtsConfig.RecipientRN,
		Name:    &s.mtsConfig.RecipientName,
		CdbCode: &s.mtsConfig.RecipientCdbCode,
	}
	createOrder.WithDefaults(mtsOperator)

	if err := s.ftpService.CheckDocumentAccess(ctx, createOrder.Contract.DocumentURL); err != nil {
		return err
	}

	return nil
}

func (s *ValidationService) ValidatePortInPatchEvent(
	ctx context.Context,
	patch *mnpevent.PortInPatch) (*containers.OrderGetByContainer, error) {
	validator := validators.NewPortInPatchEventValidator(patch)
	if err := validator.Validate(ctx); err != nil {
		return nil, err
	}

	orderIDStr := converters.WithoutPinPrefix(patch.Data.OrderID)
	cdbIDStr := strings.TrimSpace(patch.Data.CDBProcessID)

	switch {
	case orderIDStr != "":
		ordID, err := strconv.ParseInt(orderIDStr, 10, 64)
		if err != nil {
			return nil, model.ErrInvalidParameterValue("orderId")
		}

		return &containers.OrderGetByContainer{
			Key:       ordID,
			GetByType: containers.OrderID,
		}, nil

	case cdbIDStr != "":
		cdbPID, err := strconv.ParseInt(cdbIDStr, 10, 64)
		if err != nil {
			return nil, model.ErrInvalidParameterValue("cdbProcessId")
		}

		return &containers.OrderGetByContainer{
			Key:       cdbPID,
			GetByType: containers.CdbProcessID,
		}, nil

	default:
		return nil, model.ErrMandatoryParameterMissing("orderId|cdbProcessId")
	}
}

func (s *ValidationService) ValidatePortInPatchDate(event *mnpevent.PortInPatch, order *entities.PortInOrderEntity) *model.APIError {
	lastChange := order.CreationDate
	if order.ChangingDate != nil {
		lastChange = *order.ChangingDate
	}

	if !event.Date.After(lastChange) {
		return model.ErrInvalidParameterValue("date (not newer)")
	}

	return nil
}
