-- name: GetScorecardCache :one
SELECT * FROM scorecard_cache
WHERE org = $1 AND repo = $2;

-- name: UpsertScorecardCache :exec
INSERT INTO scorecard_cache (org, repo, grade, total, breakdown, computed_at)
VALUES ($1, $2, $3, $4, $5, NOW())
ON CONFLICT (org, repo) DO UPDATE
SET grade = EXCLUDED.grade,
    total = EXCLUDED.total,
    breakdown = EXCLUDED.breakdown,
    computed_at = NOW();

-- name: GetRepoCompliance :one
SELECT * FROM repo_compliance
WHERE org = $1 AND repo = $2;

-- name: GetLatestQACoverageScore :one
SELECT coverage_score
FROM qa_replay_jobs
WHERE repo_id = (SELECT id FROM repositories WHERE org_id = (SELECT id FROM organizations WHERE github_org_name = $1) AND name = $2)
ORDER BY created_at DESC
LIMIT 1;

-- name: HasRepoGuides :one
SELECT EXISTS (
    SELECT 1 FROM repo_guides
    WHERE org = $1 AND repo = $2
);
