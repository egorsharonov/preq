# portin-requests

Сервис обработки входящих заявок на портацию (PortIn)

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
1. Без указания флагов запускается основной сервис portin-requests.
2. **-help (-h)**
    Вызов справки с описанием флагов и их аргументов: ```portin-requests -h```.
3. **-migration [args]**  
    Флаг для применения миграций PostgreSQL. Требует установки [переменных окружения](#переменные-окружения) с настройками БД и пути до директории с файлами миграций.

    Для работы с миграциями используется библиотека [goose](https://github.com/pressly/goose). Для ее корректной работы в `.sql` скриптах миграций требуется использовать определенные [атрибуты](https://github.com/pressly/goose?tab=readme-ov-file#sql-migrations).

    Принимает следующие аргументы:

    - status - проверка текущего статуса применения миграций, вывод информации через стандартный логгер;

    - up [version] - "накатывание" миграций. В SQL такие миграции помечаются аннотацией `--+goose Up`. Например:
        - ```portin-requests -migration up``` - применение **всех** доступных миграций;
        - ```portin-requests -migration up 20250422124556``` - "накатить" миграций до указанной версии.

    - up-by-one - "накатывание" только следующей миграции:
        - ```portin-requests -migration up-by-one```

    - down [version] - откат миграций. В SQL такие миграции помечаются аннотацией `--+goose Down`. Например:
        - ```portin-requests -migration down``` - "откатить" до **предыдущей** версии;
        - ```portin-requests -migration down 20250422124556``` - "откатить" миграции до указанной версии.
    - ⚠️ down-all -  откат **ВСЕХ** миграций (до версии 0):
        - ```portin-requests -migration down-all```

    Пример вызова из Docker:
    ```shell
    /app/portin-requests -migration up
    ```

    Локально можно запустить миграции командами из Makefile: ```make migration-up```, ```make migration-down``` (аналогично запуску без указания версии).  
    Для локального запуска нужно настроить [переменные окружения](#переменные-окружения), пример есть в local-debug.

## Переменные окружения
- [Локальный запуск](https://gitlab.services.mts.ru/salsa/mnp-hub/local-debug/-/blob/master/config/portin-requests)
- [develop](https://gitlab.services.mts.ru/salsa/mnp-hub/kubernetes/-/tree/master/develop/portin-requests/envs)
- [pre-production](https://gitlab.services.mts.ru/salsa/mnp-hub/kubernetes/-/tree/master/pre-production/portin-requests/envs)
- [prod0000s7](https://gitlab.services.mts.ru/salsa/mnp-hub/kubernetes/-/tree/master/prod0000s7/portin-requests/envs)
- [prod0300s3](https://gitlab.services.mts.ru/salsa/mnp-hub/kubernetes/-/tree/master/prod0300s3/portin-requests/envs)

## Consul Configuration

Сервис использует Consul для централизованного управления конфигурацией. Некоторые настройки загружаются **исключительно из Consul** и не могут быть переопределены через переменные окружения.

Подробная инструкция по настройке Consul доступна в [CONSUL_SETUP.md](CONSUL_SETUP.md)

**Важно:** Приложение не запустится без корректной конфигурации в Consul!

## Публикация новой версии на pre-production, production стенды
Для публикации новой версии каждого сервиса необходимо после внесения исправлений:
&ensp;1. Создать новый тег в репозитории сервиса увеличив номер версии https://gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/-/tags

### Нумерация версий
Номер версии формируется в формате vXX.YY.ZZ
где ZZ - номер сборки в спринте, YY - условный номер спринта

Команда разработки Email: МТС ИТ MNP HUB mailto:mnphub@mts.ru
