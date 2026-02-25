-- +goose Up
-- Make orders_bru trigger function not override NEW.changing_date when it is already set

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION orders_bru()
RETURNS TRIGGER AS $$
BEGIN
  IF NEW.changing_date IS NULL THEN
     NEW.changing_date := NOW();
  END IF;

  INSERT INTO orders_log(order_id, version_date, changed_by_user, state, order_data_log)
  VALUES(OLD.order_id, COALESCE(OLD.changing_date, OLD.creation_date), OLD.changed_by_user, OLD.state, OLD.order_data);

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose Down
-- Restore previous behavior (always set changing_date on update)

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION orders_bru()
RETURNS TRIGGER AS $$
BEGIN
  NEW.changing_date := NOW();

  INSERT INTO orders_log(order_id, version_date, changed_by_user, state, order_data_log)
  VALUES(OLD.order_id, COALESCE(OLD.changing_date, OLD.creation_date), OLD.changed_by_user, OLD.state, OLD.order_data);

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd
