CREATE TABLE users (
  id     BIGINT PRIMARY KEY,
  emails TEXT NOT NULL
);
-- 'email' → 'emails': high similarity, same type TEXT → COLUMN_RENAMED
