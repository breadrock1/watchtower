# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Watchtower — Go-сервис для мониторинга S3-событий (создание/загрузка файлов) и их дальнейшей AI-обработки: извлечение текста из PDF, DOCX, TXT (через OCR и LLM), построение графа сущностей и сохранение результата в Doc-Search сервис.

Ключевые зависимости: S3 (MinIO), RabbitMQ (очередь задач), Redis (хранение статусов задач), OpenTelemetry (трейсинг/логи), Prometheus (метрики).

## Архитектура

Чистая архитектура (Domain → Application → Infrastructure) с разделением на два домена:

- **`internal/core/cloud/`** — File Storage domain (S3): управление бакетами и объектами
  - `domain/` — интерфейсы: `ICloudStorage`, `IBucketManager`, `IObjectManager`, `IObjectWalker`, `IShareManager`
  - `application/` — use cases: `StorageUseCase` (CRUD бакетов/объектов) и `StoragePool` (пул инстансов по организациям)
  - `infrastructure/s3/` — реализация через MinIO SDK

- **`internal/support/task/`** — Task Processing domain: управление задачами обработки
  - `domain/` — модель `Task`, статусы (Failed/Received/Pending/Processing/Successful), интерфейсы `ITaskManager`, `ITaskQueue`
  - `application/` — `TaskUseCase` (публикация, обновление статусов, распознавание, сохранение документа)
  - `infrastructure/redis/` — хранение/статусы задач
  - `infrastructure/rmq/` — очередь через RabbitMQ
  - `infrastructure/docparser/` — клиент для внешнего сервиса распознавания текста
  - `infrastructure/docsearch/` — клиент для Doc-Search сервиса

- **`internal/process/`** — Orchestrator: связывает оба домена. `LaunchListener()` читает из RabbitMQ, обрабатывает задачи (скачать → распознать → сохранить документ).

- **`internal/shared/`** — общие компоненты:
  - `kernel/` — type aliases (`BucketID`, `ObjectID`, `TaskID`, `OrganizationID`), контекст
  - `metrics/` — Prometheus-метрики (counter-ы и гистограммы)

- **`cmd/watchtower/httpserver/`** — HTTP-сервер на Fiber (v2)
  - Ручки: `/api/v1/system/*`, `/api/v1/tasks/*`, `/api/v1/buckets/*`, `/api/v1/objects/*`
  - Middleware: `mw/organization.go` — извлекает `X-Organization-Id` для маршрутизации к нужному S3-пулу
  - Prometheus-метрики на `/api/metrics`, Swagger на `/api/swagger/*`

- **`cmd/config.go`** — конфигурация через Viper/Enviper (TOML + env vars, префикс `WATCHTOWER__`)

### Организация по организациям (multi-tenant)

Проект поддерживает пул S3-подключений по организациям через `StoragePool`. Клиент передаёт `X-Organization-Id` header, middleware извлекает его, и запрос направляется к соответствующему S3-инстансу.

## Команды

```bash
# Сборка
make build                          # go build -o ./bin/watchtower ./cmd/watchtower
go build -v -o ./bin/watchtower ./cmd/watchtower/watchtower.go

# Запуск с dotenv (.env + toml)
make run                            # build + запуск с флагом -d

# Тесты (требуют Docker Compose с MinIO + RabbitMQ + Redis)
make test                           # go test -race ./tests/...
go test -v -count=1 -race ./...     # полный прогон (CI)

# Линтер
golangci-lint run ./...             # конфиг в .golangci.yml

# Docker
docker compose up -d watchtower     # запуск всех зависимостей + сервиса
docker compose -f docker-compose.test.yml up -d   # тестовое окружение
```

### Тесты

Тесты лежат в `tests/` и требуют внешних зависимостей (MinIO, RabbitMQ, Redis). Запускаются через:

```bash
docker compose -f docker-compose.test.yml up -d
sleep 10  # дать время на поднятие сервисов
WATCHTOWER__RUN_MODE=integration go test -v -count=1 -race ./tests/...
```

Тестовое окружение содержит два MinIO-инстанса (основной + внешний для тестирования multi-org), RabbitMQ и Redis. Стартовая конфигурация: `configs/integration.toml`.

## Конфигурация

- Формат: TOML
- Файлы: `configs/development.toml` (по умолчанию), `configs/integration.toml` (для тестов), `configs/production.toml`
- Переменные окружения: префикс `WATCHTOWER__`, разделитель `__` (двойное подчёркивание). Полный маппинг в `cmd/config.go`

## CI/CD

- GitHub Actions: `pull-request.yml` (build → lint → test)
- Drone CI: `.drone/drone.yml` (3 пайплайна: PR, merged, release)
- Dockerfile в корне для сборки образа
