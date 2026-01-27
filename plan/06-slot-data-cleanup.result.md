# 06-slot-data-cleanup — результат

## Статус
Проверил реализацию в трех репозиториях: все ключевые механизмы уже внедрены, дополнительные правки не потребовались.

## Что уже сделано (проверено)

### /home/s/projects/codexctl
- Есть конфигурация `dataPaths` в `services.yaml` и типы/резолвер:
  - `/home/s/projects/codexctl/internal/config/data_paths.go`
- Создание/очистка/удаление реализованы в CLI:
  - `/home/s/projects/codexctl/internal/cli/data_paths.go`
- Очистка данных при пересоздании слота:
  - `/home/s/projects/codexctl/internal/cli/ensure_helpers.go` (`dataPathClean`)
- Удаление данных при cleanup и GC:
  - `/home/s/projects/codexctl/internal/cli/manage_env.go` (`dataPathDelete`)
- Хук `codex.ensure-data-dirs` уже создает директории по `dataPaths`:
  - `/home/s/projects/codexctl/internal/hooks/hooks.go`

### /home/s/projects/project-example
- `dataPaths` описаны (root/envDir/dirs), используются для postgres/redis:
  - `/home/s/projects/project-example/services.yaml`
- Хук `codex.ensure-data-dirs` подключен в `hooks.beforeAll`:
  - `/home/s/projects/project-example/services.yaml`

### /home/s/projects/Alimentor
- `dataPaths` описаны (root/envDir/dirs), включают postgres/redis/rabbitmq:
  - `/home/s/projects/Alimentor/services.yaml`
- Хук `codex.ensure-data-dirs` подключен в `hooks.beforeAll`:
  - `/home/s/projects/Alimentor/services.yaml`

## Итог
Очистка и удаление данных слотов уже обеспечены: данные удаляются при `cleanup/gc`, очищаются при пересоздании слота и создаются перед apply через встроенный хук. Новых изменений не требуется.
