# Дополнительные изменения (вне *.result.md)

- /home/s/projects/codexctl/internal/cli/doctor.go: уточнены проверки наличия утилит (только реально используемые), убран self‑check `codexctl`.
- /home/s/projects/project-example/services.yaml и /home/s/projects/Alimentor/services.yaml: `versions.codexctl` переведён на `latest`.
- /home/s/projects/project-example/deploy/codex/Dockerfile, /home/s/projects/Alimentor/deploy/codex/Dockerfile, /home/s/projects/codexctl/examples/codex-agent/Dockerfile: дефолт `CODEXCTL_VERSION` сменён на `latest`.
- /home/s/projects/codexctl/internal/prompt/templates/*.tmpl: добавлено описание `codexctl` среди базовых инструментов, а в блоках Context7 указано, что там доступна справка по `github.com/codex-k8s/codexctl`.
- /home/s/projects/codexctl/internal/prompt/templates/config_default.toml: `model` и `model_reasoning_effort` берутся из `services.yaml` через `codex.model`/`codex.modelReasoningEffort` с fallback на `gpt-5.2-codex` и `medium`.
- /home/s/projects/codexctl/internal/config/config.go: добавлены поля `codex.model` и `codex.modelReasoningEffort` для шаблонов.
- /home/s/projects/project-example/services.yaml и /home/s/projects/Alimentor/services.yaml: добавлены `codex.model` и `codex.modelReasoningEffort`.
- /home/s/projects/codexctl/internal/cli/issue_overrides.go и /home/s/projects/codexctl/internal/cli/prompt.go: добавлен override `model_reasoning_effort` по меткам Issue `ai-low|ai-medium|ai-high|ai-xhigh` в ai‑запусках.
