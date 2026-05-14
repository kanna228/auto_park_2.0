# Admin Full Access Patch

Этот патч делает администратора (`role_id = 1`) супер-пользователем на уровне middleware.

## Что изменено

### `middleware/auth_jwt.go`

`RequireRoles(...)` теперь:

1. Достаёт `role_id` более безопасно: поддерживает `int64`, `int`, `int32`, `uint`, `uint64`, `float64`, `string`.
2. Если `role_id == 1`, пользователь сразу проходит любой `RequireRoles(...)`, даже если в конкретном роуте забыли добавить администратора.

Это решает проблему, когда admin получал:

```json
{"success":false,"error":"insufficient permissions"}
```

на protected routes.

### `modules/warehouse_module/router.go`

Дополнительно явно добавлен `roleAdmin` в warehouse-группы, где раньше были только механик/заведующий:

- создание/редактирование деталей;
- создание заявок;
- установка деталей на авто.

Даже если где-то потом забудешь добавить `roleAdmin`, middleware всё равно пропустит администратора.

## После вставки

```bash
gofmt -w middleware/auth_jwt.go modules/warehouse_module/router.go
docker compose up --build
```

Потом заново запусти API test script.
