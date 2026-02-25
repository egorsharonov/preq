-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
      FROM pg_type
      WHERE typname = 'process_type_enum'
  ) THEN
    CREATE TYPE process_type_enum AS ENUM ('ShortTimePort', 'LongTimePort', 'GOS');
  END IF;
END
$$;
-- +goose StatementEnd

ALTER TABLE orders ADD COLUMN IF NOT EXISTS process_type process_type_enum DEFAULT 'ShortTimePort';

-- Добавляем индекс для нового поля
CREATE INDEX IF NOT EXISTS orders_process_type ON orders(process_type);

-- Обновляем существующие записи, устанавливая значение по умолчанию
UPDATE orders SET process_type = 'ShortTimePort' WHERE process_type IS NULL;

-- Добавляем комментарий
COMMENT ON COLUMN orders.process_type IS 'Тип процесса портации (ShortTimePort, LongTimePort, GOS)';

-- +goose Down
DROP INDEX IF EXISTS orders_process_type;
ALTER TABLE orders DROP COLUMN IF EXISTS process_type;
DROP TYPE IF EXISTS process_type_enum;
