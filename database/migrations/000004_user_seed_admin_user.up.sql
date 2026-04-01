-- Временный пароль-хэш (заглушка). Потом заменим на bcrypt.
-- Например: "admin123" -> позже сделаем нормальный bcrypt в коде и обновим запись.
WITH admin_role AS (
  SELECT id FROM roles WHERE name = 'admin'
)
INSERT INTO users (
  email, first_name, last_name, middle_name, iin, phone, password, role_id, session_token, last_seen
)
SELECT
  'admin@autopark.local',
  'Admin',
  'Autopark',
  NULL,
  '000000000000',
  NULL,
  '$2a$12$93CmzSxMMK.jo6YfoXs5G.rneemNhlMu6DHuYPV3QtX6OQa3JKLme',
  admin_role.id,
  NULL,
  NULL
FROM admin_role
ON CONFLICT (email) DO NOTHING;
