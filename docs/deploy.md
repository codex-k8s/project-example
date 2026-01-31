# Деплой и окружения в project-example

## `services.yaml` и окружения

Файл `services.yaml` является единственным источником правды
для `codexctl` и GitHub Actions:

- `project: project-example`;
- `baseDomain` — домены для `dev`, `ai-staging`, `ai`;
- `environments` — политика загрузки образов и настройки окружений;
- `images` — сборка образов Django/Go/Vue и Codex;
- `infrastructure` — группы манифестов (namespace, PostgreSQL, Redis, Jaeger и т.д.);
- `services` — прикладные сервисы и Codex‑Pod.

### Окружения

- `dev` — локальное окружение разработчика (namespace `project-example-dev`);
- `ai-staging` — общее стейдж‑окружение (namespace `project-example-ai-staging`);
- `ai` — dev‑AI слоты (`project-example-dev-<slot>`), в которых работает Codex.

Подключение к кластеру выполняется in‑cluster через service account runner’а.

## Базовой цикл деплоя (локально)

Для любого окружения (`dev`, `ai-staging`, `ai`) базовый цикл одинаков:

```bash
ENV=ai-staging
NAMESPACE=project-example-ai-staging

CODEXCTL_REGISTRY_HOST="registry.${NAMESPACE}.svc.cluster.local:5000" codexctl images mirror --env "$ENV"
CODEXCTL_REGISTRY_HOST="registry.${NAMESPACE}.svc.cluster.local:5000" codexctl images build  --env "$ENV"

codexctl apply --env "$ENV" --wait --preflight
```

Команда `apply`:

- рендерит все манифесты с учётом шаблонов и переменных окружения;
- применяет их в Kubernetes через `kubectl apply`;
- ждёт готовности деплойментов (`--wait`);
- выполняет возможные hook’и (`hooks.beforeAll`, `infrastructure.*.hooks`, `services.*.hooks`).

## Инфраструктура

Блок `infrastructure` в `services.yaml`:

- `namespace-and-config` — namespace + DNS‑config + общие ConfigMap/Secret;
- `tls-issuer` — `deploy/cluster-issuer.yaml`;
- `data-services` — PostgreSQL и Redis (`deploy/postgres.service.yaml`, `deploy/redis.service.yaml`);
- `observability` — Jaeger (`deploy/jaeger.yaml`).

Хранение данных выполняется в PVC, параметры задаются в `services.yaml`
через блок `storage` и env‑переменные `CODEXCTL_STORAGE_CLASS_*`.

## Сервисы

В `services.yaml` описаны:

- `django-backend` — деплой `services/django_backend/deploy.yaml`,
  сборка образа из `services/django_backend/Dockerfile`;
- `chat-backend` — деплой `services/chat_backend/deploy.yaml`,
  образ `services/chat_backend/Dockerfile`;
- `web-frontend` — деплой `services/web_frontend/deploy.yaml`,
  образ `services/web_frontend/Dockerfile`;
- `codex` — Pod агента Codex (`deploy/codex/codex-deploy.yaml`, `deploy/codex/ingress-dev.yaml`),
  разворачивается только в `env=ai`.

Для `dev`/`ai` окружений используются `pvcMounts`, которые монтируют исходники
из workspace PVC в контейнеры (см. `services.*.overlays`). Это позволяет Codex‑агенту
и разработчикам работать с живыми исходниками.

## GitHub Actions

### AI Staging

`.github/workflows/ai_staging_deploy.yml`:

- триггер: push в `main`;
- шаги:
  - checkout репозитория и `codexctl`;
  - сборка `codexctl`;
  - `codexctl ci images --env ai-staging --mirror --build` через Kaniko в кластерный registry;
  - `codexctl ci apply --env ai-staging --wait --preflight`.

### Dev‑AI слоты и агенты

Основные файлы:

- `ai_plan_issue.yml` — планирование задач по метке `[ai-plan]`;
- `ai_plan_review.yml` — уточнение/продолжение планирования по комментарию `[ai-plan]`;
- `ai_dev_issue.yml` — разработка по метке `[ai-dev]`:
  выделение слота, развёртывание, запуск агента, ветка `codex/issue-*`;
- `ai_pr_review.yml` — auto‑fix по ревью PR (state `changes_requested`);
- `ai_cleanup.yml` — уборка слотов при закрытии Issue/PR.

Все эти воркфлоу вызывают:

- `codexctl ci ensure-slot/ensure-ready` для управления слотами;
- `codexctl prompt run` для запуска агента с нужным `kind`;
- при необходимости — `codexctl pr review-apply` для переноса правок в PR.
