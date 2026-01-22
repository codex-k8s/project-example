# 06-slot-data-cleanup

## Суть задачи
Гарантировать создание/очистку/удаление данных слота на основе конфигурации из `services.yaml`, чтобы новые задачи не наследовали старые БД/Redis.

## Где сейчас и нюансы (по результатам анализа)
- В `/home/s/projects/project-example/deploy/postgres.service.yaml` и `/home/s/projects/project-example/deploy/redis.service.yaml` используется `hostPath` на `DATA_ROOT/slots/<slot>` (ai) или `DATA_ROOT/<env>` (dev/staging).
- Аналогично в `/home/s/projects/Alimentor/deploy/postgres.service.yaml`, `/home/s/projects/Alimentor/deploy/redis.service.yaml`, `/home/s/projects/Alimentor/deploy/rabbitmq.service.yaml`.
- `ensure-local-data-dirs` реализован как shell‑hook в `services.yaml`, но список директорий жёстко закодирован внутри скрипта и отличается по проектам.
- `manage-env cleanup` и `manage-env gc` в `/home/s/projects/codexctl/internal/cli/manage_env.go` удаляют namespace и configmap, но не затрагивают `DATA_ROOT`.

## Что меняем (что именно добавляем)
- В `/home/s/projects/codexctl/internal/config/config.go` добавляем новый блок `dataPaths`, где проект описывает список директорий данных (полные пути или шаблоны, завязанные на `DATA_ROOT`, `Env`, `Slot`).
- В `/home/s/projects/codexctl/internal/hooks/hooks.go` добавляем встроенный хук `codex.ensure-data-dirs`, который:
  - создаёт каталоги из `dataPaths`;
  - может очищать их при необходимости (например, при принудительном пересоздании слота).
- В `/home/s/projects/codexctl/internal/cli/manage_env.go` добавляем использование `dataPaths` для удаления данных при cleanup/GC.
- В `/home/s/projects/project-example/services.yaml` и `/home/s/projects/Alimentor/services.yaml` заводим `dataPaths` с полным набором путей, специфичных для проекта.

## Что значит «создавать/очищать/удалять»
- Создавать: при запуске окружения (до apply) создаём директории из `dataPaths`.
- Очищать: при явном пересоздании/repair‑режиме удаляем содержимое директорий, но оставляем сами директории.
- Удалять: при cleanup/GC удаляем директории слота полностью.

## Зачем / ожидаемый эффект
- Слоты получают чистую БД при повторном использовании.
- Исключаются «утечки» данных между задачами.
- Конфигурация данных становится прозрачной и настраиваемой в каждом проекте.
