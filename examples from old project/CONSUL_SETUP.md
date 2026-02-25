# Настройка Consul для portin-requests

## Обзор

В проект добавлена поддержка Consul для централизованного управления конфигурацией. Настройки `DisableOpenOrdersCheck` и `AllowedStatusesForNewOrder` загружаются **исключительно из Consul** и не могут быть переопределены через переменные окружения.

## Переменные окружения для Consul

Добавьте следующие переменные окружения для подключения к Consul:

```bash
CONSUL_HTTP_ADDR=http://localhost:8500
CONSUL_HTTP_TOKEN=your-consul-token-here
CONSUL_KEY=portin-requests/config
```

## Структура данных в Consul

В Consul по указанному ключу **обязательно** должен храниться JSON со следующей структурой:

```json
{
  "disableOpenOrdersCheck": false,
  "allowedStatusesForNewOrder": [
    "cdb-rejected",
    "canceled",
    "donor-rejected",
    "arbitation-pending",
    "arbitation-timeout",
    "debt-collection",
    "closed"
  ]
}
```

## Пример настройки в Consul

### Через Consul UI
1. Откройте Consul UI (обычно http://localhost:8500)
2. Перейдите в раздел Key/Value
3. Создайте ключ `portin-requests/config`
4. Добавьте JSON конфигурацию

### Через Consul CLI
```bash
consul kv put portin-requests/config '{
  "disableOpenOrdersCheck": false,
  "allowedStatusesForNewOrder": [
    "cdb-rejected",
    "canceled",
    "donor-rejected",
    "arbitation-pending",
    "arbitation-timeout",
    "debt-collection",
    "closed"
  ]
}'
```

### Через HTTP API
```bash
curl -X PUT http://localhost:8500/v1/kv/portin-requests/config \
  -d '{
    "disableOpenOrdersCheck": false,
    "allowedStatusesForNewOrder": [
      "cdb-rejected",
      "canceled",
      "donor-rejected",
      "arbitation-pending",
      "arbitation-timeout",
      "debt-collection",
      "closed"
    ]
  }'
```

## Логирование

При запуске приложение выводит в лог загруженные из Consul значения:

```
DisableOpenOrdersCheck: false
AllowedStatusesForNewOrder: [cdb_rejected canceled donor_rejected arbitation_pending arbitation_timeout debt_collection closed]
```

## Обработка ошибок

- Если Consul недоступен, приложение завершится с ошибкой
- Если ключ не найден в Consul, приложение завершится с ошибкой
- Если JSON в Consul некорректный, приложение завершится с ошибкой

**Важно:** Приложение не запустится без корректной конфигурации в Consul!

