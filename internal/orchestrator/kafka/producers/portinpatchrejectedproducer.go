package producers

import (
	"context"
	"encoding/json"
	"fmt"

	"gitlab.services.mts.ru/salsa/go-base/application/diagnostics"
	"gitlab.services.mts.ru/salsa/go-base/application/infrastructure/kafka"
	kafkaproducer "gitlab.services.mts.ru/salsa/go-base/application/infrastructure/kafka/producer"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"

	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/orchestrator/kafka/mnpevent"
)

const (
	portInPatchRejectedProducerLogName = "portin-patch-rejected producer"
	portInPatchRejectedProducerCfgName = "portin-patch-rejected"
)

type IPortInPatchRejectedProducer interface {
	PublishEvent(
		ctx context.Context,
		eventKey string,
		event *mnpevent.PortInPatch) error
}

type PortInPatchRejectedProducer struct {
	producer kafka.Producer
}

func NewPortInPatchRejectedProducer(kafkaClient *kafka.Kafka) (*PortInPatchRejectedProducer, error) {
	producer, err := kafkaproducer.NewProducer(
		kafkaClient,
		kafkaproducer.WithNamedWriter(portInPatchRejectedProducerCfgName),
	)
	if err != nil {
		return nil, fmt.Errorf("%s failed to cratee producer: %w", portInPatchRejectedProducerLogName, err)
	}

	return &PortInPatchRejectedProducer{
		producer: producer,
	}, nil
}

func (p *PortInPatchRejectedProducer) PublishEvent(ctx context.Context, eventKey string, event *mnpevent.PortInPatch) error {
	if event == nil {
		return fmt.Errorf("%s: failed to publish event: given nil event", portInPatchRejectedProducerLogName)
	}

	tracer := diagnostics.TracerFromContext(ctx)

	spanCxt, span := tracer.Start(ctx, "PortInPatchRejectedProducer.PublishEvent")
	defer span.End()

	span.SetAttributes(
		attribute.String("event.id", event.ID),
		attribute.String("event.type", event.EventType),
		attribute.String("portin.order.id", event.Data.OrderID),
		attribute.String("cdb.process.id", event.Data.CDBProcessID),
	)

	log := diagnostics.LoggerFromContext(ctx).Named(portInPatchRejectedProducerLogName)

	rawEvent, err := json.Marshal(event)
	if err != nil {
		log.Error(
			"failed to serialize portin-patch message",
			zap.Error(err))

		return fmt.Errorf("%s: failed to serialize portin-patch message: %w", portInPatchRejectedProducerLogName, err)
	}

	kafkaMessage := &kafka.Message{
		Value: rawEvent,
		Key:   []byte(eventKey),
		Headers: []kafka.Header{
			{Key: "Content-Type", Value: []byte("application/json")},
			{Key: "MessageId", Value: []byte(event.ID)},
		},
	}

	if err := p.producer.SendMessage(spanCxt, kafkaMessage); err != nil {
		log.Error(
			"failed to publish event",
			zap.Error(err))

		return fmt.Errorf("%s: failed to publish event: %w", portInPatchRejectedProducerLogName, err)
	}

	return nil
}
