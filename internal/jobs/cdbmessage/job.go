package cdbmessage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.uber.org/zap"

	"gitlab.services.mts.ru/salsa/mnp-hub/mnp-datamart/internal/target"
)

var tracer = otel.Tracer("gitlab.services.mts.ru/salsa/mnp-datamart/internal/jobs/cdbmessage")

type Config struct {
	Lookback  time.Duration
	BatchSize int
	Prefix    string
}

type Job struct {
	cfg      Config
	sourceDB *sql.DB
	targetDB *sql.DB
	store    *target.Store
	logger   *zap.Logger
}

func NewJob(cfg Config, sourceDB, targetDB *sql.DB, store *target.Store, logger *zap.Logger) *Job {
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 5000
	}

	return &Job{cfg: cfg, sourceDB: sourceDB, targetDB: targetDB, store: store, logger: logger.Named("cdb-message-job")}
}

func (j *Job) Run(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "Run")
	defer span.End()
	locked, err := j.store.TryLockJob(ctx, "cdb-message-dag")
	if err != nil {
		return err
	}
	if !locked {
		j.logger.Info("job already running")
		return nil
	}
	defer j.store.UnlockJob(context.Background(), "cdb-message-dag")

	depth, err := j.store.MaxRawRequestTime(ctx)
	if err != nil {
		return err
	}
	if depth != nil {
		t := depth.Add(1 * time.Second).Add(-j.cfg.Lookback)
		depth = &t
	}

	rows, err := j.sourceDB.QueryContext(ctx, `
SELECT m.message_id, p.order_id, m.request_data, m.message_data, m.message_type, m.message_direction
FROM mnp_message m
JOIN mnp_process p ON p.process_id = m.process_id
WHERE ($1::timestamp is null or m.message_date > $1)
ORDER BY m.message_date, m.message_id
LIMIT $2`, depth, j.cfg.BatchSize)
	if err != nil {
		return err
	}
	defer rows.Close()

	tx, err := j.targetDB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	for rows.Next() {
		var (
			id          int64
			orderID     string
			requestTime time.Time
			messageData sql.NullString
			messageType sql.NullString
			direction   int
		)
		if err := rows.Scan(&id, &orderID, &requestTime, &messageData, &messageType, &direction); err != nil {
			return err
		}

		source := "MNPHUB"
		dest := "CDB"
		if direction == 1 {
			source = "CDB"
			dest = "MNPHUB"
		}

		reqID := orderID
		if j.cfg.Prefix != "" && len(orderID) >= len(j.cfg.Prefix) && orderID[:len(j.cfg.Prefix)] != j.cfg.Prefix {
			reqID = fmt.Sprintf("%s%s", j.cfg.Prefix, orderID)
		}

		if err := j.store.UpsertRawRequest(ctx, tx, target.RawRequest{
			ID:            id,
			ReqID:         reqID,
			RequestTime:   requestTime,
			XMLMessage:    messageData.String,
			OperationInfo: messageType.String,
			SystemSource:  source,
			SystemDest:    dest,
		}); err != nil {
			return err
		}
	}

	if err := rows.Err(); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
