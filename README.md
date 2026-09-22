# jobparser

Парсер вакансий с hh.ru/hh.kz, career.habr.com и qyzmet.enbek.kz с выдачей через
Telegram-бота и веб API.

> По исходному запросу источник "git" не идентифицирован однозначно (публичного
> джоб-борда с таким названием нет) — реализованы 3 явно названных источника:
> **hh**, **habr career**, **qyzmet.enbek.kz**. Если имелся в виду другой сайт —
> добавить его так же просто, как остальные: реализовать интерфейс `parsers.Parser`
> (см. `internal/parsers/interface.go`) и зарегистрировать в `cmd/worker/main.go`.

## Стек

- Go 1.22
- Web: Fiber
- Telegram: go-telegram-bot-api
- PostgreSQL + GORM (миграции продублированы в `migrations/` под golang-migrate)
- Meilisearch — полнотекстовый/фильтруемый поиск
- Redis — дедупликация вакансий между запусками + rate-limit
- Colly — скрейпинг habr career и qyzmet.enbek.kz
- Redis/cron — периодический запуск через `internal/scheduler`

## Архитектура

Три независимых бинарника, могут разворачиваться и масштабироваться отдельно:

- **cmd/worker** — по расписанию (`PARSE_INTERVAL`) обходит все источники,
  сохраняет вакансии в Postgres, индексирует в Meilisearch, дедуплицирует через
  Redis и рассылает уведомления подписчикам Telegram по новым вакансиям.
- **cmd/api** — веб API (`GET /api/vacancies`, `GET /api/vacancies/:id`) для
  фронтенда/сайта.
- **cmd/bot** — Telegram-бот: `/search`, `/subscribe`, `/unsubscribe`.

## О реализации источников

- **hh.ru / hh.kz** — используется официальный публичный API `api.hh.ru`
  (ключ не требуется), реализация полностью рабочая "из коробки"
  (`internal/parsers/hh.go`). Код региона задаётся через `HH_AREA_ID`
  (159 — Казахстан целиком, полный список — `GET https://api.hh.ru/areas`).
- **career.habr.com** и **qyzmet.enbek.kz** — публичных API нет, поэтому
  используется HTML-скрейпинг через `colly`. CSS-селекторы в
  `internal/parsers/habr.go` и `internal/parsers/qyzmet.go` — стартовая точка:
  **перед продакшен-использованием откройте страницы поиска в браузере,
  через DevTools проверьте актуальную вёрстку (или наличие XHR/JSON запроса
  со списком вакансий — если он есть, лучше дёргать его напрямую вместо
  парсинга HTML) и поправьте селекторы**. Сайты меняют вёрстку без
  предупреждения, поэтому это ожидаемая точка поддержки скрейпера.

## Запуск через Docker Compose

```bash
cp .env.example .env
# впишите TELEGRAM_TOKEN
docker compose up -d postgres redis meilisearch
docker compose run --rm api sh -c "true" # прогреть образ (миграции применятся автомиграцией GORM при старте api/worker)
docker compose up -d api worker bot
```

## Применение миграций через golang-migrate (альтернатива AutoMigrate)

```bash
migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/jobparser?sslmode=disable" up
```

## Локальный запуск без Docker

```bash
go mod tidy
export $(cat .env.example | xargs)   # или используйте direnv/godotenv
go run ./cmd/worker &
go run ./cmd/api &
go run ./cmd/bot &
```

## Веб API

```
GET /health
GET /api/vacancies?q=golang&city=Алматы&source=hh&limit=20&offset=0
GET /api/vacancies/:id
```

## Telegram-бот

```
/start
/search Golang backend
/subscribe Python
/unsubscribe
```

## Дальнейшие доработки (по желанию)

- sqlc вместо GORM, если нужен полный контроль над SQL и типобезопасность
- JWT + Redis-сессии, если веб API станет закрытым (личный кабинет)
- Вебхуки вместо long-polling для Telegram в проде
- Более умная дедупликация зарплатных вилок / нормализация городов
- Индекс `skills` в Meilisearch как facets для фильтра по стеку
