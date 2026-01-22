# 01-workflows-reuse

## Суть задачи
Сделать `/home/s/projects/codexctl` единым местом для повторяемой CI‑логики и вынести туда повторяющиеся shell‑хуки из `services.yaml`.

## Где сейчас и нюансы (по результатам анализа)
- `/home/s/projects/project-example/.github/workflows/staging_deploy_main.yml` — есть ручные ретраи `codexctl apply` и отдельный `kubectl wait` с backoff.
- `/home/s/projects/Alimentor/.github/workflows/staging_deploy_main.yml` — похожий пайплайн, но без ретраев и с другой логикой ожидания.
- `/home/s/projects/project-example/.github/workflows/ai_*` и `/home/s/projects/Alimentor/.github/workflows/ai_*` — повторы checkout/build codexctl, `manage-env ensure-ready`, `prompt run`, `cleanup`.
- В `/home/s/projects/codexctl/internal/hooks/hooks.go` есть ограниченный набор встроенных хуков (`kubectl.wait`, `github.comment`, `sleep`, `preflight`), но нет доменных хуков из `services.yaml`.
- В `/home/s/projects/Alimentor/services.yaml` есть shell‑хуки:
  - `ensure-local-data-dirs` (beforeAll) — создание каталогов в `DATA_ROOT`.
  - `ensure-codex-secrets` (codex.beforeApply) — создание `openai-secret`, `github-secret`, `context7-secret`.
  - `check-dev-host-ports` (codex.afterApply) — проверка локальных портов 80/443 и ingress.
  - `reuse-dev-tls-secret` (codex.afterApply) — копирование TLS‑секрета между namespace.
- В `/home/s/projects/project-example/services.yaml` есть `ensure-local-data-dirs` и `reuse-dev-tls-secret`, но они реализованы через `run: |`.

## Что меняем (что именно переносим в codexctl)
### 1) CI‑логика
- В `/home/s/projects/codexctl` добавляем группу `codexctl ci ...`, которая инкапсулирует:
  - images mirror/build;
  - apply + wait с едиными ретраями/таймаутами;
  - manage-env ensure‑ready + prompt run;
  - cleanup/gc.
- В `/home/s/projects/project-example/.github/workflows/*` и `/home/s/projects/Alimentor/.github/workflows/*` заменяем длинные скрипты на короткие вызовы `codexctl ci ...`.

### 2) Доменные хуки из services.yaml
- В `/home/s/projects/codexctl/internal/hooks/hooks.go` добавляем встроенные хуки, которые заменят shell‑скрипты:
  - `codex.ensure-data-dirs` — аналог `ensure-local-data-dirs` (создание каталогов для postgres/redis/rabbitmq в `DATA_ROOT`, с учетом ai‑слотов).
  - `codex.ensure-codex-secrets` — аналог `ensure-codex-secrets` (создание `openai-secret`, `github-secret`, `context7-secret`, если значения есть в ENV).
  - `codex.check-dev-host-ports` — аналог `check-dev-host-ports` (best‑effort проверки портов 80/443 и HTTP‑probe).
  - `codex.reuse-dev-tls-secret` — аналог `reuse-dev-tls-secret` (копирование и ожидание TLS‑секрета между namespace).
- В `/home/s/projects/Alimentor/services.yaml` и `/home/s/projects/project-example/services.yaml` заменяем `run: |` на `use: codex.*` и передаём параметры через `with:`.

## Где именно менять конфигурации
- `/home/s/projects/Alimentor/services.yaml`:
  - `hooks.beforeAll.ensure-local-data-dirs` -> `use: codex.ensure-data-dirs`.
  - `services.codex.hooks.beforeApply.ensure-codex-secrets` -> `use: codex.ensure-codex-secrets`.
  - `services.codex.hooks.afterApply.check-dev-host-ports` -> `use: codex.check-dev-host-ports`.
  - `services.codex.hooks.afterApply.reuse-dev-tls-secret` -> `use: codex.reuse-dev-tls-secret`.
- `/home/s/projects/project-example/services.yaml`:
  - `hooks.beforeAll.ensure-local-data-dirs` -> `use: codex.ensure-data-dirs`.
  - `services.codex.hooks.afterApply.reuse-dev-tls-secret` -> `use: codex.reuse-dev-tls-secret`.

## Зачем / ожидаемый эффект
- Снижение дублирования логики и расхождений между репозиториями.
- Более стабильные пайплайны (единые ретраи/таймауты).
- Упрощение `services.yaml` за счёт встроенных хуков codexctl.
