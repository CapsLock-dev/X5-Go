# Order Service API

Сервис заказов на Go 

## Стек
Go 1.22+, chi, slog, swaggo, testify.

## Запуск
```bash
go mod tidy
go run ./cmd/api
```

По умолчанию сервис слушает `:8080`. Адрес переопределяется переменной
окружения `HTTP_ADDR`:

```bash
HTTP_ADDR=:9090 go run ./cmd/api
```

## Тесты
```bash
go test ./...
```

## Генерация Swagger
```bash
swag init -g ./cmd/api/main.go --parseInternal
```

## API
| Метод | Путь | Назначение | Успешный код |
|---|---|---|---:|
| `GET` | `/health` | Проверка доступности | `200` |
| `POST` | `/api/v1/orders` | Создание заказа | `201` |
| `GET` | `/api/v1/orders` | Список заказов | `200` |
| `GET` | `/api/v1/orders/{id}` | Заказ по ID | `200` |
| `PATCH` | `/api/v1/orders/{id}/status` | Смена статуса | `200` |
| `GET` | `/api/v1/orders/{id}/history` | История статусов | `200` |

Swagger UI: http://localhost:8080/swagger/index.html

## Ошибки
| code | HTTP | Когда |
|---|---:|---|
| INVALID_JSON | 400 | Невалидный JSON в Body |
| VALIDATION_ERROR | 400 | Неправильный заказ, айди, статус |
| NOT_FOUND | 404 | Заказ не найден |
| INVALID_STATUS_TRANSITION | 409 | Неправильный переход статус |
| INTERNAL_ERROR | 500 | Внутренняя ошибка |

## Модель данных

### Item — позиция заказа

| Поле | Тип | JSON | Описание |
|---|---|---|---|
| `Name` | `string` | `name` | Название товара, не пустое |
| `Quantity` | `int` | `quantity` | Количество, > 0 |
| `Price` | `int64` | `price` | Цена за единицу в копейках, > 0 |

Цена хранится целым числом в минимальных денежных единицах.

### StatusChange — запись истории

| Поле | Тип | JSON |
|---|---|---|
| `Status` | `Status` | `status` |
| `ChangedAt` | `time.Time` | `changed_at` |

### Order — заказ

| Поле | Тип | JSON | Описание |
|---|---|---|---|
| `ID` | `int64` | `id` | Уникальный, назначается репозиторием |
| `Items` | `[]Item` | `items` | Непустой список позиций |
| `TotalPrice` | `int64` | `total_price` | Рассчитанная сумма |
| `Status` | `Status` | `status` | Текущий статус |
| `CreatedAt` | `time.Time` | `created_at` | Момент создания |
| `History` | `[]StatusChange` | — | История изменения статуса |

## Статусы и переходы

Допустимые статусы:
- `new`
- `processing`
- `completed`
- `cancelled`

Новый заказ автоматически получает `new`.

Разрешённые переходы:
- `new → processing`;
- `new → cancelled`;
- `processing → completed`;
- `processing → cancelled`.

Статусы `completed` и `cancelled` являются конечными.

## Архитектура

```text
HTTP request → handler → service → repository → in-memory storage
```

- `handler` читает HTTP-запрос, валидирует его формат, вызывает сервис и преобразует результат в HTTP-ответ;
- `service` содержит бизнес-логику и не зависит от конкретного хранилища;
- `repository` хранит и возвращает данные;

## Структура
```text
.
├── cmd
│   └── api
│       └── main.go
├── docs
│   ├── docs.go
│   ├── README.md
│   ├── swagger.json
│   └── swagger.yaml
├── go.mod
├── go.sum
├── internal
│   ├── handler
│   │   ├── middleware.go
│   │   └── router.go
│   ├── order
│   │   ├── errors.go
│   │   ├── model.go
│   │   └── repository.go
│   ├── repository
│   │   └── memory
│   │       └── order_repository.go
│   └── service
│       ├── order_service.go
│       └── order_service_test.go
├── Makefile
├── README.md
└── README-task.md
```