package producers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gitlab.services.mts.ru/salsa/go-base/application/diagnostics"
	"gitlab.services.mts.ru/salsa/go-base/application/infrastructure/kafka"
	kafkaproducer "gitlab.services.mts.ru/salsa/go-base/application/infrastructure/kafka/producer"

	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"

	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/converters"
	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/orchestrator/kafka/mnpevent"
)

const (
	mnpPortinEventProducerLogName = "mnp-event-portin producer"
	mnpPortinEventProducerCfgName = "mnp-event-portin"
)

type IMnpEventPortInProducer interface {
	PublishEvent(
		ctx context.Context,
		messageKey string,
		message *mnpevent.PortIn) error
	PublishCreatedEvent(
		ctx context.Context,
		orderID string,
		creationDate time.Time) error
	PublishStatusChangedEvent(
		ctx context.Context,
		orderState *mnpevent.OrderStateDTO,
		patch *mnpevent.PortInPatch,
		orderID string) error
	PublishDueDateChangedEvent(
		ctx context.Context,
		patch *mnpevent.PortInPatch,
		orderID string) error
}

type MnpEventPortInProducer struct {
	producer kafka.Producer
}

func NewMnpEventPortInProducer(kafkaClient *kafka.Kafka) (*MnpEventPortInProducer, error) {
	producer, err := kafkaproducer.NewProducer(
		kafkaClient,
		kafkaproducer.WithNamedWriter(mnpPortinEventProducerCfgName),
	)
	if err != nil {
		return nil, fmt.Errorf("%s failed to create producer: %w", mnpPortinEventProducerLogName, err)
	}

	return &MnpEventPortInProducer{
		producer: producer,
	}, nil
}

func (p *MnpEventPortInProducer) PublishCreatedEvent(
	ctx context.Context,
	orderID string,
	creationDate time.Time,
) error {
	tracer := diagnostics.TracerFromContext(ctx)

	spanCtx, span := tracer.Start(ctx, "MnpEventPortInProducer.PublishCreatedEvent")
	defer span.End()

	orderID = converters.WithPinPrefix(orderID)

	eventID := mnpevent.NewEventID()

	event := &mnpevent.PortIn{
		ID:          eventID,
		EventType:   mnpevent.CreatedEventType,
		Date:        time.Now().Format(time.RFC3339),
		ProcessType: mnpevent.PortInProcessType,
		Source:      mnpevent.ServiceSource,
		Data: mnpevent.PortInData{
			OrderID: orderID,
			OrderState: &mnpevent.OrderStateDTO{
				Code:       mnpevent.CreatedEventType,
				Message:    converters.ToPtr("Заявление создано"),
				StatusDate: converters.StrToPtr(creationDate.Format(time.RFC3339)),
			},
		},
	}

	return p.PublishEvent(spanCtx, eventID, event)
}

func (p *MnpEventPortInProducer) PublishStatusChangedEvent(
	ctx context.Context,
	orderState *mnpevent.OrderStateDTO,
	patch *mnpevent.PortInPatch,
	orderID string,
) error {
	tracer := diagnostics.TracerFromContext(ctx)

	spanCtx, span := tracer.Start(ctx, "MnpEventPortInProducer.PublishStatusChangedEvent")
	defer span.End()

	orderID = converters.WithPinPrefix(orderID)

	eventID := mnpevent.NewEventID()

	event := &mnpevent.PortIn{
		ID:          eventID,
		EventType:   patch.EventType,
		Date:        patch.Date.Format(time.RFC3339),
		ProcessType: mnpevent.PortInProcessType,
		Source:      mnpevent.ServiceSource,
		Data: mnpevent.PortInData{
			OrderID:          orderID,
			OrderState:       orderState,
			PortationNumbers: patch.Data.PortationNumbers,
		},
	}

	// При отправке в случае обновления state у номеров без обновления state заявки, глоабльный orderState может не передаваться в patch
	if orderState != nil {
		event.EventType = orderState.Code
	}

	return p.PublishEvent(spanCtx, eventID, event)
}

func (p *MnpEventPortInProducer) PublishDueDateChangedEvent(
	ctx context.Context,
	patch *mnpevent.PortInPatch,
	orderID string,
) error {
	tracer := diagnostics.TracerFromContext(ctx)

	spanCtx, span := tracer.Start(ctx, "MnpEventPortInProducer.PublishDueDateChangedEvent")
	defer span.End()

	orderID = converters.WithPinPrefix(orderID)

	eventID := mnpevent.NewEventID()

	parsedDueDate, err := time.Parse(time.RFC3339, patch.Data.DueDate)
	if err != nil {
		return fmt.Errorf("failed to parse dueDate for event: %w", err)
	}

	event := &mnpevent.PortIn{
		ID:          eventID,
		EventType:   mnpevent.DueDateChangedEventType,
		Date:        patch.Date.Format(time.RFC3339),
		ProcessType: mnpevent.PortInProcessType,
		Source:      mnpevent.ServiceSource,
		Data: mnpevent.PortInData{
			OrderID: orderID,
			DueDate: &parsedDueDate,
		},
	}

	return p.PublishEvent(spanCtx, eventID, event)
}

func (p *MnpEventPortInProducer) PublishEvent(ctx context.Context, eventKey string, event *mnpevent.PortIn) error {
	if event == nil {
		return fmt.Errorf("%s: failed to publish event: given nil event", mnpPortinEventProducerLogName)
	}

	tracer := diagnostics.TracerFromContext(ctx)

	spanCxt, span := tracer.Start(ctx, "MnpEventPortInProducer.PublishEvent")
	defer span.End()

	span.SetAttributes(
		attribute.String("event.id", event.ID),
		attribute.String("event.type", event.EventType),
		attribute.String("portin.order.id", event.Data.OrderID),
	)

	log := diagnostics.LoggerFromContext(ctx).Named(mnpPortinEventProducerLogName)

	rawEvent, err := json.Marshal(event)
	if err != nil {
		log.Error(
			"failed to serialize mnp-event-portin message",
			zap.Error(err))

		return fmt.Errorf("%s: failed to serialize mnp-event-portin message: %w", mnpPortinEventProducerLogName, err)
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

		return fmt.Errorf("%s: failed to publish event: %w", mnpPortinEventProducerLogName, err)
	}

	return nil
}
