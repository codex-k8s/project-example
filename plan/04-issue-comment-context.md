# 04-issue-comment-context

## Суть задачи
Перенести сбор Issue/PR‑контекста в codexctl и передавать агенту только актуальные (не скрытые и не резолвнутые) замечания.

## Где сейчас и нюансы (по результатам анализа)
- Шаблоны промптов (`/home/s/projects/codexctl/internal/prompt/templates/*.tmpl`) требуют от агента вручную вычитывать комментарии через `gh`.
- В `codexctl` нет централизованного сборщика комментариев; `prompt` получает только `ISSUE_NUMBER`/`PR_NUMBER` через inline vars (`/home/s/projects/codexctl/internal/cli/prompt.go`).
- Для фильтрации нужны поля `minimizedReason` и `isResolved`, которые удобнее получать через GraphQL (REST не даёт полной картины тредов).
- `codexctl` уже использует `gh` в `/home/s/projects/codexctl/internal/cli/plan.go` и `/home/s/projects/codexctl/internal/hooks/hooks.go`, поэтому паттерн работы с GitHub API уже есть.

## Что меняем (что именно добавляем)
- В `/home/s/projects/codexctl` добавляем GitHub‑клиент (новый пакет `internal/github/*`) для получения:
  - Issue‑комментариев (с `minimizedReason`);
  - PR review‑тредов (с `isResolved`).
- Расширяем `TemplateContext` в `/home/s/projects/codexctl/internal/config/config.go` для передачи отфильтрованных комментариев в шаблоны.
- В `/home/s/projects/codexctl/internal/cli/prompt.go` добавляем сбор комментариев на старте `prompt run` и добавляем их в контекст шаблонов.
- В `/home/s/projects/codexctl/internal/prompt/templates/*.tmpl` заменяем инструкции «сам найди комментарии» на готовый список нерешённых замечаний с ID.

## Зачем / ожидаемый эффект
- Агент видит только актуальные задачи и не отвечает на закрытые/скрытые комментарии.
- Появляется стабильный формат «список замечаний + ID для ответа».
- Повышается точность и скорость работы при review/plan‑сценариях.
