# Go‑сервисы в project-example

В текущей версии проекта реализован один Go‑сервис:
`services/chat_backend` — HTTP API общего чата.

## `chat_backend`

Назначение:

- регистрация и аутентификация пользователей;
- выдача сессионных токенов (Redis);
- приём и выдача сообщений общего чата.

Основные зависимости:

- `github.com/jackc/pgx/v5/pgxpool` — подключение к PostgreSQL;
- `github.com/redis/go-redis/v9` — клиент Redis;
- `golang.org/x/crypto/bcrypt` — хеширование паролей.

### Архитектура

Файл `services/chat_backend/main.go` содержит:

- инициализацию подключения к Postgres/Redis (`initPostgres`, `initRedis`);
- структуру `server` с полями `db`, `redis`;
- HTTP‑маршруты:
  - `GET /health/livez`, `GET /health/readyz`;
  - `POST /api/register`;
  - `POST /api/login`;
  - `GET /api/messages`;
  - `POST /api/messages`.

Модели данных в коде не дублируют Django‑модели,
а работают напрямую с таблицами:

- `chat_user` (`id`, `nickname`, `password_hash`, `created_at`);
- `chat_message` (`id`, `user_id`, `text`, `created_at`).

### Авторизация и сессии

- при логине:
  - сервис проверяет хеш пароля;
  - создаёт токен (случайный hex‑строка);
  - записывает `session:<token> -> user_id` в Redis с TTL 24 часа;
- при отправке сообщения:
  - токен читается из заголовка `Authorization: Bearer <token>`;
  - сервис извлекает `user_id` из Redis и проверяет наличие пользователя;
  - сообщение записывается в БД.

### Наблюдаемость

- health‑эндпоинты:
  - `/health/livez` — быстрая проверка “жив ли процесс”;
  - `/health/readyz` — дополнительная проверка доступности Postgres и Redis;
- логи пишутся в stdout через стандартный `log` пакет
  (сообщения только на английском).

### Деплой

Манифест: `services/chat_backend/deploy.yaml`:

- `Deployment`:
  - init‑контейнеры ждут доступности Postgres и Redis;
  - основной контейнер `chat-backend` слушает порт `8080`;
  - probes `/health/livez` и `/health/readyz`;
  - переменные окружения берутся из `db-secret`, `redis-secret`,
    `project-example-config`, `db-config`, `redis-config`;
- `Service` `chat-backend` (порт 8080);
- `Ingress` `chat-backend-ingress`:
  - маршрут `/api` на сервис `chat-backend`.

