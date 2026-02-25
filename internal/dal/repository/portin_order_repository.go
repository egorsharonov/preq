package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"gitlab.services.mts.ru/salsa/go-base/application/diagnostics"
	"go.opentelemetry.io/otel/attribute"

	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/dal/containers"
	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/dal/entities"
	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/model"
)

type QueriableConnection interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryRow(query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	Query(query string, args ...any) (*sql.Rows, error)
}

type IPortInOrderRepository interface {
	Create(
		ctx context.Context, order *entities.PortInOrderEntity, portationNumbers []model.PortationNumber) (*entities.PortInOrderEntity, error)

	FindOpenOrdersByMSISDN(ctx context.Context, msisdns []string, allowedStateCodes []int) ([]*entities.PortInOrderEntity, error)
	Search(ctx context.Context, filters map[string]any) ([]*entities.PortInOrderEntity, int, error)
	GetByID(ctx context.Context, orderID int64, forUpdate bool, tx *sql.Tx) (*entities.PortInOrderEntity, error)
	GetByCDBProcessID(ctx context.Context, cdbProcessID int64, forUpdate bool, tx *sql.Tx) (*entities.PortInOrderEntity, error)

	UpdateOrderPatch(ctx context.Context, tx *sql.Tx, applyPatch *containers.OrderPatchUpdate) error
	CreateAndOpenTransaction(ctx context.Context, opts *sql.TxOptions) (tx *sql.Tx, rollback func(), err error)
}

type PortInOrderRepository struct {
	db *sql.DB
}

func NewPortInOrderRepository(db *sql.DB) *PortInOrderRepository {
	return &PortInOrderRepository{
		db: db,
	}
}

func (r *PortInOrderRepository) CreateAndOpenTransaction(
	ctx context.Context, opts *sql.TxOptions) (tx *sql.Tx, rollback func(), err error) {
	tx, err = r.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open transaction: %w", err)
	}

	rollback = func() { _ = tx.Rollback() }

	return tx, rollback, nil
}

func (r *PortInOrderRepository) Create(
	ctx context.Context,
	order *entities.PortInOrderEntity,
	portationNumbers []model.PortationNumber,
) (*entities.PortInOrderEntity, error) {
	tx, rollback, err := r.CreateAndOpenTransaction(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return nil, err
	}

	defer rollback()

	const orderQuery = `
INSERT INTO orders (due_date, state, order_type, customer_id, contact_phone, order_data, changed_by_user, process_type)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	RETURNING order_id, creation_date, changing_date`

	var createdOrder entities.PortInOrderEntity

	err = tx.QueryRowContext(ctx, orderQuery,
		order.DueDate,
		order.State,
		order.OrderType,
		order.CustomerID,
		order.ContactPhone,
		order.OrderData,
		order.ChangedByUser,
		order.ProcessType,
	).Scan(&createdOrder.ID, &createdOrder.CreationDate, &createdOrder.ChangingDate)
	if err != nil {
		return nil, fmt.Errorf("insert order: %w", err)
	}

	createdOrder.DueDate = order.DueDate
	createdOrder.State = order.State
	createdOrder.OrderType = order.OrderType
	createdOrder.CustomerID = order.CustomerID
	createdOrder.ContactPhone = order.ContactPhone
	createdOrder.OrderData = order.OrderData
	createdOrder.ChangedByUser = order.ChangedByUser
	createdOrder.ProcessType = order.ProcessType

	if err := r.insertOrderPortationNumbers(ctx, tx, createdOrder.ID, portationNumbers); err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return &createdOrder, nil
}

