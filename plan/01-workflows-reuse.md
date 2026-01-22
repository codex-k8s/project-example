# 01-workflows-reuse

## Суть задачи
Сделать `codexctl` единым местом для повторяемой CI‑логики и максимально упростить воркфлоу в проектах.

## Где сейчас и нюансы (по результатам анализа)
- `/home/s/projects/project-example/.github/workflows/staging_deploy_main.yml` — ручные ретраи `codexctl apply` и отдельный `kubectl wait` с backoff; логика отличается от Alimentor.
- `/home/s/projects/Alimentor/.github/workflows/staging_deploy_main.yml` — те же шаги, но без ретраев и без унифицированных таймаутов.
- `/home/s/projects/project-example/.github/workflows/ai_dev_issue.yml`, `/home/s/projects/project-example/.github/workflows/ai_plan_issue.yml`, `/home/s/projects/project-example/.github/workflows/ai_plan_review.yml`, `/home/s/projects/project-example/.github/workflows/ai_pr_review.yml`, `/home/s/projects/project-example/.github/workflows/ai_cleanup.yml` — повторяются checkout codexctl, build codexctl, `manage-env ensure-ready`, `prompt run`, `manage-env cleanup`.
- Аналогичная дублирующая структура в `/home/s/projects/Alimentor/.github/workflows/*`.
- В `codexctl` уже есть рабочие примитивы: `/home/s/projects/codexctl/internal/cli/images.go`, `/home/s/projects/codexctl/internal/cli/apply.go`, `/home/s/projects/codexctl/internal/cli/manage_env.go`, `/home/s/projects/codexctl/internal/cli/prompt.go`, `/home/s/projects/codexctl/internal/cli/pr.go`.
- `applyStack` в `/home/s/projects/codexctl/internal/cli/apply.go` содержит ретрай только на admission webhook ingress‑nginx; общие ретраи/таймауты делаются в YAML вручную.
- `kubectl wait` в `/home/s/projects/codexctl/internal/kube/kube.go` не принимает request‑timeout и не имеет backoff‑логики.

## Что меняем (что именно переносим в codexctl)
- В `/home/s/projects/codexctl` добавляем группу `codexctl ci ...`, которая инкапсулирует:
  - подготовку/проверку codexctl и окружения;
  - `images mirror/build` с едиными параметрами;
  - `apply` с ретраями и единым поведением ожидания;
  - `manage-env ensure-ready` и `prompt run` с едиными выводами/форматами;
  - `cleanup/gc` для ai‑слотов.
- В `/home/s/projects/project-example/.github/workflows/*` и `/home/s/projects/Alimentor/.github/workflows/*` заменяем длинные блоки на один‑два вызова `codexctl ci ...`.
- В `codexctl` документируем новый слой `ci` и список поддерживаемых параметров и выходных данных.

## Зачем / ожидаемый эффект
- Воркфлоу становятся короче и стабильнее; логика ретраев и таймаутов централизована.
- Проекты перестают расходиться по поведению деплоя и запуска агентов.
- Изменение CI‑поведения делается в одном месте — `/home/s/projects/codexctl`.
