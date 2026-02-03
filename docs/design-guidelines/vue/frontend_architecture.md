# Frontend: архитектура и размещение

## Где живёт фронтенд
- Приложения для внешних пользователей (public) размещаются в `services/external/*`.
- Приложения для сотрудников/админки размещаются в `services/staff/*`.
- Dev-only фронтенд (панели тестирования/отладки) размещается в `services/dev/*`.
- Приложение (frontend) — единица деплоя и версионирования так же, как сервис.

## Технологический стек (фиксирован, но может быть дополнен по необходимости)
- Vue 3
- TypeScript
- Pinia (state)
- Axios (HTTP)
- Vite + `@vitejs/plugin-vue` + `vite-plugin-pwa`
- `vue-i18n`
- `vue-router`
- `vue3-cookies`

## Границы ответственности
- Frontend не содержит доменной логики backend-сервисов: он отображает состояние и оркестрирует UI, а не “изобретает” бизнес-правила.
- Контракт UI с backend: HTTP/OpenAPI (`api/server/api.yaml`) и стабильный формат ошибок (см. `docs/design-guidelines/go/error_handling.md`).
- Контракт async сообщений (WebSocket) описывается в AsyncAPI YAML: `api/server/asyncapi.yaml` (если используется).
- Взаимодействие с backend через единый HTTP-клиент (axios instance) и слой адаптеров/кастеров.

## Рекомендуемая структура приложения (Vite)
Внутри `services/<zone>/<app-name>/`:
- `index.html`, `package.json`, `vite.config.ts`.
- `public/` — статические файлы.
- `src/` — исходники:
  - `src/app/` — composition root: создание app, подключение router/pinia/i18n, регистрация PWA.
  - `src/router/` — маршруты, guards, route meta.
  - `src/i18n/` — конфиг `vue-i18n`, словари и ключи.
  - `src/shared/` — переиспользуемые куски внутри приложения (без привязки к конкретной странице):
    - `src/shared/api/` — axios client + типы ошибок + маппинг ответов.
    - `src/shared/ws/` — WebSocket client + типы сообщений + обработчики (если используется).
    - `src/shared/ui/` — базовые UI-компоненты приложения.
    - `src/shared/lib/` — утилиты (TS), форматтеры, parsers/casters.
  - `src/features/` — фичи (модули UI), каждая со своим state (Pinia), компонентами и API-адаптерами.
  - `src/pages/` — страницы (route-level) и композиция фич.
  - `src/widgets/` (опционально) — крупные блоки страниц, если нужно разделение.

Запрещено:
- прямые вызовы `axios` из компонентов/страниц (только через единый клиент и слой `shared/api`/фичи);
- хранить “общие” утилиты в каждом приложении вместо `libs/{vue|ts|js}` (если используется >= 2 приложениями);
- дублировать i18n строки (ключи должны быть стабильны и переиспользуемы).
