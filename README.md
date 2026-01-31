# Warehouse (учебный проект)

Требования (выполнено):
- CRUD инвентаря: POST/GET/PUT/DELETE /api/items (DELETE только admin)
- JWT авторизация (роль в токене) + RBAC на каждом запросе
- История изменений: таблица items_history, заполнение ТОЛЬКО через триггеры Postgres (антипаттерн)
- UI: логин по роли, список товаров, CRUD по правам, история по товару, фильтры, экспорт CSV, diff для update

## Запуск через Docker (db + api)
```bash
docker compose up --build
```

Открыть UI:
- http://localhost:8080/web/

## Локальный запуск (если Postgres уже поднят)
1) env
```bash
cp .env.example .env
```

2) deps
```bash
go mod tidy
```

3) run
```bash
go run ./cmd/server
```

## API
- POST /api/auth/login  {username, role} -> {token}
- GET  /api/items?search=...
- POST /api/items
- PUT  /api/items/{id}
- DELETE /api/items/{id}   (admin)
- GET /api/items/{id}/history?from=&to=&user=&action=&includeChanges=1
- GET /api/items/{id}/history.csv?from=&to=&user=&action=
