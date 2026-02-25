package portin

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"gitlab.services.mts.ru/salsa/go-base/application/diagnostics"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"

	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/converters"
	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/dal/containers"
	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/dal/entities"
	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/model"
	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/model/mappers"
)

func (s *Service) GetPortInOrderByID(ctx context.Context, orderID string) (*model.PortInOrder, error) {
	orderID = converters.WithoutPinPrefix(orderID)

	id, err := strconv.ParseInt(orderID, 10, 64)
	if err != nil {
		return nil, model.ErrInvalidParameterValue("id")
	}

	getBy := &containers.OrderGetByContainer{
		GetByType: containers.OrderID,
		Key:       id,
	}

	orderEntity, err := s.getPortInOrder(ctx, nil, getBy)
	if err != nil {
		return nil, err
	}

	return mappers.PortInEntityToModel(orderEntity)
}

func (s *Service) getPortInOrder(
	ctx context.Context, tx *sql.Tx, getBy *containers.OrderGetByContainer) (*entities.PortInOrderEntity, error) {
	tracer := diagnostics.TracerFromContext(ctx)

	spanCtx, span := tracer.Start(ctx, "PortInService.GetPortInOrder")
	defer span.End()

	log := s.getNamedLogger(spanCtx)

	span.SetAttributes(
		attribute.String("search.field", getBy.GetByType.String()),
		attribute.Int64("search.key", getBy.Key),
	)

	var (
		err         error
		orderEntity *entities.PortInOrderEntity
	)

	switch getBy.GetByType {
	case containers.OrderID:
		orderEntity, err = s.repo.GetByID(ctx, getBy.Key, getBy.ForUpdate, tx)
	case containers.CdbProcessID:
		orderEntity, err = s.repo.GetByCDBProcessID(ctx, getBy.Key, getBy.ForUpdate, tx)
	default:
		err = fmt.Errorf("invalid get by type: %d - %s", getBy.GetByType, getBy.GetByType)
	}

	if err != nil {
		log.Error(
			"failed to get PortIn order",
			zap.String("search.field", getBy.GetByType.String()),
			zap.Int64("search.key", getBy.Key),
			zap.Error(err),
		)

		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrOrderIDNotFound
		}

		return nil, fmt.Errorf("failed to get PortIn order by %s - %d : %w", getBy.GetByType.String(), getBy.Key, err)
	}

	return orderEntity, nil
}

func (s *Service) SearchPortInOrders(
	ctx context.Context,
	portNum, tempNum, cdbProcessID *string,
) ([]*model.PortInOrder, error) {
	tracer := diagnostics.TracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "PortInService.SearchPortInOrders")
	defer span.End()

	log := s.getNamedLogger(ctx)

	filters := make(map[string]interface{})

	// Параметры уже отвалидированы через openapi схему
	if portNum != nil {
		filters["msisdn"] = *portNum
	}

	if tempNum != nil {
		filters["tempnumber"] = *tempNum
	}

	if cdbProcessID != nil {
		filters["cdbProcessID"] = *cdbProcessID
	}

	ordersDB, totalCount, err := s.repo.Search(ctx, filters)
	if err != nil {
		log.Error("search orders failed", zap.Error(err))
		return nil, fmt.Errorf("search orders: %w", err)
	}

	log.Info("found orders", zap.Int("count", totalCount))

	if len(ordersDB) == 0 {
		return []*model.PortInOrder{}, model.ErrOrderIDNotFound
	}

	orders := make([]*model.PortInOrder, 0, totalCount)

	for _, orderDB := range ordersDB {
		order, err := mappers.PortInEntityToModel(orderDB)
		if err != nil {
			log.Error("Failed to convert order", zap.Error(err), zap.Int64("orderId", orderDB.ID))
			continue
		}

		orders = append(orders, order)
	}

	return orders, nil
}
