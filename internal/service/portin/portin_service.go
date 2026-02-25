package portin

import (
	"context"

	"gitlab.services.mts.ru/salsa/go-base/application/diagnostics"
	"go.uber.org/zap"

	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/config"
	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/dal/repository"
	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/model"
	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/orchestrator/kafka/mnpevent"
	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/orchestrator/kafka/producers"
	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/service"
)

type IPortInService interface {
	CreatePortInOrder(ctx context.Context, createOrder *model.CreatePortinOrder) (*string, error)
	GetPortInOrderByID(ctx context.Context, orderID string) (*model.PortInOrder, error)
	SearchPortInOrders(ctx context.Context, portNum, tempNum, cdbProcessID *string) ([]*model.PortInOrder, error)
	UpdatePortInOrder(ctx context.Context, patch *mnpevent.PortInPatch) error
}

const (
	serviceLogName = "port-in-service"
)

type Service struct {
	repo               repository.IPortInOrderRepository
	validator          service.IValidationService
	mnpEventProducer   producers.IMnpEventPortInProducer
	portInOrdersConfig *config.PortInOrdersConfig
}

func NewPortInService(
	repo repository.IPortInOrderRepository,
	validator service.IValidationService,
	mnpEventProducer producers.IMnpEventPortInProducer,
	portInOrdersConfig *config.PortInOrdersConfig,
) *Service {
	return &Service{
		repo:               repo,
		validator:          validator,
		mnpEventProducer:   mnpEventProducer,
		portInOrdersConfig: portInOrdersConfig,
	}
}

func (s *Service) getNamedLogger(ctx context.Context) *zap.Logger {
	return diagnostics.LoggerFromContext(ctx).Named(serviceLogName)
}
