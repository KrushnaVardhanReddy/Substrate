CREATE TABLE IF NOT EXISTS preview_sessions (
    token UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pr_number INT NOT NULL,
    diff_id UUID NOT NULL REFERENCES diff_reports(id),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
