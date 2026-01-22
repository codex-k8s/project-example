# 03-ai-repair-flow

## Суть задачи
Добавить repair‑режим, чтобы агент запускался даже при сбое развёртывания и мог восстановить инфраструктуру.

## Где сейчас и нюансы (по результатам анализа)
- `prompt run` в `/home/s/projects/codexctl/internal/cli/prompt.go` ждёт `rollout status deploy/codex`. Если codex‑деплой не поднялся, агент не стартует.
- `manage-env ensure-ready` в `/home/s/projects/codexctl/internal/cli/manage_env.go` применяет манифесты только для новых/пересозданных слотов; при «сломавшемся» окружении без пересоздания apply пропускается.
- `applyStack` в `/home/s/projects/codexctl/internal/cli/apply.go` рендерит весь стек через `/home/s/projects/codexctl/internal/engine/engine.go`; фильтрации по инфраструктуре нет.
- `codex`‑деплой в `/home/s/projects/project-example/deploy/codex/codex-deploy.yaml` требует `serviceAccountName: codex-sa`, но в проекте нет отдельного RBAC‑манифеста (в Alimentor он есть: `/home/s/projects/Alimentor/deploy/codex/rbac.yaml`).
- `codex` зависит от секретов/ConfigMap из `/home/s/projects/project-example/deploy/secret.yaml` и `/home/s/projects/project-example/deploy/configmap.yaml` (группа `namespace-and-config` в `/home/s/projects/project-example/services.yaml`).

## Что меняем (что именно добавляем)
- В `/home/s/projects/codexctl/internal/prompt/templates/*` добавляем новые шаблоны `repair_issue`/`repair_plan` с инструкцией диагностики и восстановления.
- В `/home/s/projects/codexctl/internal/prompt/prompt.go` регистрируем новые виды промптов.
- В `/home/s/projects/codexctl/internal/cli/manage_env.go` вводим режим `ensure-repair`, который применяет только минимальный набор манифестов (namespace/config/secrets + codex).
- В `/home/s/projects/codexctl/internal/engine/engine.go` или в модели инфраструктуры добавляем возможность рендерить/применять только часть инфраструктуры (нужна фильтрация по `infrastructure.name`).
- В `/home/s/projects/project-example` добавляем RBAC для codex (аналог `/home/s/projects/Alimentor/deploy/codex/rbac.yaml`) или корректируем `codex-deploy.yaml`, чтобы сервис‑аккаунт существовал в repair‑режиме.
- В `/home/s/projects/project-example/.github/workflows/ai_*.yml` и `/home/s/projects/Alimentor/.github/workflows/ai_*.yml` добавляем:
  - fallback на repair‑промпт при неуспешном ensure/apply;
  - отдельный workflow на метку `[ai-repair]`.

## Зачем / ожидаемый эффект
- Агент запускается даже при проблемах развёртывания и может их исправить.
- Ремонтный запуск становится стандартным сценариям (ai‑repair и auto‑fallback).
- Меньше ручной диагностики и простоя.
