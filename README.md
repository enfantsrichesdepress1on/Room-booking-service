# Room Booking Service

Сервис бронирования переговорок на **Go + PostgreSQL + Docker Compose**.

Проект реализует обязательные пункты тестового задания и большую часть дополнительных:
- JWT-авторизация с обязательным `POST /dummyLogin`
- создание переговорок и расписаний
- генерация и хранение слотов в PostgreSQL
- бронирование только для роли `user`
- идемпотентная отмена брони
- список всех броней с пагинацией для `admin`
- список будущих броней текущего пользователя для `user`
- дополнительные `POST /register` и `POST /login`
- опциональная генерация conference link через mock external service
- unit-тесты и HTTP integration tests
- Docker Compose, Makefile, линтер-конфиг

---

## 1. Бизнес-логика и роли

В системе есть две роли:

### `admin`
Может:
- создавать переговорки
- создавать расписание доступности переговорки
- просматривать список всех броней с пагинацией
- просматривать комнаты и доступные слоты

Не может:
- создавать бронь
- отменять чужие или свои брони как `admin`

### `user`
Может:
- просматривать комнаты
- просматривать доступные слоты
- создавать бронь на слот от своего имени
- отменять **только свою** бронь
- просматривать **только свои будущие** брони

Не может:
- создавать переговорки
- создавать расписания
- просматривать список всех броней

---

## 2. Что реализовано

### Обязательная часть
- `POST /dummyLogin`
- `GET /_info`
- создание переговорок (`admin`)
- создание расписания один раз (`admin`)
- автоматическая генерация 30-минутных слотов
- список комнат (`admin`, `user`)
- список доступных слотов по комнате и дате (`admin`, `user`)
- создание брони (`user`)
- идемпотентная отмена своей брони (`user`)
- список всех броней с пагинацией (`admin`)
- список будущих броней текущего пользователя (`user`)
- хранение и передача времени в UTC
- unit-тесты
- integration / E2E сценарии

### Дополнительная часть
Реализовано:
- `POST /register`
- `POST /login`
- опциональный `createConferenceLink`
- mock Conference Service
- `Makefile`
- `Dockerfile`
- `docker-compose.yml`
- `.golangci.yml`
- Swagger scaffold в `docs/swagger.md`


---

## 3. Архитектурные решения

## Почему слоты хранятся в БД
По условию у слота должен быть **стабильный UUID**, потому что бронь создаётся по `slotId`.

Из этого следует, что слоты нельзя считать просто временными объектами «на лету». Поэтому в проекте слоты:
- генерируются системой
- сохраняются в PostgreSQL
- имеют стабильный ID
- могут быть безопасно использованы при создании брони


## Почему используется PostgreSQL
PostgreSQL здесь решает три важные задачи:

1. **Хранение данных и связей**
   - пользователи
   - переговорки
   - расписания
   - слоты
   - брони

2. **Ограничения целостности на уровне БД**
   - одно расписание на комнату
   - один активный booking на слот
   - уникальность слота


## Работа со временем
Все даты и время:
- хранятся в UTC
- передаются в UTC
- возвращаются в UTC

---

## 4. Структура проекта

```text
cmd/
  server/               
  seed/                

internal/
  app/                  
  auth/                 
  conference/           
  config/               
  db/                  
  httpapi/              
  middleware/           
  models/               
  repository/
    memory/             
    postgres/           
  service/              

migrations/            
api/                    
docs/                  
tests/                  
```

---

## 5. Локальный запуск

## Что должно быть установлено
Нужно иметь:
- Go 1.22+
- Docker
- Docker Compose
- Git

Проверка:

```bash
go version
docker --version
docker compose version
git --version
```


## 6. Запуск проекта

Из корня проекта:

```bash
go mod tidy
go build ./...
go test ./... -cover
```

Запуск через Docker Compose:

```bash
docker compose up --build
```

После запуска сервис будет доступен по адресу:

```text
http://localhost:8080
```

Проверка health endpoint:

```bash
curl -i http://localhost:8080/_info
```

Ожидается `200 OK`.

Остановка проекта:

```bash
docker compose down
```

Если нужно удалить volume с БД:

```bash
docker compose down -v
```

---

## 7. Makefile

Доступные команды:

```bash
make up
make seed
make test
make lint
```

### `make up`
Поднимает PostgreSQL и приложение.

### `make seed`
Заполняет БД тестовыми данными.

---

## 8. Авторизация и права доступа

## Как получить токен
Для тестирования обязательной части используется `POST /dummyLogin`.

### Получить admin token
```bash
curl -s -X POST http://localhost:8080/dummyLogin \
  -H "Content-Type: application/json" \
  -d '{"role":"admin"}'
```

### Получить user token
```bash
curl -s -X POST http://localhost:8080/dummyLogin \
  -H "Content-Type: application/json" \
  -d '{"role":"user"}'
```

