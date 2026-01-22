# 02-agent-images-codexctl — результат

## Что сделано

### /home/s/projects/project-example
- /home/s/projects/project-example/services.yaml: добавлена версия `codexctl` (`versions.codexctl`) и передача `CODEXCTL_VERSION` в build-args образа `codex`.
- /home/s/projects/project-example/deploy/codex/Dockerfile: добавлена установка `codexctl` через `go install ...@${CODEXCTL_VERSION}`.
- /home/s/projects/project-example/README_RU.md: уточнено, что версия `codexctl` для образа агента задаётся через `services.yaml`.

### /home/s/projects/Alimentor
- /home/s/projects/Alimentor/services.yaml: добавлена версия `codexctl` и build-arg `CODEXCTL_VERSION` для образа `codex`.
- /home/s/projects/Alimentor/deploy/codex/Dockerfile: добавлена установка `codexctl` через `go install ...@${CODEXCTL_VERSION}`.
- /home/s/projects/Alimentor/README.md: добавлена ремарка, что `codexctl` доступен внутри образа агента и версия задаётся в `services.yaml`.

### /home/s/projects/codexctl
- /home/s/projects/codexctl/examples/codex-agent/Dockerfile: переведена установка `codexctl` на `go install` по версии `CODEXCTL_VERSION` (без локальной сборки).

## Зачем
- Унифицирована установка `codexctl` внутри контейнеров агента в трёх репозиториях.
- Версия `codexctl` фиксируется в `services.yaml` и воспроизводимо передаётся в сборку образа.

