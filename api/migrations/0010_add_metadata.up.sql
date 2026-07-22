ALTER TABLE repositories ADD COLUMN metadata JSONB DEFAULT '{}'::jsonb;
