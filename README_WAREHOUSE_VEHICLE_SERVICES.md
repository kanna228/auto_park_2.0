# Warehouse vehicle services addition

Добавлено обслуживание машины без списания деталей со склада и без установки новых деталей.

## Таблицы

### `parts_collection`
Справочник частей автомобиля, к которым может относиться обслуживание.

Поля:
- `id`
- `name`
- `description` nullable
- `created_at`
- `updated_at`

Начальные данные: передний/задний бампер, двери, крылья, капот, крыша, стекла, фары, зеркала, весь кузов.

### `service_types`
Справочник типов обслуживания.

Поля:
- `id`
- `name`
- `description` nullable
- `created_at`
- `updated_at`

Начальные данные: полировка, тонировка, защитная пленка, антигравийная защита, химчистка, детейлинг, покраска, рихтовка, мойка двигателя, антикор.

### `vehicle_services`
История обслуживания автомобиля.

Поля:
- `id`
- `type_id` -> `service_types.id`
- `part_id` -> `parts_collection.id`
- `vehicle_id` -> `vehicles.id`
- `service_date`
- `created_at`
- `updated_at`

## Routes

### Parts collection
```text
GET    /api/warehouse/service-parts
GET    /api/warehouse/service-parts/:id
POST   /api/warehouse/service-parts
PUT    /api/warehouse/service-parts/:id
DELETE /api/warehouse/service-parts/:id
```

### Service types
```text
GET    /api/warehouse/service-types
GET    /api/warehouse/service-types/:id
POST   /api/warehouse/service-types
PUT    /api/warehouse/service-types/:id
DELETE /api/warehouse/service-types/:id
```

### Vehicle services
```text
GET    /api/warehouse/vehicle-services
GET    /api/warehouse/vehicle-services/:id
POST   /api/warehouse/vehicle-services
PUT    /api/warehouse/vehicle-services/:id
DELETE /api/warehouse/vehicle-services/:id
```

## Vehicle services filters

```text
vehicle_id
part_id
type_id
part_name
type_name
date_from
date_to
limit
offset
sort_by=date|created_at|updated_at|vehicle_id|part_id|type_id
order=asc|desc
```

## Passport response

`GET /api/vehicles/:id` now returns:

```json
"services": [
  {
    "id": 1,
    "type_id": 1,
    "type": {
      "id": 1,
      "name": "Полировка",
      "description": "Восстановление блеска и удаление мелких дефектов поверхности"
    },
    "part_id": 20,
    "part": {
      "id": 20,
      "name": "Весь кузов",
      "description": "Обслуживание всего кузова автомобиля"
    },
    "vehicle_id": 12,
    "date": "2026-05-14T00:00:00Z",
    "created_at": "2026-05-14T10:00:00Z",
    "updated_at": "2026-05-14T10:00:00Z"
  }
]
```

## Example create

```json
{
  "type_id": 1,
  "part_id": 20,
  "vehicle_id": 12,
  "date": "2026-05-14"
}
```
