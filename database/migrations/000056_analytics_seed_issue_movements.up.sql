-- Analytics monthly "parts" uses stock issue movements. Demo installations are
-- real part usage, so mirror them into stock movements when the movement is absent.

INSERT INTO part_stock_movements (
    part_id,
    type,
    quantity,
    vehicle_id,
    part_request_id,
    document_number,
    actor_user_id,
    created_at
)
SELECT
    vi.part_id,
    'issue',
    vi.quantity,
    vi.vehicle_id,
    NULL,
    'AN-DEMO-ISSUE-' || vi.id,
    vi.installed_by_user_id,
    vi.created_at
FROM vehicle_part_installations vi
JOIN parts_catalog pc ON pc.id = vi.part_id
LEFT JOIN vehicles v ON v.id = vi.vehicle_id
WHERE (pc.part_id LIKE 'DEMO-%' OR v.board_number BETWEEN 'B-101' AND 'B-120')
  AND NOT EXISTS (
      SELECT 1
      FROM part_stock_movements psm
      WHERE psm.document_number = 'AN-DEMO-ISSUE-' || vi.id
  );
