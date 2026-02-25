package orchestrator

import (
	"context"

	"go.uber.org/zap"

	"gitlab.services.mts.ru/salsa/go-base/application/diagnostics"
	"gitlab.services.mts.ru/salsa/go-base/application/infrastructure/kafka"

	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/orchestrator/kafka/consumers"
	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/orchestrator/kafka/mnpevent"
	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/orchestrator/kafka/producers"
	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/service/portin"
)

const (
	processorLogName = "message-processor"
)

type MessageProcessor struct {
	portInPatchConsumer consumers.IPortInPathConsumer
	rejectedProducer    producers.IPortInPatchRejectedProducer
	portInService       portin.IPortInService
}

func NewMessageProcessor(
	portInPatchConsumer consumers.IPortInPathConsumer,
	rejectedProducer producers.IPortInPatchRejectedProducer,
	portInService portin.IPortInService,
) *MessageProcessor {
	return &MessageProcessor{
		portInPatchConsumer: portInPatchConsumer,
		rejectedProducer:    rejectedProducer,
		portInService:       portInService,
	}
}

func (mp *MessageProcessor) handlePortInPatchEvent(ctx context.Context, msg *kafka.MessageResult[mnpevent.PortInPatch]) error {
	tracer := diagnostics.TracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "MessageProcessor.HandlePortinPatch")
	defer span.End()

	log := mp.getNamedLogger(ctx)

	if err := mp.portInService.UpdatePortInOrder(ctx, msg.Body); err != nil {
		if pubErr := mp.rejectedProducer.PublishEvent(ctx, msg.Body.ID, msg.Body); pubErr != nil {
			log.Error(
				"failed to publish patch rejected",
				zap.Error(pubErr))

			return pubErr
		}

		log.Warn("portin-patch published to rejected topic", zap.Error(err))

		return nil
	}

	return nil
}

func (mp *MessageProcessor) Start(ctx context.Context) error {
	log := mp.getNamedLogger(ctx)

	log.Info("start processing")
	mp.portInPatchConsumer.Start(ctx, mp.handlePortInPatchEvent)
	<-ctx.Done()

	_ = mp.portInPatchConsumer.Stop()

	log.Info("stopped processing")

	return nil
}

func (mp *MessageProcessor) getNamedLogger(ctx context.Context) *zap.Logger {
	return diagnostics.LoggerFromContext(ctx).Named(processorLogName)
}
