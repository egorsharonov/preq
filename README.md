# mnp-datamart

Сервис для выгрузки заявок из БД и передачи их в витрины.

[US01.AV](https://confluence.mts.ru/pages/viewpage.action?pageId=1724031378)

# Содержание

- [Сборка](#сборка)
    - [Taskfile](#taskfile)
        - [Установка](#установка)
        - [Использование](#использование)
    - [Описание переменных окружения проекта](#описание-переменных-окружения-проекта)
    - [Локальная отладка с использованием Docker](#локальная-отладка-с-использованием-docker)
    - [Зависимости](#зависимости)
    - [Скачивание зависимостей из GitLab MTS](#скачивание-зависимостей-из-gitlab-mts)
- [Флаги запуска](#флаги-запуска)
- [Переменные окружения](#переменные-окружения)
- [Consul Configuration](#consul-configuration)

## Сборка

Используемая версия Go: **1.25.7**

### Taskfile

#### Установка

[Инструкция по установке для разных ОС](https://taskfile.dev/docs/installation)

#### Использование

В терминале достаточно написать `task {команда}` для вызова скриптов из Taskfile.yaml:

`task all` - запуск генерации кода, линтера, тестов и сборки  
`task generate` - запуск генерации openapi http клинетов и сервера  
`task lint` - запуск только линтера  
`task build` - запуск сборки бинарника  
`task clean` - запуск очистики собраных бинарников и временных файлов с тестов  
`task mock-config` - запуск генерации конфигурации mockery  
`task mock` - запуск генерации моков  
`task test` - запуск тестов с отчетом покрытия
`task test-cover` - запуск тестов с анализом покрытия  
`task migration:status` - проверка статуса применения миграций  
`task migration:up` - применение всех доступных миграций  
`task migration:down` - откат до предыдущей версии

В качестве линтера используется `golangci-lint`

Для VS Code есть [расширение](https://marketplace.visualstudio.com/items?itemName=task.vscode-task) для отображения списка задач и их запуска.

### Описание переменных окружения проекта

Для локального запуска в переменных окружения проекта необходимо указать  значения из пример - [.env.template](./.env.template)


Переменные для стендов указываются в [репозитории kubernetes](https://gitlab.services.mts.ru/salsa/mnp-hub/kubernetes/-/tree/master/)

### Локальная отладка с использованием Docker

[Репозиторий local-debug](https://gitlab.services.mts.ru/salsa/mnp-hub/local-debug)

### Зависимости

Для скачивания зависимостей нужно установить переменные среды в системе (и не забыть перезагрузить комп)

| Переменная   | Значение                                                                 |
|--------------|--------------------------------------------------------------------------|
| `GONOPROXY`  | `gitlab.services.mts.ru/*`                                               |
| `GONOSUMDB`  | `gitlab.services.mts.ru/*`                                               |
| `GOPRIVATE`  | `gitlab.services.mts.ru`                                                 |
| `GOPROXY`    | `https://nexus.services.mts.ru/repository/go-proxy/`                     |
| `GOSUMDB`    | `sum.golang.org https://nexus.services.mts.ru/repository/go-sum`         |


### Скачивание зависимостей из GitLab MTS

Для подтягивания зависимостей из корп. гитлаба нужно настроить файл .netrc.
На  Windows файл назвать **_netrc** и положить в папку %USERPROFILE%

Сгенерировать токен GitLab со скоупом **api** [тут](https://gitlab.services.mts.ru/-/profile/personal_access_tokens)

В файл добавить следующее содержимое:

> machine gitlab.services.mts.ru  
login \*логин\*  
password \*API-токен\*

## Флаги запуска
1. Без указания флагов запускается основной сервис mnp-datamart.
2. **-help (-h)**
   Вызов справки с описанием флагов и их аргументов: ```mnp-datamart -h```.
3. **-migration [args]**  
   Флаг для применения миграций PostgreSQL. Требует установки [переменных окружения](#переменные-окружения) с настройками БД и пути до директории с файлами миграций.

   Для работы с миграциями используется библиотека [goose](https://github.com/pressly/goose). Для ее корректной работы в `.sql` скриптах миграций требуется использовать определенные [атрибуты](https://github.com/pressly/goose?tab=readme-ov-file#sql-migrations).

   Принимает следующие аргументы:

    - status - проверка текущего статуса применения миграций, вывод информации через стандартный логгер;

    - up [version] - "накатывание" миграций. В SQL такие миграции помечаются аннотацией `--+goose Up`. Например:
        - ```mnp-datamart -migration up``` - применение **всех** доступных миграций;
        - ```mnp-datamart -migration up 20250422124556``` - "накатить" миграций до указанной версии.

    - up-by-one - "накатывание" только следующей миграции:
        - ```mnp-datamart -migration up-by-one```

    - down [version] - откат миграций. В SQL такие миграции помечаются аннотацией `--+goose Down`. Например:
        - ```mnp-datamart -migration down``` - "откатить" до **предыдущей** версии;
        - ```mnp-datamart -migration down 20250422124556``` - "откатить" миграции до указанной версии.
    - ⚠️ down-all -  откат **ВСЕХ** миграций (до версии 0):
        - ```mnp-datamart -migration down-all```

   Пример вызова из Docker:
    ```shell
    /app/mnp-datamart -migration up
    ```

   Локально можно запустить миграции командами из Makefile: ```make migration-up```, ```make migration-down``` (аналогично запуску без указания версии).  
   Для локального запуска нужно настроить [переменные окружения](#переменные-окружения), пример есть в local-debug.

## Переменные окружения
- [Локальный запуск](https://gitlab.services.mts.ru/salsa/mnp-hub/local-debug/-/blob/master/config/mnp-datamart)
- [develop](https://gitlab.services.mts.ru/salsa/mnp-hub/kubernetes/-/tree/master/develop/mnp-datamart/envs)
- [pre-production](https://gitlab.services.mts.ru/salsa/mnp-hub/kubernetes/-/tree/master/pre-production/mnp-datamart/envs)
- [prod0000s7](https://gitlab.services.mts.ru/salsa/mnp-hub/kubernetes/-/tree/master/prod0000s7/mnp-datamart/envs)
- [prod0300s3](https://gitlab.services.mts.ru/salsa/mnp-hub/kubernetes/-/tree/master/prod0300s3/mnp-datamart/envs)

## Публикация новой версии на pre-production, production стенды
Для публикации новой версии каждого сервиса необходимо после внесения исправлений:
1. Создать новый тег в репозитории сервиса увеличив номер версии https://gitlab.services.mts.ru/salsa/mnp-hub/mnp-datamart/-/tags

### Нумерация версий
Номер версии формируется в формате vXX.YY.ZZ
где ZZ - номер сборки в спринте, YY - условный номер спринта

Команда разработки Email: МТС ИТ MNP HUB mailto:mnphub@mts.ru

## Архитектура mnp-datamart (MNP HUB)

Сервис запускает две ETL-джобы:
- `portin-dag` — перенос заявок `orders`, истории `orders_log` и номеров `portationNumbers` в `mnp_request`, `mnp_request_h`, `req_number`.
- `cdb-message-dag` — перенос `mnp_message` + `mnp_process` в `mnp_raw_request`.

Джобы работают по расписанию (по умолчанию ежечасно) и могут быть вызваны вручную:
- `POST /jobs/portin/run`
- `POST /jobs/cdb-message/run`

Health endpoints:
- `GET /health/live`
- `GET /health/ready`

Ключевые правила:
- Загружается только `order_type='portin'` и только физлица (`subscriber_type=Person`).
- В `mnp_request` используется upsert (`order_number`).
- В `mnp_request_h` используется idempotent insert (`order_id, from_date`).
- В `req_number` используется upsert (`req_id, msisdn`).
- В `mnp_raw_request` используется upsert (`id`).
- Мэппинг статусов выполняется на стороне mnp-datamart.
- `reject_reason` заполняется только первым числовым кодом.

### Контракт с DataHouse по техполям

`mnp-datamart-db` хранит бизнес-данные витрины. Технические поля DataHouse (`raw_dt`, `raw_ts`, `processed_dttm`, `etl_run_id` и т.п.) заполняются downstream ETL-процессами DataHouse (RDB2HADOOP/Airflow).
