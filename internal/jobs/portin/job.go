package portin

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"time"

	"go.opentelemetry.io/otel"
	"go.uber.org/zap"

	"gitlab.services.mts.ru/salsa/mnp-hub/mnp-datamart/internal/target"
	"gitlab.services.mts.ru/salsa/mnp-hub/mnp-datamart/internal/transform"
)

var tracer = otel.Tracer("gitlab.services.mts.ru/salsa/mnp-datamart/internal/jobs/portin")

type Config struct {
	Lookback    time.Duration
	BatchSize   int
	Prefix      string
	CancelTable string
}

type Job struct {
	cfg      Config
	sourceDB *sql.DB
	cancelDB *sql.DB
	targetDB *sql.DB
	store    *target.Store
	logger   *zap.Logger
}

func NewJob(cfg Config, sourceDB, cancelDB, targetDB *sql.DB, store *target.Store, logger *zap.Logger) *Job {
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 5000
	}

	return &Job{cfg: cfg, sourceDB: sourceDB, cancelDB: cancelDB, targetDB: targetDB, store: store, logger: logger.Named("portin-job")}
}

func (j *Job) Run(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "Run")
	defer span.End()
	locked, err := j.store.TryLockJob(ctx, "portin-dag")
	if err != nil {
		return err
	}
	if !locked {
		j.logger.Info("job already running")
		return nil
	}
	defer j.store.UnlockJob(context.Background(), "portin-dag")

	depth, err := j.store.MaxFromDate(ctx)
	if err != nil {
		return err
	}
	if depth != nil {
		t := depth.Add(-j.cfg.Lookback)
		depth = &t
	}

	cancelMap, err := j.loadCancelStatuses(ctx, depth)
	if err != nil {
		return err
	}

	tx, err := j.targetDB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	if err := j.processOrders(ctx, tx, depth, cancelMap); err != nil {
		return err
	}
	if err := j.processOrderHistory(ctx, tx, depth, cancelMap); err != nil {
		return err
	}

	return tx.Commit()
}

type sourceOrder struct {
	OrderID      int64
	State        int
	CreationDate sql.NullTime
	DueDate      sql.NullTime
	ChangingDate time.Time
	CDBProcessID sql.NullString
	OrderType    string
	OrderData    []byte
}

func (j *Job) processOrders(ctx context.Context, tx *sql.Tx, depth *time.Time, cancelMap map[int64]bool) error {
	query := `SELECT order_id, state, creation_date, due_date, changing_date, cdb_process_id, order_type, order_data
FROM orders
WHERE ($1::timestamp is null or changing_date > $1)
ORDER BY changing_date, order_id
LIMIT $2`
	rows, err := j.sourceDB.QueryContext(ctx, query, depth, j.cfg.BatchSize)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var o sourceOrder
		if err := rows.Scan(&o.OrderID, &o.State, &o.CreationDate, &o.DueDate, &o.ChangingDate, &o.CDBProcessID, &o.OrderType, &o.OrderData); err != nil {
			return err
		}
		if o.OrderType != "portin" {
			continue
		}

		payload, err := transform.ParseOrderPayload(o.OrderData)
		if err != nil {
			return err
		}
		subscriberType := transform.SubscriberType(payload)
		if subscriberType != "Person" {
			continue
		}
		statusID := mapStatus(o.State)
		if cancelMap[o.OrderID] {
			statusID = 11
		}

		request := target.Request{
			OrderNumber:     fmt.Sprintf("%s%d", j.cfg.Prefix, o.OrderID),
			RequestStatusID: statusID,
			RequestDate:     nullTime(o.CreationDate),
			ContractDate:    transform.ParseContractDate(payload.Contract.DocumentDate),
			PortDate:        nullTime(o.DueDate),
			FromDate:        o.ChangingDate,
			CDBID:           o.CDBProcessID.String,
			ProcessType:     payload.ProcessType,
			PortType:        o.OrderType,
			SubscriberType:  subscriberType,
			MessageCode:     payload.Status.Code,
			RejectReason:    transform.ParseRejectReason(o.State, payload.Status.Message),
			OrderID:         o.OrderID,
		}
		if err := j.store.UpsertRequest(ctx, tx, request); err != nil {
			return err
		}

		for _, n := range payload.PortationNumbers {
			if n.MSISDN == "" {
				continue
			}
			err = j.store.UpsertReqNumber(ctx, tx, target.RequestNumber{
				ReqID:       request.OrderNumber,
				RecipientID: payload.Recipient.CDBCode,
				MSISDN:      n.MSISDN,
				RN:          n.RN,
			})
			if err != nil {
				return err
			}
		}
	}

	return rows.Err()
}

