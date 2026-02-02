# 9 — Сквозные примеры и шаблоны

## Общая схема (мнемоника)
- Внизу: ошибка возникает и оборачивается контекстом
- В домене: typed error только когда нужен
- В transport handler: просто `return err`
- На границе: маппинг + логирование + контракт ответа

---

## Пример A: HTTP Not Found
Repository:
- `sql.ErrNoRows` -> `errs.NotFound{Entity:"user", ID:id}`
  Service:
- добавляет контекст `"get user: %w"`
  Handler:
- `return err`
  HTTP error boundary:
- `errs.NotFound` -> 404 + стабильное сообщение

---

## Пример B: HTTP бизнес-валидация
Service:
- при нарушении правила -> `errs.Validation{Field:"age", Msg:"must be >= 18"}`
  HTTP boundary:
- `Validation` -> 400 + field (+ safe details)

---

## Пример C: HTTP OpenAPI/schema validation
Middleware:
- ловит schema/type ошибки
  HTTP boundary:
- 400 + `{message, loc, field}` без внутренних деталей в prod

---

## Пример D: gRPC Not Found
Service:
- возвращает `errs.NotFound`
  gRPC interceptor (boundary):
- `errs.NotFound` -> `status.Error(codes.NotFound, "not found")`

---

## Пример E: gRPC Validation
Service:
- возвращает `errs.Validation{Field:"limit", Msg:"must be <= 100"}`
  gRPC interceptor:
- `InvalidArgument` + (опционально) структурированные details про поле

---

## Пример F: параллельность cancel-on-first
- `errgroup.WithContext`
- каждая задача добавляет контекст
- итоговая ошибка возвращается наверх, логирование на границе

---

## Пример G: параллельность collect-all
- задачи независимы
- собираем все ошибки и объединяем (например, join)
- логируем один раз на границе

---

## Пример H: panic в goroutine -> error
- `recover()` внутри обёртки превращает panic в error
- дальше правило то же: возвращаем наверх, логируем на границе

---

## Шаблон “как писать новый код”
1) Repository: `%w` + нормализация инфраструктуры -> домен
2) Service: бизнес-валидация -> `errs.Validation`, осмысленный контекст
3) Handler: `return err` без форматирования и логов
4) Boundary (HTTP/gRPC): маппинг + логирование + стабильный контракт
