# Vue чек-лист перед PR

## Стек и размещение
- Используется стек: Vue 3 + TypeScript + Pinia + Axios + Vite (`@vitejs/plugin-vue`) + `vite-plugin-pwa` + `vue-router` + `vue-i18n` + `vue3-cookies`.
- Приложение лежит в правильной зоне: `services/external/*` (public) / `services/staff/*` (staff) / `services/dev/*` (dev-only).

## Структура и зависимости
- Структура `src/` выдержана (см. `docs/design-guidelines/vue/frontend_architecture.md`): `src/app`, `src/router`, `src/i18n`, `src/shared/{api,ui,lib,ws}`, `src/features`, `src/pages`.
- Нет циклических зависимостей; `shared/*` не импортирует `features/pages`.

## HTTP (Axios)
- Нет прямых вызовов `axios` из компонентов/страниц; есть единый axios instance + интерсепторы + нормализация ошибок в `src/shared/api/*`.
- UI не показывает сырые сообщения backend; используется `messageKey`/i18n для пользовательских сообщений.

## WebSocket (если есть)
- Есть единый WS клиент в `src/shared/ws/*` (не “подключение в каждом компоненте”).
- Формат сообщений описан в AsyncAPI (`api/server/asyncapi.yaml`) и соблюдён на клиенте; парсинг/валидация входящих сообщений безопасны.
- Учтены reconnect/backoff, heartbeat (ping/pong или аналог), cancel/cleanup на unmount/route change.

## State (Pinia)
- Store по фичам; side-effects в actions/feature services; нет “одного глобального store на всё”.

## Router/i18n/cookies/PWA
- Router: named routes + meta для auth/ролей; guards без “тяжёлых” запросов без стратегии.
- i18n: UI тексты через ключи; форматирование дат/чисел централизовано.
- Cookies: доступ через единый адаптер; нет “магических строк”; чувствительные данные не хранятся без необходимости.
- PWA: кеширование не ломает auth и не выдаёт устаревшие API данные как актуальные; есть UX обновления версии.

## libs (если код нужен >= 2 приложениям)
- Код вынесен в `libs/{vue|ts|js}` по правилам `docs/design-guidelines/common/libraries_reusable_code_requirements.md` и `docs/design-guidelines/vue/libraries.md`.

