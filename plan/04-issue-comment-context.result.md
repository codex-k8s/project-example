# 04-issue-comment-context — результат

## Что сделано

### /home/s/projects/codexctl
- /home/s/projects/codexctl/internal/githubapi/client.go: добавлен GitHub GraphQL‑клиент для получения Issue‑комментариев и PR review‑комментариев с фильтрацией `minimized` и `isResolved`.
- /home/s/projects/codexctl/internal/githubapi/types.go: вынесены структуры ответов/типов и добавлен TODO о переходе на официальный Go‑клиент GitHub.
- /home/s/projects/codexctl/internal/promptctx/types.go: новые модели `IssueComment`/`ReviewComment` для контекста промптов.
- /home/s/projects/codexctl/internal/config/config.go: `TemplateContext` теперь хранит комментарии через `promptctx` (без привязки к services.yaml).
- /home/s/projects/codexctl/internal/cli/issue_context.go: добавлен сбор контекста Issue/PR и запись в TemplateContext; поддержан `FOCUS_ISSUE_NUMBER`.
- /home/s/projects/codexctl/internal/cli/prompt.go: подключён сбор контекста Issue/PR перед рендером промптов.
- /home/s/projects/codexctl/internal/prompt/templates/dev_issue_*.tmpl, plan_issue_*.tmpl, plan_review_*.tmpl:
  - добавлены списки актуальных Issue‑комментариев (не скрытых) с ID и URL;
  - добавлены инструкции ссылаться на ID в ответах.
- /home/s/projects/codexctl/internal/prompt/templates/dev_review_*.tmpl:
  - добавлен список нерешённых PR review‑комментариев с ID и URL;
  - добавлена инструкция отвечать по ID через `gh api`.

### /home/s/projects/project-example
- Изменения только через codexctl/templates (промпты и контекст), файлов проекта не добавлялось.

### /home/s/projects/Alimentor
- Изменения только через codexctl/templates (промпты и контекст), файлов проекта не добавлялось.

## Для чего
- Агент получает готовый список актуальных (не скрытых) комментариев Issue и нерешённых review‑комментариев PR.
- Убирается повторное чтение/дублирование контекста и снижается риск ответа на скрытые/закрытые замечания.
- Появляется единый формат с ID, по которым можно отвечать на комментарии.
