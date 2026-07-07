CREATE TABLE orders (
  id         BIGINT PRIMARY KEY,
  total      NUMERIC(10,2),
  legacy_ref TEXT
);