func (r *PortInOrderRepository) insertOrderPortationNumbers(
	ctx context.Context,
	tx *sql.Tx,
	orderID int64,
	portationNumbers []model.PortationNumber) error {
	const portNumsQueryBase = `
INSERT INTO portation_numbers (order_id, msisdn, telco_id, temp_msisdn) 
	VALUES %s`

	portNumsPlaceholders := make([]string, 0, len(portationNumbers))
	portNumsArgs := make([]any, 0, len(portationNumbers)*4)

	i := 0

	for _, portNum := range portationNumbers {
		msisdnValue, err := parseInt64String(portNum.Msisdn, "msisdn")
		if err != nil {
			return fmt.Errorf("invalid portationNumbers[%d].msisdn: %w", i, err)
		}

		portNumsPlaceholders = append(portNumsPlaceholders, fmt.Sprintf("($%d, $%d, $%d, $%d)", i*4+1, i*4+2, i*4+3, i*4+4))
		portNumsArgs = append(portNumsArgs,
			orderID,
			msisdnValue,
			portNum.TelcoAccount.ID,
			portNum.TelcoAccount.Msisdn)
		i++
	}

	portNumsQuery := fmt.Sprintf(portNumsQueryBase, strings.Join(portNumsPlaceholders, ","))

	res, err := tx.ExecContext(ctx, portNumsQuery, portNumsArgs...)
	if err != nil {
		return fmt.Errorf("failed to save order portation numbres: %w", err)
	}

	rowsAff, err := res.RowsAffected()
	if err == nil {
		if int(rowsAff) != len(portationNumbers) {
			return fmt.Errorf("failed to save all %d order portation numbres: saved only %d - rollback", len(portationNumbers), rowsAff)
		}
	}

	return nil
}

func (r *PortInOrderRepository) FindOpenOrdersByMSISDN(
	ctx context.Context, msisdns []string, allowedStateCodes []int) ([]*entities.PortInOrderEntity, error) {
	if len(msisdns) == 0 {
		return []*entities.PortInOrderEntity{}, nil
	}

	placeholders, args, err := buildMSISDNFilter(msisdns)
	if err != nil {
		return nil, err
	}

	notInClause, stateArgs := buildStateFilterClause(allowedStateCodes, len(placeholders)+1)
	args = append(args, stateArgs...)

	query := buildOpenOrdersQuery(placeholders, notInClause)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query open orders: %w", err)
	}
	defer rows.Close()

	orders, err := r.scanOrderRows(rows)
	if err != nil {
		return nil, err
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return orders, nil
}

func buildOpenOrdersQuery(msisdnPlaceholders []string, notInClause string) string {
	var queryBuilder strings.Builder

	queryBuilder.WriteString(`
		SELECT DISTINCT o.order_id, o.creation_date, o.due_date, o.state, o.order_type, 
		       o.customer_id, o.contact_phone, o.order_data, o.changing_date, o.changed_by_user, o.process_type
		FROM orders o
		JOIN portation_numbers pn ON o.order_id = pn.order_id
		WHERE pn.msisdn IN (`)
	queryBuilder.WriteString(strings.Join(msisdnPlaceholders, ","))
	queryBuilder.WriteString(")")
	queryBuilder.WriteString(notInClause)

	return queryBuilder.String()
}

func buildStateFilterClause(allowedStateCodes []int, startIndex int) (clause string, args []interface{}) {
	if len(allowedStateCodes) == 0 {
		return "", nil
	}

	statusPlaceholders := make([]string, len(allowedStateCodes))
	args = make([]interface{}, len(allowedStateCodes))

	for i, st := range allowedStateCodes {
		statusPlaceholders[i] = fmt.Sprintf("$%d", startIndex+i)
		args[i] = st
	}

	clause = " AND o.state NOT IN (" + strings.Join(statusPlaceholders, ",") + ")"

	return clause, args
}

func buildMSISDNFilter(msisdns []string) (placeholders []string, args []interface{}, err error) {
	placeholders = make([]string, len(msisdns))
	args = make([]interface{}, len(msisdns))

	for i, msisdn := range msisdns {
		placeholders[i] = fmt.Sprintf("$%d", i+1)

		var msisdnValue int64

		msisdnValue, err = parseInt64String(msisdn, "msisdn")
		if err != nil {
			return nil, nil, fmt.Errorf("invalid msisdn at position %d: %w", i, err)
		}

		args[i] = msisdnValue
	}

	return placeholders, args, nil
}

