# 06-slot-data-cleanup

## Суть задачи
Гарантировать удаление данных слота при cleanup/GC, чтобы новые задачи не наследовали старые БД/Redis.

## Где сейчас и нюансы (по результатам анализа)
- В `/home/s/projects/project-example/deploy/postgres.service.yaml` и `/home/s/projects/project-example/deploy/redis.service.yaml` используется `hostPath` на `DATA_ROOT/slots/<slot>` (ai) или `DATA_ROOT/<env>` (dev/staging).
- Аналогично в `/home/s/projects/Alimentor/deploy/postgres.service.yaml`, `/home/s/projects/Alimentor/deploy/redis.service.yaml`, `/home/s/projects/Alimentor/deploy/rabbitmq.service.yaml`.
- `manage-env cleanup` и `manage-env gc` в `/home/s/projects/codexctl/internal/cli/manage_env.go` удаляют namespace и configmap, но не затрагивают `DATA_ROOT`.
- Хук `ensure-local-data-dirs` в `/home/s/projects/project-example/services.yaml` и `/home/s/projects/Alimentor/services.yaml` создаёт директории, но нет обратной очистки.

## Что меняем (что именно добавляем)
- В `/home/s/projects/codexctl/internal/cli/manage_env.go` добавляем удаление каталога данных:
  - `DATA_ROOT/slots/<slot>` для ai‑слотов (основной кейс);
  - опционально `DATA_ROOT/<env>` для dev/staging (если будет необходимо).
- Очистку вызываем в `cleanup` и `gc` после удаления namespace/configmap.
- Документируем поведение в `/home/s/projects/project-example/README_RU.md`.

## Зачем / ожидаемый эффект
- Слоты получают чистую БД при повторном использовании.
- Исключаются трудновоспроизводимые баги из‑за старых данных.
