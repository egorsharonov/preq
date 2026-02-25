package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"
	"time"

	"gitlab.services.mts.ru/salsa/go-base/application/app"
	appConfig "gitlab.services.mts.ru/salsa/go-base/application/config"
	"gitlab.services.mts.ru/salsa/go-base/application/diagnostics"
	"gitlab.services.mts.ru/salsa/go-base/application/httphandler"
	"gitlab.services.mts.ru/salsa/go-base/application/httphandler/oapi"
	"gitlab.services.mts.ru/salsa/go-base/migration/migrate"

	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/cmd/dependencies"
	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/config"
	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/openapi/portin"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := appConfig.NewConfig[config.Config](ctx)
	if err != nil {
		log.Printf("%v", err)
		return
	}

	if migrate.HandleMigrationFlag(
		ctx,
		stop,
		cfg.PortInDB.GetMigrationConnectionString(),
		cfg.MigrationsPath,
		cfg.MigrationsVersionTable,
		10) {
		return
	}

	runApplication(ctx, stop)
}

func runApplication(ctx context.Context, stop context.CancelFunc) {
	a, err := app.NewApp[config.Config](ctx)
	if err != nil {
		panic(fmt.Errorf("create app: %w", err))
	}

	ctx = diagnostics.ContextWithLogger(ctx, a.Logger)
	ctx = diagnostics.ContextWithTracer(ctx, a.Tracer)

	enrichedCfg, err := config.EnrichConfig(ctx, &a.Config)
	if err != nil {
		panic(fmt.Errorf("failed to enrich configuration: %w", err))
	}

	a.Config = *enrichedCfg

	portInDB := dependencies.MustInitDB(ctx, &a.Config.PortInDB)
	defer portInDB.Close()

	mnpEventKafkaClient := dependencies.MustInitKafkaClient(&a.Config.MnpEventKafka)
	portInService := dependencies.MustInitPortInService(mnpEventKafkaClient, portInDB, &a.Config)
	messageProcessor := dependencies.MustInitMessageProcessor(ctx, mnpEventKafkaClient, portInService)
	server := portin.NewServer(portInService)

	httpServer := httphandler.CreateBuilder(nil).
		WithHealthCheck(
			httphandler.WithPerCheckTimeout(3*time.Second),
			httphandler.WithDB(portInDB),
			httphandler.WithKafka(mnpEventKafkaClient, "MNP Event Kafka"),
		).
		WithRecoveryMessage("panic occurred. Check logs for details", a.Logger).
		WithLoggingAndTracing(a.Logger.Named("http-server")).
		WithOpenAPI(
			server,
			oapi.WithSwagger(),
			oapi.WithValidation(portin.MapValidationErrToResponse, nil)).
		Build(":" + a.Config.HTTP.Port)

	a.AddStarter(mnpEventKafkaClient)
	a.AddStarter(messageProcessor)
	a.AddStarter(httpServer)

	go a.Start(ctx, stop)

	<-ctx.Done()

	a.Logger.Info("Shutdown complete")
}
