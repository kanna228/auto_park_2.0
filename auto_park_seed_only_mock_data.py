#!/usr/bin/env python3
"""
auto_park_seed_only_mock_data.py

Заполняет проект mock-данными БЕЗ удаления.
Создаёт по 3 записи там, где это возможно через текущие API:

- users (3)
- drivers (3)
- vehicles (3)
- tire_places (3)
- tires (3)
- tripsheets (3)
- tripsheet_trips (3)

Использует admin login:
    email=admin@autopark.local
    password=admin123
    iin=000000000000

Запуск:
    pip install requests
    python auto_park_seed_only_mock_data.py

Env:
    AUTO_PARK_BASE_URL=http://localhost:8080
    AUTO_PARK_ADMIN_EMAIL=admin@autopark.local
    AUTO_PARK_ADMIN_PASSWORD=admin123
    AUTO_PARK_ADMIN_IIN=000000000000
    AUTO_PARK_TIMEOUT=20
"""

from __future__ import annotations

import json
import os
import random
import string
import sys
import time
from dataclasses import dataclass, field
from typing import Any, Dict, List, Optional, Tuple

import requests

BASE_URL = os.getenv("AUTO_PARK_BASE_URL", "http://localhost:8080").rstrip("/")
ADMIN_EMAIL = os.getenv("AUTO_PARK_ADMIN_EMAIL", "admin@autopark.local")
ADMIN_PASSWORD = os.getenv("AUTO_PARK_ADMIN_PASSWORD", "admin123")
ADMIN_IIN = os.getenv("AUTO_PARK_ADMIN_IIN", "000000000000")
TIMEOUT = int(os.getenv("AUTO_PARK_TIMEOUT", "20"))


def rand_suffix(n: int = 6) -> str:
    chars = string.ascii_lowercase + string.digits
    return "".join(random.choice(chars) for _ in range(n))


def now_ms() -> int:
    return int(time.time() * 1000)


@dataclass
class SeedResult:
    entity: str
    success: bool
    status_code: Optional[int]
    name: str
    id_value: Optional[int] = None
    error: Optional[str] = None
    response_preview: Optional[str] = None


@dataclass
class SeedContext:
    token: Optional[str] = None
    admin_user_id: Optional[int] = None

    users: List[int] = field(default_factory=list)
    drivers: List[int] = field(default_factory=list)
    vehicles: List[int] = field(default_factory=list)
    tire_places: List[int] = field(default_factory=list)
    tires: List[int] = field(default_factory=list)
    tripsheets: List[int] = field(default_factory=list)
    tripsheet_trips: List[int] = field(default_factory=list)


