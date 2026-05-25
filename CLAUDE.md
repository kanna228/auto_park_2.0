# Claude Guide: ДТП -> Механик -> Заявка на детали -> Склад -> Покупка -> Ремонт

Этот файл нужен для Claude/следующего агента. Он описывает правильную бизнес-логику, какие API использовать, кто их вызывает, какие side effects должны происходить в базе и что нельзя ломать.

## Главная бизнес-логика

Нужный процесс:

1. Диспетчер/админ создает ДТП или другой инцидент.
2. Backend отправляет уведомление назначенному механику.
3. Механик получает уведомление, лично смотрит машину.
4. После осмотра механик создает заявку на детали для ремонта/обслуживания.
5. Backend отправляет уведомление завскладу о новой заявке.
6. Завсклада открывает заявку и смотрит остаток детали.
7. Если деталей хватает, завсклада approve-ит заявку механика.
8. При approve backend автоматически списывает детали со склада и уведомляет механика, что можно работать.
9. Если деталей не хватает, approve блокируется, завсклада получает уведомление о нехватке.
10. Завсклада создает заявку на покупку недостающих деталей.
11. Когда покупку подтвердили, backend автоматически добавляет детали в складской остаток.
12. После этого завсклада возвращается к заявке механика и approve-ит ее.
13. Механик получает уведомление об approve и выполняет ремонт/обслуживание.
14. Выполненную работу механик фиксирует через API обслуживания или установки детали.

Главная цепочка:

```text
POST /api/incidents
  -> notification to mechanic
  -> mechanic inspection
POST /api/warehouse/part-requests
  -> notification to warehouse manager
PATCH /api/warehouse/part-requests/{id}/status
  -> if enough stock: stock issue + mechanic notification
  -> if not enough: 409 + shortage notification to warehouse
POST /api/warehouse/purchase-requests
PATCH /api/warehouse/purchase-requests/{id}/confirm
  -> stock arrival
PATCH /api/warehouse/part-requests/{id}/status
  -> stock issue + mechanic notification
PATCH /api/warehouse/part-requests/{id}/repair-status
  -> when status=completed: backend creates vehicle_part_installations without second stock write-off
POST /api/warehouse/vehicle-services
  -> optional service record if this work also needs service history
```

## Роли

Текущие важные роли:

- `1` admin
- `2` и `3` могут создавать/редактировать инциденты
- `4` duty mechanic, дежурный механик
- `5` warehouse manager, завсклада

Все защищенные запросы идут с JWT:

```http
Authorization: Bearer <token>
```

Уведомления доступны ролям `1,2,3,4,5`.

## Инциденты / ДТП

### Создать инцидент

API:

```http
POST /api/incidents
```

Кто вызывает:

- admin
- dispatcher/manager роли, которые уже имеют доступ к созданию инцидентов

Пример body:

```json
{
  "incident_type_id": 1,
  "vehicle_id": 12,
  "driver_id": 8,
  "mechanic_id": 5,
  "mechanic_shift_id": 3,
  "tripsheet_id": 20,
  "date": "2026-05-25",
  "time": "14:30",
  "location": "Кызылорда, ул. Абая 10",
  "description": "ДТП, поврежден передний бампер и фара"
}
```

Что важно:

- `incident_type_id = 1` обычно ДТП.
- `mechanic_id` обязателен: именно этому механику уйдет уведомление.
- `mechanic_shift_id` обязателен и должен принадлежать этому механику.
- После создания backend переводит машину в статус ТО/ремонтного процесса по существующей логике репозитория инцидентов.
- После создания backend вызывает уведомление типа `incident_created`.

Ожидаемый response:

```json
{
  "success": true,
  "data": {
    "id": 123
  }
}
```

### Получить список типов инцидентов

```http
GET /api/incidents/types
```

Нужно фронту, чтобы выбрать ДТП/Поломка/Повреждение.

### Получить инциденты

```http
GET /api/incidents
GET /api/incidents/{id}
```

Полезные query:

```text
incident_type_id
vehicle_id
driver_id
mechanic_id
mechanic_shift_id
tripsheet_id
date_from
date_to
limit
offset
sort_by
order
```

## Уведомления

### WebSocket уведомлений

```http
GET /api/notifications/ws
```

Использовать на фронте для realtime. При подключении backend пушит snapshot непрочитанных уведомлений и счетчик.

