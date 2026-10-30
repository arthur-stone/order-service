# 📦 Order Service (`order-service`)

[![CI/CD Pipeline](https://github.com/arthur-stone/order-service/actions/workflows/ci-cd.yml/badge.svg)](https://github.com/arthur-stone/order-service/actions/workflows/ci-cd.yml)

Легковесный, production-ready микросервис обработки и управления заказами на языке **Go**, спроектированный по принципам **Чистой Архитектуры** и **Domain-Driven Design (DDD)**. 

Сервис выступает базовым ядром для сквозных практических заданий курса по микросервисам (CI/CD, Docker, gRPC, OpenTelemetry, Redis, Kafka, Saga).

---

## 🏗 Архитектура и структура проекта

Проект организован по стандарту `Standard Go Project Layout`:

```
order-service/
├── cmd/
│   └── server/
│       └── main.go              # Точка входа, DI, перехват OS-сигналов (Graceful Shutdown)
├── internal/
│   ├── app/
│   │   └── app.go               # Инициализация HTTP-сервера, маршрутов и логгера
│   ├── config/
│   │   └── config.go            # Чтение конфигурации из переменных окружения
│   ├── domain/                  # Ядро DDD (не зависит от внешних библиотек)
│   │   ├── order.go             # Агрегат Order, Entity OrderItem, инварианты и методы
│   │   └── order_test.go        # Юнит-тесты бизнес-правил и переходов статусов
│   ├── handler/                 # Delivery-слой (HTTP хэндлеры)
│   │   ├── handler.go           # Вспомогательные функции сериализации JSON и ошибок
│   │   ├── health.go            # Пробы /healthz, /ready, /version
│   │   ├── order.go             # REST API для управления заказами
│   │   └── order_test.go        # HTTP-тесты эндпоинтов
│   ├── repository/              # Слой адаптеров хранилища (Ports & Adapters)
│   │   ├── repository.go        # Интерфейс OrderRepository
│   │   └── memory.go            # Потокобезопасная in-memory реализация (sync.RWMutex)
│   └── service/                 # Usecase / Слой бизнес-логики приложения
│       ├── service.go           # Сервис заказов (CRUD, оплата, отмена)
│       └── service_test.go      # Тесты слоя сценариев использования
├── docs/
│   └── hw05_report.md           # Отчет о выполнении домашнего задания №5
├── .github/
│   └── workflows/
│       └── ci-cd.yml            # Полный GitHub Actions пайплайн (Lint, Test, Build, Docker, Deploy)
├── .dockerignore
├── .golangci.yml                # Строгие правила линтинга
├── Dockerfile                   # Оптимизированный multi-stage образ (Alpine, non-root)
├── docker-compose.yml           # Локальный запуск сервиса в контейнере
├── Makefile                     # Автоматизация типовых команд разработчика
├── go.mod
└── README.md
```

---

## 🚀 Быстрый старт (Локально)

### Требования
* Go >= 1.24 (или 1.26)
* `golangci-lint` (для линтинга)
* Docker / Docker Compose (для контейнеризации)

### 1. Форматирование и проверка кода
```bash
# Форматирование исходного кода
make fmt

# Запуск линтеров
make lint
```

### 2. Запуск тестов
```bash
# Прогон модульных тестов с детектором гонок (-race) и замером покрытия
make test
```

### 3. Сборка бинарника
```bash
make build
# Бинарник будет скомпилирован в bin/order-service с внедрением версии, commit и даты сборки
```

### 4. Локальный запуск
```bash
make run
# Или запуск собранного бинарника:
./bin/order-service
```

---

## 🐳 Сборка и запуск в Docker

### Сборка через Makefile:
```bash
make docker-build
```

### Запуск контейнера:
```bash
make docker-run
```

Или через `docker compose`:
```bash
docker compose up -d
docker compose logs -f
```

---

## 📡 API Эндпоинты

| Метод | Путь | Описание |
|---|---|---|
| `GET` | `/healthz` | Liveness-проба для Kubernetes / Docker Healthcheck |
| `GET` | `/ready` | Readiness-проба готовности обрабатывать трафик |
| `GET` | `/version` | Информация о версии, Git-хэше и дате сборки |
| `POST` | `/api/v1/orders` | Создание нового заказа |
| `GET` | `/api/v1/orders/{id}` | Получение информации о заказе по ID |
| `GET` | `/api/v1/orders` | Пагинированный список заказов (`?limit=10&offset=0`) |
| `POST` | `/api/v1/orders/{id}/pay` | Оплата заказа (`CREATED -> PAID`) |
| `POST` | `/api/v1/orders/{id}/cancel` | Отмена заказа (`CREATED -> CANCELLED`) |

### Примеры запросов:

**Проверка здоровья:**
```bash
curl -s http://localhost:8080/healthz
# {"status":"UP"}
```

**Создание заказа:**
```bash
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "cust_101",
    "items": [
      {"product_id": "laptop_pro", "title": "MacBook Pro", "quantity": 1, "price": 1999.99},
      {"product_id": "mouse_m1", "title": "Wireless Mouse", "quantity": 2, "price": 49.50}
    ]
  }'
```

---

## 🔄 CI/CD Пайплайн

В проекте настроен автоматический пайплайн GitHub Actions (`.github/workflows/ci-cd.yml`):
1. **`lint`**: Статический анализ кода с помощью `golangci-lint`.
2. **`test`**: Выполнение всех тестов с детекцией гонок (`-race`) и отчетом о покрытии.
3. **`build`**: Кросс-компиляция статически слинкованного Linux-бинарника.
4. **`docker-build-push`**: Multi-stage сборка Docker-образа и публикация в **GitHub Container Registry (GHCR)**: `ghcr.io/arthur-stone/order-service`.
5. **`deploy`**: Автоматический pull опубликованного контейнера из GHCR, запуск изолированного контейнера, проверка `/healthz`, `/version` и вызов тестового API.

Подробный отчет и скриншоты представлены в [docs/hw05_report.md](docs/hw05_report.md).
