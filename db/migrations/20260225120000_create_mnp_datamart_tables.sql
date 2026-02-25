-- +goose Up

CREATE TABLE IF NOT EXISTS mnp_request (
  id                BIGSERIAL PRIMARY KEY,
  order_number      VARCHAR(64) NOT NULL,
  request_status_id INTEGER,
  request_date      TIMESTAMP,
  contract_date     TIMESTAMP,
  port_date         TIMESTAMP,
  from_date         TIMESTAMP,
  to_date           TIMESTAMP,
  change_date       TIMESTAMP NOT NULL DEFAULT NOW(),
  deleted           INTEGER   NOT NULL DEFAULT 0,
  cdb_id            VARCHAR(20),
  process_type      VARCHAR(20),
  port_type         VARCHAR(20) NOT NULL DEFAULT 'portin',
  subscriber_type   VARCHAR(20),
  message_code      VARCHAR(50),
  reject_reason     INTEGER,
  order_id          BIGINT NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS mnp_request_order_number_uk ON mnp_request(order_number);
CREATE UNIQUE INDEX IF NOT EXISTS mnp_request_port_type_order_id_uk ON mnp_request(port_type, order_id);
CREATE INDEX IF NOT EXISTS mnp_request_request_status_id_idx ON mnp_request(request_status_id);
CREATE INDEX IF NOT EXISTS mnp_request_request_date_idx ON mnp_request(request_date);
CREATE INDEX IF NOT EXISTS mnp_request_from_date_idx ON mnp_request(from_date);
CREATE INDEX IF NOT EXISTS mnp_request_change_date_idx ON mnp_request(change_date);

CREATE TABLE IF NOT EXISTS mnp_request_h (
  id                BIGSERIAL PRIMARY KEY,
  order_number      VARCHAR(64) NOT NULL,
  request_status_id INTEGER,
  request_date      TIMESTAMP,
  contract_date     TIMESTAMP,
  port_date         TIMESTAMP,
  from_date         TIMESTAMP,
  to_date           TIMESTAMP,
  change_date       TIMESTAMP NOT NULL DEFAULT NOW(),
  deleted           INTEGER   NOT NULL DEFAULT 0,
  cdb_id            VARCHAR(20),
  process_type      VARCHAR(20),
  port_type         VARCHAR(20) NOT NULL DEFAULT 'portin',
  subscriber_type   VARCHAR(20),
  message_code      VARCHAR(50),
  reject_reason     INTEGER,
  order_id          BIGINT NOT NULL
);

CREATE INDEX IF NOT EXISTS mnp_request_h_order_number_idx ON mnp_request_h(order_number);
CREATE UNIQUE INDEX IF NOT EXISTS ux_mnp_request_h_order_ver ON mnp_request_h(order_id, from_date);
CREATE INDEX IF NOT EXISTS mnp_request_h_order_id_idx ON mnp_request_h(order_id);
CREATE INDEX IF NOT EXISTS mnp_request_h_request_status_id_idx ON mnp_request_h(request_status_id);
CREATE INDEX IF NOT EXISTS mnp_request_h_from_date_idx ON mnp_request_h(from_date);
CREATE INDEX IF NOT EXISTS mnp_request_h_to_date_idx ON mnp_request_h(to_date);
CREATE INDEX IF NOT EXISTS mnp_request_h_change_date_idx ON mnp_request_h(change_date);

CREATE TABLE IF NOT EXISTS req_number (
  id           BIGSERIAL PRIMARY KEY,
  req_id       VARCHAR(64) NOT NULL,
  recipient_id VARCHAR(50),
  msisdn       VARCHAR(20) NOT NULL,
  rn           CHAR(5),
  change_date  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_req_number_req_msisdn ON req_number(req_id, msisdn);
CREATE INDEX IF NOT EXISTS req_number_req_id_idx ON req_number(req_id);
CREATE INDEX IF NOT EXISTS req_number_msisdn_idx ON req_number(msisdn);
CREATE INDEX IF NOT EXISTS req_number_change_date_idx ON req_number(change_date);

-- +goose StatementBegin
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'req_number_req_id_fk'
  ) THEN
    ALTER TABLE req_number
      ADD CONSTRAINT req_number_req_id_fk
      FOREIGN KEY (req_id)
      REFERENCES mnp_request(order_number)
      ON UPDATE CASCADE
      ON DELETE CASCADE;
  END IF;
END
$$;
-- +goose StatementEnd

CREATE TABLE IF NOT EXISTS mnp_raw_request (
  id             BIGINT PRIMARY KEY,
  req_id         VARCHAR(64),
  request_time   TIMESTAMP,
  xml_message    TEXT,
  operation_info VARCHAR(50),
  system_source  VARCHAR(50),
  system_dest    VARCHAR(50),
  change_date    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS mnp_raw_request_req_id_idx ON mnp_raw_request(req_id);
CREATE INDEX IF NOT EXISTS mnp_raw_request_request_time_idx ON mnp_raw_request(request_time);
CREATE INDEX IF NOT EXISTS mnp_raw_request_change_date_idx ON mnp_raw_request(change_date);


CREATE TABLE IF NOT EXISTS etl_state (
  job_name VARCHAR(64) PRIMARY KEY,
  watermark TIMESTAMPTZ,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down

DROP TABLE IF EXISTS etl_state;
DROP TABLE IF EXISTS mnp_raw_request;
ALTER TABLE req_number DROP CONSTRAINT IF EXISTS req_number_req_id_fk;
DROP TABLE IF EXISTS req_number;
DROP TABLE IF EXISTS mnp_request_h;
DROP TABLE IF EXISTS mnp_request;
