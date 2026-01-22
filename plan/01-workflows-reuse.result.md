Сделано
- В /home/s/projects/codexctl добавлен CI‑слой:
  - /home/s/projects/codexctl/internal/cli/ci.go — новые команды `codexctl ci images`, `codexctl ci apply`, `codexctl ci ensure-slot`, `codexctl ci ensure-ready`.
  - /home/s/projects/codexctl/internal/cli/ensure_helpers.go — общий код для ensure‑slot/ensure‑ready.
  - /home/s/projects/codexctl/internal/cli/manage_env.go — переиспользует общий код, добавлена очистка/удаление dataPaths.
  - /home/s/projects/codexctl/internal/cli/data_paths.go — безопасная обработка create/clean/delete для dataPaths.
- В /home/s/projects/codexctl добавлена поддержка dataPaths и новых хуков:
  - /home/s/projects/codexctl/internal/config/config.go и /home/s/projects/codexctl/internal/config/data_paths.go — новый блок dataPaths.
  - /home/s/projects/codexctl/internal/hooks/hooks.go — хуки `codex.ensure-data-dirs`, `codex.ensure-codex-secrets`, `codex.check-dev-host-ports`, `codex.reuse-dev-tls-secret`.
- Обновлены workflow’ы, чтобы использовать новые команды:
  - /home/s/projects/project-example/.github/workflows/staging_deploy_main.yml
  - /home/s/projects/project-example/.github/workflows/ai_dev_issue.yml
  - /home/s/projects/project-example/.github/workflows/ai_plan_issue.yml
  - /home/s/projects/project-example/.github/workflows/ai_plan_review.yml
  - /home/s/projects/project-example/.github/workflows/ai_pr_review.yml
  - /home/s/projects/Alimentor/.github/workflows/staging_deploy_main.yml
  - /home/s/projects/Alimentor/.github/workflows/ai_dev_issue.yml
  - /home/s/projects/Alimentor/.github/workflows/ai_plan_issue.yml
  - /home/s/projects/Alimentor/.github/workflows/ai_plan_review.yml
  - /home/s/projects/Alimentor/.github/workflows/ai_pr_review.yml
- Обновлены services.yaml для dataPaths и встроенных хуков:
  - /home/s/projects/project-example/services.yaml
  - /home/s/projects/Alimentor/services.yaml

Для чего
- Централизовать повторяемую логику CI (apply/wait, images, ensure‑slot/ready) в codexctl.
- Убрать дублирующиеся shell‑хуки из services.yaml, заменить их встроенными хуками.
- Сделать управление каталогами данных явным через dataPaths и единым для всех проектов.
