# Архитектура проекта project-example

## Технологический стек

- **Backend (админка и схема БД)**: Django (`services/django_backend`);
- **Backend (основная логика)**: Go‑сервис `chat_backend`;
- **Frontend**: SPA на Vue3 + Pinia (`services/web_frontend`);
- **База данных**: PostgreSQL;
- **Кэш и сессии**: Redis;
- **Наблюдаемость**:
  - Jaeger (in‑memory) для приёма OTLP‑трейсов;
  - health‑эндпоинты и логирование в stdout;
- **Оркестрация**: Kubernetes + `codexctl` + GitHub Actions.

## Основные компоненты

### Django backend (`services/django_backend`)

- Отвечает за схему БД и админку.
- Модели:
  - `ChatUser` — пользователь чата (ник, хеш пароля, время регистрации);
  - `Message` — сообщение (автор, текст, время создания).
- Используется только для административного управления и миграций;
  рабочее HTTP‑API чата реализовано в Go‑сервисе.

### Go chat backend (`services/chat_backend`)

- HTTP‑API для простого общего чата:
  - `POST /api/register` — регистрация по нику и паролю;
  - `POST /api/login` — получение сессионного токена (Redis);
  - `POST /api/messages` — отправка сообщения (с токеном);
  - `GET /api/messages` — получение последних сообщений.
- Работает с таблицами `chat_user` и `chat_message`
  (создаются Django‑миграцией).
- Авторизация по заголовку `Authorization: Bearer <token>`.

### Web frontend (`services/web_frontend`)

- Vite + Vue3 + Pinia.
- Регистрация, вход и интерфейс общего чата.
- Периодически опрашивает `GET /api/messages` и отправляет
  сообщения через `POST /api/messages`.

### Инфраструктура (`deploy/`)

- `namespace.yaml` — создание namespace по `{{ .Namespace }}`;
- `configmap.yaml` — общие переменные окружения (HOST, STAGE),
  конфиг PostgreSQL и Redis;
- `secret.yaml` — секреты БД, Redis, Django, OpenAI, Context7, GitHub;
- `postgres.service.yaml` — Deployment/Service для PostgreSQL
  с хранением данных на hostPath (`.data/.../postgres`);
- `redis.service.yaml` — Deployment/Service для Redis
  (`.data/.../redis`);
- `jaeger.yaml` — in‑memory Jaeger (ConfigMap + Deployment + Service);
- `dns.configmap.yaml` — конфигурация CoreDNS для кластера;
- `cluster-issuer.yaml` — ClusterIssuer для Let’s Encrypt;
- `echo-probe.yaml` — временный echo‑сервис для проверки доступности домена перед выпуском сертификата;
- `ingress-nginx.controller.yaml` — ingress‑контроллер;
- `codex/*` — Pod Codex, ingress для dev‑слотов (`dev-<slot>.baseDomain.ai`) и RBAC для service account `codex-sa` (включая ai-staging‑repair).

### Codex и dev‑AI слоты

`services.yaml` описывает:

- `project: project-example`;
- `baseDomain.dev/ai-staging/ai` — домены;
- `environments.dev/ai-staging/ai` — kubeconfig и registry;
- `images` — сборку образов сервисов и Codex;
- `infrastructure` — группы манифестов (`namespace-and-config`, `data-services`, `observability` и т.д.);
- `services` — приложения (`django-backend`, `chat-backend`, `web-frontend`, `codex`).

Dev‑AI‑слоты (`env=ai`) создаются через `codexctl ci ensure-slot/ensure-ready`,
а метаданные/очистка выполняются через `codexctl manage-env` (set/comment/cleanup)
из GitHub Actions — см. `.github/workflows/ai_*.yml`.
Режимы:
- `[ai-dev]` — обычная разработка агентом;
- `[ai-plan]` — планирование задач агентом;
- `[ai-repair]` — восстановление стейджинга агентом.

## Структура директорий (основное)

```text
.
├── AGENTS.md                  // правила для Codex‑агента
├── README.md                  // краткое описание
├── README_RU.md               // полный гайд по установке
├── services.yaml              // конфигурация codexctl
├── deploy/                    // инфраструктурные манифесты K8s
│   ├── namespace.yaml
│   ├── configmap.yaml
│   ├── secret.yaml
│   ├── postgres.service.yaml
│   ├── redis.service.yaml
│   ├── jaeger.yaml
│   ├── dns.configmap.yaml
│   ├── cluster-issuer.yaml
│   ├── echo-probe.yaml
│   ├── ingress-nginx.controller.yaml
│   └── codex/
│       ├── Dockerfile
│       ├── rbac.yaml
│       ├── rbac-ai-repair.yaml
│       ├── codex-deploy.yaml
│       └── ingress-dev.yaml
├── services/
│   ├── django_backend/        // Django‑админка и миграции
│   ├── chat_backend/          // Go‑сервис HTTP API чата
│   └── web_frontend/          // фронтенд Vue3+Pinia
├── .github/
│   └── workflows/             // CI/CD и AI‑воркфлоу
└── docs/
    ├── architecture_project.md
    ├── deploy.md
    ├── libs.md
    ├── models.md
    ├── migrations_and_fixtures.md
    ├── go_services.md
    └── observability.md
```
