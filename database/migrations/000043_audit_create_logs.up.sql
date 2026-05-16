CREATE TABLE IF NOT EXISTS audit_logs (
    id BIGSERIAL PRIMARY KEY,
    level VARCHAR(16) NOT NULL CHECK (level IN ('system', 'info', 'success', 'warning', 'error')),
    function VARCHAR(32) NOT NULL CHECK (function IN ('request', 'arrival', 'tripsheet', 'shift', 'incident', 'vehicle', 'user')),
    from_status VARCHAR(64) NULL,
    to_status VARCHAR(64) NULL,
    actor VARCHAR(128) NOT NULL,
    message TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at
    ON audit_logs(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_audit_logs_function
    ON audit_logs(function);

CREATE INDEX IF NOT EXISTS idx_audit_logs_level
    ON audit_logs(level);
