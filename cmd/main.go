package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"gitlab.services.mts.ru/salsa/go-base/application/app"
	appConfig "gitlab.services.mts.ru/salsa/go-base/application/config"
	"gitlab.services.mts.ru/salsa/go-base/application/diagnostics"
	"gitlab.services.mts.ru/salsa/go-base/application/httphandler"
	"gitlab.services.mts.ru/salsa/go-base/migration/migrate"
	"go.uber.org/zap"

	"gitlab.services.mts.ru/salsa/mnp-hub/mnp-datamart/cmd/dependencies"
	"gitlab.services.mts.ru/salsa/mnp-hub/mnp-datamart/config"
	"gitlab.services.mts.ru/salsa/mnp-hub/mnp-datamart/internal/jobs/cdbmessage"
	"gitlab.services.mts.ru/salsa/mnp-hub/mnp-datamart/internal/jobs/portin"
	"gitlab.services.mts.ru/salsa/mnp-hub/mnp-datamart/internal/target"
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
		cfg.TargetDB.GetMigrationConnectionString(),
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

	targetDB := dependencies.MustInitDB(ctx, &a.Config.TargetDB)
	defer targetDB.Close()

	portInDB := dependencies.MustInitDB(ctx, &a.Config.PortInOrdersDB)
	defer portInDB.Close()

	cancelDB := dependencies.MustInitDB(ctx, &a.Config.PortInCancelDB)
	defer cancelDB.Close()

	cdbDB := dependencies.MustInitDB(ctx, &a.Config.CDBMessagingDB)
	defer cdbDB.Close()

	store := target.NewStore(targetDB)
	portInJob := portin.NewJob(portin.Config{
		Lookback:    a.Config.LookbackDuration,
		BatchSize:   a.Config.BatchSize,
		Prefix:      a.Config.PortInPrefix,
		CancelTable: a.Config.PortInCancelTable,
	}, portInDB, cancelDB, targetDB, store, a.Logger)
	cdbJob := cdbmessage.NewJob(cdbmessage.Config{
		Lookback:  a.Config.LookbackDuration,
		BatchSize: a.Config.BatchSize,
		Prefix:    a.Config.PortInPrefix,
	}, cdbDB, targetDB, store, a.Logger)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /jobs/portin/run", runJobHandler(a.Logger.Named("http.portin-run"), portInJob.Run))
	mux.HandleFunc("POST /jobs/cdb-message/run", runJobHandler(a.Logger.Named("http.cdb-message-run"), cdbJob.Run))

	httpServer := httphandler.CreateBuilder(mux).
		WithHealthCheck(
			httphandler.WithPerCheckTimeout(3*time.Second),
			httphandler.WithDB(targetDB),
			httphandler.WithDB(portInDB),
			httphandler.WithDB(cancelDB),
			httphandler.WithDB(cdbDB),
		).
		WithRecoveryMessage("panic occurred. Check logs for details", a.Logger).
		WithLoggingAndTracing(a.Logger.Named("http-server")).
		Build(":" + a.Config.HTTP.Port)

	go runTicker(ctx, a.Config.PortInJobInterval, "portin", a.Logger.Named("scheduler.portin"), portInJob.Run)
	go runTicker(ctx, a.Config.CDBMessageJobInterval, "cdb-message", a.Logger.Named("scheduler.cdb-message"), cdbJob.Run)

	a.AddStarter(httpServer)

	go a.Start(ctx, stop)

	<-ctx.Done()

	a.Logger.Info("Shutdown complete")
}

func runTicker(ctx context.Context, interval time.Duration, name string, logger *zap.Logger, run func(context.Context) error) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			runCtx, cancel := context.WithTimeout(ctx, interval)
			err := run(runCtx)
			cancel()
			if err != nil {
				logger.Error("job execution failed", zap.String("job", name), zap.Error(err))
			}
		}
	}
}

func runJobHandler(logger *zap.Logger, run func(context.Context) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := run(r.Context()); err != nil {
			logger.Error("job run failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		w.WriteHeader(http.StatusAccepted)
	}
}