func (j *Job) processOrderHistory(ctx context.Context, tx *sql.Tx, depth *time.Time, cancelMap map[int64]bool) error {
	query := `SELECT l.order_id, l.state, l.creation_date, l.due_date, l.version_date, l.cdb_process_id, l.order_type, l.order_data_log,
(
	coalesce(
		(select min(l2.version_date) from orders_log l2 where l2.order_id = l.order_id and l2.version_date > l.version_date),
		o.changing_date
	) - interval '1 second'
) as to_date
FROM orders_log l
JOIN orders o ON o.order_id = l.order_id
WHERE ($1::timestamp is null or l.version_date > $1)
ORDER BY l.version_date, l.order_id
LIMIT $2`
	rows, err := j.sourceDB.QueryContext(ctx, query, depth, j.cfg.BatchSize)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var o sourceOrder
		var versionDate time.Time
		var toDate sql.NullTime
		if err := rows.Scan(&o.OrderID, &o.State, &o.CreationDate, &o.DueDate, &versionDate, &o.CDBProcessID, &o.OrderType, &o.OrderData, &toDate); err != nil {
			return err
		}
		if o.OrderType != "portin" {
			continue
		}
		payload, err := transform.ParseOrderPayload(o.OrderData)
		if err != nil {
			return err
		}
		subscriberType := transform.SubscriberType(payload)
		if subscriberType != "Person" {
			continue
		}
		statusID := mapStatus(o.State)
		if cancelMap[o.OrderID] {
			statusID = 11
		}
		request := target.Request{
			OrderNumber:     fmt.Sprintf("%s%d", j.cfg.Prefix, o.OrderID),
			RequestStatusID: statusID,
			RequestDate:     nullTime(o.CreationDate),
			ContractDate:    transform.ParseContractDate(payload.Contract.DocumentDate),
			PortDate:        nullTime(o.DueDate),
			FromDate:        versionDate,
			ToDate:          nullTime(toDate),
			CDBID:           o.CDBProcessID.String,
			ProcessType:     payload.ProcessType,
			PortType:        o.OrderType,
			SubscriberType:  subscriberType,
			MessageCode:     payload.Status.Code,
			RejectReason:    transform.ParseRejectReason(o.State, payload.Status.Message),
			OrderID:         o.OrderID,
		}
		if err := j.store.InsertRequestHistory(ctx, tx, request); err != nil {
			return err
		}
	}

	return rows.Err()
}

func (j *Job) loadCancelStatuses(ctx context.Context, depth *time.Time) (map[int64]bool, error) {
	table := j.cfg.CancelTable
	if table == "" {
		table = "orders"
	}
	if !regexp.MustCompile(`^[a-zA-Z0-9_\.\"]+$`).MatchString(table) {
		return nil, fmt.Errorf("unsafe cancel table name: %s", table)
	}

	query := fmt.Sprintf(`SELECT order_id, status FROM %s WHERE ($1::timestamp is null or changing_date > $1)`, table)
	rows, err := j.cancelDB.QueryContext(ctx, query, depth)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make(map[int64]bool)
	for rows.Next() {
		var orderID int64
		var status int
		if err := rows.Scan(&orderID, &status); err != nil {
			return nil, err
		}
		if status == 50 {
			res[orderID] = true
		}
	}

	return res, rows.Err()
}

func nullTime(ts sql.NullTime) *time.Time {
	if !ts.Valid {
		return nil
	}

	return &ts.Time
}

func mapStatus(v int) int {
	mapping := map[int]int{0: 1, 1: 2, -1: 3, 2: 4, -2: 12, 3: 4, -3: 3, 4: 4, -4: 5, 5: 6, 6: 6, 7: 7, 8: 7, 9: 7, 20: 8, 21: 8, 22: 9}
	if mapped, ok := mapping[v]; ok {
		return mapped
	}

	return v
}
