package dependencies

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"time"

	appcfg "gitlab.services.mts.ru/salsa/go-base/application/config"
	"gitlab.services.mts.ru/salsa/go-base/application/diagnostics"
	"gitlab.services.mts.ru/salsa/go-base/application/infrastructure/kafka"
	"go.uber.org/zap"

	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/config"
	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/dal/repository"
	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/orchestrator"
	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/orchestrator/kafka/consumers"
	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/orchestrator/kafka/producers"
	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/service"
	portinservice "gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/service/portin"
)

func MustInitMessageProcessor(
	ctx context.Context,
	kafkaClient *kafka.Kafka,
	portInService portinservice.IPortInService,
) *orchestrator.MessageProcessor {
	portInPatchConsumer := mustInitPortInPatchConsumer(ctx, kafkaClient)
	portInPatchRejectedProducer := mustInitPortinPatchRejectedProducer(kafkaClient)

	processor := orchestrator.NewMessageProcessor(
		portInPatchConsumer,
		portInPatchRejectedProducer,
		portInService,
	)

	return processor
}

func MustInitPortInService(
	kafkaClient *kafka.Kafka,
	db *sql.DB,
	cfg *config.Config) *portinservice.Service {
	portInRepo := mustInitPortInOrderRepository(db)

	mnpEventProducer := mustInitMnpEventPortInProducer(kafkaClient)

	ftpSerivce := mustInitFTPService(&cfg.FTP)
	validationService := service.NewValidationService(ftpSerivce, &cfg.MTS)

	portInServce := portinservice.NewPortInService(
		portInRepo,
		validationService,
		mnpEventProducer,
		&cfg.PortInOrders,
	)

	return portInServce
}

func MustInitKafkaClient(cfg *appcfg.KafkaConfig) *kafka.Kafka {
	kafkaClient, err := kafka.InitKafka(cfg)
	if err != nil {
		panic(fmt.Errorf("failed to init kafka client: %w", err))
	}

	return kafkaClient
}

func MustInitDB(ctx context.Context, cfg *config.PostgresConfig) *sql.DB {
	db, err := sql.Open("postgres", cfg.GetAppConnectionString())
	if err != nil {
		panic(fmt.Errorf("failed to open database: %w", err))
	}

	if err := pingWithRetry(ctx, db, cfg.MaxConnectionRetries); err != nil {
		defer db.Close()

		panic(fmt.Errorf("failed to ping database: %w", err))
	}

	return db
}

// Services.
func mustInitFTPService(cfg *appcfg.FTPConfig) *service.FTPService {
	ftpSerivce, err := service.NewFTPService(cfg)
	if err != nil {
		panic(fmt.Errorf("failed to init FTP service: %w", err))
	}

	return ftpSerivce
}

// Kafka.
func mustInitPortInPatchConsumer(ctx context.Context, kafkaClient *kafka.Kafka) *consumers.PortInPatchConsumer {
	consumer, err := consumers.NewPortInPatchConsumer(ctx, kafkaClient)
	if err != nil {
		panic(fmt.Errorf("failed to init portin-patch consumer: %w", err))
	}

	return consumer
}

func mustInitMnpEventPortInProducer(kafkaClient *kafka.Kafka) *producers.MnpEventPortInProducer {
	producer, err := producers.NewMnpEventPortInProducer(kafkaClient)
	if err != nil {
		panic(fmt.Errorf("failed to init mnp-event-portin producer: %w", err))
	}

	return producer
}

func mustInitPortinPatchRejectedProducer(kafkaClient *kafka.Kafka) *producers.PortInPatchRejectedProducer {
	producer, err := producers.NewPortInPatchRejectedProducer(kafkaClient)
	if err != nil {
		panic(fmt.Errorf("failed to init portin-patch-rejected producer: %w", err))
	}

	return producer
}

// DB.
func mustInitPortInOrderRepository(db *sql.DB) *repository.PortInOrderRepository {
	return repository.NewPortInOrderRepository(db)
}

func pingWithRetry(ctx context.Context, db *sql.DB, maxRetries int) error {
	var err error

	log := diagnostics.LoggerFromContext(ctx)

	for retry := range maxRetries {
		err = db.PingContext(ctx)
		if err == nil {
			return nil
		}

		if retry < maxRetries {
			backoffTime := time.Duration(math.Pow(2, float64(retry))*100) * time.Millisecond

			log.Warn("failed to ping database. Retrying...",
				zap.Int("ping.attempt", retry+1),
				zap.Int("ping.max_retries", maxRetries),
				zap.Duration("ping.retry_in", backoffTime),
				zap.Error(err))

			time.Sleep(backoffTime)
		}
	}

	return err
}