### Список моих уведомлений

```http
GET /api/notifications
GET /api/notifications?only_unread=true
GET /api/notifications/unread
GET /api/notifications/unread/count
```

### Прочитать уведомление

```http
PATCH /api/notifications/{id}/read
PATCH /api/notifications/read-all
```

Типы уведомлений, важные для этого процесса:

- `incident_created`: механику назначен новый инцидент, нужно осмотреть машину.
- `part_request_created`: завскладу пришла новая заявка механика на деталь.
- `part_request_stock_shortage`: завскладу пришло уведомление, что деталей не хватает.
- `part_request_approved`: механику одобрили заявку, детали списаны, можно работать.
- `part_request_rejected`: механику отклонили заявку.

## Заявка механика на детали

Это НЕ заявка на покупку. Это заявка механика на детали со склада под ремонт/обслуживание.

Таблица:

```text
part_requests
```

Статусы:

```text
1 new       новая
2 rejected  отклонена
3 approved  утверждена
```

### Создать заявку механика

API:

```http
POST /api/warehouse/part-requests
```

Кто вызывает:

- duty mechanic
- admin

Пример body:

```json
{
  "part_id": 1,
  "quantity": 2,
  "mechanic_comment": "После ДТП нужно заменить левую фару и крепления"
}
```

Что происходит автоматически:

- создается запись в `part_requests` со статусом `new`;
- создается первая запись истории в `part_request_history`;
- завсклада получает уведомление `part_request_created`;
- заявка видна завскладу в списке новых заявок.

Ожидаемый response:

```json
{
  "success": true,
  "data": {
    "id": 55
  }
}
```

### Завсклада смотрит заявки

API:

```http
GET /api/warehouse/part-requests
```

Кто вызывает:

- warehouse manager
- admin
- mechanic, но mechanic видит только свои заявки

Полезные query:

```text
status_code=new
status_code=approved
status_code=rejected
author_user_id=5
part_id=1
date_from=2026-05-01
date_to=2026-05-25
limit=50
offset=0
sort_by=created_at
order=desc
```

Пример:

```http
GET /api/warehouse/part-requests?status_code=new&limit=50&offset=0
```

В response по каждой заявке есть важное поле:

```json
{
  "id": 55,
  "part_id": 1,
  "part": {
    "id": 1,
    "catalog_part_id": "LAMP-H7",
    "name": "Фара левая",
    "category": "Оптика",
    "available_quantity": 1
  },
  "quantity": 2,
  "status": {
    "id": 1,
    "code": "new",
    "name": "Новая"
  }
}
```

`part.available_quantity` - текущий остаток на складе. Фронт завсклада должен показывать это поле рядом с запрошенным количеством.

### Получить одну заявку

```http
GET /api/warehouse/part-requests/{id}
```

Response содержит саму заявку и историю изменений.

### Approve заявки механика

API:

```http
PATCH /api/warehouse/part-requests/{id}/status
```

Кто вызывает:

- warehouse manager
- admin

Body для approve:

```json
{
  "status_id": 3,
  "comment": "Детали есть на складе, выдать механику"
}
```

Если деталей хватает:

- статус заявки становится `approved`;
- backend атомарно уменьшает `parts_catalog.quantity`;
- backend пишет движение склада в `part_stock_movements` с `type = 'issue'`;
- `part_stock_movements.part_request_id` ссылается на заявку механика;
- механику уходит уведомление `part_request_approved`;
- повторный approve не должен списать детали второй раз.

Успешный response:

```json
{
  "success": true,
  "data": {
    "id": 55
  }
}
```

Если деталей не хватает:

- статус НЕ должен стать `approved`;
- остаток НЕ должен измениться;
- backend возвращает `409 Conflict`;
- завскладу создается уведомление `part_request_stock_shortage`;
- после этого завсклада должен создать заявку на покупку.

Response при нехватке:

```json
{
  "success": false,
  "error": "not enough part quantity in stock"
}
```

### Reject заявки механика

```http
PATCH /api/warehouse/part-requests/{id}/status
```

Body:

```json
{
  "status_id": 2,
  "rejection_comment": "Нужна дополнительная фотофиксация повреждения"
}
```

Что происходит:

- статус становится `rejected`;
- остаток деталей не меняется;
- механику уходит уведомление `part_request_rejected`.

## Заявка на покупку деталей