Ответ:

```json
{"token":"<jwt>"}
```

### Сохранить токены в переменные shell
```bash
ADMIN_TOKEN='сюда_вставь_admin_token'
USER_TOKEN='сюда_вставь_user_token'
```

### Что содержит JWT
JWT содержит как минимум:
- `user_id`
- `role`

Это важно, потому что:
- сервер определяет роль пользователя по токену
- `user_id` для брони берётся из токена, а не из request body

---

## 9. Ручной сценарий проверки API

Ниже — полный сценарий, который можно повторить вручную.

## 9.1 Создание комнаты (`admin`)

```bash
curl -s -X POST http://localhost:8080/rooms/create \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name":"Room A",
    "description":"Main room",
    "capacity":8
  }'
```

Пример ответа:

```json
{
  "room": {
    "id": "f2543680-c481-4a8e-8fba-04add0b1eba8",
    "name": "Room A",
    "description": "Main room",
    "capacity": 8,
    "createdAt": "2026-03-25T16:52:31.515028Z"
  }
}
```

Сохранить ID комнаты:

```bash
ROOM_ID='f2543680-c481-4a8e-8fba-04add0b1eba8'
```

## 9.2 Создание расписания (`admin`)

```bash
curl -s -X POST http://localhost:8080/rooms/$ROOM_ID/schedule/create \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "roomId":"'"$ROOM_ID"'",
    "daysOfWeek":[1,2,3,4,5],
    "startTime":"09:00",
    "endTime":"18:00"
  }'
```

Что важно:
- расписание можно создать только один раз
- повторный вызов должен вернуть ошибку `SCHEDULE_EXISTS`
- `daysOfWeek` принимает значения только от `1` до `7`

## 9.3 Получение списка комнат (`admin` и `user`)

```bash
curl -s http://localhost:8080/rooms/list \
  -H "Authorization: Bearer $USER_TOKEN"
```

## 9.4 Получение доступных слотов (`admin` и `user`)

```bash
curl -s "http://localhost:8080/rooms/$ROOM_ID/slots/list?date=2026-03-26" \
  -H "Authorization: Bearer $USER_TOKEN"
```

В ответе придёт список свободных слотов на эту дату.

Пример первого слота:

```json
{
  "id": "23701be5-2641-55ba-975e-aabdfdb3e6fd",
  "roomId": "f2543680-c481-4a8e-8fba-04add0b1eba8",
  "start": "2026-03-26T09:00:00Z",
  "end": "2026-03-26T09:30:00Z"
}
```

Сохранить ID слота:

```bash
SLOT_ID='23701be5-2641-55ba-975e-aabdfdb3e6fd'
```

## 9.5 Создание брони (`user`)

```bash
curl -s -X POST http://localhost:8080/bookings/create \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "slotId":"'"$SLOT_ID"'",
    "createConferenceLink": true
  }'
```

Что важно:
- бронь доступна только роли `user`
- `admin` не должен мочь забронировать слот
- если слот уже занят, вернётся `409 SLOT_ALREADY_BOOKED`
- если слот в прошлом, вернётся `400 INVALID_REQUEST`

## 9.6 Список своих броней (`user`)

```bash
curl -s http://localhost:8080/bookings/my \
  -H "Authorization: Bearer $USER_TOKEN"
```

Что важно:
- возвращаются только **будущие** брони
- `userId` берётся из JWT

## 9.7 Список всех броней (`admin`)

```bash
curl -s "http://localhost:8080/bookings/list?page=1&pageSize=20" \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

Что важно:
- доступно только `admin`
- работает пагинация
- `pageSize` ограничен 100

## 9.8 Отмена своей брони (`user`)

```bash
BOOKING_ID='сюда_вставь_booking_id'

curl -s -X POST http://localhost:8080/bookings/$BOOKING_ID/cancel \
  -H "Authorization: Bearer $USER_TOKEN"
```

### Идемпотентность отмены
Повторный вызов той же команды:

```bash
curl -s -X POST http://localhost:8080/bookings/$BOOKING_ID/cancel \
  -H "Authorization: Bearer $USER_TOKEN"
```

должен снова вернуть `200 OK` и бронь со статусом `cancelled`.

---

## 10. Дополнительные endpoint'ы

## Регистрация
```bash
curl -s -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{
    "email":"user@example.com",
    "password":"strong-password",
    "role":"user"
  }'
```

## Логин
```bash
curl -s -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{
    "email":"user@example.com",
    "password":"strong-password"
  }'
```

---

## 11. Тесты

## Unit tests
Покрывают бизнес-логику сервисов.

Запуск:

```bash
go test ./... -cover
```

## HTTP integration / E2E tests
Покрывают сценарии:
- создание комнаты
- создание расписания
- создание брони
- отмена брони


---

