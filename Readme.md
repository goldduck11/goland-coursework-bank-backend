# Banking Service — REST API

Банковский REST API на Go: регистрация пользователей, управление счетами, виртуальные карты, кредиты, переводы и финансовая аналитика.

---

## Содержание

- [Технологический стек](#технологический-стек)
- [Структура проекта](#структура-проекта)
- [Переменные окружения](#переменные-окружения)
- [Запуск через Docker Compose](#запуск-через-docker-compose)
- [Запуск локально (без Docker)](#запуск-локально-без-docker)
- [API — эндпоинты](#api--эндпоинты)
- [Аутентификация](#аутентификация)
- [Шифрование и безопасность](#шифрование-и-безопасность)
- [Шедулер платежей](#шедулер-платежей)
- [Интеграции](#интеграции)

---

## Технологический стек

| Компонент | Технология |
|---|---|
| Язык | Go 1.25+ |
| Маршрутизация | gorilla/mux |
| База данных | PostgreSQL 17 + pgcrypto |
| Драйвер БД | lib/pq |
| Аутентификация | JWT (golang-jwt/jwt/v5) |
| Логирование | logrus (JSON-формат) |
| Шифрование | bcrypt, HMAC-SHA256, OpenPGP |
| Email | gomail.v2 (SMTP) |
| XML/SOAP | beevik/etree |
| Контейнеризация | Docker + Docker Compose |

---

## Структура проекта

```
.
├── main.go                          # Точка входа: инициализация, запуск сервера
├── Dockerfile
├── docker-compose.yaml
├── go.mod / go.sum
├── .env                             # Переменные окружения (не коммитить в git!)
├── .env.example                     # Шаблон .env
├── scripts/
│   └── init.sql                     # SQL-схема БД (выполняется при старте автоматически)
└── internal/
    ├── config/        config.go      # Чтение конфигурации из env
    ├── core/logger/   logger.go      # Инициализация logrus
    ├── database/      connect.go     # Подключение к PostgreSQL
    ├── model/                        # Структуры данных (User, Account, Card, Credit, ...)
    ├── repository/
    │   ├── interfaces.go             # Интерфейсы репозиториев
    │   └── postgres/                 # Реализации: user, account, card, credit, transaction
    ├── service/                      # Бизнес-логика: auth, account, card, credit, analytics
    │   └── external/                 # Интеграции: ЦБ РФ (SOAP), Email (SMTP)
    ├── handlers/                     # HTTP-хендлеры
    ├── middleware/                   # JWT middleware
    └── router/                      # Маршрутизация
```

---

## Переменные окружения

Скопируйте `.env.example` в `.env` и заполните:

```bash
cp .env.example .env
```

| Переменная | Обязательна | Описание |
|---|---|---|
| `DATABASE_URL` | **да** | DSN для PostgreSQL. Пример: `postgres://postgres:postgres@localhost:5432/banking?sslmode=disable` |
| `JWT_SECRET` | **да** | Секрет для подписи JWT-токенов (произвольная строка, минимум 32 символа) |
| `SERVER_PORT` | **да** | Порт сервера. Пример: `:8080` |
| `LOG_LEVEL` | нет | Уровень логирования: `debug`, `info`, `warn`, `error`. По умолчанию: `info` |
| `HMAC_SECRET` | нет | Секрет для HMAC-подписи данных карт. По умолчанию: `default-hmac-secret-change-me` |
| `PGP_PUBLIC_KEY` | нет | Публичный PGP-ключ в ASCII-armor формате для шифрования карт. Если не задан — генерируется автоматически |
| `PGP_PRIVATE_KEY` | нет | Приватный PGP-ключ в ASCII-armor формате |
| `PGP_PASSPHRASE` | нет | Пароль для PGP приватного ключа |
| `SMTP_HOST` | нет | SMTP-хост. По умолчанию: `smtp.gmail.com` |
| `SMTP_PORT` | нет | SMTP-порт. По умолчанию: `587` |
| `SMTP_USER` | нет | Логин SMTP |
| `SMTP_PASSWORD` | нет | Пароль SMTP |
| `SMTP_FROM` | нет | Адрес отправителя email. По умолчанию совпадает с `SMTP_USER` |

> **Важно:** если `PGP_PUBLIC_KEY` и `PGP_PRIVATE_KEY` не заданы, приложение автоматически генерирует временную пару ключей при старте. Это удобно для разработки, но **не подходит для продакшена** — при перезапуске контейнера старые зашифрованные данные карт нельзя будет расшифровать.

---

## Запуск через Docker Compose

Самый простой способ запустить всё окружение.

**1. Клонируйте репозиторий и перейдите в папку:**
```bash
git clone <url>
cd awesomeProject
```

**2. Создайте `.env` (минимальный вариант для разработки):**
```bash
DATABASE_URL=postgres://postgres:postgres@transaction-bank-postgres:5432/banking?sslmode=disable
JWT_SECRET=my-super-secret-jwt-key-change-me
SERVER_PORT=:8080
LOG_LEVEL=info
HMAC_SECRET=my-hmac-secret-change-me
```

**3. Запустите:**
```bash
docker compose up --build
```

После запуска:
- API доступен на `http://localhost:8080`
- PostgreSQL слушает на `localhost:5431` (порт проброшен из контейнера)
- Схема БД применяется автоматически при старте сервиса

**Остановка:**
```bash
docker compose down
```

**Остановка с удалением данных БД:**
```bash
docker compose down -v
```

---

## Запуск локально (без Docker)

**Требования:**
- Go 1.25+
- PostgreSQL 17 с расширением `pgcrypto`

**1. Установите зависимости:**
```bash
go mod download
```

**2. Создайте БД и примените схему:**
```bash
psql -U postgres -c "CREATE DATABASE banking;"
psql -U postgres -d banking -f scripts/init.sql
```

**3. Задайте переменные окружения:**
```bash
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/banking?sslmode=disable"
export JWT_SECRET="my-super-secret-jwt-key"
export SERVER_PORT=":8080"
export HMAC_SECRET="my-hmac-secret"
```

**4. Запустите:**
```bash
go run main.go
```

---

## API — эндпоинты

### Публичные (без JWT)

| Метод | Путь | Описание |
|---|---|---|
| `POST` | `/register` | Регистрация пользователя |
| `POST` | `/login` | Аутентификация, получение JWT-токена |

### Защищённые (требуют `Authorization: Bearer <token>`)

Все защищённые эндпоинты расположены под префиксом `/api/v1`.

#### Счета

| Метод | Путь | Описание |
|---|---|---|
| `POST` | `/api/v1/accounts` | Создать банковский счёт |
| `GET` | `/api/v1/accounts` | Список своих счетов |
| `POST` | `/api/v1/accounts/{accountId}/deposit` | Пополнить счёт |
| `POST` | `/api/v1/accounts/{accountId}/withdraw` | Снять средства |
| `GET` | `/api/v1/accounts/{accountId}/predict?days=N` | Прогноз баланса на N дней (макс. 365) |

#### Карты

| Метод | Путь | Описание |
|---|---|---|
| `POST` | `/api/v1/cards` | Выпустить виртуальную карту |
| `GET` | `/api/v1/cards` | Список своих карт |
| `POST` | `/api/v1/cards/pay` | Оплатить картой |

#### Переводы

| Метод | Путь | Описание |
|---|---|---|
| `POST` | `/api/v1/transfer` | Перевод между счетами |

#### Кредиты

| Метод | Путь | Описание |
|---|---|---|
| `POST` | `/api/v1/credits` | Оформить кредит |
| `GET` | `/api/v1/credits/{creditId}/schedule` | График платежей по кредиту |

#### Аналитика

| Метод | Путь | Описание |
|---|---|---|
| `GET` | `/api/v1/analytics` | Статистика доходов/расходов за текущий месяц |

---

## Примеры запросов

### Регистрация
```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"username": "ivan", "email": "ivan@example.com", "password": "secret123"}'
```
Ответ:
```json
{"user_id": "uuid...", "status": "success"}
```

### Логин
```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email": "ivan@example.com", "password": "secret123"}'
```
Ответ:
```json
{"token": "eyJhbGciOi..."}
```

### Создать счёт
```bash
curl -X POST http://localhost:8080/api/v1/accounts \
  -H "Authorization: Bearer <token>"
```

### Пополнить счёт
```bash
curl -X POST http://localhost:8080/api/v1/accounts/<accountId>/deposit \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"amount": 10000}'
```

### Перевод между счетами
```bash
curl -X POST http://localhost:8080/api/v1/transfer \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"sender_account_id": "uuid1", "receiver_account_id": "uuid2", "amount": 500}'
```

### Выпустить карту
```bash
curl -X POST http://localhost:8080/api/v1/cards \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"account_id": "uuid..."}'
```

### Оплата картой
```bash
curl -X POST http://localhost:8080/api/v1/cards/pay \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"card_id": "uuid...", "amount": 250}'
```

### Оформить кредит
```bash
curl -X POST http://localhost:8080/api/v1/credits \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"account_id": "uuid...", "principal": 50000, "term_months": 12}'
```

### Аналитика за месяц
```bash
curl http://localhost:8080/api/v1/analytics \
  -H "Authorization: Bearer <token>"
```
Ответ:
```json
{
  "month": "2026-06",
  "income": 15000.00,
  "expense": 3200.00,
  "net_flow": 11800.00,
  "credit_load": 4500.00,
  "total_balance": 45300.00
}
```

---

## Аутентификация

Все защищённые эндпоинты требуют JWT-токен в заголовке:

```
Authorization: Bearer <token>
```

Токен выдаётся при успешном логине (`POST /login`) и действует **24 часа**. Алгоритм подписи — HS256.

---

## Шифрование и безопасность

| Данные | Метод защиты |
|---|---|
| Пароли пользователей | bcrypt (cost 10) |
| Номер карты | PGP-шифрование (публичный ключ) |
| Срок действия карты | PGP-шифрование (публичный ключ) |
| CVV карты | bcrypt-хеш |
| Целостность данных карты | HMAC-SHA256 |
| Маршруты | JWT middleware (проверка подписи + срока действия) |
| Доступ к счетам/картам | Проверка владельца (запрет операций с чужими счетами) |

Для каждой карты при выпуске вычисляется HMAC-подпись над `(номер + срок)`. При расшифровке подпись проверяется — это защищает от подделки данных в БД.

---

## Шедулер платежей

При старте приложения запускается фоновый планировщик, который **каждые 12 часов** обрабатывает просроченные платежи:

- Если на счёте достаточно средств — платёж списывается, статус меняется на `paid`, отправляется email-уведомление.
- Если средств недостаточно — к сумме платежа начисляется штраф **+10%**, статус меняется на `overdue`, отправляется уведомление о просрочке.

---

## Интеграции

### Центральный банк РФ (ЦБ РФ)

При оформлении кредита сервис делает SOAP-запрос к `https://www.cbr.ru/DailyInfoWebServ/DailyInfo.asmx` для получения актуальной ключевой ставки. Если запрос не удался (нет сети, таймаут), используется запасное значение **16%**.

### SMTP (Email-уведомления)

Отправляются при:
- Успешной регистрации пользователя
- Списании очередного платежа по кредиту
- Просрочке платежа и начислении штрафа

Если SMTP не настроен (`SMTP_USER` или `SMTP_PASSWORD` не заданы), уведомления пропускаются без ошибки — приложение продолжает работу.

---

## Схема базы данных

БД содержит следующие таблицы (создаются автоматически из `scripts/init.sql`):

| Таблица | Описание |
|---|---|
| `users` | Пользователи (UUID, username, email, bcrypt-хеш пароля) |
| `accounts` | Банковские счета (UUID, номер счёта, валюта RUB, баланс) |
| `cards` | Карты (зашифрованный номер PGP, зашифрованный срок PGP, bcrypt CVV, HMAC) |
| `transactions` | История операций (deposit, withdraw, transfer, payment) |
| `credits` | Кредиты (сумма, ставка, срок, статус) |
| `payment_schedules` | График платежей (дата, сумма, части основного долга и процентов, штраф, статус) |