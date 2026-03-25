# Room Booking Service

Тестовое задание: сервис бронирования переговорок на Go + PostgreSQL + Docker Compose.

## Статус проекта

Репозиторий доведён до состояния, в котором есть:
- рабочая архитектурная основа под PostgreSQL;
- SQL migration files;
- unit-тесты бизнес-логики;
- HTTP E2E flow-тесты;
- отдельный PostgreSQL integration test scaffold;
- Docker Compose и команды для локального прогона.

Важно: в этой среде не было доступа к сети для скачивания Go-модулей, поэтому финальная компиляция и запуск должны быть подтверждены локально командами из раздела **Проверка**.

## Что реализовано

### Обязательное
- `/dummyLogin` с JWT и фиксированными UUID для `admin` и `user`.
- `/_info`.
- Создание переговорок (`admin`).
- Создание расписания один раз (`admin`).
- Генерация 30-минутных слотов системой.
- Список переговорок.
- Список доступных слотов по комнате и дате.
- Создание брони только для `user`.
- Идемпотентная отмена своей брони.
- Список всех броней с пагинацией (`admin`).
- Список будущих броней текущего пользователя (`user`).
- Хранение времени в UTC.
- Unit-тесты и HTTP integration/E2E тесты.

### Дополнительно
- `/register` и `/login`.
- `createConferenceLink` через mock conference service.
- `Makefile`.
- Docker Compose.
- Конфиг линтера.
- SQL migration files.
- Swagger scaffold.

## Архитектурные решения

### Генерация слотов
Выбран гибридный подход:
- при создании расписания сервис генерирует слоты на окно вперёд (`14` дней);
- при запросе слотов на конкретную дату сервис лениво дозаполняет отсутствующие слоты именно на эту дату.

Почему так:
- у слотов есть стабильные UUID и они хранятся в БД;
- горячий endpoint `/rooms/{roomId}/slots/list` работает по предсгенерированным данным;
- при нагрузке это дешевле, чем каждый раз вычислять слоты на лету.

UUID слота детерминирован: `UUIDv5(roomId + startTimeUTC)`, поэтому повторная генерация не ломает ссылочную целостность.

### Ограничения консистентности
- одно расписание на комнату: `UNIQUE (room_id)` в `schedules`;
- не более одной активной брони на слот: partial unique index на `bookings(slot_id) WHERE status='active'`;
- уникальность слота: `UNIQUE (room_id, start_time, end_time)`.

### Время
- все значения времени хранятся и отдаются в UTC;
- для слотов используется `TIMESTAMPTZ`.

### PostgreSQL
Слои работы с БД:
- `internal/db` — pool и применение миграций;
- `internal/repository/postgres` — SQL-репозитории;
- `internal/service` — бизнес-логика без SQL в handlers.

Миграции лежат в:
- `migrations/` — видимые SQL файлы в корне проекта;
- `internal/db/migrations/` — embedded-копия для применения приложением.

## Структура проекта

```text
cmd/server            # запуск HTTP API
cmd/seed              # наполнение тестовыми данными
internal/app          # сборка зависимостей
internal/auth         # JWT
internal/config       # env config
internal/db           # pool + migrations
internal/httpapi      # handlers и ответы
internal/middleware   # auth middleware
internal/models       # доменные модели
internal/repository   # postgres + memory repos
internal/service      # бизнес-логика
internal/conference   # mock integration
migrations            # SQL migration files
tests                 # HTTP и DB tests
```

## Запуск

```bash
make up
```

Сервис будет доступен на `http://localhost:8080`.

### Наполнение тестовыми данными

```bash
make seed
```

## Проверка

### 1. Подтянуть зависимости

```bash
go mod tidy
```

### 2. Проверить форматирование и сборку

```bash
gofmt -w ./cmd ./internal ./tests
go build ./...
```

### 3. Прогнать unit и HTTP tests

```bash
go test ./... -cover
```

### 4. Поднять сервис с Postgres

```bash
docker compose up --build
```

### 5. Опционально прогнать PostgreSQL integration test

```bash
RUN_DB_TESTS=1 go test ./tests -run TestPostgresRepositories -v
```

## Основные ручки
- `POST /dummyLogin`
- `POST /register`
- `POST /login`
- `GET /rooms/list`
- `POST /rooms/create`
- `POST /rooms/{roomId}/schedule/create`
- `GET /rooms/{roomId}/slots/list?date=YYYY-MM-DD`
- `POST /bookings/create`
- `GET /bookings/list?page=1&pageSize=20`
- `GET /bookings/my`
- `POST /bookings/{bookingId}/cancel`
- `GET /_info`

## Conference link
Если при создании брони указан `createConferenceLink=true`, сервис вызывает mock conference service и сохраняет ссылку в брони.

Принятое поведение при сбоях внешнего сервиса:
- если внешняя интеграция не вернула ссылку, бронь не создаётся;
- бронь и ссылка сохраняются атомарно в рамках операции создания брони.

## Swagger
В `docs/swagger.md` добавлен scaffold для генерации Swagger через `swaggo/swag`.

## Что осталось подтвердить локально
- `go mod tidy`;
- успешный `go build ./...`;
- успешный `go test ./...`;
- успешный `docker compose up --build`;
- ручной smoke test через `dummyLogin -> create room -> create schedule -> list slots -> create booking -> cancel booking`.

## Что можно улучшить дальше
- полноценные versioned migrations через `golang-migrate`;
- реальная генерация Swagger из аннотаций и публикация UI;
- нагрузочный сценарий и краткий отчёт;
- observability: structured logging, tracing, metrics.
