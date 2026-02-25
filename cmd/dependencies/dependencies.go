package dependencies

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"time"

	"gitlab.services.mts.ru/salsa/go-base/application/diagnostics"
	"go.uber.org/zap"

	"gitlab.services.mts.ru/salsa/mnp-hub/mnp-datamart/config"
)

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
