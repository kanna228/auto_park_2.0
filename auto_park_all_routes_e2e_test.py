#!/usr/bin/env python3
"""
auto_park_all_routes_e2e_test.py

End-to-end smoke/seed test for Auto Park backend.
It creates test data, checks the main API routes, leaves useful records in DB,
and writes a JSON report.

Install:
    pip install -r requirements_auto_park_api_tests.txt

Run from anywhere while docker compose backend is running:
    python auto_park_all_routes_e2e_test.py

Useful env:
    AUTO_PARK_BASE_URL=http://localhost:8080
    AUTO_PARK_DB_HOST=localhost
    AUTO_PARK_DB_PORT=5433
    AUTO_PARK_DB_USER=autopark_user
    AUTO_PARK_DB_PASSWORD=autopark_password
    AUTO_PARK_DB_NAME=auto_park
    AUTO_PARK_TEST_PASSWORD=admin123
    AUTO_PARK_TIMEOUT=20

Notes:
- The script uses DB only to create deterministic test users for every role
  and to insert one notification for notification routes.
- Business data is created mostly through API routes.
- It intentionally does NOT clean up all data, because the goal is to fill DB.
- DELETE routes are tested on separate disposable records.
"""
from __future__ import annotations

import json
import os
import random
import string
import sys
import tempfile
import time
from dataclasses import dataclass, asdict
from pathlib import Path
from typing import Any, Dict, Iterable, List, Optional, Tuple

import requests

try:
    import psycopg2
    import psycopg2.extras
except Exception:  # pragma: no cover
    psycopg2 = None

BASE_URL = os.getenv("AUTO_PARK_BASE_URL", "http://localhost:8080").rstrip("/")
TIMEOUT = int(os.getenv("AUTO_PARK_TIMEOUT", "20"))
TEST_PASSWORD = os.getenv("AUTO_PARK_TEST_PASSWORD", "admin123")

# bcrypt hash for password: admin123
# This is the same default password hash used by your admin migration.
DEFAULT_BCRYPT_HASH = "$2a$12$93CmzSxMMK.jo6YfoXs5G.rneemNhlMu6DHuYPV3QtX6OQa3JKLme"

DB_CONFIG = {
    "host": os.getenv("AUTO_PARK_DB_HOST", "localhost"),
    "port": int(os.getenv("AUTO_PARK_DB_PORT", "5433")),
    "dbname": os.getenv("AUTO_PARK_DB_NAME", "auto_park"),
    "user": os.getenv("AUTO_PARK_DB_USER", "autopark_user"),
    "password": os.getenv("AUTO_PARK_DB_PASSWORD", "autopark_password"),
}

DATE = os.getenv("AUTO_PARK_TEST_DATE", "2026-05-14")
DATE2 = os.getenv("AUTO_PARK_TEST_DATE_2", "2026-05-15")

ROLE_ADMIN = 1
ROLE_MANAGER = 2
ROLE_DISPATCHER = 3
ROLE_DUTY_MECHANIC = 4
ROLE_WAREHOUSE_MANAGER = 5

ROLE_USERS = {
    "admin": {
        "email": "apitest.admin@autopark.local",
        "iin": "900000000001",
        "role_id": ROLE_ADMIN,
        "first_name": "ApiTest",
        "last_name": "Admin",
    },
    "manager": {
        "email": "apitest.manager@autopark.local",
        "iin": "900000000002",
        "role_id": ROLE_MANAGER,
        "first_name": "ApiTest",
        "last_name": "Manager",
    },
    "dispatcher": {
        "email": "apitest.dispatcher@autopark.local",
        "iin": "900000000003",
        "role_id": ROLE_DISPATCHER,
        "first_name": "ApiTest",
        "last_name": "Dispatcher",
    },
    "mechanic": {
        "email": "apitest.mechanic@autopark.local",
        "iin": "900000000004",
        "role_id": ROLE_DUTY_MECHANIC,
        "first_name": "ApiTest",
        "last_name": "Mechanic",
    },
    "warehouse": {
        "email": "apitest.warehouse@autopark.local",
        "iin": "900000000005",
        "role_id": ROLE_WAREHOUSE_MANAGER,
        "first_name": "ApiTest",
        "last_name": "Warehouse",
    },
}

ALL_ROUTE_SIGNATURES: List[Tuple[str, str]] = [
    ("GET", "/healthz"),
    ("GET", "/api/tripsheet/health"),
    ("GET", "/api/fuel/health"),

    ("POST", "/api/users/login"),
    ("GET", "/api/users"),
    ("POST", "/api/users"),
    ("GET", "/api/users/{id}"),
    ("PUT", "/api/users/{id}"),
    ("DELETE", "/api/users/{id}"),
    ("GET", "/api/users/drivers"),
    ("POST", "/api/users/drivers"),
    ("GET", "/api/users/drivers/{id}"),
    ("PUT", "/api/users/drivers/{id}"),
    ("DELETE", "/api/users/drivers/{id}"),
    ("POST", "/api/users/drivers/{id}/photo"),
    ("PUT", "/api/users/drivers/{id}/photo"),
    ("DELETE", "/api/users/drivers/{id}/photo"),
    ("GET", "/api/users/mechanic-shifts"),
    ("POST", "/api/users/mechanic-shifts"),
    ("GET", "/api/users/mechanic-shifts/{id}"),
    ("PUT", "/api/users/mechanic-shifts/{id}"),
    ("PATCH", "/api/users/mechanic-shifts/{id}/activity"),
    ("DELETE", "/api/users/mechanic-shifts/{id}"),

    ("GET", "/api/vehicles/statuses"),
    ("GET", "/api/vehicles/status-history"),
    ("GET", "/api/vehicles/status-history/{id}"),
    ("GET", "/api/vehicles"),
    ("POST", "/api/vehicles"),
    ("GET", "/api/vehicles/{id}"),
    ("PUT", "/api/vehicles/{id}"),
    ("DELETE", "/api/vehicles/{id}"),
    ("POST", "/api/vehicles/{id}/photo"),
    ("PUT", "/api/vehicles/{id}/photo"),
    ("DELETE", "/api/vehicles/{id}/photo"),
    ("GET", "/api/vehicles/tire-places"),
    ("POST", "/api/vehicles/tire-places"),
    ("GET", "/api/vehicles/tire-places/{id}"),
    ("PUT", "/api/vehicles/tire-places/{id}"),
    ("DELETE", "/api/vehicles/tire-places/{id}"),
    ("GET", "/api/vehicles/tires"),
    ("POST", "/api/vehicles/tires"),
    ("GET", "/api/vehicles/tires/{id}"),
    ("PUT", "/api/vehicles/tires/{id}"),
    ("DELETE", "/api/vehicles/tires/{id}"),
    ("GET", "/api/vehicles/vehicle-tires/{vehicle_id}"),
    ("GET", "/api/vehicles/insurance"),
    ("POST", "/api/vehicles/insurance"),
    ("GET", "/api/vehicles/insurance/{id}"),
    ("PUT", "/api/vehicles/insurance/{id}"),
    ("DELETE", "/api/vehicles/insurance/{id}"),
    ("POST", "/api/vehicles/insurance/{id}/file"),
    ("DELETE", "/api/vehicles/insurance/{id}/file"),
    ("GET", "/api/vehicles/technical-inspections"),
    ("POST", "/api/vehicles/technical-inspections"),
    ("GET", "/api/vehicles/technical-inspections/{id}"),
    ("PUT", "/api/vehicles/technical-inspections/{id}"),
    ("DELETE", "/api/vehicles/technical-inspections/{id}"),
    ("POST", "/api/vehicles/technical-inspections/{id}/file"),
    ("DELETE", "/api/vehicles/technical-inspections/{id}/file"),

    ("GET", "/api/tripsheet"),
    ("POST", "/api/tripsheet"),
    ("GET", "/api/tripsheet/{id}"),
    ("PUT", "/api/tripsheet/{id}"),
    ("DELETE", "/api/tripsheet/{id}"),
    ("GET", "/api/tripsheet/trips"),
    ("POST", "/api/tripsheet/trips"),
    ("GET", "/api/tripsheet/trips/{id}"),
    ("PUT", "/api/tripsheet/trips/{id}"),
    ("DELETE", "/api/tripsheet/trips/{id}"),
    ("GET", "/api/tripsheet/trips/by-tripsheet/{tripsheet_id}"),

    ("GET", "/api/incidents/types"),
    ("GET", "/api/incidents"),
    ("POST", "/api/incidents"),
    ("GET", "/api/incidents/{id}"),
    ("PUT", "/api/incidents/{id}"),
    ("DELETE", "/api/incidents/{id}"),

    ("GET", "/api/fuel/refills"),
    ("POST", "/api/fuel/refills"),
    ("GET", "/api/fuel/refills/{id}"),
    ("PUT", "/api/fuel/refills/{id}"),
    ("DELETE", "/api/fuel/refills/{id}"),
    ("GET", "/api/fuel/refills/by-tripsheet/{tripsheet_id}"),
    ("GET", "/api/fuel/refills/by-vehicle/{vehicle_id}"),
    ("GET", "/api/fuel/reports/driver/{driver_id}"),
    ("GET", "/api/fuel/reports/vehicle/{vehicle_id}"),
    ("GET", "/api/fuel/reports/tripsheet/{tripsheet_id}"),

    ("GET", "/api/notifications"),
    ("GET", "/api/notifications/unread"),
    ("GET", "/api/notifications/unread/count"),
    ("PATCH", "/api/notifications/{id}/read"),
    ("PATCH", "/api/notifications/read-all"),
    ("GET", "/api/notifications/ws"),

    ("GET", "/api/warehouse/parts"),
    ("POST", "/api/warehouse/parts"),
    ("GET", "/api/warehouse/parts/{id}"),
    ("PUT", "/api/warehouse/parts/{id}"),
    ("DELETE", "/api/warehouse/parts/{id}"),
    ("GET", "/api/warehouse/part-request-statuses"),
    ("GET", "/api/warehouse/part-request-history"),
    ("GET", "/api/warehouse/part-requests"),
    ("POST", "/api/warehouse/part-requests"),
    ("GET", "/api/warehouse/part-requests/{id}"),
    ("PUT", "/api/warehouse/part-requests/{id}"),
    ("PATCH", "/api/warehouse/part-requests/{id}/status"),
    ("GET", "/api/warehouse/part-requests/{id}/history"),
    ("DELETE", "/api/warehouse/part-requests/{id}"),
    ("GET", "/api/warehouse/vehicle-part-installations"),
    ("POST", "/api/warehouse/vehicle-part-installations"),
    ("GET", "/api/warehouse/vehicle-part-installations/{id}"),
    ("PUT", "/api/warehouse/vehicle-part-installations/{id}"),
    ("PATCH", "/api/warehouse/vehicle-part-installations/{id}/activity"),
    ("DELETE", "/api/warehouse/vehicle-part-installations/{id}"),
    ("GET", "/api/warehouse/service-parts"),
    ("POST", "/api/warehouse/service-parts"),
    ("GET", "/api/warehouse/service-parts/{id}"),
    ("PUT", "/api/warehouse/service-parts/{id}"),
    ("DELETE", "/api/warehouse/service-parts/{id}"),
    ("GET", "/api/warehouse/service-types"),
    ("POST", "/api/warehouse/service-types"),
    ("GET", "/api/warehouse/service-types/{id}"),
    ("PUT", "/api/warehouse/service-types/{id}"),
    ("DELETE", "/api/warehouse/service-types/{id}"),
    ("GET", "/api/warehouse/vehicle-services"),
    ("POST", "/api/warehouse/vehicle-services"),
    ("GET", "/api/warehouse/vehicle-services/{id}"),
    ("PUT", "/api/warehouse/vehicle-services/{id}"),
    ("DELETE", "/api/warehouse/vehicle-services/{id}"),
]