// sonar:ignore go:S4144
//
//nolint:dupl // .
func (r *PortInOrderRepository) GetByID(
	ctx context.Context, orderID int64, forUpdate bool, tx *sql.Tx) (*entities.PortInOrderEntity, error) {
	var connection QueriableConnection = r.db
	if tx != nil {
		connection = tx
	}

	const query = `
SELECT * FROM orders o
	WHERE o.order_id = $1;`

	const forUpdateQ = `
SELECT * FROM orders o
	WHERE o.order_id = $1
	FOR UPDATE;`

	q := query
	if forUpdate {
		q = forUpdateQ
	}

	var order entities.PortInOrderEntity

	err := connection.QueryRowContext(ctx, q, orderID).Scan(
		&order.ID,
		&order.CdbProcessID,
		&order.CreationDate,
		&order.DueDate,
		&order.State,
		&order.OrderType,
		&order.CustomerID,
		&order.ContactPhone,
		&order.OrderData,
		&order.ChangingDate,
		&order.ChangedByUser,
		&order.ProcessType,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("order not found: %w", err)
		}

		return nil, fmt.Errorf("get order by id: %w", err)
	}

	return &order, nil
}

// sonar:ignore go:S4144
//
//nolint:dupl // .
func (r *PortInOrderRepository) GetByCDBProcessID(
	ctx context.Context, cdbProcessID int64, forUpdate bool, tx *sql.Tx) (*entities.PortInOrderEntity, error) {
	var connection QueriableConnection = r.db
	if tx != nil {
		connection = tx
	}

	const query = `
SELECT * FROM orders o
	WHERE o.cdb_process_id = $1
	LIMIT 1;`

	const forUpdateQ = `
SELECT * FROM orders o
	WHERE o.cdb_process_id = $1
	FOR UPDATE
	LIMIT 1`

	q := query
	if forUpdate {
		q = forUpdateQ
	}

	var order entities.PortInOrderEntity

	err := connection.QueryRowContext(ctx, q, cdbProcessID).Scan(
		&order.ID,
		&order.CdbProcessID,
		&order.CreationDate,
		&order.DueDate,
		&order.State,
		&order.OrderType,
		&order.CustomerID,
		&order.ContactPhone,
		&order.OrderData,
		&order.ChangingDate,
		&order.ChangedByUser,
		&order.ProcessType,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("order not found: %w", err)
		}

		return nil, fmt.Errorf("get order by cdb_process_id: %w", err)
	}

	return &order, nil
}

func (r *PortInOrderRepository) buildSearchQuery() *strings.Builder {
	queryBuilder := &strings.Builder{}

	queryBuilder.WriteString(`
		SELECT DISTINCT o.order_id, o.creation_date, o.due_date, o.state, o.order_type, 
		       o.customer_id, o.contact_phone, o.order_data, o.changing_date, o.changed_by_user, o.process_type
		FROM orders o
		LEFT JOIN portation_numbers pn ON o.order_id = pn.order_id
		WHERE 1=1`)

	return queryBuilder
}

func (r *PortInOrderRepository) applyFilters(
	filters map[string]interface{},
	whereClause *strings.Builder,
) (args []interface{}, paramIndex int, err error) {
	args = make([]interface{}, 0)
	paramIndex = 1

	for key, value := range filters {
		switch key {
		case "msisdn":
			var normalized int64

			normalized, err = parseNumericFilterValue(value, "msisdn")
			if err != nil {
				return nil, 0, err
			}

			fmt.Fprintf(whereClause, " AND pn.msisdn = $%d", paramIndex)

			args = append(args, normalized)
			paramIndex++
		case "tempnumber":
			fmt.Fprintf(whereClause, " AND pn.temp_msisdn = $%d", paramIndex)

			args = append(args, value)
			paramIndex++
		case "cdbProcessID":
			var normalized int64

			normalized, err = parseNumericFilterValue(value, "cdbProcessId")
			if err != nil {
				return nil, 0, err
			}

			fmt.Fprintf(whereClause, " AND o.cdb_process_id = $%d", paramIndex)

			args = append(args, normalized)
			paramIndex++
		}
	}

	return args, paramIndex, nil
}

