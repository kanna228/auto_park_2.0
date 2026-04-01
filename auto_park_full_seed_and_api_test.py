#!/usr/bin/env python3
from __future__ import annotations
import json, os, random, string, sys, time
from dataclasses import dataclass
from typing import Any, Dict, List, Optional, Set, Tuple
import requests

BASE_URL = os.getenv("AUTO_PARK_BASE_URL", "http://localhost:8080").rstrip("/")
ADMIN_EMAIL = os.getenv("AUTO_PARK_ADMIN_EMAIL", "admin@autopark.local")
ADMIN_PASSWORD = os.getenv("AUTO_PARK_ADMIN_PASSWORD", "admin123")
ADMIN_IIN = os.getenv("AUTO_PARK_ADMIN_IIN", "000000000000")
TIMEOUT = int(os.getenv("AUTO_PARK_TIMEOUT", "20"))

EXPECTED_ROUTES: Set[Tuple[str, str]] = {
    ("GET", "/healthz"), ("GET", "/api/tripsheet/health"), ("POST", "/api/users/login"),
    ("GET", "/api/users"), ("GET", "/api/users/{id}"), ("POST", "/api/users"), ("PUT", "/api/users/{id}"), ("DELETE", "/api/users/{id}"),
    ("GET", "/api/users/drivers"), ("GET", "/api/users/drivers/{id}"), ("POST", "/api/users/drivers"), ("PUT", "/api/users/drivers/{id}"), ("DELETE", "/api/users/drivers/{id}"),
    ("POST", "/api/vehicles"), ("GET", "/api/vehicles/{id}"), ("GET", "/api/vehicles"), ("PUT", "/api/vehicles/{id}"), ("DELETE", "/api/vehicles/{id}"),
    ("POST", "/api/vehicles/tires"), ("GET", "/api/vehicles/tires/{id}"), ("GET", "/api/vehicles/tires"), ("PUT", "/api/vehicles/tires/{id}"), ("DELETE", "/api/vehicles/tires/{id}"),
    ("POST", "/api/vehicles/tire-places"), ("GET", "/api/vehicles/tire-places/{id}"), ("GET", "/api/vehicles/tire-places"), ("PUT", "/api/vehicles/tire-places/{id}"), ("DELETE", "/api/vehicles/tire-places/{id}"),
    ("GET", "/api/vehicles/vehicle-tires/{vehicle_id}"),
    ("GET", "/api/tripsheet"), ("GET", "/api/tripsheet/{id}"), ("POST", "/api/tripsheet"), ("PUT", "/api/tripsheet/{id}"), ("DELETE", "/api/tripsheet/{id}"),
    ("GET", "/api/tripsheet/trips"), ("GET", "/api/tripsheet/trips/{id}"), ("GET", "/api/tripsheet/trips/by-tripsheet/{tripsheet_id}"),
    ("POST", "/api/tripsheet/trips"), ("PUT", "/api/tripsheet/trips/{id}"), ("DELETE", "/api/tripsheet/trips/{id}"),
}

def rand_suffix(n:int=6)->str:
    a = string.ascii_lowercase + string.digits
    return "".join(random.choice(a) for _ in range(n))

def now_ms()->int:
    return int(time.time()*1000)

@dataclass
class TestResult:
    name:str
    method:str
    path:str
    success:bool
    status_code:Optional[int]
    elapsed_ms:int
    expected:Tuple[int,...]
    error:Optional[str]=None
    response_preview:Optional[str]=None

