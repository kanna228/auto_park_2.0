CREATE TABLE IF NOT EXISTS maintenance_executions (
    id BIGSERIAL PRIMARY KEY,
    schedule_id BIGINT NOT NULL REFERENCES maintenance_schedules(id) ON DELETE CASCADE,
    board VARCHAR(16) NOT NULL,
    action VARCHAR(48) NOT NULL CHECK (action IN ('serviced', 'serviced_replaced', 'defect_parked')),
    comment TEXT NULL,
    responsible_mechanic VARCHAR(128) NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_maintenance_executions_schedule_id
    ON maintenance_executions(schedule_id);

CREATE INDEX IF NOT EXISTS idx_maintenance_executions_board
    ON maintenance_executions(board);
