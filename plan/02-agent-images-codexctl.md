# 02-agent-images-codexctl

## Суть задачи
Сделать `codexctl` доступным внутри контейнера агента во всех проектах, чтобы агент мог вызывать его напрямую.

## Где сейчас и нюансы (по результатам анализа)
- `/home/s/projects/project-example/deploy/codex/Dockerfile` и `/home/s/projects/Alimentor/deploy/codex/Dockerfile` не устанавливают `codexctl`.
- `/home/s/projects/codexctl/examples/codex-agent/Dockerfile` собирает `codexctl` из локального исходного кода (`COPY .` + `go build`).
- В обоих codex‑образах уже есть Go и `PATH`/`GOBIN` настроены так, чтобы `go install` мог класть бинарник в `/usr/local/bin`.

## Что меняем (что именно добавляем)
- В `/home/s/projects/project-example/services.yaml` и `/home/s/projects/Alimentor/services.yaml` добавляем версию `versions.codexctl` и передаём `CODEXCTL_VERSION` как build‑arg.
- В `/home/s/projects/project-example/deploy/codex/Dockerfile` и `/home/s/projects/Alimentor/deploy/codex/Dockerfile` добавляем установку `codexctl` через `go install ...@${CODEXCTL_VERSION}`.
- В `/home/s/projects/codexctl/examples/codex-agent/Dockerfile` переводим установку на `go install` (без локальной сборки).
- В `/home/s/projects/project-example/README_RU.md` и `/home/s/projects/Alimentor/README_RU.md` фиксируем, что `codexctl` доступен внутри агента и как задаётся версия.

## Зачем / ожидаемый эффект
- Агент получает единый CLI‑инструмент (`codexctl`) в любом слоте.
- Версия фиксирована и воспроизводима (через `CODEXCTL_VERSION`).
- Сборка образов становится проще и стабильнее.
