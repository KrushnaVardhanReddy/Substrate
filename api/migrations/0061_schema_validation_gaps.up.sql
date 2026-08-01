CREATE TABLE IF NOT EXISTS schema_validation_gaps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    method VARCHAR(10) NOT NULL,
    path TEXT NOT NULL,
    payload JSONB NOT NULL,
    issue TEXT NOT NULL,
    severity VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
