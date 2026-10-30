# Отчет по домашнему заданию №5: «Собираем приложение»

## 🎯 Цель работы
Спроектировать архитектуру микросервиса на Go, настроить локальную сборку, линтинг, форматирование, контейнеризацию в multi-stage `Dockerfile`, создать автоматический CI/CD-пайплайн с публикацией в контейнерный реестр (GHCR) и автоматическим деплоем на сервер с верификацией работоспособности.

---

## 1. Структура приложения

Микросервис **`order-service`** реализован по принципам Чистой Архитектуры (Clean Architecture) и DDD:
* **`cmd/server/`** — точка входа (`main.go`), инициализация зависимостей, перехват системных прерываний (`SIGINT`, `SIGTERM`) и Graceful Shutdown.
* **`internal/domain/`** — ядро бизнес-логики: Агрегат `Order`, сущности `OrderItem`, перечисления статусов (`CREATED`, `PAID`, `CANCELLED`), инварианты валидации и переходы состояний.
* **`internal/repository/`** — порты и адаптеры хранилища: потокобезопасная in-memory реализация с использованием `sync.RWMutex`.
* **`internal/service/`** — сценарии использования (Usecases): создание заказа, оплата, отмена, выборка с пагинацией.
* **`internal/handler/`** — HTTP-контроллеры: эндпоинты API заказов, а также пробы `/healthz`, `/ready` и `/version`.
* **`internal/config/`** — конфигурация сервиса с чтением переменных окружения и значениями по умолчанию.

---

## 2. Локальная сборка и проверка качества

Для автоматизации локальной разработки подготовлен `Makefile`:
* `make fmt` — автоматическое форматирование кода (`go fmt`).
* `make lint` — запуск линтера `golangci-lint` со строгими правилами (`govet`, `errcheck`, `staticcheck`, `revive`, `ineffassign`, `misspell`).
* `make test` — запуск модульных тестов с флагом `-race` и замером покрытия.
* `make build` — компиляция бинарника с передачей версии, git commit хэша и даты через `-ldflags`.
* `make docker-build` / `make docker-run` — сборка и запуск контейнера.

Конфигурация линтера зафиксирована в файле `.golangci.yml`.

---

## 3. Контейнеризация (Multi-stage Dockerfile)

Dockerfile разделен на два изолированных этапа для минимизации размера и повышения безопасности:
1. **Builder Stage (`golang:1.24-alpine`)**:
   * Кэширование слоев `go.mod` и зависимостей.
   * Статическая компиляция бинарника (`CGO_ENABLED=0 GOOS=linux`) с удалением отладочной информации (`-ldflags="-s -w"`).
2. **Runtime Stage (`alpine:3.21`)**:
   * Создание непривилегированного пользователя и группы `appuser:appgroup` (non-root execution).
   * Установка системных сертификатов `ca-certificates` и часовых поясов `tzdata`.
   * Копирование только скомпилированного бинарного файла.
   * Настройка встроенного `HEALTHCHECK` с опросом эндпоинта `/healthz`.
   * Итоговый размер образа составляет менее 20 MB.

---

## 4. CI/CD Пайплайн (GitHub Actions)

В файле `.github/workflows/ci-cd.yml` настроен сквозной пайплайн, запускающийся при пушах и создании Pull Request:

1. **Этап `lint`**:
   * Проверка исходного кода официальным `golangci/golangci-lint-action@v6`.
2. **Этап `test`**:
   * Запуск тестов с детектором гонок (`go test -v -race -coverprofile=coverage.out ./...`).
3. **Этап `build`**:
   * Проверка сборки исполняемого файла под целевую архитектуру Linux.
4. **Этап `docker-build-push`**:
   * Авторизация в **GitHub Container Registry (`ghcr.io`)** через нативный `${{ secrets.GITHUB_TOKEN }}`.
   * Сборка образа с кэшированием слоев через GitHub Actions Cache (`type=gha`).
   * Публикация образов с тегами `sha-<commit>` и `latest`.
5. **Этап `deploy`**:
   * Автоматический pull опубликованного образа из реестра `ghcr.io`.
   * Запуск контейнера в продакшн-режиме (`docker run -d -p 8080:8080 ...`).
   * Ожидание перехода в статус готовности (Healthcheck `/healthz`).
   * Проверка API: опрос `/version` и создание тестового заказа через `curl`.

---

## 5. Результаты и ссылки

* **Репозиторий проекта:** [github.com/arthur-stone/order-service](https://github.com/arthur-stone/order-service)
* **Ссылка на Pull Request:** [PR #1: CI/CD Pipeline & Application Build](https://github.com/arthur-stone/order-service/pull/1)
* **Контейнер в реестре GHCR:** `ghcr.io/arthur-stone/order-service:latest`

### Скриншоты:
1. *Успешное выполнение всех стадий CI/CD пайплайна в GitHub Actions*:
   `![CI/CD Pipeline](screenshots/ci_cd_pipeline_success.png)`
2. *Локальный/серверный запуск и проверка эндпоинтов `/healthz`, `/version`*:
   `![Server Run](screenshots/server_run_healthz.png)`
