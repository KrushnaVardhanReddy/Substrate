CREATE TABLE IF NOT EXISTS repo_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    repo_id UUID NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    predictive_risk_score INTEGER NOT NULL DEFAULT 0,
    calculated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_repo_metrics_repo_id ON repo_metrics(repo_id);
CREATE INDEX IF NOT EXISTS idx_repo_metrics_calculated_at ON repo_metrics(calculated_at);