Это отдельная сущность для завсклада. Она нужна только когда деталей на складе не хватает.

Таблица:

```text
part_purchase_requests
```

API base:

```http
/api/warehouse/purchase-requests
```

Статусы:

```text
new        новая заявка на покупку
confirmed  покупка подтверждена, детали добавлены на склад
cancelled  заявка отменена
```

### Создать заявку на покупку

API:

```http
POST /api/warehouse/purchase-requests
```

Кто вызывает:

- warehouse manager
- admin

Когда вызывать:

- после того как approve заявки механика вернул `409 not enough part quantity in stock`;
- или когда завсклада заранее видит, что `available_quantity < quantity`.

Body:

```json
{
  "part_id": 1,
  "quantity": 2,
  "source_part_request_id": 55,
  "comment": "Не хватает деталей для ремонта после ДТП"
}
```

Поля:

- `part_id`: какая деталь покупается.
- `quantity`: сколько нужно купить.
- `source_part_request_id`: необязательная, но очень желательная ссылка на заявку механика.
- `comment`: комментарий завсклада.

Что происходит:

- создается запись `part_purchase_requests`;
- статус `new`;
- складской остаток пока НЕ увеличивается;
- это именно заявка/намерение купить.

Response:

```json
{
  "success": true,
  "data": {
    "id": 77
  }
}
```

### Список заявок на покупку

```http
GET /api/warehouse/purchase-requests
```

Query:

```text
status=new
status=confirmed
status=cancelled
part_id=1
source_part_request_id=55
limit=50
offset=0
sort_by=created_at
order=desc
```

Пример:

```http
GET /api/warehouse/purchase-requests?status=new&source_part_request_id=55
```

### Получить одну заявку на покупку

```http
GET /api/warehouse/purchase-requests/{id}
```

### Подтвердить покупку

API:

```http
PATCH /api/warehouse/purchase-requests/{id}/confirm
```

Кто вызывает:

- warehouse manager
- admin

Body не нужен.

Что происходит автоматически:

- `part_purchase_requests.status` становится `confirmed`;
- `confirmed_by_user_id` заполняется текущим пользователем;
- `confirmed_at` заполняется текущим временем;
- `parts_catalog.quantity` увеличивается на `part_purchase_requests.quantity`;
- создается движение склада в `part_stock_movements` с `type = 'arrival'`;
- `document_number` будет вида `purchase-request-{id}`;
- если была связь `source_part_request_id`, movement будет связан с исходной заявкой механика через `part_request_id`.

Успешный response:

```json
{
  "success": true,
  "data": {
    "id": 77
  }
}
```

Важно:

- Confirm - это момент, когда детали реально добавляются в базу.
- Нельзя добавлять остаток в момент создания purchase request.
- Повторный confirm уже confirmed-заявки не должен добавить детали второй раз.
- После confirm завсклада должен снова approve-ить исходную заявку механика.

### Отменить заявку на покупку

API:

```http
PATCH /api/warehouse/purchase-requests/{id}/cancel
```

Body:

```json
{
  "comment": "Покупка не нужна, нашли аналог на складе"
}
```

Что происходит:

- статус становится `cancelled`;
- остаток деталей не меняется.

## Складские остатки и движения

Основная таблица остатков:

```text
parts_catalog.quantity
```

Движения склада:

```text
part_stock_movements
```

Типы движений:

```text
arrival  приход / подтвержденная покупка / поступление
issue    выдача детали под заявку механика
return   возврат
writeoff списание
```

Для этой логики:

- confirm purchase request создает `arrival`;
- approve mechanic part request создает `issue`;
- ручное изменение количества детали может создавать `arrival` или `writeoff` через существующую логику.

Посмотреть движения по детали:

```http
GET /api/warehouse/parts/{id}/movements
```

## Каталог деталей

### Список деталей

```http
GET /api/warehouse/parts
```

Query:

```text
part_id
name
category
limit
offset
sort_by
order
```

### Получить деталь

```http
GET /api/warehouse/parts/{id}
```

### Создать/обновить деталь

```http
POST /api/warehouse/parts
PUT /api/warehouse/parts/{id}
```

Это для завсклада/admin. В рамках ДТП-сценария обычно детали уже есть в каталоге, а меняется только `quantity`.

## Фиксация работы механика после approve