class Seeder:
    def __init__(self) -> None:
        self.session = requests.Session()
        self.ctx = SeedContext()
        self.results: List[SeedResult] = []

    def headers(self, auth: bool = False) -> Dict[str, str]:
        h = {"Content-Type": "application/json"}
        if auth and self.ctx.token:
            h["Authorization"] = f"Bearer {self.ctx.token}"
        return h

    def request(self, method: str, path: str, *, auth: bool = False, json_body: Optional[dict] = None,
                expected: Tuple[int, ...] = (200, 201)) -> requests.Response:
        url = f"{BASE_URL}{path}"
        resp = self.session.request(
            method=method,
            url=url,
            headers=self.headers(auth=auth),
            json=json_body,
            timeout=TIMEOUT,
        )
        if resp.status_code not in expected:
            raise AssertionError(f"{method} {path}: expected {expected}, got {resp.status_code}: {resp.text[:300]}")
        return resp

    def safe_create(self, entity: str, name: str, method: str, path: str, *,
                    auth: bool = True, json_body: Optional[dict] = None,
                    expected: Tuple[int, ...] = (200, 201)) -> Optional[dict]:
        try:
            resp = self.request(method, path, auth=auth, json_body=json_body, expected=expected)
            data = self.try_json(resp)
            item_id = self.extract_id(data)
            self.results.append(SeedResult(
                entity=entity,
                success=True,
                status_code=resp.status_code,
                name=name,
                id_value=item_id,
                response_preview=resp.text[:300],
            ))
            return data if isinstance(data, dict) else {"raw": data}
        except Exception as e:
            self.results.append(SeedResult(
                entity=entity,
                success=False,
                status_code=None,
                name=name,
                error=str(e),
            ))
            return None

    def wait_api(self, seconds: int = 60) -> None:
        deadline = time.time() + seconds
        while time.time() < deadline:
            try:
                self.request("GET", "/healthz", expected=(200,))
                return
            except Exception:
                time.sleep(2)
        raise RuntimeError("API did not become healthy in time")

    def try_json(self, resp: requests.Response) -> Any:
        try:
            return resp.json()
        except Exception:
            return None

    def extract_nested(self, data: Any, *keys: str) -> Any:
        cur = data
        for key in keys:
            if not isinstance(cur, dict) or key not in cur:
                return None
            cur = cur[key]
        return cur

    def extract_id(self, data: Any) -> Optional[int]:
        if isinstance(data, dict):
            for key in ("id", "ID", "user_id", "vehicle_id", "tripsheet_id", "tire_id", "driver_id"):
                if key in data and isinstance(data[key], int):
                    return data[key]
            for value in data.values():
                found = self.extract_id(value)
                if found is not None:
                    return found
        elif isinstance(data, list):
            for item in data:
                found = self.extract_id(item)
                if found is not None:
                    return found
        return None

    def login(self) -> None:
        payload = {
            "email": ADMIN_EMAIL,
            "password": ADMIN_PASSWORD,
            "iin": ADMIN_IIN,
        }
        resp = self.request("POST", "/api/users/login", json_body=payload, expected=(200,))
        data = self.try_json(resp)
        token = self.extract_nested(data, "data", "token")
        user_id = self.extract_nested(data, "data", "user_id")
        if not token:
            raise RuntimeError(f"Login succeeded but token not found in response: {data}")
        self.ctx.token = token
        self.ctx.admin_user_id = user_id

    def seed_users(self, count: int = 3) -> None:
        for i in range(count):
            suffix = rand_suffix()
            payload = {
                "email": f"mock_user_{suffix}@example.com",
                "first_name": f"Mock{i+1}",
                "last_name": "User",
                "middle_name": "Seed",
                "iin": f"99{random.randint(10,99)}{random.randint(10,99)}{random.randint(100000,999999)}",
                "phone": f"+7700{random.randint(1000000,9999999)}",
                "role_id": 2,
            }
            data = self.safe_create(
                "users",
                f"user_{i+1}",
                "POST",
                "/api/users",
                auth=True,
                json_body=payload,
                expected=(201,),
            )
            item_id = self.extract_id(data)
            if item_id:
                self.ctx.users.append(item_id)

    def seed_drivers(self, count: int = 3) -> None:
        for i in range(count):
            suffix = rand_suffix()
            payload = {
                "iin": f"01{random.randint(10,99)}{random.randint(10,99)}{random.randint(100000,999999)}",
                "name": f"Driver{i+1}",
                "surname": f"Seed{suffix}",
                "middlename": "A",
                "phone": f"+7700{random.randint(1000000,9999999)}",
                "mail": f"mock_driver_{suffix}@example.com",
            }
            data = self.safe_create(
                "drivers",
                f"driver_{i+1}",
                "POST",
                "/api/users/drivers",
                auth=True,
                json_body=payload,
                expected=(201,),
            )
            item_id = self.extract_id(data)
            if item_id:
                self.ctx.drivers.append(item_id)

    def seed_tire_places(self, count: int = 3) -> None:
        for i in range(count):
            payload = {
                "name": f"mock-tire-place-{i+1}-{rand_suffix()}",
            }
            data = self.safe_create(
                "tire_places",
                f"tire_place_{i+1}",
                "POST",
                "/api/vehicles/tire-places",
                auth=True,
                json_body=payload,
                expected=(201,),
            )
            item_id = self.extract_id(data)
            if item_id:
                self.ctx.tire_places.append(item_id)

    def seed_vehicles(self, count: int = 3) -> None:
        available_driver_ids = self.ctx.drivers[:]
        if not available_driver_ids:
            return

        for i in range(count):
            suffix = rand_suffix()
            payload = {
                "board_number": f"B-{random.randint(100,999)}",
                "technical_passport_number": f"TP-{random.randint(10000,99999)}",
                "state_number": f"777{random.choice(['AAA','BBB','CCC'])}{random.randint(100,999)}",
                "vin": f"VIN{suffix.upper()}123456789",
                "brand_model": random.choice(["Toyota Camry", "Hyundai Elantra", "Kia K5"]),
                "manufacture_year": random.choice([2020, 2021, 2022, 2023]),
                "received_date": "2026-03-31T00:00:00Z",
                "empty_weight_kg": random.choice([1450.0, 1500.0, 1600.0]),
                "max_weight_kg": random.choice([2100.0, 2200.0, 2300.0]),
                "engine_volume_cc": random.choice([1800, 2000, 2500]),
                "insurance_policy_number": f"POL-{random.randint(1000,9999)}",
                "insurance_expiry_date": "2027-03-31T00:00:00Z",
                "mileage": random.choice([5000, 12000, 18000]),
                "current_fuel": random.choice([20.0, 35.5, 42.0]),
                "drivers_ids": [available_driver_ids[i % len(available_driver_ids)]],
            }
            data = self.safe_create(
                "vehicles",
                f"vehicle_{i+1}",
                "POST",
                "/api/vehicles",
                auth=True,
                json_body=payload,
                expected=(201,),
            )
            item_id = self.extract_id(data)
            if item_id:
                self.ctx.vehicles.append(item_id)

    def seed_tires(self, count: int = 3) -> None:
        if not self.ctx.vehicles or not self.ctx.tire_places:
            return

        for i in range(count):
            payload = {
                "place_id": self.ctx.tire_places[i % len(self.ctx.tire_places)],
                "vehicle_id": self.ctx.vehicles[i % len(self.ctx.vehicles)],
                "tire": random.choice(["Michelin Primacy 4", "Pirelli Cinturato", "Bridgestone Turanza"]),
                "mileage": random.choice([3000, 7000, 12000]),
                "max_usage": random.choice([50000, 60000, 65000]),
            }
            data = self.safe_create(
                "tires",
                f"tire_{i+1}",
                "POST",
                "/api/vehicles/tires",
                auth=True,
                json_body=payload,
                expected=(201,),
            )
            item_id = self.extract_id(data)
            if item_id:
                self.ctx.tires.append(item_id)

    def seed_tripsheets(self, count: int = 3) -> None:
        if not self.ctx.vehicles or not self.ctx.drivers:
            return

        for i in range(count):
            payload = {
                "tripsheet_number": f"TS-{random.randint(1000,9999)}-{rand_suffix(3)}",
                "tripsheet_date": "2026-03-31",
                "vehicle_id": self.ctx.vehicles[i % len(self.ctx.vehicles)],
                "vehicle_brand": random.choice(["Toyota Camry", "Hyundai Elantra", "Kia K5"]),
                "vehicle_plate_number": f"777{random.choice(['AAA','BBB','CCC'])}{random.randint(100,999)}",
                "driver_last_name": "Driver",
                "driver_first_name": f"Mock{i+1}",
                "driver_middle_name": "A",
                "driver_id": self.ctx.drivers[i % len(self.ctx.drivers)],
                "start_time": "2026-03-31T08:00:00Z",
                "end_time": "2026-03-31T18:00:00Z",
                "mileage_start": random.choice([10000, 12000, 15000]),
                "mileage_end": random.choice([10120, 12150, 15180]),
                "fuel_start": random.choice([20, 30, 40]),
                "fuel_issued": random.choice([10, 15, 20]),
                "fuel_consumption_theoretical": random.choice([8, 10, 12]),
                "fuel_consumption_actual": random.choice([9, 11, 13]),
                "status_id": 1,
            }
            data = self.safe_create(
                "tripsheets",
                f"tripsheet_{i+1}",
                "POST",
                "/api/tripsheet",
                auth=True,
                json_body=payload,
                expected=(201,),
            )
            item_id = self.extract_id(data)
            if item_id:
                self.ctx.tripsheets.append(item_id)

    def seed_tripsheet_trips(self, count: int = 3) -> None:
        if not self.ctx.tripsheets:
            return

        for i in range(count):
            payload = {
                "tripsheet_id": self.ctx.tripsheets[i % len(self.ctx.tripsheets)],
                "route_description": random.choice([
                    "Warehouse -> Service center -> Garage",
                    "Garage -> Fuel station -> Office",
                    "Depot -> Inspection point -> Garage",
                ]),
                "start_time": "2026-03-31T09:00:00Z",
                "end_time": "2026-03-31T10:30:00Z",
                "distance_passed": random.choice([25, 42, 58]),
                "status_id": 1,
            }
            data = self.safe_create(
                "tripsheet_trips",
                f"tripsheet_trip_{i+1}",
                "POST",
                "/api/tripsheet/trips",
                auth=True,
                json_body=payload,
                expected=(201,),
            )
            item_id = self.extract_id(data)
            if item_id:
                self.ctx.tripsheet_trips.append(item_id)

    def save_report(self) -> None:
        report = {
            "base_url": BASE_URL,
            "created": {
                "users": self.ctx.users,
                "drivers": self.ctx.drivers,
                "vehicles": self.ctx.vehicles,
                "tire_places": self.ctx.tire_places,
                "tires": self.ctx.tires,
                "tripsheets": self.ctx.tripsheets,
                "tripsheet_trips": self.ctx.tripsheet_trips,
            },
            "results": [r.__dict__ for r in self.results],
        }
        with open("auto_park_seed_only_report.json", "w", encoding="utf-8") as f:
            json.dump(report, f, ensure_ascii=False, indent=2)

    def print_summary(self) -> int:
        total = len(self.results)
        passed = sum(1 for r in self.results if r.success)
        failed = total - passed

        print("\n" + "=" * 110)
        print("AUTO PARK SEED ONLY SUMMARY")
        print("=" * 110)
        print(f"Base URL:                 {BASE_URL}")
        print(f"Total create attempts:    {total}")
        print(f"Successful creates:       {passed}")
        print(f"Failed creates:           {failed}")
        print("-" * 110)
        print(f"Users created:            {len(self.ctx.users)} -> {self.ctx.users}")
        print(f"Drivers created:          {len(self.ctx.drivers)} -> {self.ctx.drivers}")
        print(f"Vehicles created:         {len(self.ctx.vehicles)} -> {self.ctx.vehicles}")
        print(f"Tire places created:      {len(self.ctx.tire_places)} -> {self.ctx.tire_places}")
        print(f"Tires created:            {len(self.ctx.tires)} -> {self.ctx.tires}")
        print(f"Tripsheets created:       {len(self.ctx.tripsheets)} -> {self.ctx.tripsheets}")
        print(f"Tripsheet trips created:  {len(self.ctx.tripsheet_trips)} -> {self.ctx.tripsheet_trips}")
        print("-" * 110)

        for r in self.results:
            mark = "PASS" if r.success else "FAIL"
            code = r.status_code if r.status_code is not None else "-"
            print(f"[{mark}] {r.entity:<17} {r.name:<18} status={code:<4} id={str(r.id_value):<6}")
            if not r.success and r.error:
                print(f"       error: {r.error}")

        print("-" * 110)
        print("Saved report: auto_park_seed_only_report.json")
        print("=" * 110)
        return 0 if failed == 0 else 1

    def run(self) -> int:
        self.wait_api()
        self.login()

        self.seed_users(3)
        self.seed_drivers(3)
        self.seed_tire_places(3)
        self.seed_vehicles(3)
        self.seed_tires(3)
        self.seed_tripsheets(3)
        self.seed_tripsheet_trips(3)

        self.save_report()
        return self.print_summary()


if __name__ == "__main__":
    try:
        sys.exit(Seeder().run())
    except KeyboardInterrupt:
        print("\nInterrupted.")
        sys.exit(130)
