# TrackLinker — backend

Бэкенд сервиса, который по ссылке на трек или альбом из одного стриминга собирает карточку с метаданными и ссылками на тот же релиз во всех поддерживаемых сервисах.

Поддерживаемые платформы: **Spotify, Apple Music, Яндекс Музыка, Deezer**.

## Архитектура

Два микросервиса на Go, общаются через RabbitMQ:

```
клиент ──HTTP──► parsing ──RabbitMQ──► collection
                   │                      │
                   ├─ PostgreSQL          └─ PostgreSQL
                   ├─ S3 (превью треков)
                   └─ API стримингов (Spotify, Apple, Яндекс, Deezer)
```

**parsing** — разбирает входящую ссылку, определяет платформу, параллельно ищет релиз в остальных сервисах, нормализует метаданные (трек, альбом, артисты, жанры, дата релиза) в единую модель, сохраняет в Postgres, 30-секундные превью треков загружает в S3.

**collection** — пользовательские коллекции и прослушивания: добавленный контент, теги, реферальные ссылки и клики по ним. Данные о треках получает от parsing через очередь.

Код разделён по слоям (ports & adapters): `transport` → `service` → `repository`, внешние зависимости (БД, S3, MQ, другие сервисы) спрятаны за интерфейсами в `internal/ports`.

## API (parsing)

Базовый путь `/api/v1`, полная схема в Swagger (`parsing/docs`).

| Метод | Путь | Что делает |
|-|-|-|
| `POST` | `/music/search` | Поиск по ссылке или по названию и артисту |
| `GET` | `/track/:id` | Карточка трека |
| `GET` | `/album/:album_id` | Карточка альбома |
| `GET` | `/album/:album_id/track/:track_id` | Трек в составе альбома |

collection: `POST /api/v1/action/listen` — зафиксировать прослушивание.

## Стек

Go 1.23, Gin, GORM + PostgreSQL, RabbitMQ (amqp), S3 (aws-sdk-go), Swagger (swaggo), Docker, GitLab CI.

## Запуск

Конфигурация читается из `.env.dev` в папке сервиса:

```env
DATABASE_URI=postgres://user:pass@localhost:5432/tracklinker
AMQP_URL=amqp://guest:guest@localhost:5672/
SPOTIFY_CLIENT_ID=...
SPOTIFY_CLIENT_SECRET=...
S3_ENDPOINT=...
S3_BUCKET=...
S3_REGION=...
S3_ACCESS_KEY=...
S3_SECRET_KEY=...
SWAGGER_HOST=localhost:8080
```

```bash
cd parsing
docker-compose --env-file ./.env.dev up -d   # Postgres и RabbitMQ
go run ./cmd
```

`collection` запускается так же из своей папки.

## Тесты

```bash
cd parsing
go test ./...                                  # unit-тесты: разбор ссылок, сервис парсинга на моках (testify)
go test -tags integration ./tests/...          # интеграционные: реальные API Spotify, Apple, Яндекса, Deezer
```

Для интеграционных тестов Spotify нужны `SPOTIFY_CLIENT_ID` и `SPOTIFY_CLIENT_SECRET` в `.env.dev`.

## Релиз

Сборка Docker-образа в GitLab CI запускается по тегу:

```bash
git tag -a parsing/v1.0.0 -m "parsing/v1.0.0"
git push --tags
```
