# 05-workflow-allowlist

## Суть задачи
Ограничить запуск AI‑воркфлоу списком разрешенных пользователей и исключить автотриггеры от бота.

## Где сейчас и нюансы (по результатам анализа)
- В `/home/s/projects/project-example/.github/workflows/ai_plan_review.yml` есть проверка `github.actor != CODEX_GH_USERNAME`, но она не покрывает остальные воркфлоу.
- Остальные AI‑воркфлоу (`ai_dev_issue.yml`, `ai_plan_issue.yml`, `ai_pr_review.yml`) запускаются без проверки автора.
- В `/home/s/projects/Alimentor/.github/workflows/*` аналогичная ситуация.

## Что меняем (что именно добавляем)
- В обоих репозиториях добавляем переменную `AI_ALLOWED_USERS` (список GitHub‑логинов).
- В воркфлоу добавляем единый guard‑блок, который:
  - блокирует запуск, если автор — `CODEX_GH_USERNAME`;
  - разрешает запуск только если автор входит в `AI_ALLOWED_USERS`.
- Изменения в файлах:
  - `/home/s/projects/project-example/.github/workflows/ai_dev_issue.yml`
  - `/home/s/projects/project-example/.github/workflows/ai_plan_issue.yml`
  - `/home/s/projects/project-example/.github/workflows/ai_plan_review.yml`
  - `/home/s/projects/project-example/.github/workflows/ai_pr_review.yml`
  - `/home/s/projects/project-example/.github/workflows/ai_repair_issue.yml` (новый, если добавится)
  - те же файлы в `/home/s/projects/Alimentor/.github/workflows/*`.
  - `/home/s/projects/project-example/README_RU.md`

## Зачем / ожидаемый эффект
- Запуски AI‑тасков контролируются и предсказуемы.
- Бот не триггерит сам себя.
