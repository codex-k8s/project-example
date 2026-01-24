# 03-ai-repair-flow

## Суть задачи
Ввести три режима запуска агента: `ai-dev`, `ai-plan`, `ai-repair`. В первых двух — пытаться поднять инфраструктуру, но при таймауте ожидания всё равно запускать агента с флагом «инфра не готова». В `ai-repair` — запускать агента с отдельным промптом на восстановление.

## Где сейчас и нюансы (по результатам анализа)
- `/home/s/projects/project-example/.github/workflows/ai_dev_issue.yml` и `/home/s/projects/project-example/.github/workflows/ai_plan_issue.yml`:
  - `manage-env ensure-ready` вызывается с `--apply` и `--prepare-images`.
  - При ошибке `ensure-ready` пайплайн падает, агент не стартует.
- Аналогично в `/home/s/projects/Alimentor/.github/workflows/ai_dev_issue.yml` и `/home/s/projects/Alimentor/.github/workflows/ai_plan_issue.yml`.
- `ensure-ready` в `/home/s/projects/codexctl/internal/cli/manage_env.go`:
  - применяет манифесты только для созданных/пересозданных окружений (иначе apply пропускается);
  - не умеет «мягко» переживать ошибки ожидания.
- `applyStack` в `/home/s/projects/codexctl/internal/cli/apply.go`:
  - ждёт `WaitForDeployments` и возвращает ошибку при таймауте;
  - нет режима «wait‑timeout, но продолжить».
- `prompt run` в `/home/s/projects/codexctl/internal/cli/prompt.go`:
  - ждёт готовность `deploy/codex` и завершает работу при ошибке.
- Промпты (`/home/s/projects/codexctl/internal/prompt/templates/*.tmpl`) не знают о состоянии инфраструктуры.

## Что меняем (что именно добавляем)
### 1) Флаг «инфра не готова» для ai-dev/ai-plan
- В `/home/s/projects/codexctl/internal/cli/manage_env.go` добавляем режим «мягкого ожидания»:
  - возвращаем JSON‑признак `infraReady` или `waitFailed` без фатального падения.
  - добавляем флаги вроде `--wait-timeout` и `--wait-soft-fail`.
  - добавляем опцию «force apply» для существующих слотов (иначе `ensure-ready` может ничего не применить).
- В `/home/s/projects/codexctl/internal/cli/prompt.go` добавляем флаг (например `--infra-unhealthy`), который пробрасывает в шаблоны `INFRA_UNHEALTHY=1`.
- В `/home/s/projects/codexctl/internal/prompt/templates/dev_issue_*.tmpl` и `plan_issue_*.tmpl` добавляем условную приписку, если `INFRA_UNHEALTHY=1`:
  - «инфраструктура не поднялась, сначала проверь pod’ы/деплойменты, восстанови сервисы, лкализуй и исправь проблему, убедись что исправления будут в PR».
- В `/home/s/projects/project-example/.github/workflows/ai_dev_issue.yml` и `/home/s/projects/project-example/.github/workflows/ai_plan_issue.yml`:
  - читаем JSON‑выход `ensure-ready` и, если `infraReady=false`, всё равно запускаем `prompt run` с `--infra-unhealthy`.
- Аналогичные изменения в `/home/s/projects/Alimentor/.github/workflows/ai_dev_issue.yml` и `/home/s/projects/Alimentor/.github/workflows/ai_plan_issue.yml`.

### 2) Режим ai-repair
- Добавляем новый workflow (например `/home/s/projects/project-example/.github/workflows/ai_repair_issue.yml` и аналог в Alimentor) с триггером `[ai-repair]`.
- В `ai-repair`:
  - выполняем `ensure-ready` (с коротким `--wait-timeout`, чтобы не блокировать пайплайн);
  - всегда запускаем `prompt run` с отдельным видом промпта `repair_issue`.
- В `/home/s/projects/codexctl/internal/prompt/prompt.go` регистрируем `repair_issue` (и при необходимости `repair_plan`).
- В `/home/s/projects/codexctl/internal/prompt/templates/repair_issue_*.tmpl` добавляем жёсткую инструкцию:
  - разобраться с проблемой инфраструктуры;
  - восстановить сервисы;
  - подготовить и отправить PR с исправлениями.

## Дополнительные технические нюансы
- `ensure-ready` сейчас пропускает apply для «живых» слотов; для repair‑флоу нужен флаг «force apply».
- `prompt run` должен получать флаг состояния инфраструктуры из `--vars`/`--infra-unhealthy`, иначе в шаблонах нет доступа к этому состоянию.
- Для надёжности нужно учитывать, что `deploy/codex` может не стартовать без RBAC (в Alimentor есть `/home/s/projects/Alimentor/deploy/codex/rbac.yaml`, в project-example пока нет аналога).

## Зачем / ожидаемый эффект
- ai‑dev/ai‑plan не падают из‑за неготовой инфраструктуры.
- ai‑repair даёт отдельный сценарий восстановления и фиксирует проблему через PR.
- Агент получает явный контекст о проблеме инфраструктуры.

## Результат

### /home/s/projects/codexctl
- /home/s/projects/codexctl/internal/cli/ensure_helpers.go: добавлены флаги `forceApply`, `waitTimeout`, `waitSoftFail`, и признак `infraReady`; ожидание деплойментов вынесено из `applyStack` в `ensureReady` с мягким фейлом.
- /home/s/projects/codexctl/internal/cli/manage_env.go: новые флаги `--force-apply`, `--wait-timeout`, `--wait-soft-fail`; JSON‑вывод `ensure-ready` включает `infraReady`.
- /home/s/projects/codexctl/internal/cli/ci.go: новые флаги и `infraReady` в JSON для `ci ensure-ready`.
- /home/s/projects/codexctl/internal/cli/prompt.go: добавлен флаг `--infra-unhealthy`, пробрасывающий `INFRA_UNHEALTHY=1` в шаблоны; зарегистрирован новый kind `repair_issue`.
- /home/s/projects/codexctl/internal/prompt/prompt.go: добавлен `KindRepairIssue`.
- /home/s/projects/codexctl/internal/prompt/templates/dev_issue_*.tmpl и plan_issue_*.tmpl: условная приписка при `INFRA_UNHEALTHY=1`.
- /home/s/projects/codexctl/internal/prompt/templates/repair_issue_ru.tmpl и repair_issue_en.tmpl: новые промпты для ai‑repair.

### /home/s/projects/project-example
- /home/s/projects/project-example/.github/workflows/ai_dev_issue.yml и ai_plan_issue.yml: `ci ensure-ready` теперь с `--force-apply --wait-soft-fail --wait-timeout`, JSON‑парсинг `infraReady`, передача `--infra-unhealthy` в `prompt run` при проблемах.
- /home/s/projects/project-example/.github/workflows/ai_repair_issue.yml: новый workflow `[ai-repair]` с коротким ожиданием, запуском `repair_issue` и стандартной цепочкой commit/push/PR‑комментариев.
- /home/s/projects/project-example/docs/architecture_project.md: добавлены режимы ai‑dev/ai‑plan/ai‑repair.

### /home/s/projects/Alimentor
- /home/s/projects/Alimentor/.github/workflows/ai_dev_issue.yml и ai_plan_issue.yml: аналогичные изменения (`infraReady` + `--infra-unhealthy`).
- /home/s/projects/Alimentor/.github/workflows/ai_repair_issue.yml: новый workflow `[ai-repair]`.
- /home/s/projects/Alimentor/docs/architecture_project.md: отмечены AI‑флоу, включая `ai_repair_issue.yml`.
