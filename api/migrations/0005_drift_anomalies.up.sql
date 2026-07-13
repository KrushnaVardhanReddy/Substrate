CREATE TABLE IF NOT EXISTS drift_anomalies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_name TEXT NOT NULL,
    repo_name TEXT NOT NULL,
    method TEXT NOT NULL,
    path TEXT NOT NULL,
    error_message TEXT NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
