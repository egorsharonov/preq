package consumers

import (
	"context"
	"fmt"

	"gitlab.services.mts.ru/salsa/go-base/application/diagnostics"
	"gitlab.services.mts.ru/salsa/go-base/application/infrastructure/kafka"
	kafkaconsumer "gitlab.services.mts.ru/salsa/go-base/application/infrastructure/kafka/consumer"

	"go.uber.org/zap"

	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/orchestrator/kafka/mnpevent"
)

const (
	consumerLogName = "portin-patch consumer"
	consumerCfgName = "portin-patch"
)

type IPortInPathConsumer interface {
	Start(ctx context.Context, handler kafka.MessageHandler[mnpevent.PortInPatch])
	Stop() error
}

type PortInPatchConsumer struct {
	consumer kafka.Consumer[mnpevent.PortInPatch]
}

func NewPortInPatchConsumer(ctx context.Context, kafkaClient *kafka.Kafka) (*PortInPatchConsumer, error) {
	log := diagnostics.LoggerFromContext(ctx)

	consumer, err := kafkaconsumer.NewConsumer(
		kafkaClient,
		kafkaconsumer.WithNamedConfig[mnpevent.PortInPatch](consumerCfgName),
		kafkaconsumer.WithErrorLogger[mnpevent.PortInPatch](log),
	)
	if err != nil {
		return nil, fmt.Errorf("%s failed to init consumer: %w", consumerLogName, err)
	}

	return &PortInPatchConsumer{
		consumer: consumer,
	}, nil
}

func (c *PortInPatchConsumer) Start(ctx context.Context, handler kafka.MessageHandler[mnpevent.PortInPatch]) {
	log := diagnostics.LoggerFromContext(ctx).Named(consumerLogName)
	log.Info(
		"consumer started",
		zap.String("kafka.topic", consumerCfgName))

	c.consumer.Start(ctx, handler)
}

func (c *PortInPatchConsumer) Stop() error {
	return c.consumer.Close()
}
