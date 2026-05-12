# Notifications patch

## Что добавлено

- Отдельный `notification_module`.
- WebSocket endpoint: `GET /api/notifications/ws`.
- REST endpoints для непрочитанных уведомлений и отметки прочтения.
- Таблицы `notification_types` и `notifications`.
- Интеграция с `part_requests`:
  - при создании заявки уведомляются все пользователи с ролью `warehouse_manager` / fallback role_id = 5;
  - при утверждении заявки уведомляется механик-автор заявки;
  - при отклонении заявки уведомляется механик-автор заявки.

## WebSocket события

```json
{"event":"notification.created","data":{...}}
{"event":"notification.unread_snapshot","data":{"items":[...],"total":1,"limit":50,"offset":0}}
{"event":"notification.unread_count","data":{"count":1}}
```

## Авторизация WebSocket

Лучше через cookie `session_token`.

Также поддержано:

```text
/api/notifications/ws?access_token=<JWT>
/api/notifications/ws?token=<JWT>
```

Обычные HTTP endpoints работают через cookie или `Authorization: Bearer <JWT>`.

## Важно

Добавлен новый dependency:

```go
github.com/gorilla/websocket v1.5.3
```

После вставки файлов выполни:

```bash
go mod tidy
```

или просто:

```bash
docker compose up --build
```

Docker должен скачать dependency и обновить `go.sum`.
