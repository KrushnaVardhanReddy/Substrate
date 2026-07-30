ALTER TABLE repositories ADD COLUMN IF NOT EXISTS metadata JSONB DEFAULT '{}'::jsonb;