class ApiTestRunner:
    def __init__(self)->None:
        self.s=requests.Session()
        self.results:List[TestResult]=[]
        self.tested_routes:Set[Tuple[str,str]]=set()
        self.token:Optional[str]=None
        self.admin_user_id:Optional[int]=None
        self.created_user_id:Optional[int]=None
        self.driver_main_id:Optional[int]=None
        self.driver_temp_id:Optional[int]=None
        self.vehicle_id:Optional[int]=None
        self.tire_place_id:Optional[int]=None
        self.tire_id:Optional[int]=None
        self.tripsheet_id:Optional[int]=None
        self.tripsheet_trip_id:Optional[int]=None

    def headers(self, auth:bool=False)->Dict[str,str]:
        h={"Content-Type":"application/json"}
        if auth and self.token:
            h["Authorization"]=f"Bearer {self.token}"
        return h

    def record(self, name:str, method:str, path:str, success:bool, status_code:Optional[int], elapsed_ms:int, expected:Tuple[int,...], error:Optional[str]=None, response_preview:Optional[str]=None)->None:
        self.results.append(TestResult(name, method, path, success, status_code, elapsed_ms, expected, error, response_preview[:500] if response_preview else None))

    def request(self, name:str, method:str, path:str, *, auth:bool=False, json_body:Optional[dict]=None, params:Optional[dict]=None, expected:Tuple[int,...]=(200,201), route_signature:Optional[Tuple[str,str]]=None)->requests.Response:
        started=now_ms()
        try:
            resp=self.s.request(method=method, url=f"{BASE_URL}{path}", headers=self.headers(auth), json=json_body, params=params, timeout=TIMEOUT)
            ok=resp.status_code in expected
            self.record(name, method, path, ok, resp.status_code, now_ms()-started, expected, None if ok else f"expected {expected}, got {resp.status_code}", resp.text)
            if route_signature: self.tested_routes.add(route_signature)
            if not ok:
                raise AssertionError(f"{name}: expected {expected}, got {resp.status_code}: {resp.text[:300]}")
            return resp
        except Exception as e:
            if isinstance(e, AssertionError):
                raise
            self.record(name, method, path, False, None, now_ms()-started, expected, str(e), None)
            if route_signature: self.tested_routes.add(route_signature)
            raise

    def nested(self, obj:Any, *keys:str)->Any:
        cur=obj
        for k in keys:
            if not isinstance(cur, dict) or k not in cur:
                return None
            cur=cur[k]
        return cur

    def wait_api(self)->None:
        deadline=time.time()+60
        while time.time()<deadline:
            try:
                self.request("healthz wait","GET","/healthz", expected=(200,), route_signature=("GET","/healthz"))
                return
            except Exception:
                time.sleep(2)
        raise RuntimeError("API did not become healthy in time")

    def login(self)->None:
        resp=self.request("login admin","POST","/api/users/login", json_body={"email":ADMIN_EMAIL,"password":ADMIN_PASSWORD,"iin":ADMIN_IIN}, expected=(200,), route_signature=("POST","/api/users/login"))
        data=resp.json()
        self.token=self.nested(data,"data","token")
        self.admin_user_id=self.nested(data,"data","user_id")
        if not self.token:
            raise RuntimeError(f"token not found in login response: {data}")

    def run_open_routes(self)->None:
        self.request("healthz","GET","/healthz", expected=(200,), route_signature=("GET","/healthz"))
        self.request("tripsheet health","GET","/api/tripsheet/health", expected=(200,), route_signature=("GET","/api/tripsheet/health"))

    def users_crud(self)->None:
        self.request("list users","GET","/api/users", auth=True, expected=(200,), route_signature=("GET","/api/users"))
        if self.admin_user_id:
            self.request("get admin user","GET",f"/api/users/{self.admin_user_id}", auth=True, expected=(200,), route_signature=("GET","/api/users/{id}"))
        suffix=rand_suffix()
        resp=self.request("create user","POST","/api/users", auth=True, expected=(201,), route_signature=("POST","/api/users"), json_body={
            "email":f"user_{suffix}@example.com","first_name":"Test","last_name":"User","middle_name":"Smoke","iin":f"990101{random.randint(100000,999999)}","phone":"+77001234567","role_id":2
        })
        self.created_user_id=self.nested(resp.json(),"data","id")
        self.request("get created user","GET",f"/api/users/{self.created_user_id}", auth=True, expected=(200,), route_signature=("GET","/api/users/{id}"))
        self.request("update user","PUT",f"/api/users/{self.created_user_id}", auth=True, expected=(200,), route_signature=("PUT","/api/users/{id}"), json_body={"first_name":"Updated","last_name":"User","phone":"+77009998877"})

    def drivers_crud(self)->None:
        self.request("list drivers initial","GET","/api/users/drivers", auth=True, expected=(200,), route_signature=("GET","/api/users/drivers"))
        suffix=rand_suffix()
        resp=self.request("create main driver","POST","/api/users/drivers", auth=True, expected=(201,), route_signature=("POST","/api/users/drivers"), json_body={
            "iin":f"010203{random.randint(100000,999999)}","name":"Main","surname":f"Driver{suffix}","middlename":"A","phone":"+77001112233","mail":f"main_driver_{suffix}@example.com"
        })
        self.driver_main_id=self.nested(resp.json(),"data","id")
        resp=self.request("create temp driver","POST","/api/users/drivers", auth=True, expected=(201,), route_signature=("POST","/api/users/drivers"), json_body={
            "iin":f"020304{random.randint(100000,999999)}","name":"Temp","surname":f"Driver{suffix}","middlename":"B","phone":"+77004445566","mail":f"temp_driver_{suffix}@example.com"
        })
        self.driver_temp_id=self.nested(resp.json(),"data","id")
        self.request("get main driver","GET",f"/api/users/drivers/{self.driver_main_id}", auth=True, expected=(200,), route_signature=("GET","/api/users/drivers/{id}"))
        self.request("update temp driver","PUT",f"/api/users/drivers/{self.driver_temp_id}", auth=True, expected=(200,), route_signature=("PUT","/api/users/drivers/{id}"), json_body={"phone":"+77007776655","mail":f"updated_driver_{suffix}@example.com"})
        self.request("list drivers after create","GET","/api/users/drivers", auth=True, expected=(200,), route_signature=("GET","/api/users/drivers"))

    def vehicles_crud(self)->None:
        suffix=rand_suffix()
        create_payload={
            "board_number":f"B-{random.randint(100,999)}","technical_passport_number":f"TP-{random.randint(10000,99999)}","state_number":f"777AAA{random.randint(100,999)}","vin":f"VIN{suffix.upper()}123456789",
            "brand_model":"Toyota Camry","manufacture_year":2022,"received_date":"2026-03-31T00:00:00Z","empty_weight_kg":1500.0,"max_weight_kg":2200.0,"engine_volume_cc":2500,
            "insurance_policy_number":f"POL-{random.randint(1000,9999)}","insurance_expiry_date":"2027-03-31T00:00:00Z","mileage":12000,"current_fuel":35.5,"drivers_ids":[self.driver_main_id]
        }
        resp=self.request("create vehicle","POST","/api/vehicles", auth=True, expected=(201,), route_signature=("POST","/api/vehicles"), json_body=create_payload)
        self.vehicle_id=self.nested(resp.json(),"data","id")
        self.request("list vehicles","GET","/api/vehicles", auth=True, expected=(200,), route_signature=("GET","/api/vehicles"))
        self.request("get vehicle","GET",f"/api/vehicles/{self.vehicle_id}", auth=True, expected=(200,), route_signature=("GET","/api/vehicles/{id}"))
        upd=dict(create_payload); upd["brand_model"]="Toyota Camry Restyled"; upd["mileage"]=13000; upd["current_fuel"]=30.0
        self.request("update vehicle","PUT",f"/api/vehicles/{self.vehicle_id}", auth=True, expected=(200,), route_signature=("PUT","/api/vehicles/{id}"), json_body=upd)

    def tire_places_crud(self)->None:
        self.request("list tire places initial","GET","/api/vehicles/tire-places", auth=True, expected=(200,), route_signature=("GET","/api/vehicles/tire-places"))
        resp=self.request("create tire place","POST","/api/vehicles/tire-places", auth=True, expected=(201,), route_signature=("POST","/api/vehicles/tire-places"), json_body={"name":f"special-place-{rand_suffix()}"})
        self.tire_place_id=self.nested(resp.json(),"data","id")
        self.request("get tire place","GET",f"/api/vehicles/tire-places/{self.tire_place_id}", auth=True, expected=(200,), route_signature=("GET","/api/vehicles/tire-places/{id}"))
        self.request("update tire place","PUT",f"/api/vehicles/tire-places/{self.tire_place_id}", auth=True, expected=(200,), route_signature=("PUT","/api/vehicles/tire-places/{id}"), json_body={"name":f"special-place-updated-{rand_suffix()}"})

    def tires_crud(self)->None:
        create_payload={"place_id":self.tire_place_id,"vehicle_id":self.vehicle_id,"tire":"Michelin Primacy 4","mileage":10000,"max_usage":60000}
        resp=self.request("create tire","POST","/api/vehicles/tires", auth=True, expected=(201,), route_signature=("POST","/api/vehicles/tires"), json_body=create_payload)
        self.tire_id=self.nested(resp.json(),"data","id")
        self.request("list tires","GET","/api/vehicles/tires", auth=True, expected=(200,), route_signature=("GET","/api/vehicles/tires"))
        self.request("get tire","GET",f"/api/vehicles/tires/{self.tire_id}", auth=True, expected=(200,), route_signature=("GET","/api/vehicles/tires/{id}"))
        upd={"place_id":self.tire_place_id,"vehicle_id":self.vehicle_id,"tire":"Michelin Primacy 4 Updated","mileage":15000,"max_usage":65000}
        self.request("update tire","PUT",f"/api/vehicles/tires/{self.tire_id}", auth=True, expected=(200,), route_signature=("PUT","/api/vehicles/tires/{id}"), json_body=upd)
        self.request("vehicle tires","GET",f"/api/vehicles/vehicle-tires/{self.vehicle_id}", auth=True, expected=(200,), route_signature=("GET","/api/vehicles/vehicle-tires/{vehicle_id}"))

    def tripsheets_crud(self)->None:
        create_payload={"tripsheet_number":f"TS-{random.randint(1000,9999)}","tripsheet_date":"2026-03-31","vehicle_id":self.vehicle_id,"vehicle_brand":"Toyota Camry Restyled","vehicle_plate_number":"777AAA111","driver_last_name":"Driver","driver_first_name":"Main","driver_middle_name":"A","driver_id":self.driver_main_id,"start_time":"2026-03-31T08:00:00Z","end_time":"2026-03-31T18:00:00Z","mileage_start":13000,"mileage_end":13120,"fuel_start":30,"fuel_issued":15,"fuel_consumption_theoretical":10,"fuel_consumption_actual":11,"status_id":1}
        resp=self.request("create tripsheet","POST","/api/tripsheet", auth=True, expected=(201,), route_signature=("POST","/api/tripsheet"), json_body=create_payload)
        self.tripsheet_id=self.nested(resp.json(),"data","id")
        self.request("list tripsheets","GET","/api/tripsheet", auth=True, expected=(200,), route_signature=("GET","/api/tripsheet"))
        self.request("get tripsheet","GET",f"/api/tripsheet/{self.tripsheet_id}", auth=True, expected=(200,), route_signature=("GET","/api/tripsheet/{id}"))
        upd=dict(create_payload); upd["status_id"]=2; upd["fuel_consumption_actual"]=12
        self.request("update tripsheet","PUT",f"/api/tripsheet/{self.tripsheet_id}", auth=True, expected=(200,), route_signature=("PUT","/api/tripsheet/{id}"), json_body=upd)

    def tripsheet_trips_crud(self)->None:
        create_payload={"tripsheet_id":self.tripsheet_id,"route_description":"Warehouse -> Service center -> Garage","start_time":"2026-03-31T09:00:00Z","end_time":"2026-03-31T10:30:00Z","distance_passed":42,"status_id":1}
        resp=self.request("create tripsheet trip","POST","/api/tripsheet/trips", auth=True, expected=(201,), route_signature=("POST","/api/tripsheet/trips"), json_body=create_payload)
        self.tripsheet_trip_id=self.nested(resp.json(),"data","id")
        self.request("list tripsheet trips","GET","/api/tripsheet/trips", auth=True, expected=(200,), route_signature=("GET","/api/tripsheet/trips"))
        self.request("get tripsheet trip","GET",f"/api/tripsheet/trips/{self.tripsheet_trip_id}", auth=True, expected=(200,), route_signature=("GET","/api/tripsheet/trips/{id}"))
        self.request("get trips by tripsheet","GET",f"/api/tripsheet/trips/by-tripsheet/{self.tripsheet_id}", auth=True, expected=(200,), route_signature=("GET","/api/tripsheet/trips/by-tripsheet/{tripsheet_id}"))
        upd={"tripsheet_id":self.tripsheet_id,"route_description":"Warehouse -> Fuel station -> Garage","start_time":"2026-03-31T09:00:00Z","end_time":"2026-03-31T11:00:00Z","distance_passed":55,"status_id":2}
        self.request("update tripsheet trip","PUT",f"/api/tripsheet/trips/{self.tripsheet_trip_id}", auth=True, expected=(200,), route_signature=("PUT","/api/tripsheet/trips/{id}"), json_body=upd)

    def cleanup(self)->None:
        if self.tripsheet_trip_id: self.request("delete tripsheet trip","DELETE",f"/api/tripsheet/trips/{self.tripsheet_trip_id}", auth=True, expected=(200,), route_signature=("DELETE","/api/tripsheet/trips/{id}"))
        if self.tripsheet_id: self.request("delete tripsheet","DELETE",f"/api/tripsheet/{self.tripsheet_id}", auth=True, expected=(200,), route_signature=("DELETE","/api/tripsheet/{id}"))
        if self.tire_id: self.request("delete tire","DELETE",f"/api/vehicles/tires/{self.tire_id}", auth=True, expected=(200,), route_signature=("DELETE","/api/vehicles/tires/{id}"))
        if self.tire_place_id: self.request("delete tire place","DELETE",f"/api/vehicles/tire-places/{self.tire_place_id}", auth=True, expected=(200,), route_signature=("DELETE","/api/vehicles/tire-places/{id}"))
        if self.vehicle_id: self.request("delete vehicle","DELETE",f"/api/vehicles/{self.vehicle_id}", auth=True, expected=(200,), route_signature=("DELETE","/api/vehicles/{id}"))
        if self.driver_temp_id: self.request("delete temp driver","DELETE",f"/api/users/drivers/{self.driver_temp_id}", auth=True, expected=(200,), route_signature=("DELETE","/api/users/drivers/{id}"))
        if self.driver_main_id: self.request("delete main driver","DELETE",f"/api/users/drivers/{self.driver_main_id}", auth=True, expected=(200,), route_signature=("DELETE","/api/users/drivers/{id}"))
        if self.created_user_id: self.request("delete created user","DELETE",f"/api/users/{self.created_user_id}", auth=True, expected=(200,), route_signature=("DELETE","/api/users/{id}"))

    def print_summary(self)->int:
        total=len(self.results); passed=sum(1 for r in self.results if r.success); failed=total-passed
        missing=sorted(EXPECTED_ROUTES - self.tested_routes); covered=len(EXPECTED_ROUTES)-len(missing)
        print("\n"+"="*120); print("AUTO PARK FULL SEED + API TEST SUMMARY"); print("="*120)
        print(f"Base URL:            {BASE_URL}"); print(f"Total requests:      {total}"); print(f"Passed requests:     {passed}"); print(f"Failed requests:     {failed}")
        print(f"Expected routes:     {len(EXPECTED_ROUTES)}"); print(f"Covered routes:      {covered}"); print(f"Missing route tests: {len(missing)}"); print("-"*120)
        for r in self.results:
            mark="PASS" if r.success else "FAIL"; code=r.status_code if r.status_code is not None else "-"
            print(f"[{mark}] {r.method:<6} {code:<4} {r.elapsed_ms:>5} ms | {r.name} | {r.path}")
            if not r.success:
                if r.error: print(f"       error: {r.error}")
                if r.response_preview: print(f"       resp : {r.response_preview[:300].replace(chr(10), ' ')}")
        print("-"*120)
        if missing:
            print("MISSING ROUTE TESTS:")
            for method, path in missing: print(f"  - {method} {path}")
        else:
            print("All current routes were covered by the test script.")
        with open("auto_park_api_test_report.json","w",encoding="utf-8") as f:
            json.dump({"base_url":BASE_URL,"total_requests":total,"passed_requests":passed,"failed_requests":failed,"expected_routes":sorted(list(EXPECTED_ROUTES)),"tested_routes":sorted(list(self.tested_routes)),"missing_routes":missing,"results":[r.__dict__ for r in self.results]}, f, ensure_ascii=False, indent=2)
        print("-"*120); print("Saved report: auto_park_api_test_report.json"); print("="*120)
        return 0 if failed == 0 and not missing else 1

    def run(self)->int:
        self.wait_api(); self.run_open_routes(); self.login(); self.users_crud(); self.drivers_crud(); self.vehicles_crud(); self.tire_places_crud(); self.tires_crud(); self.tripsheets_crud(); self.tripsheet_trips_crud(); self.cleanup(); return self.print_summary()

if __name__ == "__main__":
    try:
        sys.exit(ApiTestRunner().run())
    except KeyboardInterrupt:
        print("\nInterrupted."); sys.exit(130)