После того как механику пришло уведомление `part_request_approved`, фронт должен дать ему возможность зафиксировать выполненную работу.

Есть два похожих API:

1. `/api/warehouse/vehicle-services` - запись факта обслуживания/ремонтной услуги.
2. `/api/warehouse/vehicle-part-installations` - запись установки/замены конкретной детали на конкретной машине.

### Зафиксировать обслуживание

API:

```http
POST /api/warehouse/vehicle-services
```

Body:

```json
{
  "type_id": 1,
  "part_id": 1,
  "vehicle_id": 12,
  "date": "2026-05-25"
}
```

Справочники:

```http
GET /api/warehouse/service-types
GET /api/warehouse/service-parts
```

Использовать, если нужно зафиксировать тип работы: ремонт, замена, диагностика, обслуживание и т.п.

### Зафиксировать установку детали

API:

```http
POST /api/warehouse/vehicle-part-installations
```

Body:

```json
{
  "part_id": 1,
  "vehicle_id": 12,
  "mechanic_shift_id": 3,
  "installed_at": "2026-05-25",
  "planned_replacement_at": "2026-11-25",
  "quantity": 1
}
```

Использовать, если деталь реально поставили на машину и нужно вести историю установленной детали/плановую замену.

Важно:

- Списание со склада уже происходит на approve заявки механика.
- Не надо второй раз списывать деталь при `vehicle-part-installations`.
- Установка детали - это фиксация факта, что механик установил уже выданную/списанную деталь.

## Рекомендуемый frontend flow

### Экран диспетчера/админа

1. Пользователь создает инцидент через `POST /api/incidents`.
2. В форме обязательно выбрать:
   - тип инцидента;
   - машину;
   - водителя;
   - механика;
   - смену механика;
   - дату, время, место.
3. После success показать, что механик уведомлен.

### Экран механика

1. Механик слушает `/api/notifications/ws` или периодически читает `/api/notifications/unread`.
2. Видит `incident_created`.
3. Открывает карточку инцидента через `GET /api/incidents/{id}`.
4. Идет физически смотреть машину.
5. После осмотра выбирает деталь из `GET /api/warehouse/parts`.
6. Создает заявку через `POST /api/warehouse/part-requests`.
7. Ждет approve.
8. Когда приходит `part_request_approved`, выполняет работу.
9. Фиксирует работу через:
   - `POST /api/warehouse/vehicle-services`, если это услуга/обслуживание;
   - `POST /api/warehouse/vehicle-part-installations`, если это установка/замена детали.

### Экран завсклада

1. Завсклада слушает уведомления или открывает список:

```http
GET /api/warehouse/part-requests?status_code=new
```

2. Открывает заявку.
3. Сравнивает:

```text
requested quantity = item.quantity
available stock = item.part.available_quantity
```

4. Если `available_quantity >= quantity`, нажимает approve:

```http
PATCH /api/warehouse/part-requests/{id}/status
```

5. Если `available_quantity < quantity`, создает заявку на покупку:

```http
POST /api/warehouse/purchase-requests
```

6. После покупки/поступления подтверждает:

```http
PATCH /api/warehouse/purchase-requests/{id}/confirm
```

7. Возвращается к заявке механика и approve-ит ее.

## Что нельзя ломать

- Нельзя списывать детали при создании заявки механика.
- Нельзя списывать детали при создании purchase request.
- Нельзя добавлять детали на склад при создании purchase request.
- Детали добавляются только при confirm purchase request.
- Детали списываются только при approve заявки механика.
- Если деталей не хватает, approve должен быть заблокирован.
- Если approve уже был выполнен, повторный approve не должен списать детали второй раз.
- Если purchase request уже confirmed, повторный confirm не должен добавить детали второй раз.
- У механика и завсклада должны оставаться уведомления по ключевым этапам.

## Быстрый end-to-end пример

1. Создать ДТП:

```http
POST /api/incidents
```

2. Механик получил `incident_created`.

3. Механик создал заявку:

```http
POST /api/warehouse/part-requests
```

4. Завсклада получил `part_request_created`.

5. Завсклада пытается approve:

```http
PATCH /api/warehouse/part-requests/55/status
```

Body:

```json
{
  "status_id": 3,
  "comment": "Проверил склад"
}
```

6. Если получил `409`, создает покупку:

```http
POST /api/warehouse/purchase-requests
```

Body:

