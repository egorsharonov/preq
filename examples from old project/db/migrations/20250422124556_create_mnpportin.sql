-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
      FROM pg_type
      WHERE typname = 'order_type_enum'
  ) THEN
    CREATE TYPE order_type_enum AS ENUM ('portin');
  END IF;
END
$$;
-- +goose StatementEnd

CREATE TABLE IF NOT EXISTS orders(
  order_id        BIGSERIAL       PRIMARY KEY,
  cdb_process_id  BIGINT,
  creation_date   TIMESTAMPTZ     NOT NULL,
  due_date        TIMESTAMPTZ,
  state           INTEGER         NOT NULL,
  order_type      order_type_enum NOT NULL,
  customer_id     VARCHAR(128),
  contact_phone   VARCHAR(20),
  order_data      JSONB,
  changing_date   TIMESTAMPTZ,
  changed_by_user VARCHAR(128)
) PARTITION BY RANGE (order_id);

CREATE TABLE IF NOT EXISTS orders_1 PARTITION OF orders
  FOR VALUES FROM (0) to (1000000);
CREATE TABLE IF NOT EXISTS orders_2 PARTITION OF orders
  FOR VALUES FROM (1000001) to (2000000);

CREATE INDEX IF NOT EXISTS orders_cdb_process_id ON orders(cdb_process_id);
CREATE INDEX IF NOT EXISTS orders_creation_date ON orders(creation_date);
CREATE INDEX IF NOT EXISTS orders_due_date ON orders(due_date);
CREATE INDEX IF NOT EXISTS orders_state ON orders(state);
CREATE INDEX IF NOT EXISTS orders_order_type ON orders(order_type);
CREATE INDEX IF NOT EXISTS orders_customer_id ON orders(customer_id);
CREATE INDEX IF NOT EXISTS orders_contact_phone ON orders(contact_phone);
CREATE INDEX IF NOT EXISTS orders_changing_date ON orders(changing_date);
CREATE INDEX IF NOT EXISTS orders_changed_by_user ON orders(changed_by_user);
CREATE INDEX IF NOT EXISTS orders_lastName ON orders USING GIN ((order_data -> 'person.lastName') jsonb_path_ops);
CREATE INDEX IF NOT EXISTS orders_docNumber ON orders USING GIN ((order_data -> 'idDocuments.docNumber') jsonb_path_ops);
CREATE INDEX IF NOT EXISTS orders_contractDate ON orders USING GIN ((order_data -> 'contract.contractDate') jsonb_path_ops);

COMMENT ON TABLE orders IS 'Заявки на портацию';
COMMENT ON COLUMN orders.order_id IS 'ИД заявки на портацию.';
COMMENT ON COLUMN orders.cdb_process_id IS 'Идентификатор процесса БДПН.';
COMMENT ON COLUMN orders.creation_date IS 'Дата создания заявки на портацию. Автоматически заполняется датой создания записи.';
COMMENT ON COLUMN orders.due_date IS 'Запланированная дата портации.';
COMMENT ON COLUMN orders.state IS 'ИД состояния ЖЦ (0-created, 100-in_progress, 200-completed, 400-failed)';
COMMENT ON COLUMN orders.order_type IS 'Тип заявки (portin)';
COMMENT ON COLUMN orders.customer_id IS 'Идентификатор экосистемного клиента.';
COMMENT ON COLUMN orders.contact_phone IS 'Контактный телефон.';
COMMENT ON COLUMN orders.order_data IS 'Полный объект заявки.';
COMMENT ON COLUMN orders.changing_date IS 'Дата последнего изменения.';
COMMENT ON COLUMN orders.changed_by_user IS 'Кем произведено изменение. Доменное NTLM-имя оператора сервиса, под котором идет обращение.';

CREATE TABLE IF NOT EXISTS portation_numbers(
  order_id    BIGINT      NOT NULL REFERENCES orders(order_id) ON DELETE CASCADE ON UPDATE CASCADE,
  msisdn      VARCHAR(20) NOT NULL,
  telco_id    VARCHAR(128),
  temp_msisdn VARCHAR(20) NOT NULL
);

CREATE INDEX IF NOT EXISTS portation_numbers_order_id ON portation_numbers(order_id);
CREATE INDEX IF NOT EXISTS portation_numbers_msisdn ON portation_numbers(msisdn);
CREATE INDEX IF NOT EXISTS portation_numbers_telco_id ON portation_numbers(telco_id);
CREATE INDEX IF NOT EXISTS portation_numbers_temp_msisdn ON portation_numbers(temp_msisdn);

