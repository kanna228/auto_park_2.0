CREATE SEQUENCE IF NOT EXISTS invoice_number_seq START 223780;

INSERT INTO part_request_statuses (code, name)
VALUES ('issued', U&'\0412\044B\0434\0430\043D\043E')
ON CONFLICT (code) DO UPDATE SET
    name = EXCLUDED.name;

CREATE TABLE IF NOT EXISTS invoices (
    id BIGSERIAL PRIMARY KEY,
    invoice_number VARCHAR(32) NOT NULL UNIQUE,
    invoice_date DATE NOT NULL,
    part_request_id BIGINT NULL REFERENCES part_requests(id) ON DELETE SET NULL,
    request_number VARCHAR(32) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_invoices_invoice_date
    ON invoices(invoice_date DESC);

CREATE INDEX IF NOT EXISTS idx_invoices_part_request_id
    ON invoices(part_request_id);
