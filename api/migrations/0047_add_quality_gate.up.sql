ALTER TABLE organizations ADD COLUMN IF NOT EXISTS quality_gate VARCHAR(50) DEFAULT 'standard';
