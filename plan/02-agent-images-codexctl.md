# 02-agent-images-codexctl

## Цель
Сделать `codexctl` доступным внутри контейнеров агента, чтобы агент мог использовать утилиту напрямую.

## Выбранный подход
Установка через `go install` с закрепленной версией в `CODEXCTL_VERSION`.

## Что делаем и где
- Добавляем версию `codexctl` в версии образов.
  - Файлы: `project-example/services.yaml`, `Alimentor/services.yaml` (поле `versions.codexctl`).
- Прокидываем `CODEXCTL_VERSION` в сборку codex‑образа.
  - Файлы: `project-example/deploy/codex/Dockerfile`, `Alimentor/deploy/codex/Dockerfile`.
- Переводим `codexctl/examples/codex-agent/Dockerfile` на установку через `go install` вместо локальной сборки.
  - Файл: `codexctl/examples/codex-agent/Dockerfile`.
- Обновляем документацию о версии и доступности `codexctl` внутри агента.
  - Файлы: `project-example/README_RU.md`, `Alimentor/README_RU.md`.

## Зачем
- Агент получает единый инструмент для всех операций (`apply`, `prompt`, `manage-env` и т.д.).
- Версия фиксирована и воспроизводима.
- Упрощается диагностика и повторяемость в облачных окружениях.

## Ожидаемый результат
- Внутри контейнера агента доступен `codexctl` нужной версии.
- Сборка образов не зависит от локального кода codexctl.
