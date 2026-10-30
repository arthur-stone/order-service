# Отчет по домашнему заданию №5: «Собираем приложение»

## 🎯 Цель работы
Спроектировать архитектуру микросервиса на Go, настроить локальную сборку, линтинг, форматирование, контейнеризацию в multi-stage `Dockerfile`, создать автоматический CI/CD-пайплайн с публикацией в контейнерный реестр (GHCR) и автоматическим деплоем на сервер с верификацией работоспособности.

---

## 1. Ссылка на Pull Request и репозиторий

* **Репозиторий проекта:** [https://github.com/arthur-stone/order-service](https://github.com/arthur-stone/order-service)
* **Pull Request c результатами:** [PR #1: Microservice Architecture, Dockerization, and Full CI/CD Pipeline](https://github.com/arthur-stone/order-service/pull/1)
* **Опубликованный контейнер в GHCR:** `ghcr.io/arthur-stone/order-service:latest`

---

## 2. Структура приложения

Микросервис **`order-service`** реализован по принципам Чистой Архитектуры (Clean Architecture) и DDD:
* **`cmd/server/main.go`** — точка входа, DI, перехват системных прерываний (`SIGINT`, `SIGTERM`) и Graceful Shutdown.
* **`internal/domain/`** — ядро бизнес-логики: Агрегат `Order`, сущности `OrderItem`, перечисления статусов (`CREATED`, `PAID`, `CANCELED`), инварианты валидации и переходы состояний.
* **`internal/repository/`** — порты и адаптеры хранилища: потокобезопасная in-memory реализация с использованием `sync.RWMutex`.
* **`internal/service/`** — сценарии использования (Usecases): создание заказа, оплата, отмена, выборка с пагинацией.
* **`internal/handler/`** — HTTP-контроллеры: эндпоинты API заказов, а также пробы `/healthz`, `/ready` и `/version`.
* **`internal/config/`** — конфигурация сервиса с чтением переменных окружения и значениями по умолчанию.

---

## 3. Локальная сборка и качество кода

Для автоматизации локальной разработки подготовлен `Makefile`:
* `make fmt` — автоматическое форматирование кода (`go fmt`).
* `make lint` — запуск линтера `golangci-lint` со строгими правилами (`govet`, `errcheck`, `staticcheck`, `revive`, `ineffassign`, `misspell`).
* `make test` — запуск модульных тестов с флагом `-race` и замером покрытия.
* `make build` — компиляция бинарника с передачей версии, git commit хэша и даты через `-ldflags`.
* `make docker-build` / `make docker-run` — сборка и запуск контейнера.

Конфигурация линтера зафиксирована в файле `.golangci.yml`.

---

## 4. Контейнеризация (Multi-stage Dockerfile)

Dockerfile разделен на два изолированных этапа для минимизации размера и повышения безопасности:
1. **Builder Stage (`golang:1.24-alpine`)**:
   * Кэширование слоев `go.mod` и зависимостей.
   * Статическая компиляция бинарника (`CGO_ENABLED=0 GOOS=linux`) с удалением отладочной информации (`-ldflags="-s -w"`).
2. **Runtime Stage (`alpine:3.21`)**:
   * Создание непривилегированного пользователя и группы `appuser:appgroup` (non-root execution).
   * Установка системных сертификатов `ca-certificates` и часовых поясов `tzdata`.
   * Копирование скомпилированного бинарного файла.
   * Настройка встроенного `HEALTHCHECK` с опросом эндпоинта `/healthz`.
   * Итоговый размер образа составляет менее 20 MB.

---

## 5. CI/CD Пайплайн (GitHub Actions + Self-Hosted Runner)

В файле `.github/workflows/ci-cd.yml` настроен сквозной пайплайн, запускающийся при пушах и создании Pull Request:

1. **Этап `lint` (`Code Linting`)**:
   * Проверка исходного кода линтером `golangci-lint run ./...` — пройден без единой ошибки и предупреждения.
2. **Этап `test` (`Unit & Domain Tests`)**:
   * Запуск тестов с детектором гонок (`go test -v -race -coverprofile=coverage.out ./...`) — 100% тестов пройдены.
3. **Этап `build` (`Compile Binary`)**:
   * Проверка кросс-компиляции бинарника под Linux с внедрением версии и commit sha.
4. **Этап `docker-build-push` (`Build & Push Docker Image`)**:
   * Авторизация в **GitHub Container Registry (`ghcr.io`)** через `${{ secrets.GITHUB_TOKEN }}`.
   * Сборка Docker-образа и публикация тегов `sha-<commit>` и `latest` в реестр.
5. **Этап `deploy` (`Deploy & Healthcheck Verification`)**:
   * Деплой на целевой сервер (через Self-Hosted Runner):
   * Авторизация в GHCR, `docker pull` опубликованного образа;
   * Запуск контейнера в фоновом режиме: `docker run -d --name order-service-prod -p 8080:8080`;
   * Ожидание готовности по `/healthz`;
   * Проверка эндпоинтов `/version`, `/healthz`, создание тестового заказа через `curl`;
   * Корректный cleanup контейнера после верификации.

---

## 6. Результаты выполнения пайплайна (Лог прогона)

```text
✓ Code Linting in 14s (ID 124302511449)
✓ Unit & Domain Tests in 13s (ID 124302511723)
✓ Compile Binary in 12s (ID 124303047616)
✓ Build & Push Docker Image in 18s (ID 124303287451)
✓ Deploy & Healthcheck Verification in 12s (ID 124303607546)
```

### Лог верификации на сервере деплоя:
```text
Deploying image: ghcr.io/arthur-stone/order-service:sha-74fd07f
Status: Image is up to date for ghcr.io/arthur-stone/order-service:sha-74fd07f
8c2693793f34ecb6bc255f3cfefb08523a9ac074875ee96be1c525503f8a6f27

Waiting for service to be healthy...
{"status":"UP"}
Service is healthy!

=== 1. Health Probe ===
{"status":"UP"}

=== 2. Version Info ===
{"version":"1.0.0","commit":"dc84256","build_date":"2026-10-30T20:45:00Z"}

=== 3. Create Sample Order ===
{"id":"ord_5cb893593466b0b0","customer_id":"cust_test_01","items":[{"id":"item_81f2a27ae98f2231","product_id":"item_1","title":"Go Book","quantity":2,"price":49.99}],"total_amount":99.98,"status":"CREATED","created_at":"2026-10-30T20:45:52.878785365Z","updated_at":"2026-10-30T20:45:52.878785365Z"}

=== 4. Check Container Status ===
CONTAINER ID   IMAGE                                            COMMAND                CREATED                  STATUS                                     PORTS                                         NAMES
8c2693793f34   ghcr.io/arthur-stone/order-service:sha-74fd07f   "/app/order-service"   Less than a second ago   Up Less than a second (health: starting)   0.0.0.0:8080->8080/tcp, [::]:8080->8080/tcp   order-service-prod
```
