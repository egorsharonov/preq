package portin

import (
	"context"
	"fmt"
	"strconv"

	"gitlab.services.mts.ru/salsa/go-base/application/diagnostics"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"

	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/dal/entities"
	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/model"
)

func (s *Service) CreatePortInOrder(ctx context.Context, createOrder *model.CreatePortinOrder) (*string, error) {
	tracer := diagnostics.TracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "PortInService.CreatePortInOrder")
	defer span.End()

	log := s.getNamedLogger(ctx)

	if err := s.validator.ValidateAndPrepareCreatePortInOrder(ctx, createOrder); err != nil {
		log.Error("validateAndPrepareRequest.fail", zap.Error(err))
		return nil, err
	}

	if err := s.checkExistingOrders(ctx, createOrder); err != nil {
		log.Error("checkExistingOrders.fail", zap.Error(err))
		return nil, err
	}

	createdOrder, err := s.createOrderInDB(ctx, createOrder)
	if err != nil {
		log.Error("createOrderInDB.fail", zap.Error(err))
		return nil, err
	}

	err = s.mnpEventProducer.PublishCreatedEvent(ctx, createdOrder.StringID(), createdOrder.CreationDate)
	if err != nil {
		log.Error("publishCreatedEvent.fail", zap.Error(err), zap.Int64("orderId", createdOrder.ID))
		return nil, err
	}

	orderID := strconv.FormatInt(createdOrder.ID, 10)
	log.Info("successful create", zap.String("order_ref", orderID))

	return &orderID, nil
}

func (s *Service) checkExistingOrders(ctx context.Context, createOrder *model.CreatePortinOrder) error {
	if s.portInOrdersConfig.DisableOpenOrdersCheck {
		return nil
	}

	tracer := diagnostics.TracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "PortInService.CheckExistingOrders")
	defer span.End()

	span.SetAttributes(
		attribute.StringSlice("portationNumbers", createOrder.GetMSISDNs()),
	)

	msisdns := createOrder.GetMSISDNs()

	existingOrders, err := s.repo.FindOpenOrdersByMSISDN(ctx, msisdns, s.portInOrdersConfig.AllowedStatusesForNewOrder)
	if err != nil {
		return fmt.Errorf("find existing orders: %w", err)
	}

	if len(existingOrders) > 0 {
		return model.ErrPortationRequestExists
	}

	return nil
}

func (s *Service) createOrderInDB(
	ctx context.Context, createOrderModel *model.CreatePortinOrder) (*entities.PortInOrderEntity, error) {
	orderEntity, err := createOrderModel.ToEntity()
	if err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	createdOrder, err := s.repo.Create(ctx, orderEntity, createOrderModel.PortationNumbers)
	if err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	return createdOrder, nil
}