```json
{
  "part_id": 1,
  "quantity": 2,
  "source_part_request_id": 55,
  "comment": "Докупить для ремонта после ДТП"
}
```

7. Подтверждает покупку:

```http
PATCH /api/warehouse/purchase-requests/77/confirm
```

8. Снова approve заявки механика:

```http
PATCH /api/warehouse/part-requests/55/status
```

Body:

```json
{
  "status_id": 3,
  "comment": "Детали поступили, можно выдавать"
}
```

9. Механик получил `part_request_approved`.

10. Механик фиксирует работу:

```http
POST /api/warehouse/vehicle-part-installations
```

или

```http
POST /api/warehouse/vehicle-services
```

## Где смотреть код

- Инцидент и уведомление механику:
  - `modules/incident_module/service/incident_service.go`
  - `modules/incident_module/router.go`

- Уведомления:
  - `modules/notification_module/service/notification_service.go`
  - `modules/notification_module/repository/notification_repository.go`
  - `modules/notification_module/router.go`

- Заявка механика на деталь:
  - `modules/warehouse_module/service/part_request_service.go`
  - `modules/warehouse_module/repository/part_request_repository.go`
  - `modules/warehouse_module/handlers/part_request_handler.go`

- Заявка на покупку:
  - `modules/warehouse_module/service/purchase_request_service.go`
  - `modules/warehouse_module/repository/purchase_request_repository.go`
  - `modules/warehouse_module/handlers/purchase_request_handler.go`
  - `modules/warehouse_module/dto/purchase_request_dto.go`
  - `modules/warehouse_module/models/purchase_request.go`

- Складские остатки/приходы/движения:
  - `modules/warehouse_module/service/part_service.go`
  - `modules/warehouse_module/repository/part_repository.go`

- Роуты склада:
  - `modules/warehouse_module/router.go`

- Миграции:
  - `database/migrations/000050_incident_and_stock_notifications.up.sql`
  - `database/migrations/000051_warehouse_create_purchase_requests.up.sql`
  - `database/migrations/000052_part_requests_repair_completion.up.sql`

## Важно: completed ремонта и установка детали

Новая правильная логика такая:

1. `PATCH /api/warehouse/part-requests/{id}/status` со `status_id = 3` означает approve завсклада.
2. На approve backend списывает деталь со склада: уменьшает `parts_catalog.quantity` и пишет `part_stock_movements.type = issue`.
3. На approve backend НЕ создает `vehicle_part_installations`.
4. После approve у заявки становится `repair_status = in_progress`.
5. Когда механик реально закончил работу, фронт вызывает:

```http
PATCH /api/warehouse/part-requests/{id}/repair-status
```

Body:

```json
{
  "status": "completed",
  "vehicle_id": 12,
  "mechanic_shift_id": 3,
  "installed_at": "2026-05-25",
  "planned_replacement_at": "2026-11-25",
  "comment": "Repair completed and part installed"
}
```

Что делает backend при `status = completed`:

- проверяет, что заявка уже `approved`;
- берет `part_id` и `quantity` из заявки механика;
- берет `vehicle_id`, `mechanic_shift_id`, `planned_replacement_at` из body или из сохраненных полей заявки;
- создает запись в `vehicle_part_installations`;
- ставит `part_requests.repair_status = completed`;
- заполняет `part_requests.completed_at`, `completed_by_user_id`, `vehicle_part_installation_id`;
- НЕ списывает склад второй раз и НЕ создает еще один `part_stock_movements.issue`.

Если не передали и заранее не сохранили `vehicle_id`, `mechanic_shift_id`, `planned_replacement_at`, backend вернет `400`.

В ДТП-flow не надо напрямую дергать `POST /api/warehouse/vehicle-part-installations` после approve, иначе старый API установки может списать склад как отдельную операцию. Для заявки механика установка детали должна идти только через `PATCH /api/warehouse/part-requests/{id}/repair-status`.

## Короткое правило для Claude

Если пользователь просит "сделай логику ДТП -> механик -> склад -> покупка -> ремонт", не придумывать новый flow. Использовать ровно этот:

```text
incident_created notification
mechanic part_request
warehouse approve or shortage
purchase_request if shortage
purchase_request confirm adds stock
part_request approve issues stock
mechanic notification
part_request repair-status completed creates vehicle_part_installations
vehicle service after work if service history is also needed
```
