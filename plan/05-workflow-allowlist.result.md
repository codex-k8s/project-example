# 05-workflow-allowlist — результат

## Что сделано

### /home/s/projects/project-example
- /home/s/projects/project-example/.github/workflows/ai_dev_issue.yml: добавлен allowlist через `AI_ALLOWED_USERS` и блокировка запуска от `CODEX_GH_USERNAME`.
- /home/s/projects/project-example/.github/workflows/ai_plan_issue.yml: добавлен allowlist и блокировка запуска от бота.
- /home/s/projects/project-example/.github/workflows/ai_repair_issue.yml: добавлен allowlist и блокировка запуска от бота.
- /home/s/projects/project-example/.github/workflows/ai_plan_review.yml: allowlist + блокировка запуска от бота на уровне job; снят шаг‑guard.
- /home/s/projects/project-example/.github/workflows/ai_pr_review.yml: allowlist + блокировка запуска от бота на уровне job.
- /home/s/projects/project-example/README_RU.md: `AI_ALLOWED_USERS` и `CODEX_GH_USERNAME` перенесены в список Repository Variables, `CODEX_GH_USERNAME` убран из Secrets.

### /home/s/projects/Alimentor
- /home/s/projects/Alimentor/.github/workflows/ai_dev_issue.yml: добавлен allowlist и блокировка запуска от бота.
- /home/s/projects/Alimentor/.github/workflows/ai_plan_issue.yml: добавлен allowlist и блокировка запуска от бота.
- /home/s/projects/Alimentor/.github/workflows/ai_repair_issue.yml: добавлен allowlist и блокировка запуска от бота.
- /home/s/projects/Alimentor/.github/workflows/ai_plan_review.yml: allowlist + блокировка запуска от бота на уровне job; снят шаг‑guard.
- /home/s/projects/Alimentor/.github/workflows/ai_pr_review.yml: allowlist + блокировка запуска от бота на уровне job.

## Для чего
- AI‑воркфлоу запускаются только для заранее разрешённых пользователей.
- Исключены самозапуски от бота (`CODEX_GH_USERNAME`).