SKIPPED_ROUTE_SIGNATURES = {("GET", "/api/notifications/ws")}


def rand_suffix(n: int = 8) -> str:
    chars = string.ascii_lowercase + string.digits
    return "".join(random.choice(chars) for _ in range(n))


def make_iin(prefix: str = "91") -> str:
    return (prefix + "".join(random.choice(string.digits) for _ in range(10)))[:12]


def now_ms() -> int:
    return int(time.time() * 1000)


@dataclass
class TestResult:
    name: str
    method: str
    path: str
    success: bool
    status_code: Optional[int]
    elapsed_ms: int
    expected: Tuple[int, ...]
    actor: str = "-"
    error: Optional[str] = None
    response_preview: Optional[str] = None
    skipped: bool = False


class AutoParkE2ETest:
    def __init__(self) -> None:
        self.session = requests.Session()
        self.results: List[TestResult] = []
        self.tested: set[Tuple[str, str]] = set()
        self.tokens: Dict[str, str] = {}
        self.users: Dict[str, int] = {}
        self.ids: Dict[str, Any] = {}
        self.suffix = rand_suffix()
        self.tmp_dir = tempfile.TemporaryDirectory(prefix="auto_park_api_test_")
        self.photo_path = Path(self.tmp_dir.name) / "test_photo.jpg"
        self.pdf_path = Path(self.tmp_dir.name) / "test_doc.pdf"
        self.photo_path.write_bytes(b"\xff\xd8\xff\xe0" + b"testjpg" * 100 + b"\xff\xd9")
        self.pdf_path.write_bytes(b"%PDF-1.4\n1 0 obj<</Type/Catalog>>endobj\n%%EOF")

    # ------------------------ DB seed helpers ------------------------
    def db_conn(self):
        if psycopg2 is None:
            raise RuntimeError(
                "psycopg2 is not installed. Run: pip install requests psycopg2-binary"
            )
        return psycopg2.connect(**DB_CONFIG)

    def seed_role_users_in_db(self) -> None:
        with self.db_conn() as conn:
            with conn.cursor() as cur:
                for actor, info in ROLE_USERS.items():
                    cur.execute(
                        """
                        INSERT INTO users (
                            email, first_name, last_name, middle_name, iin, phone, password, role_id, created_at, updated_at
                        ) VALUES (%s,%s,%s,NULL,%s,%s,%s,%s,NOW(),NOW())
                        ON CONFLICT (email) DO UPDATE SET
                            first_name = EXCLUDED.first_name,
                            last_name = EXCLUDED.last_name,
                            iin = EXCLUDED.iin,
                            phone = EXCLUDED.phone,
                            password = EXCLUDED.password,
                            role_id = EXCLUDED.role_id,
                            updated_at = NOW()
                        RETURNING id;
                        """,
                        (
                            info["email"],
                            info["first_name"],
                            info["last_name"],
                            info["iin"],
                            "+77000000000",
                            DEFAULT_BCRYPT_HASH,
                            info["role_id"],
                        ),
                    )
                    self.users[actor] = int(cur.fetchone()[0])
                conn.commit()

    def seed_notification_for_admin(self) -> Optional[int]:
        try:
            with self.db_conn() as conn:
                with conn.cursor() as cur:
                    cur.execute(
                        """
                        INSERT INTO notifications (user_id, notification_type_id, title, message, context, is_readed, created_at, updated_at)
                        VALUES (%s, 1, %s, %s, %s::jsonb, FALSE, NOW(), NOW())
                        RETURNING id;
                        """,
                        (
                            self.users.get("admin"),
                            "API test notification",
                            f"Smoke test notification {self.suffix}",
                            json.dumps({"source": "auto_park_all_routes_e2e_test", "suffix": self.suffix}),
                        ),
                    )
                    nid = int(cur.fetchone()[0])
                    conn.commit()
                    self.ids["notification_id"] = nid
                    return nid
        except Exception as e:
            self.add_result(
                "seed notification in DB",
                "DB",
                "notifications",
                False,
                None,
                0,
                (),
                actor="db",
                error=str(e),
            )
            return None

    # ------------------------ HTTP helpers ------------------------
    def headers(self, actor: str = "admin", json_content: bool = True) -> Dict[str, str]:
        h: Dict[str, str] = {}
        if json_content:
            h["Content-Type"] = "application/json"
        token = self.tokens.get(actor)
        if token:
            h["Authorization"] = f"Bearer {token}"
        return h

    def add_result(
        self,
        name: str,
        method: str,
        path: str,
        success: bool,
        status_code: Optional[int],
        elapsed_ms: int,
        expected: Tuple[int, ...],
        actor: str = "-",
        error: Optional[str] = None,
        response_preview: Optional[str] = None,
        skipped: bool = False,
        route_signature: Optional[Tuple[str, str]] = None,
    ) -> None:
        if route_signature:
            self.tested.add(route_signature)
        self.results.append(
            TestResult(
                name=name,
                method=method,
                path=path,
                success=success,
                status_code=status_code,
                elapsed_ms=elapsed_ms,
                expected=expected,
                actor=actor,
                error=error,
                response_preview=(response_preview or "")[:700] if response_preview else None,
                skipped=skipped,
            )
        )

    def request(
        self,
        name: str,
        method: str,
        path: str,
        *,
        actor: str = "admin",
        json_body: Optional[dict] = None,
        params: Optional[dict] = None,
        expected: Tuple[int, ...] = (200, 201),
        route_signature: Optional[Tuple[str, str]] = None,
        raise_on_fail: bool = True,
    ) -> Optional[requests.Response]:
        started = now_ms()
        try:
            resp = self.session.request(
                method=method,
                url=f"{BASE_URL}{path}",
                headers=self.headers(actor),
                json=json_body,
                params=params,
                timeout=TIMEOUT,
            )
            ok = resp.status_code in expected
            self.add_result(
                name,
                method,
                path,
                ok,
                resp.status_code,
                now_ms() - started,
                expected,
                actor=actor,
                error=None if ok else f"expected {expected}, got {resp.status_code}",
                response_preview=resp.text,
                route_signature=route_signature,
            )
            if not ok and raise_on_fail:
                raise AssertionError(f"{name}: expected {expected}, got {resp.status_code}: {resp.text[:500]}")
            return resp
        except Exception as e:
            if isinstance(e, AssertionError):
                if raise_on_fail:
                    raise
                return None
            self.add_result(
                name,
                method,
                path,
                False,
                None,
                now_ms() - started,
                expected,
                actor=actor,
                error=str(e),
                route_signature=route_signature,
            )
            if raise_on_fail:
                raise
            return None

    def upload_file(
        self,
        name: str,
        method: str,
        path: str,
        field_name: str,
        file_path: Path,
        *,
        actor: str = "admin",
        expected: Tuple[int, ...] = (200, 201),
        route_signature: Optional[Tuple[str, str]] = None,
        raise_on_fail: bool = True,
    ) -> Optional[requests.Response]:
        started = now_ms()
        try:
            with file_path.open("rb") as f:
                files = {field_name: (file_path.name, f, "application/octet-stream")}
                resp = self.session.request(
                    method=method,
                    url=f"{BASE_URL}{path}",
                    headers=self.headers(actor, json_content=False),
                    files=files,
                    timeout=TIMEOUT,
                )
            ok = resp.status_code in expected
            self.add_result(
                name,
                method,
                path,
                ok,
                resp.status_code,
                now_ms() - started,
                expected,
                actor=actor,
                error=None if ok else f"expected {expected}, got {resp.status_code}",
                response_preview=resp.text,
                route_signature=route_signature,
            )
            if not ok and raise_on_fail:
                raise AssertionError(f"{name}: expected {expected}, got {resp.status_code}: {resp.text[:500]}")
            return resp
        except Exception as e:
            if isinstance(e, AssertionError):
                if raise_on_fail:
                    raise
                return None
            self.add_result(
                name,
                method,
                path,
                False,
                None,
                now_ms() - started,
                expected,
                actor=actor,
                error=str(e),
                route_signature=route_signature,
            )
            if raise_on_fail:
                raise
            return None

    def skip(self, route_signature: Tuple[str, str], reason: str) -> None:
        method, path = route_signature
        self.add_result(
            f"SKIP {method} {path}",
            method,
            path,
            True,
            None,
            0,
            (),
            actor="-",
            error=reason,
            skipped=True,
            route_signature=route_signature,
        )

    @staticmethod
    def as_json(resp: Optional[requests.Response]) -> Any:
        if resp is None:
            return None
        try:
            return resp.json()
        except Exception:
            return None

    @staticmethod
    def nested(obj: Any, *keys: str) -> Any:
        cur = obj
        for key in keys:
            if not isinstance(cur, dict) or key not in cur:
                return None
            cur = cur[key]
        return cur

    def extract_id(self, obj: Any) -> Optional[int]:
        if isinstance(obj, dict):
            for key in ("id", "ID", "user_id", "driver_id", "vehicle_id", "tripsheet_id"):
                val = obj.get(key)
                if isinstance(val, int):
                    return val
            for key in ("data", "item"):
                found = self.extract_id(obj.get(key))
                if found is not None:
                    return found
            for val in obj.values():
                found = self.extract_id(val)
                if found is not None:
                    return found
        elif isinstance(obj, list):
            for item in obj:
                found = self.extract_id(item)
                if found is not None:
                    return found
        return None

    def extract_items(self, obj: Any) -> List[Any]:
        data = obj.get("data") if isinstance(obj, dict) else obj
        if isinstance(data, dict) and isinstance(data.get("items"), list):
            return data["items"]
        if isinstance(data, list):
            return data
        return []

    def get_id_from_response(self, resp: Optional[requests.Response]) -> Optional[int]:
        return self.extract_id(self.as_json(resp))

    def wait_api(self) -> None:
        deadline = time.time() + 90
        last_error = ""
        while time.time() < deadline:
            try:
                resp = self.request(
                    "wait healthz",
                    "GET",
                    "/healthz",
                    expected=(200,),
                    route_signature=("GET", "/healthz"),
                    raise_on_fail=False,
                )
                if resp is not None and resp.status_code == 200:
                    return
            except Exception as e:
                last_error = str(e)
            time.sleep(2)
        raise RuntimeError(f"API is not healthy at {BASE_URL}: {last_error}")

    def login(self, actor: str) -> None:
        info = ROLE_USERS[actor]
        resp = self.request(
            f"login {actor}",
            "POST",
            "/api/users/login",
            json_body={"email": info["email"], "password": TEST_PASSWORD, "iin": info["iin"]},
            expected=(200,),
            route_signature=("POST", "/api/users/login"),
        )
        data = self.as_json(resp)
        token = self.nested(data, "data", "token")
        user_id = self.nested(data, "data", "user_id")
        if not token:
            raise RuntimeError(f"token not found for {actor}: {data}")
        self.tokens[actor] = token
        if isinstance(user_id, int):
            self.users[actor] = user_id

    def login_all_roles(self) -> None:
        for actor in ROLE_USERS:
            self.login(actor)

    def lookup_first_id(self, actor: str, path: str, name: str, params: Optional[dict] = None) -> Optional[int]:
        resp = self.request(name, "GET", path, actor=actor, params=params, expected=(200,), raise_on_fail=False)
        items = self.extract_items(self.as_json(resp))
        if items:
            return self.extract_id(items[0])
        return None

    # ------------------------ Test sections ------------------------
    def test_health(self) -> None:
        self.request("healthz", "GET", "/healthz", expected=(200,), route_signature=("GET", "/healthz"), raise_on_fail=False)
        self.request("tripsheet health", "GET", "/api/tripsheet/health", expected=(200,), route_signature=("GET", "/api/tripsheet/health"), raise_on_fail=False)
        self.request("fuel health", "GET", "/api/fuel/health", expected=(200,), route_signature=("GET", "/api/fuel/health"), raise_on_fail=False)

    def test_users(self) -> None:
        self.request("list users", "GET", "/api/users", actor="admin", expected=(200,), route_signature=("GET", "/api/users"), raise_on_fail=False)
        admin_id = self.users["admin"]
        self.request("get test admin user", "GET", f"/api/users/{admin_id}", actor="admin", expected=(200,), route_signature=("GET", "/api/users/{id}"), raise_on_fail=False)

        # Disposable user for POST/PUT/DELETE route coverage.
        iin = make_iin("92")
        email = f"api.created.{self.suffix}@example.com"
        resp = self.request(
            "create disposable user",
            "POST",
            "/api/users",
            actor="admin",
            json_body={
                "email": email,
                "first_name": "Api",
                "last_name": "Created",
                "middle_name": "Test",
                "iin": iin,
                "phone": "+77001234567",
                "role_id": ROLE_MANAGER,
            },
            expected=(201,),
            route_signature=("POST", "/api/users"),
            raise_on_fail=False,
        )
        uid = self.get_id_from_response(resp)
        if uid:
            self.ids["created_user_delete_id"] = uid
            self.request(
                "update disposable user",
                "PUT",
                f"/api/users/{uid}",
                actor="admin",
                json_body={"first_name": "ApiUpdated", "phone": "+77009998877"},
                expected=(200,),
                route_signature=("PUT", "/api/users/{id}"),
                raise_on_fail=False,
            )
            self.request("delete disposable user", "DELETE", f"/api/users/{uid}", actor="admin", expected=(200,), route_signature=("DELETE", "/api/users/{id}"), raise_on_fail=False)

    def test_drivers(self) -> None:
        self.request("list drivers", "GET", "/api/users/drivers", actor="admin", expected=(200,), route_signature=("GET", "/api/users/drivers"), raise_on_fail=False)
        main_payload = {
            "iin": make_iin("93"),
            "name": "Main",
            "surname": f"Driver{self.suffix}",
            "middlename": "Smoke",
            "phone": "+77001112233",
            "mail": f"main.driver.{self.suffix}@example.com",
        }
        resp = self.request("create main driver", "POST", "/api/users/drivers", actor="admin", json_body=main_payload, expected=(201,), route_signature=("POST", "/api/users/drivers"), raise_on_fail=False)
        self.ids["driver_id"] = self.get_id_from_response(resp)

        disposable_payload = {
            "iin": make_iin("94"),
            "name": "Delete",
            "surname": f"Driver{self.suffix}",
            "middlename": "Smoke",
            "phone": "+77002223344",
            "mail": f"delete.driver.{self.suffix}@example.com",
        }
        resp2 = self.request("create disposable driver", "POST", "/api/users/drivers", actor="admin", json_body=disposable_payload, expected=(201,), route_signature=("POST", "/api/users/drivers"), raise_on_fail=False)
        disposable_id = self.get_id_from_response(resp2)
        driver_id = self.ids.get("driver_id")
        if driver_id:
            self.request("get driver", "GET", f"/api/users/drivers/{driver_id}", actor="admin", expected=(200,), route_signature=("GET", "/api/users/drivers/{id}"), raise_on_fail=False)
            self.request(
                "update driver",
                "PUT",
                f"/api/users/drivers/{driver_id}",
                actor="admin",
                json_body={"phone": "+77007776655", "mail": f"main.driver.updated.{self.suffix}@example.com"},
                expected=(200,),
                route_signature=("PUT", "/api/users/drivers/{id}"),
                raise_on_fail=False,
            )
            self.upload_file("upload driver photo", "POST", f"/api/users/drivers/{driver_id}/photo", "photo", self.photo_path, actor="admin", expected=(200,), route_signature=("POST", "/api/users/drivers/{id}/photo"), raise_on_fail=False)
            self.upload_file("update driver photo", "PUT", f"/api/users/drivers/{driver_id}/photo", "photo", self.photo_path, actor="admin", expected=(200,), route_signature=("PUT", "/api/users/drivers/{id}/photo"), raise_on_fail=False)
            self.request("delete driver photo", "DELETE", f"/api/users/drivers/{driver_id}/photo", actor="admin", expected=(200,), route_signature=("DELETE", "/api/users/drivers/{id}/photo"), raise_on_fail=False)
        if disposable_id:
            self.request("delete disposable driver", "DELETE", f"/api/users/drivers/{disposable_id}", actor="admin", expected=(200,), route_signature=("DELETE", "/api/users/drivers/{id}"), raise_on_fail=False)

    def test_mechanic_shifts(self) -> None:
        mech_id = self.users["mechanic"]
        resp = self.request(
            "create mechanic shift",
            "POST",
            "/api/users/mechanic-shifts",
            actor="admin",
            json_body={
                "user_id": mech_id,
                "shift_date": DATE,
                "time_from": "08:00",
                "time_to": "20:00",
                "comment": f"API test shift {self.suffix}",
                "is_active": True,
            },
            expected=(201,),
            route_signature=("POST", "/api/users/mechanic-shifts"),
            raise_on_fail=False,
        )
        self.ids["mechanic_shift_id"] = self.get_id_from_response(resp)
        shift_id = self.ids.get("mechanic_shift_id")
        self.request("list mechanic shifts", "GET", "/api/users/mechanic-shifts", actor="admin", params={"user_id": mech_id, "date_from": DATE, "date_to": DATE2, "is_active": "true", "sort_by": "shift_date", "order": "desc"}, expected=(200,), route_signature=("GET", "/api/users/mechanic-shifts"), raise_on_fail=False)
        if shift_id:
            self.request("get mechanic shift", "GET", f"/api/users/mechanic-shifts/{shift_id}", actor="admin", expected=(200,), route_signature=("GET", "/api/users/mechanic-shifts/{id}"), raise_on_fail=False)
            self.request("update mechanic shift", "PUT", f"/api/users/mechanic-shifts/{shift_id}", actor="admin", json_body={"comment": f"updated shift {self.suffix}", "time_to": "21:00"}, expected=(200,), route_signature=("PUT", "/api/users/mechanic-shifts/{id}"), raise_on_fail=False)
            self.request("update mechanic shift activity", "PATCH", f"/api/users/mechanic-shifts/{shift_id}/activity", actor="admin", json_body={"is_active": True}, expected=(200,), route_signature=("PATCH", "/api/users/mechanic-shifts/{id}/activity"), raise_on_fail=False)
        # disposable for delete route
        resp2 = self.request("create disposable mechanic shift", "POST", "/api/users/mechanic-shifts", actor="admin", json_body={"user_id": mech_id, "shift_date": DATE2, "time_from": "09:00", "time_to": "10:00", "comment": "delete me"}, expected=(201,), raise_on_fail=False)
        sid2 = self.get_id_from_response(resp2)
        if sid2:
            self.request("delete disposable mechanic shift", "DELETE", f"/api/users/mechanic-shifts/{sid2}", actor="admin", expected=(200,), route_signature=("DELETE", "/api/users/mechanic-shifts/{id}"), raise_on_fail=False)

    def get_status_ids(self) -> Dict[str, int]:
        resp = self.request("list vehicle statuses", "GET", "/api/vehicles/statuses", actor="admin", expected=(200,), route_signature=("GET", "/api/vehicles/statuses"), raise_on_fail=False)
        out: Dict[str, int] = {}
        for item in self.extract_items(self.as_json(resp)):
            if isinstance(item, dict) and isinstance(item.get("id"), int):
                out[str(item.get("name"))] = int(item["id"])
        if not out:
            out = {"В использовании": 1, "На ТО": 2, "Не используется": 3, "На ремонте": 4, "Списан": 5}
        return out

    def vehicle_payload(self, state_number: str, vin: str, driver_id: int, status_id: int, brand: str = "Toyota Camry") -> dict:
        return {
            "board_number": f"B-{rand_suffix(4).upper()}",
            "technical_passport_number": f"TP-{rand_suffix(8).upper()}",
            "state_number": state_number,
            "vin": vin,
            "brand_model": brand,
            "manufacture_year": 2022,
            "received_date": f"{DATE}T00:00:00Z",
            "empty_weight_kg": 1500.0,
            "max_weight_kg": 2200.0,
            "engine_volume_cc": 2500,
            "insurance_policy_number": f"POL-{rand_suffix(6).upper()}",
            "insurance_expiry_date": "2027-05-14T00:00:00Z",
            "mileage": 12000,
            "current_fuel": 35.5,
            "status_id": status_id,
            "drivers_ids": [driver_id],
        }

    def test_vehicles(self) -> None:
        driver_id = self.ids.get("driver_id")
        if not driver_id:
            self.add_result("vehicles skipped - driver missing", "SKIP", "/api/vehicles", True, None, 0, (), skipped=True)
            return
        statuses = self.get_status_ids()
        status_active = statuses.get("В использовании", 1)
        status_repair = statuses.get("На ремонте", status_active)
        self.ids["vehicle_status_active_id"] = status_active
        self.ids["vehicle_status_second_id"] = status_repair

        payload = self.vehicle_payload(f"777{self.suffix[:3].upper()}01", f"VIN{self.suffix.upper()}000001", driver_id, status_active)
        resp = self.request("create vehicle", "POST", "/api/vehicles", actor="admin", json_body=payload, expected=(201,), route_signature=("POST", "/api/vehicles"), raise_on_fail=False)
        self.ids["vehicle_id"] = self.get_id_from_response(resp)
        vehicle_id = self.ids.get("vehicle_id")
        self.ids["vehicle_plate"] = payload["state_number"]
        self.ids["vehicle_brand"] = payload["brand_model"]

        self.request("list vehicles", "GET", "/api/vehicles", actor="admin", params={"status_id": status_active, "limit": 20, "offset": 0}, expected=(200,), route_signature=("GET", "/api/vehicles"), raise_on_fail=False)
        if vehicle_id:
            self.request("get vehicle passport", "GET", f"/api/vehicles/{vehicle_id}", actor="admin", expected=(200,), route_signature=("GET", "/api/vehicles/{id}"), raise_on_fail=False)
            upd = dict(payload)
            upd["brand_model"] = "Toyota Camry Restyled"
            upd["mileage"] = 13000
            upd["status_id"] = status_repair
            self.request("update vehicle", "PUT", f"/api/vehicles/{vehicle_id}", actor="admin", json_body=upd, expected=(200,), route_signature=("PUT", "/api/vehicles/{id}"), raise_on_fail=False)
            self.upload_file("upload vehicle photo", "POST", f"/api/vehicles/{vehicle_id}/photo", "photo", self.photo_path, actor="admin", expected=(200,), route_signature=("POST", "/api/vehicles/{id}/photo"), raise_on_fail=False)
            self.upload_file("update vehicle photo", "PUT", f"/api/vehicles/{vehicle_id}/photo", "photo", self.photo_path, actor="admin", expected=(200,), route_signature=("PUT", "/api/vehicles/{id}/photo"), raise_on_fail=False)
            self.request("delete vehicle photo", "DELETE", f"/api/vehicles/{vehicle_id}/photo", actor="admin", expected=(200,), route_signature=("DELETE", "/api/vehicles/{id}/photo"), raise_on_fail=False)
            self.request("list vehicle status history", "GET", "/api/vehicles/status-history", actor="admin", params={"vehicle_id": vehicle_id, "sort_by": "start_date", "order": "desc"}, expected=(200,), route_signature=("GET", "/api/vehicles/status-history"), raise_on_fail=False)
            hist_resp = self.request("list vehicle status history find id", "GET", "/api/vehicles/status-history", actor="admin", params={"vehicle_id": vehicle_id, "limit": 1}, expected=(200,), raise_on_fail=False)
            items = self.extract_items(self.as_json(hist_resp))
            if items:
                hist_id = self.extract_id(items[0])
                if hist_id:
                    self.ids["vehicle_status_history_id"] = hist_id
                    self.request("get vehicle status history", "GET", f"/api/vehicles/status-history/{hist_id}", actor="admin", expected=(200,), route_signature=("GET", "/api/vehicles/status-history/{id}"), raise_on_fail=False)

        # disposable vehicle for DELETE route
        disposable = self.vehicle_payload(f"888{self.suffix[:3].upper()}02", f"VIN{self.suffix.upper()}000002", driver_id, status_active, brand="Delete Vehicle")
        r2 = self.request("create disposable vehicle", "POST", "/api/vehicles", actor="admin", json_body=disposable, expected=(201,), raise_on_fail=False)
        did = self.get_id_from_response(r2)
        if did:
            self.request("delete disposable vehicle", "DELETE", f"/api/vehicles/{did}", actor="admin", expected=(200,), route_signature=("DELETE", "/api/vehicles/{id}"), raise_on_fail=False)

    def test_vehicle_extra_entities(self) -> None:
        vehicle_id = self.ids.get("vehicle_id")
        if not vehicle_id:
            return
        # Tire places
        self.request("list tire places", "GET", "/api/vehicles/tire-places", actor="admin", expected=(200,), route_signature=("GET", "/api/vehicles/tire-places"), raise_on_fail=False)
        resp = self.request("create tire place", "POST", "/api/vehicles/tire-places", actor="admin", json_body={"name": f"API tire place {self.suffix}"}, expected=(201,), route_signature=("POST", "/api/vehicles/tire-places"), raise_on_fail=False)
        self.ids["tire_place_id"] = self.get_id_from_response(resp)
        place_id = self.ids.get("tire_place_id")
        if place_id:
            self.request("get tire place", "GET", f"/api/vehicles/tire-places/{place_id}", actor="admin", expected=(200,), route_signature=("GET", "/api/vehicles/tire-places/{id}"), raise_on_fail=False)
            self.request("update tire place", "PUT", f"/api/vehicles/tire-places/{place_id}", actor="admin", json_body={"name": f"API tire place updated {self.suffix}"}, expected=(200,), route_signature=("PUT", "/api/vehicles/tire-places/{id}"), raise_on_fail=False)
        resp2 = self.request("create disposable tire place", "POST", "/api/vehicles/tire-places", actor="admin", json_body={"name": f"API tire place del {self.suffix}"}, expected=(201,), raise_on_fail=False)
        del_place_id = self.get_id_from_response(resp2)
        if del_place_id:
            self.request("delete disposable tire place", "DELETE", f"/api/vehicles/tire-places/{del_place_id}", actor="admin", expected=(200,), route_signature=("DELETE", "/api/vehicles/tire-places/{id}"), raise_on_fail=False)

        # Tires
        if place_id:
            tire_payload = {"place_id": place_id, "vehicle_id": vehicle_id, "tire": f"Michelin API {self.suffix}", "mileage": 1000, "max_usage": 60000}
            resp = self.request("create tire", "POST", "/api/vehicles/tires", actor="admin", json_body=tire_payload, expected=(201,), route_signature=("POST", "/api/vehicles/tires"), raise_on_fail=False)
            self.ids["tire_id"] = self.get_id_from_response(resp)
            tire_id = self.ids.get("tire_id")
            self.request("list tires", "GET", "/api/vehicles/tires", actor="admin", params={"vehicle_id": vehicle_id}, expected=(200,), route_signature=("GET", "/api/vehicles/tires"), raise_on_fail=False)
            if tire_id:
                self.request("get tire", "GET", f"/api/vehicles/tires/{tire_id}", actor="admin", expected=(200,), route_signature=("GET", "/api/vehicles/tires/{id}"), raise_on_fail=False)
                self.request("update tire", "PUT", f"/api/vehicles/tires/{tire_id}", actor="admin", json_body={**tire_payload, "tire": f"Michelin API Updated {self.suffix}", "mileage": 1500}, expected=(200,), route_signature=("PUT", "/api/vehicles/tires/{id}"), raise_on_fail=False)
                self.request("get vehicle tires", "GET", f"/api/vehicles/vehicle-tires/{vehicle_id}", actor="admin", expected=(200,), route_signature=("GET", "/api/vehicles/vehicle-tires/{vehicle_id}"), raise_on_fail=False)
            resp3 = self.request("create disposable tire", "POST", "/api/vehicles/tires", actor="admin", json_body={**tire_payload, "tire": f"Delete Tire {self.suffix}"}, expected=(201,), raise_on_fail=False)
            del_tire_id = self.get_id_from_response(resp3)
            if del_tire_id:
                self.request("delete disposable tire", "DELETE", f"/api/vehicles/tires/{del_tire_id}", actor="admin", expected=(200,), route_signature=("DELETE", "/api/vehicles/tires/{id}"), raise_on_fail=False)

        # Insurance
        ins_payload = {"vehicle_id": vehicle_id, "name": f"Insurance API {self.suffix}", "start_date": DATE, "end_date": "2027-05-14", "is_active": True}
        resp = self.request("create insurance", "POST", "/api/vehicles/insurance", actor="admin", json_body=ins_payload, expected=(201,), route_signature=("POST", "/api/vehicles/insurance"), raise_on_fail=False)
        self.ids["insurance_id"] = self.get_id_from_response(resp)
        ins_id = self.ids.get("insurance_id")
        self.request("list insurance", "GET", "/api/vehicles/insurance", actor="admin", params={"vehicle_id": vehicle_id}, expected=(200,), route_signature=("GET", "/api/vehicles/insurance"), raise_on_fail=False)
        if ins_id:
            self.request("get insurance", "GET", f"/api/vehicles/insurance/{ins_id}", actor="admin", expected=(200,), route_signature=("GET", "/api/vehicles/insurance/{id}"), raise_on_fail=False)
            self.request("update insurance", "PUT", f"/api/vehicles/insurance/{ins_id}", actor="admin", json_body={**ins_payload, "name": f"Insurance API Updated {self.suffix}"}, expected=(200,), route_signature=("PUT", "/api/vehicles/insurance/{id}"), raise_on_fail=False)
            self.upload_file("upload insurance file", "POST", f"/api/vehicles/insurance/{ins_id}/file", "file", self.pdf_path, actor="admin", expected=(200,), route_signature=("POST", "/api/vehicles/insurance/{id}/file"), raise_on_fail=False)
            self.request("delete insurance file", "DELETE", f"/api/vehicles/insurance/{ins_id}/file", actor="admin", expected=(200,), route_signature=("DELETE", "/api/vehicles/insurance/{id}/file"), raise_on_fail=False)
        resp_del = self.request("create disposable insurance", "POST", "/api/vehicles/insurance", actor="admin", json_body={**ins_payload, "name": f"Insurance DELETE {self.suffix}"}, expected=(201,), raise_on_fail=False)
        ins_del_id = self.get_id_from_response(resp_del)
        if ins_del_id:
            self.request("delete disposable insurance", "DELETE", f"/api/vehicles/insurance/{ins_del_id}", actor="admin", expected=(200,), route_signature=("DELETE", "/api/vehicles/insurance/{id}"), raise_on_fail=False)

        # Technical inspection
        ti_payload = {"vehicle_id": vehicle_id, "name": f"TechInspection API {self.suffix}", "start_date": DATE, "end_date": "2027-05-14", "is_active": True}
        resp = self.request("create technical inspection", "POST", "/api/vehicles/technical-inspections", actor="admin", json_body=ti_payload, expected=(201,), route_signature=("POST", "/api/vehicles/technical-inspections"), raise_on_fail=False)
        self.ids["technical_inspection_id"] = self.get_id_from_response(resp)
        ti_id = self.ids.get("technical_inspection_id")
        self.request("list technical inspections", "GET", "/api/vehicles/technical-inspections", actor="admin", params={"vehicle_id": vehicle_id}, expected=(200,), route_signature=("GET", "/api/vehicles/technical-inspections"), raise_on_fail=False)
        if ti_id:
            self.request("get technical inspection", "GET", f"/api/vehicles/technical-inspections/{ti_id}", actor="admin", expected=(200,), route_signature=("GET", "/api/vehicles/technical-inspections/{id}"), raise_on_fail=False)
            self.request("update technical inspection", "PUT", f"/api/vehicles/technical-inspections/{ti_id}", actor="admin", json_body={**ti_payload, "name": f"TechInspection API Updated {self.suffix}"}, expected=(200,), route_signature=("PUT", "/api/vehicles/technical-inspections/{id}"), raise_on_fail=False)
            self.upload_file("upload technical inspection file", "POST", f"/api/vehicles/technical-inspections/{ti_id}/file", "file", self.pdf_path, actor="admin", expected=(200,), route_signature=("POST", "/api/vehicles/technical-inspections/{id}/file"), raise_on_fail=False)
            self.request("delete technical inspection file", "DELETE", f"/api/vehicles/technical-inspections/{ti_id}/file", actor="admin", expected=(200,), route_signature=("DELETE", "/api/vehicles/technical-inspections/{id}/file"), raise_on_fail=False)
        resp_del = self.request("create disposable technical inspection", "POST", "/api/vehicles/technical-inspections", actor="admin", json_body={**ti_payload, "name": f"TechInspection DELETE {self.suffix}"}, expected=(201,), raise_on_fail=False)
        ti_del_id = self.get_id_from_response(resp_del)
        if ti_del_id:
            self.request("delete disposable technical inspection", "DELETE", f"/api/vehicles/technical-inspections/{ti_del_id}", actor="admin", expected=(200,), route_signature=("DELETE", "/api/vehicles/technical-inspections/{id}"), raise_on_fail=False)

    def test_tripsheets(self) -> None:
        vehicle_id = self.ids.get("vehicle_id")
        driver_id = self.ids.get("driver_id")
        if not vehicle_id or not driver_id:
            return
        plate = self.ids.get("vehicle_plate", f"777API{self.suffix[:2].upper()}")
        brand = self.ids.get("vehicle_brand", "Toyota Camry")
        payload = {
            "tripsheet_number": f"TS-{self.suffix.upper()}",
            "tripsheet_date": DATE,
            "vehicle_id": vehicle_id,
            "vehicle_brand": brand,
            "vehicle_plate_number": plate,
            "driver_last_name": f"Driver{self.suffix}",
            "driver_first_name": "Main",
            "driver_middle_name": "Smoke",
            "driver_id": driver_id,
            "start_time": f"{DATE}T08:00:00Z",
            "end_time": f"{DATE}T18:00:00Z",
            "mileage_start": 13000,
            "mileage_end": 13120,
            "fuel_start": 30,
            "fuel_issued": 15,
            "fuel_consumption_theoretical": 10,
            "fuel_consumption_actual": 11,
            "status_id": 1,
        }
        resp = self.request("create tripsheet", "POST", "/api/tripsheet", actor="admin", json_body=payload, expected=(201,), route_signature=("POST", "/api/tripsheet"), raise_on_fail=False)
        self.ids["tripsheet_id"] = self.get_id_from_response(resp)
        ts_id = self.ids.get("tripsheet_id")
        self.request("list tripsheets", "GET", "/api/tripsheet", actor="admin", params={"vehicle_id": vehicle_id, "driver_id": driver_id}, expected=(200,), route_signature=("GET", "/api/tripsheet"), raise_on_fail=False)
        if ts_id:
            self.request("get tripsheet", "GET", f"/api/tripsheet/{ts_id}", actor="admin", expected=(200,), route_signature=("GET", "/api/tripsheet/{id}"), raise_on_fail=False)
            self.request("update tripsheet", "PUT", f"/api/tripsheet/{ts_id}", actor="admin", json_body={**payload, "status_id": 2, "fuel_consumption_actual": 12}, expected=(200,), route_signature=("PUT", "/api/tripsheet/{id}"), raise_on_fail=False)

        resp2 = self.request("create disposable tripsheet", "POST", "/api/tripsheet", actor="admin", json_body={**payload, "tripsheet_number": f"TS-DEL-{self.suffix.upper()}"}, expected=(201,), raise_on_fail=False)
        del_id = self.get_id_from_response(resp2)
        if del_id:
            self.request("delete disposable tripsheet", "DELETE", f"/api/tripsheet/{del_id}", actor="admin", expected=(200,), route_signature=("DELETE", "/api/tripsheet/{id}"), raise_on_fail=False)

    def test_tripsheet_trips(self) -> None:
        ts_id = self.ids.get("tripsheet_id")
        if not ts_id:
            return
        payload = {"tripsheet_id": ts_id, "route_description": f"Garage -> Warehouse -> Garage {self.suffix}", "start_time": f"{DATE}T09:00:00Z", "end_time": f"{DATE}T10:30:00Z", "distance_passed": 42, "status_id": 1}
        resp = self.request("create tripsheet trip", "POST", "/api/tripsheet/trips", actor="admin", json_body=payload, expected=(201,), route_signature=("POST", "/api/tripsheet/trips"), raise_on_fail=False)
        self.ids["tripsheet_trip_id"] = self.get_id_from_response(resp)
        trip_id = self.ids.get("tripsheet_trip_id")
        self.request("list tripsheet trips", "GET", "/api/tripsheet/trips", actor="admin", expected=(200,), route_signature=("GET", "/api/tripsheet/trips"), raise_on_fail=False)
        self.request("list trips by tripsheet", "GET", f"/api/tripsheet/trips/by-tripsheet/{ts_id}", actor="admin", expected=(200,), route_signature=("GET", "/api/tripsheet/trips/by-tripsheet/{tripsheet_id}"), raise_on_fail=False)
        if trip_id:
            self.request("get tripsheet trip", "GET", f"/api/tripsheet/trips/{trip_id}", actor="admin", expected=(200,), route_signature=("GET", "/api/tripsheet/trips/{id}"), raise_on_fail=False)
            self.request("update tripsheet trip", "PUT", f"/api/tripsheet/trips/{trip_id}", actor="admin", json_body={**payload, "status_id": 2, "distance_passed": 55}, expected=(200,), route_signature=("PUT", "/api/tripsheet/trips/{id}"), raise_on_fail=False)
        resp2 = self.request("create disposable tripsheet trip", "POST", "/api/tripsheet/trips", actor="admin", json_body={**payload, "route_description": f"Delete trip {self.suffix}"}, expected=(201,), raise_on_fail=False)
        del_id = self.get_id_from_response(resp2)
        if del_id:
            self.request("delete disposable tripsheet trip", "DELETE", f"/api/tripsheet/trips/{del_id}", actor="admin", expected=(200,), route_signature=("DELETE", "/api/tripsheet/trips/{id}"), raise_on_fail=False)

    def test_incidents(self) -> None:
        vehicle_id = self.ids.get("vehicle_id")
        driver_id = self.ids.get("driver_id")
        shift_id = self.ids.get("mechanic_shift_id")
        mechanic_id = self.users.get("mechanic")
        if not all([vehicle_id, driver_id, shift_id, mechanic_id]):
            return
        self.request("list incident types", "GET", "/api/incidents/types", actor="admin", expected=(200,), route_signature=("GET", "/api/incidents/types"), raise_on_fail=False)
        type_id = self.lookup_first_id("admin", "/api/incidents/types", "lookup incident type") or 1
        payload = {"incident_type_id": type_id, "vehicle_id": vehicle_id, "driver_id": driver_id, "mechanic_id": mechanic_id, "mechanic_shift_id": shift_id, "tripsheet_id": self.ids.get("tripsheet_id"), "date": DATE, "time": "12:30", "location": "API test garage", "description": f"API test incident {self.suffix}"}
        resp = self.request("create incident", "POST", "/api/incidents", actor="admin", json_body=payload, expected=(201,), route_signature=("POST", "/api/incidents"), raise_on_fail=False)
        self.ids["incident_id"] = self.get_id_from_response(resp)
        inc_id = self.ids.get("incident_id")
        self.request("list incidents", "GET", "/api/incidents", actor="admin", params={"vehicle_id": vehicle_id, "driver_id": driver_id, "mechanic_shift_id": shift_id}, expected=(200,), route_signature=("GET", "/api/incidents"), raise_on_fail=False)
        if inc_id:
            self.request("get incident", "GET", f"/api/incidents/{inc_id}", actor="admin", expected=(200,), route_signature=("GET", "/api/incidents/{id}"), raise_on_fail=False)
            self.request("update incident", "PUT", f"/api/incidents/{inc_id}", actor="admin", json_body={**payload, "description": f"Updated API test incident {self.suffix}"}, expected=(200,), route_signature=("PUT", "/api/incidents/{id}"), raise_on_fail=False)
        resp2 = self.request("create disposable incident", "POST", "/api/incidents", actor="admin", json_body={**payload, "description": f"Delete incident {self.suffix}"}, expected=(201,), raise_on_fail=False)
        del_id = self.get_id_from_response(resp2)
        if del_id:
            self.request("delete disposable incident", "DELETE", f"/api/incidents/{del_id}", actor="admin", expected=(200,), route_signature=("DELETE", "/api/incidents/{id}"), raise_on_fail=False)

    def test_fuel(self) -> None:
        ts_id = self.ids.get("tripsheet_id")
        vehicle_id = self.ids.get("vehicle_id")
        driver_id = self.ids.get("driver_id")
        if not ts_id or not vehicle_id:
            return
        payload = {"tripsheet_id": ts_id, "vehicle_id": vehicle_id, "fuel_amount": 25.5, "date": DATE, "time": "14:30:00", "location": "API Test Fuel Station"}
        resp = self.request("create fuel refill", "POST", "/api/fuel/refills", actor="admin", json_body=payload, expected=(201,), route_signature=("POST", "/api/fuel/refills"), raise_on_fail=False)
        self.ids["fuel_refill_id"] = self.get_id_from_response(resp)
        fr_id = self.ids.get("fuel_refill_id")
        self.request("list fuel refills", "GET", "/api/fuel/refills", actor="admin", params={"vehicle_id": vehicle_id, "tripsheet_id": ts_id, "date_from": DATE, "date_to": DATE2}, expected=(200,), route_signature=("GET", "/api/fuel/refills"), raise_on_fail=False)
        self.request("fuel by tripsheet", "GET", f"/api/fuel/refills/by-tripsheet/{ts_id}", actor="admin", expected=(200,), route_signature=("GET", "/api/fuel/refills/by-tripsheet/{tripsheet_id}"), raise_on_fail=False)
        self.request("fuel by vehicle", "GET", f"/api/fuel/refills/by-vehicle/{vehicle_id}", actor="admin", expected=(200,), route_signature=("GET", "/api/fuel/refills/by-vehicle/{vehicle_id}"), raise_on_fail=False)
        if fr_id:
            self.request("get fuel refill", "GET", f"/api/fuel/refills/{fr_id}", actor="admin", expected=(200,), route_signature=("GET", "/api/fuel/refills/{id}"), raise_on_fail=False)
            self.request("update fuel refill", "PUT", f"/api/fuel/refills/{fr_id}", actor="admin", json_body={**payload, "fuel_amount": 30.0, "location": "Updated Fuel Station"}, expected=(200,), route_signature=("PUT", "/api/fuel/refills/{id}"), raise_on_fail=False)
        resp2 = self.request("create disposable fuel refill", "POST", "/api/fuel/refills", actor="admin", json_body={**payload, "fuel_amount": 3.0}, expected=(201,), raise_on_fail=False)
        del_id = self.get_id_from_response(resp2)
        if del_id:
            self.request("delete disposable fuel refill", "DELETE", f"/api/fuel/refills/{del_id}", actor="admin", expected=(200,), route_signature=("DELETE", "/api/fuel/refills/{id}"), raise_on_fail=False)

        if driver_id:
            self.request("fuel report by driver", "GET", f"/api/fuel/reports/driver/{driver_id}", actor="admin", params={"format": "pdf", "date_from": DATE, "date_to": DATE2}, expected=(200,), route_signature=("GET", "/api/fuel/reports/driver/{driver_id}"), raise_on_fail=False)
        self.request("fuel report by vehicle", "GET", f"/api/fuel/reports/vehicle/{vehicle_id}", actor="admin", params={"format": "pdf", "date_from": DATE, "date_to": DATE2}, expected=(200,), route_signature=("GET", "/api/fuel/reports/vehicle/{vehicle_id}"), raise_on_fail=False)
        self.request("fuel report by tripsheet", "GET", f"/api/fuel/reports/tripsheet/{ts_id}", actor="admin", params={"format": "pdf", "date_from": DATE, "date_to": DATE2}, expected=(200,), route_signature=("GET", "/api/fuel/reports/tripsheet/{tripsheet_id}"), raise_on_fail=False)

    def test_notifications(self) -> None:
        self.seed_notification_for_admin()
        self.request("list notifications", "GET", "/api/notifications", actor="admin", expected=(200,), route_signature=("GET", "/api/notifications"), raise_on_fail=False)
        self.request("list unread notifications", "GET", "/api/notifications/unread", actor="admin", expected=(200,), route_signature=("GET", "/api/notifications/unread"), raise_on_fail=False)
        self.request("count unread notifications", "GET", "/api/notifications/unread/count", actor="admin", expected=(200,), route_signature=("GET", "/api/notifications/unread/count"), raise_on_fail=False)
        nid = self.ids.get("notification_id")
        if nid:
            self.request("mark notification read", "PATCH", f"/api/notifications/{nid}/read", actor="admin", expected=(200,), route_signature=("PATCH", "/api/notifications/{id}/read"), raise_on_fail=False)
        self.request("mark all notifications read", "PATCH", "/api/notifications/read-all", actor="admin", expected=(200,), route_signature=("PATCH", "/api/notifications/read-all"), raise_on_fail=False)
        self.skip(("GET", "/api/notifications/ws"), "WebSocket route is not tested by requests-based smoke script")

    def test_warehouse_parts_and_requests(self) -> None:
        # Parts
        self.request("list warehouse parts", "GET", "/api/warehouse/parts", actor="warehouse", expected=(200,), route_signature=("GET", "/api/warehouse/parts"), raise_on_fail=False)
        part_payload = {"part_id": f"API-PART-{self.suffix.upper()}", "name": f"API Test Part {self.suffix}", "start_quantity": 100, "category": "api_test", "dimensions": "10x10", "manufacturer": "API", "is_consumable": True}
        resp = self.request("create warehouse part", "POST", "/api/warehouse/parts", actor="warehouse", json_body=part_payload, expected=(201,), route_signature=("POST", "/api/warehouse/parts"), raise_on_fail=False)
        self.ids["warehouse_part_id"] = self.get_id_from_response(resp)
        part_id = self.ids.get("warehouse_part_id")
        if part_id:
            self.request("get warehouse part", "GET", f"/api/warehouse/parts/{part_id}", actor="warehouse", expected=(200,), route_signature=("GET", "/api/warehouse/parts/{id}"), raise_on_fail=False)
            self.request("update warehouse part", "PUT", f"/api/warehouse/parts/{part_id}", actor="warehouse", json_body={"name": f"API Test Part Updated {self.suffix}", "quantity": 120, "category": "api_test", "dimensions": "10x20", "manufacturer": "API", "is_consumable": True}, expected=(200,), route_signature=("PUT", "/api/warehouse/parts/{id}"), raise_on_fail=False)
        # disposable part for delete
        resp_del = self.request("create disposable warehouse part", "POST", "/api/warehouse/parts", actor="warehouse", json_body={**part_payload, "part_id": f"API-PART-DEL-{self.suffix.upper()}", "name": f"API Test Part Delete {self.suffix}"}, expected=(201,), raise_on_fail=False)
        part_del_id = self.get_id_from_response(resp_del)
        if part_del_id:
            self.request("delete disposable warehouse part", "DELETE", f"/api/warehouse/parts/{part_del_id}", actor="admin", expected=(200,), route_signature=("DELETE", "/api/warehouse/parts/{id}"), raise_on_fail=False)

        # Part requests
        self.request("list part request statuses", "GET", "/api/warehouse/part-request-statuses", actor="warehouse", expected=(200,), route_signature=("GET", "/api/warehouse/part-request-statuses"), raise_on_fail=False)
        self.request("list all part request history", "GET", "/api/warehouse/part-request-history", actor="warehouse", params={"limit": 10}, expected=(200,), route_signature=("GET", "/api/warehouse/part-request-history"), raise_on_fail=False)
        if part_id:
            req_payload = {"part_id": part_id, "quantity": 5, "mechanic_comment": f"Need part {self.suffix}"}
            resp = self.request("create part request", "POST", "/api/warehouse/part-requests", actor="mechanic", json_body=req_payload, expected=(201,), route_signature=("POST", "/api/warehouse/part-requests"), raise_on_fail=False)
            self.ids["part_request_id"] = self.get_id_from_response(resp)
            pr_id = self.ids.get("part_request_id")
            self.request("list part requests", "GET", "/api/warehouse/part-requests", actor="warehouse", params={"part_id": part_id, "status_code": "new", "date_from": DATE, "sort_by": "created_at", "order": "desc"}, expected=(200,), route_signature=("GET", "/api/warehouse/part-requests"), raise_on_fail=False)
            if pr_id:
                self.request("get part request", "GET", f"/api/warehouse/part-requests/{pr_id}", actor="warehouse", expected=(200,), route_signature=("GET", "/api/warehouse/part-requests/{id}"), raise_on_fail=False)
                self.request("update part request", "PUT", f"/api/warehouse/part-requests/{pr_id}", actor="warehouse", json_body={"part_id": part_id, "quantity": 6, "mechanic_comment": f"Updated request {self.suffix}", "status_id": 1, "history_comment": "API test update"}, expected=(200,), route_signature=("PUT", "/api/warehouse/part-requests/{id}"), raise_on_fail=False)
                self.request("update part request status", "PATCH", f"/api/warehouse/part-requests/{pr_id}/status", actor="warehouse", json_body={"status_id": 3, "comment": "API approved"}, expected=(200,), route_signature=("PATCH", "/api/warehouse/part-requests/{id}/status"), raise_on_fail=False)
                self.request("get part request history", "GET", f"/api/warehouse/part-requests/{pr_id}/history", actor="warehouse", expected=(200,), route_signature=("GET", "/api/warehouse/part-requests/{id}/history"), raise_on_fail=False)
            # disposable for DELETE while still new
            resp2 = self.request("create disposable part request", "POST", "/api/warehouse/part-requests", actor="mechanic", json_body={"part_id": part_id, "quantity": 1, "mechanic_comment": "delete me"}, expected=(201,), raise_on_fail=False)
            pr_del = self.get_id_from_response(resp2)
            if pr_del:
                self.request("delete disposable part request", "DELETE", f"/api/warehouse/part-requests/{pr_del}", actor="warehouse", expected=(200,), route_signature=("DELETE", "/api/warehouse/part-requests/{id}"), raise_on_fail=False)

    def test_vehicle_part_installations(self) -> None:
        part_id = self.ids.get("warehouse_part_id")
        vehicle_id = self.ids.get("vehicle_id")
        shift_id = self.ids.get("mechanic_shift_id")
        if not all([part_id, vehicle_id, shift_id]):
            return
        payload = {"part_id": part_id, "vehicle_id": vehicle_id, "mechanic_shift_id": shift_id, "installed_at": DATE, "planned_replacement_at": "2026-08-12", "quantity": 2}
        resp = self.request("create vehicle part installation", "POST", "/api/warehouse/vehicle-part-installations", actor="warehouse", json_body=payload, expected=(201,), route_signature=("POST", "/api/warehouse/vehicle-part-installations"), raise_on_fail=False)
        self.ids["vehicle_part_installation_id"] = self.get_id_from_response(resp)
        vpi_id = self.ids.get("vehicle_part_installation_id")
        self.request("list vehicle part installations", "GET", "/api/warehouse/vehicle-part-installations", actor="warehouse", params={"vehicle_id": vehicle_id, "part_id": part_id, "mechanic_shift_id": shift_id, "is_active": "true"}, expected=(200,), route_signature=("GET", "/api/warehouse/vehicle-part-installations"), raise_on_fail=False)
        if vpi_id:
            self.request("get vehicle part installation", "GET", f"/api/warehouse/vehicle-part-installations/{vpi_id}", actor="warehouse", expected=(200,), route_signature=("GET", "/api/warehouse/vehicle-part-installations/{id}"), raise_on_fail=False)
            self.request("update vehicle part installation", "PUT", f"/api/warehouse/vehicle-part-installations/{vpi_id}", actor="warehouse", json_body={**payload, "quantity": 3, "is_active": True}, expected=(200,), route_signature=("PUT", "/api/warehouse/vehicle-part-installations/{id}"), raise_on_fail=False)
            self.request("update vehicle part installation activity", "PATCH", f"/api/warehouse/vehicle-part-installations/{vpi_id}/activity", actor="warehouse", json_body={"is_active": True}, expected=(200,), route_signature=("PATCH", "/api/warehouse/vehicle-part-installations/{id}/activity"), raise_on_fail=False)
        resp2 = self.request("create disposable vehicle part installation", "POST", "/api/warehouse/vehicle-part-installations", actor="warehouse", json_body={**payload, "quantity": 1}, expected=(201,), raise_on_fail=False)
        del_id = self.get_id_from_response(resp2)
        if del_id:
            self.request("delete disposable vehicle part installation", "DELETE", f"/api/warehouse/vehicle-part-installations/{del_id}", actor="warehouse", expected=(200,), route_signature=("DELETE", "/api/warehouse/vehicle-part-installations/{id}"), raise_on_fail=False)

    def test_vehicle_services(self) -> None:
        vehicle_id = self.ids.get("vehicle_id")
        if not vehicle_id:
            return
        # service-parts
        self.request("list service parts", "GET", "/api/warehouse/service-parts", actor="warehouse", expected=(200,), route_signature=("GET", "/api/warehouse/service-parts"), raise_on_fail=False)
        resp = self.request("create service part", "POST", "/api/warehouse/service-parts", actor="warehouse", json_body={"name": f"API кузов часть {self.suffix}", "description": "API test body part"}, expected=(201,), route_signature=("POST", "/api/warehouse/service-parts"), raise_on_fail=False)
        self.ids["service_part_id"] = self.get_id_from_response(resp)
        sp_id = self.ids.get("service_part_id")
        if sp_id:
            self.request("get service part", "GET", f"/api/warehouse/service-parts/{sp_id}", actor="warehouse", expected=(200,), route_signature=("GET", "/api/warehouse/service-parts/{id}"), raise_on_fail=False)
            self.request("update service part", "PUT", f"/api/warehouse/service-parts/{sp_id}", actor="warehouse", json_body={"name": f"API кузов часть updated {self.suffix}", "description": "Updated"}, expected=(200,), route_signature=("PUT", "/api/warehouse/service-parts/{id}"), raise_on_fail=False)
        resp_del = self.request("create disposable service part", "POST", "/api/warehouse/service-parts", actor="warehouse", json_body={"name": f"API кузов часть delete {self.suffix}", "description": "delete"}, expected=(201,), raise_on_fail=False)
        sp_del_id = self.get_id_from_response(resp_del)
        if sp_del_id:
            self.request("delete disposable service part", "DELETE", f"/api/warehouse/service-parts/{sp_del_id}", actor="warehouse", expected=(200,), route_signature=("DELETE", "/api/warehouse/service-parts/{id}"), raise_on_fail=False)

        # service-types
        self.request("list service types", "GET", "/api/warehouse/service-types", actor="warehouse", expected=(200,), route_signature=("GET", "/api/warehouse/service-types"), raise_on_fail=False)
        resp = self.request("create service type", "POST", "/api/warehouse/service-types", actor="warehouse", json_body={"name": f"API Полировка {self.suffix}", "description": "API test service type"}, expected=(201,), route_signature=("POST", "/api/warehouse/service-types"), raise_on_fail=False)
        self.ids["service_type_id"] = self.get_id_from_response(resp)
        st_id = self.ids.get("service_type_id")
        if st_id:
            self.request("get service type", "GET", f"/api/warehouse/service-types/{st_id}", actor="warehouse", expected=(200,), route_signature=("GET", "/api/warehouse/service-types/{id}"), raise_on_fail=False)
            self.request("update service type", "PUT", f"/api/warehouse/service-types/{st_id}", actor="warehouse", json_body={"name": f"API Полировка updated {self.suffix}", "description": "Updated"}, expected=(200,), route_signature=("PUT", "/api/warehouse/service-types/{id}"), raise_on_fail=False)
        resp_del = self.request("create disposable service type", "POST", "/api/warehouse/service-types", actor="warehouse", json_body={"name": f"API Полировка delete {self.suffix}", "description": "delete"}, expected=(201,), raise_on_fail=False)
        st_del_id = self.get_id_from_response(resp_del)
        if st_del_id:
            self.request("delete disposable service type", "DELETE", f"/api/warehouse/service-types/{st_del_id}", actor="warehouse", expected=(200,), route_signature=("DELETE", "/api/warehouse/service-types/{id}"), raise_on_fail=False)

        # vehicle-services
        if sp_id and st_id:
            payload = {"type_id": st_id, "part_id": sp_id, "vehicle_id": vehicle_id, "date": DATE}
            resp = self.request("create vehicle service", "POST", "/api/warehouse/vehicle-services", actor="warehouse", json_body=payload, expected=(201,), route_signature=("POST", "/api/warehouse/vehicle-services"), raise_on_fail=False)
            self.ids["vehicle_service_id"] = self.get_id_from_response(resp)
            vs_id = self.ids.get("vehicle_service_id")
            self.request("list vehicle services", "GET", "/api/warehouse/vehicle-services", actor="warehouse", params={"vehicle_id": vehicle_id, "type_id": st_id, "part_id": sp_id, "date_from": DATE, "date_to": DATE2}, expected=(200,), route_signature=("GET", "/api/warehouse/vehicle-services"), raise_on_fail=False)
            if vs_id:
                self.request("get vehicle service", "GET", f"/api/warehouse/vehicle-services/{vs_id}", actor="warehouse", expected=(200,), route_signature=("GET", "/api/warehouse/vehicle-services/{id}"), raise_on_fail=False)
                self.request("update vehicle service", "PUT", f"/api/warehouse/vehicle-services/{vs_id}", actor="warehouse", json_body={**payload, "date": DATE2}, expected=(200,), route_signature=("PUT", "/api/warehouse/vehicle-services/{id}"), raise_on_fail=False)
            resp_del = self.request("create disposable vehicle service", "POST", "/api/warehouse/vehicle-services", actor="warehouse", json_body={**payload, "date": DATE2}, expected=(201,), raise_on_fail=False)
            vs_del_id = self.get_id_from_response(resp_del)
            if vs_del_id:
                self.request("delete disposable vehicle service", "DELETE", f"/api/warehouse/vehicle-services/{vs_del_id}", actor="warehouse", expected=(200,), route_signature=("DELETE", "/api/warehouse/vehicle-services/{id}"), raise_on_fail=False)

    def final_vehicle_passport_check(self) -> None:
        vehicle_id = self.ids.get("vehicle_id")
        if not vehicle_id:
            return
        self.request("final vehicle passport with linked entities", "GET", f"/api/vehicles/{vehicle_id}", actor="admin", expected=(200,), route_signature=("GET", "/api/vehicles/{id}"), raise_on_fail=False)

    # ------------------------ Summary ------------------------
    def print_summary(self) -> int:
        total = len(self.results)
        skipped = sum(1 for r in self.results if r.skipped)
        passed = sum(1 for r in self.results if r.success and not r.skipped)
        failed = sum(1 for r in self.results if not r.success)
        expected = set(ALL_ROUTE_SIGNATURES)
        missing = sorted(expected - self.tested - SKIPPED_ROUTE_SIGNATURES)

        print("\n" + "=" * 130)
        print("AUTO PARK ALL ROUTES E2E API TEST SUMMARY")
        print("=" * 130)
        print(f"Base URL:              {BASE_URL}")
        print(f"DB:                    {DB_CONFIG['host']}:{DB_CONFIG['port']}/{DB_CONFIG['dbname']}")
        print(f"Run suffix:            {self.suffix}")
        print(f"Total checks:          {total}")
        print(f"Passed checks:         {passed}")
        print(f"Failed checks:         {failed}")
        print(f"Skipped checks:        {skipped}")
        print(f"Expected route count:  {len(expected)}")
        print(f"Covered route count:   {len(self.tested - SKIPPED_ROUTE_SIGNATURES)}")
        print(f"Missing route tests:   {len(missing)}")
        print("-" * 130)

        for r in self.results:
            if r.skipped:
                mark = "SKIP"
            else:
                mark = "PASS" if r.success else "FAIL"
            code = "-" if r.status_code is None else str(r.status_code)
            print(f"[{mark}] {r.method:<6} {code:<4} {r.elapsed_ms:>5} ms | {r.actor:<10} | {r.name} | {r.path}")
            if not r.success:
                if r.error:
                    print(f"       error: {r.error}")
                if r.response_preview:
                    print(f"       resp : {r.response_preview[:500].replace(chr(10), ' ')}")
            elif r.skipped and r.error:
                print(f"       reason: {r.error}")

        print("-" * 130)
        if missing:
            print("MISSING ROUTE TESTS:")
            for method, path in missing:
                print(f"  - {method} {path}")
        else:
            print("All non-WebSocket registered routes from the script were covered.")

        report = {
            "base_url": BASE_URL,
            "db": {k: ("***" if k == "password" else v) for k, v in DB_CONFIG.items()},
            "suffix": self.suffix,
            "ids": self.ids,
            "users": self.users,
            "total_checks": total,
            "passed_checks": passed,
            "failed_checks": failed,
            "skipped_checks": skipped,
            "expected_routes": sorted(list(expected)),
            "tested_routes": sorted(list(self.tested)),
            "missing_routes": missing,
            "results": [asdict(r) for r in self.results],
        }
        report_path = Path.cwd() / f"auto_park_api_test_report_{self.suffix}.json"
        report_path.write_text(json.dumps(report, ensure_ascii=False, indent=2), encoding="utf-8")
        print(f"Saved report: {report_path}")
        print("=" * 130)
        return 0 if failed == 0 and not missing else 1

    def run(self) -> int:
        self.seed_role_users_in_db()
        self.wait_api()
        self.login_all_roles()
        self.test_health()
        self.test_users()
        self.test_drivers()
        self.test_mechanic_shifts()
        self.test_vehicles()
        self.test_vehicle_extra_entities()
        self.test_tripsheets()
        self.test_tripsheet_trips()
        self.test_incidents()
        self.test_fuel()
        self.test_notifications()
        self.test_warehouse_parts_and_requests()
        self.test_vehicle_part_installations()
        self.test_vehicle_services()
        self.final_vehicle_passport_check()
        return self.print_summary()


if __name__ == "__main__":
    try:
        sys.exit(AutoParkE2ETest().run())
    except KeyboardInterrupt:
        print("\nInterrupted")
        sys.exit(130)
    except Exception as e:
        print(f"\nFATAL: {e}")
        sys.exit(1)
