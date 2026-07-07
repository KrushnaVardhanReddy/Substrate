CREATE TABLE orders (
  id         BIGINT PRIMARY KEY,
  total      NUMERIC(10,2),
  created_at TIMESTAMP
);
-- legacy_ref removed (BREAKING) + created_at added nullable (SAFE) = 2 changes