COMMENT ON TABLE portation_numbers IS 'Номера портации.';
COMMENT ON COLUMN portation_numbers.order_id IS 'ИД заявки на портацию.';
COMMENT ON COLUMN portation_numbers.msisdn IS 'Портируемый номер в формате 79000000000';
COMMENT ON COLUMN portation_numbers.telco_id IS 'ИД аккаунта Teklo';
COMMENT ON COLUMN portation_numbers.temp_msisdn IS 'Временный номер телефона';

CREATE TABLE IF NOT EXISTS orders_log(
  order_id       BIGINT      NOT NULL REFERENCES orders(order_id) ON DELETE CASCADE ON UPDATE CASCADE,
  version_date   TIMESTAMPTZ NOT NULL,
  changed_by_user VARCHAR(128),
  state          INTEGER     NOT NULL,
  order_data_log JSONB
);

CREATE INDEX IF NOT EXISTS orders_log_order_id ON orders_log(order_id);
CREATE INDEX IF NOT EXISTS orders_log_version_date ON orders_log(version_date);
CREATE INDEX IF NOT EXISTS orders_log_changed_by_user ON orders_log(changed_by_user);
CREATE INDEX IF NOT EXISTS orders_log_state ON orders_log(state);
CREATE INDEX IF NOT EXISTS orders_log_lastName ON orders_log USING GIN ((order_data_log -> 'person.lastName') jsonb_path_ops);
CREATE INDEX IF NOT EXISTS orders_log_docNumber ON orders_log USING GIN ((order_data_log -> 'idDocuments.docNumber') jsonb_path_ops);
CREATE INDEX IF NOT EXISTS orders_log_contractDate ON orders_log USING GIN ((order_data_log -> 'contract.contractDate') jsonb_path_ops);

COMMENT ON TABLE orders_log IS 'История изменения заявок.';
COMMENT ON COLUMN orders_log.order_id IS 'ИД заявки на портацию.';
COMMENT ON COLUMN orders_log.version_date IS 'Дата с которой действовала версия заявки.';
COMMENT ON COLUMN orders_log.changed_by_user IS 'Кем создана версия. Доменное NTLM-имя оператора сервиса, под котором идет обращение.';
COMMENT ON COLUMN orders_log.state IS 'ИД состояния ЖЦ версии заявки';
COMMENT ON COLUMN orders_log.order_data_log IS 'JSON Полный объект версии заявки.';

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION orders_bri()
RETURNS TRIGGER AS $$
BEGIN
  IF (NEW.creation_date IS NULL) THEN
    NEW.creation_date := NOW();
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE OR REPLACE TRIGGER orders_bri BEFORE INSERT
ON orders
FOR EACH ROW
EXECUTE FUNCTION orders_bri();

-- Сохранение предыдущей версии заявки в orders_log перед update
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

CREATE OR REPLACE TRIGGER orders_bru BEFORE UPDATE
ON orders
FOR EACH ROW
EXECUTE FUNCTION orders_bru();

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION create_constraint_if_not_exists (
    t_name text, c_name text, constraint_sql text
)
RETURNS void AS $$
BEGIN
  -- Look for our constraint
  IF NOT EXISTS (
    SELECT constraint_name
      FROM information_schema.constraint_column_usage
      WHERE table_name = t_name
        AND constraint_name = c_name
  ) THEN
    EXECUTE constraint_sql;
  END IF;
END;
$$ LANGUAGE 'plpgsql';
-- +goose StatementEnd

SELECT create_constraint_if_not_exists(
  'orders_log',
  'orders_log_uk',
  'ALTER TABLE orders_log ADD CONSTRAINT orders_log_uk UNIQUE (order_id, version_date);'
);

SELECT create_constraint_if_not_exists(
  'portation_numbers',
  'portation_numbers_uk',
  'ALTER TABLE portation_numbers ADD CONSTRAINT portation_numbers_uk UNIQUE (order_id, msisdn);'
);

-- +goose Down
DROP TABLE IF EXISTS portation_numbers CASCADE;
DROP TABLE IF EXISTS orders CASCADE;
DROP TABLE IF EXISTS orders_log CASCADE;