package target

import (
	"context"
	"database/sql"
	"time"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store { return &Store{db: db} }

type Request struct {
	OrderNumber     string
	RequestStatusID int
	RequestDate     *time.Time
	ContractDate    *time.Time
	PortDate        *time.Time
	FromDate        time.Time
	ToDate          *time.Time
	CDBID           string
	ProcessType     string
	PortType        string
	SubscriberType  string
	MessageCode     string
	RejectReason    *int
	OrderID         int64
}

type RequestNumber struct {
	ReqID       string
	RecipientID string
	MSISDN      string
	RN          string
}

type RawRequest struct {
	ID            int64
	ReqID         string
	RequestTime   time.Time
	XMLMessage    string
	OperationInfo string
	SystemSource  string
	SystemDest    string
}

func (s *Store) TryLockJob(ctx context.Context, name string) (bool, error) {
	var ok bool
	err := s.db.QueryRowContext(ctx, `SELECT pg_try_advisory_lock(hashtext($1))`, "mnp-datamart:"+name).Scan(&ok)

	return ok, err
}

func (s *Store) UnlockJob(ctx context.Context, name string) {
	_, _ = s.db.ExecContext(ctx, `SELECT pg_advisory_unlock(hashtext($1))`, "mnp-datamart:"+name)
}

func (s *Store) MaxFromDate(ctx context.Context) (*time.Time, error) {
	var ts sql.NullTime
	if err := s.db.QueryRowContext(ctx, `SELECT max(from_date) FROM mnp_request`).Scan(&ts); err != nil {
		return nil, err
	}
	if !ts.Valid {
		return nil, nil
	}

	return &ts.Time, nil
}

func (s *Store) MaxRawRequestTime(ctx context.Context) (*time.Time, error) {
	var ts sql.NullTime
	if err := s.db.QueryRowContext(ctx, `SELECT max(request_time) FROM mnp_raw_request`).Scan(&ts); err != nil {
		return nil, err
	}
	if !ts.Valid {
		return nil, nil
	}

	return &ts.Time, nil
}

func (s *Store) UpsertRequest(ctx context.Context, tx *sql.Tx, r Request) error {
	_, err := tx.ExecContext(ctx, `
INSERT INTO mnp_request (
  order_number, request_status_id, request_date, contract_date, port_date, from_date, to_date,
  change_date, deleted, cdb_id, process_type, port_type, subscriber_type, message_code, reject_reason, order_id
) VALUES ($1,$2,$3,$4,$5,$6,$7,now(),0,$8,$9,$10,$11,$12,$13,$14)
ON CONFLICT (order_number)
DO UPDATE SET
  request_status_id = EXCLUDED.request_status_id,
  request_date = EXCLUDED.request_date,
  contract_date = EXCLUDED.contract_date,
  port_date = EXCLUDED.port_date,
  from_date = EXCLUDED.from_date,
  to_date = EXCLUDED.to_date,
  change_date = now(),
  deleted = 0,
  cdb_id = EXCLUDED.cdb_id,
  process_type = EXCLUDED.process_type,
  port_type = EXCLUDED.port_type,
  subscriber_type = EXCLUDED.subscriber_type,
  message_code = EXCLUDED.message_code,
  reject_reason = EXCLUDED.reject_reason,
  order_id = EXCLUDED.order_id
`, r.OrderNumber, r.RequestStatusID, r.RequestDate, r.ContractDate, r.PortDate, r.FromDate, r.ToDate, r.CDBID,
		r.ProcessType, r.PortType, r.SubscriberType, r.MessageCode, r.RejectReason, r.OrderID)

	return err
}

func (s *Store) InsertRequestHistory(ctx context.Context, tx *sql.Tx, r Request) error {
	_, err := tx.ExecContext(ctx, `
INSERT INTO mnp_request_h (
  order_number, request_status_id, request_date, contract_date, port_date, from_date, to_date,
  change_date, deleted, cdb_id, process_type, port_type, subscriber_type, message_code, reject_reason, order_id
) VALUES ($1,$2,$3,$4,$5,$6,$7,now(),0,$8,$9,$10,$11,$12,$13,$14)
ON CONFLICT (order_id, from_date)
DO UPDATE SET
  request_status_id = EXCLUDED.request_status_id,
  request_date = EXCLUDED.request_date,
  contract_date = EXCLUDED.contract_date,
  port_date = EXCLUDED.port_date,
  to_date = EXCLUDED.to_date,
  change_date = now(),
  cdb_id = EXCLUDED.cdb_id,
  process_type = EXCLUDED.process_type,
  port_type = EXCLUDED.port_type,
  subscriber_type = EXCLUDED.subscriber_type,
  message_code = EXCLUDED.message_code,
  reject_reason = EXCLUDED.reject_reason
`, r.OrderNumber, r.RequestStatusID, r.RequestDate, r.ContractDate, r.PortDate, r.FromDate, r.ToDate, r.CDBID,
		r.ProcessType, r.PortType, r.SubscriberType, r.MessageCode, r.RejectReason, r.OrderID)

	return err
}

func (s *Store) UpsertReqNumber(ctx context.Context, tx *sql.Tx, n RequestNumber) error {
	_, err := tx.ExecContext(ctx, `
INSERT INTO req_number(req_id, recipient_id, msisdn, rn, change_date)
VALUES ($1,$2,$3,$4,now())
ON CONFLICT (req_id, msisdn)
DO UPDATE SET recipient_id = EXCLUDED.recipient_id, rn = EXCLUDED.rn, change_date = now()
`, n.ReqID, nullIfEmpty(n.RecipientID), n.MSISDN, nullIfEmpty(n.RN))

	return err
}

func (s *Store) UpsertRawRequest(ctx context.Context, tx *sql.Tx, rr RawRequest) error {
	_, err := tx.ExecContext(ctx, `
INSERT INTO mnp_raw_request(id, req_id, request_time, xml_message, operation_info, system_source, system_dest, change_date)
VALUES ($1,$2,$3,$4,$5,$6,$7,now())
ON CONFLICT (id)
DO UPDATE SET
  req_id=EXCLUDED.req_id,
  request_time=EXCLUDED.request_time,
  xml_message=EXCLUDED.xml_message,
  operation_info=EXCLUDED.operation_info,
  system_source=EXCLUDED.system_source,
  system_dest=EXCLUDED.system_dest,
  change_date=now()
`, rr.ID, rr.ReqID, rr.RequestTime, rr.XMLMessage, rr.OperationInfo, rr.SystemSource, rr.SystemDest)

	return err
}

func nullIfEmpty(v string) any {
	if v == "" {
		return nil
	}

	return v
}