func (r *PortInOrderRepository) scanOrderRows(rows *sql.Rows) ([]*entities.PortInOrderEntity, error) {
	var orders []*entities.PortInOrderEntity

	for rows.Next() {
		var order entities.PortInOrderEntity

		err := rows.Scan(
			&order.ID,
			&order.CreationDate,
			&order.DueDate,
			&order.State,
			&order.OrderType,
			&order.CustomerID,
			&order.ContactPhone,
			&order.OrderData,
			&order.ChangingDate,
			&order.ChangedByUser,
			&order.ProcessType,
		)
		if err != nil {
			return nil, fmt.Errorf("scan order: %w", err)
		}

		orders = append(orders, &order)
	}

	return orders, nil
}

func (r *PortInOrderRepository) Search(
	ctx context.Context,
	filters map[string]any,
) (orders []*entities.PortInOrderEntity, totalCount int, err error) {
	queryBuilder := r.buildSearchQuery()

	whereClause := &strings.Builder{}

	args, _, err := r.applyFilters(filters, whereClause)
	if err != nil {
		return nil, 0, fmt.Errorf("apply filters: %w", err)
	}

	queryBuilder.WriteString(whereClause.String())
	queryBuilder.WriteString(" ORDER BY o.creation_date DESC")

	var rows *sql.Rows

	rows, err = r.db.QueryContext(ctx, queryBuilder.String(), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("search orders: %w", err)
	}

	defer rows.Close()

	orders, err = r.scanOrderRows(rows)
	if err != nil {
		return nil, 0, err
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}

	return orders, len(orders), nil
}

func (r *PortInOrderRepository) UpdateOrderPatch(ctx context.Context, tx *sql.Tx, applyPatch *containers.OrderPatchUpdate) error {
	tracer := diagnostics.TracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "PortInOrdersRepository.UpdateOrder")
	defer span.End()

	span.SetAttributes(
		attribute.Int64("orderId", applyPatch.OrderID),
		attribute.Bool("statusChanged", applyPatch.UpdateRes.StatusChanged),
		attribute.Bool("dueDateChanged", applyPatch.UpdateRes.DueDateChanged),
		attribute.Bool("cdbProcessIDChanged", applyPatch.UpdateRes.CDBIDChanged),
		attribute.Bool("portationNumbersStateChanged", applyPatch.UpdateRes.PortationNumbersStateChaned),
	)

	const query = `
UPDATE orders
  SET
  	state = $1,
    cdb_process_id = $2,
	due_date = $3,
	order_data = $4,
	changing_date = $5,
	changed_by_user = $6
  WHERE order_id = $7;`

	res, err := tx.ExecContext(ctx, query,
		applyPatch.State,
		applyPatch.CdbProcessID,
		applyPatch.DueDate,
		applyPatch.OrderDataJSON,
		applyPatch.ChangingDate,
		applyPatch.ChangedBy,
		applyPatch.OrderID,
	)
	if err != nil {
		return fmt.Errorf("failed to update order with id %d: %w", applyPatch.OrderID, err)
	}

	nAffected, _ := res.RowsAffected()
	if nAffected != 1 {
		return fmt.Errorf("order with id %d not updated", applyPatch.OrderID)
	}

	return nil
}

func parseInt64String(value, field string) (int64, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0, fmt.Errorf("%s is empty", field)
	}

	parsed, err := strconv.ParseInt(trimmed, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", field, err)
	}

	return parsed, nil
}

func parseNumericFilterValue(value any, field string) (int64, error) {
	switch v := value.(type) {
	case string:
		return parseInt64String(v, field)
	case *string:
		if v == nil {
			return 0, fmt.Errorf("%s is nil", field)
		}

		return parseInt64String(*v, field)
	case int:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	default:
		return 0, fmt.Errorf("unsupported %s type %T", field, value)
	}
}
