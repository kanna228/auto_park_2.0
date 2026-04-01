CREATE TABLE IF NOT EXISTS roles (
  id   BIGSERIAL PRIMARY KEY,
  name TEXT NOT NULL UNIQUE
);

-- 5 ролей из ТЗ
INSERT INTO roles (name) VALUES
  ('admin'),
  ('manager'),
  ('garage_dispatcher'),
  ('duty_mechanic'),
  ('warehouse_manager')
ON CONFLICT (name) DO NOTHING;
